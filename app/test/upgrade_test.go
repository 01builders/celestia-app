package app_test

import (
	"fmt"
	"strings"
	"testing"

	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/math"
	"cosmossdk.io/x/params/types/proposal"
	app "github.com/celestiaorg/celestia-app/v3/app"
	"github.com/celestiaorg/celestia-app/v3/app/encoding"
	"github.com/celestiaorg/celestia-app/v3/pkg/appconsts"
	v1 "github.com/celestiaorg/celestia-app/v3/pkg/appconsts/v1"
	v2 "github.com/celestiaorg/celestia-app/v3/pkg/appconsts/v2"
	v3 "github.com/celestiaorg/celestia-app/v3/pkg/appconsts/v3"
	"github.com/celestiaorg/celestia-app/v3/pkg/user"
	"github.com/celestiaorg/celestia-app/v3/test/util"
	"github.com/celestiaorg/celestia-app/v3/test/util/genesis"
	"github.com/celestiaorg/celestia-app/v3/test/util/testnode"
	blobstreamtypes "github.com/celestiaorg/celestia-app/v3/x/blobstream/types"
	"github.com/celestiaorg/celestia-app/v3/x/minfee"
	signaltypes "github.com/celestiaorg/celestia-app/v3/x/signal/types"
	"github.com/celestiaorg/go-square/v2/share"
	"github.com/celestiaorg/go-square/v2/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	// packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v9/packetforward/types"
	"cosmossdk.io/log"
	icahosttypes "github.com/cosmos/ibc-go/v9/modules/apps/27-interchain-accounts/host/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
)

func TestAppUpgradeV3(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestAppUpgradeV3 in short mode")
	}

	testApp, genesis := SetupTestAppWithUpgradeHeight(t, 3)
	upgradeFromV1ToV2(t, testApp)

	ctx := testApp.NewContext(true)
	validators, err := testApp.StakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	valAddr, err := sdk.ValAddressFromBech32(validators[0].OperatorAddress)
	require.NoError(t, err)
	record, err := genesis.Keyring().Key(testnode.DefaultValidatorAccountName)
	require.NoError(t, err)
	accAddr, err := record.GetAddress()
	require.NoError(t, err)
	encCfg := encoding.MakeConfig(app.ModuleEncodingRegisters...)
	account := testApp.AuthKeeper.GetAccount(ctx, accAddr)
	signer, err := user.NewSigner(
		genesis.Keyring(), encCfg.TxConfig, testApp.ChainID(), v3.Version,
		user.NewAccount(testnode.DefaultValidatorAccountName, account.GetAccountNumber(), account.GetSequence()),
	)
	require.NoError(t, err)

	upgradeTx, err := signer.CreateTx(
		[]sdk.Msg{
			signaltypes.NewMsgSignalVersion(valAddr, 3),
			signaltypes.NewMsgTryUpgrade(accAddr),
		},
		user.SetGasLimitAndGasPrice(100_000, appconsts.DefaultMinGasPrice),
	)
	require.NoError(t, err)
	testApp.BeginBlock(abci.RequestBeginBlock{
		Header: tmproto.Header{
			ChainID: genesis.ChainID,
			Height:  3,
			Version: tmversion.Consensus{App: 2},
		},
	})

	deliverTxResp := testApp.DeliverTx(abci.RequestDeliverTx{
		Tx: upgradeTx,
	})
	require.Equal(t, abci.CodeTypeOK, deliverTxResp.Code, deliverTxResp.Log)

	endBlockResp := testApp.EndBlock(abci.RequestEndBlock{
		Height: 3,
	})
	require.Equal(t, v2.Version, endBlockResp.ConsensusParamUpdates.Version.AppVersion)
	require.Equal(t, appconsts.GetTimeoutCommit(v2.Version),
		endBlockResp.Timeouts.TimeoutCommit)
	require.Equal(t, appconsts.GetTimeoutPropose(v2.Version),
		endBlockResp.Timeouts.TimeoutPropose)
	testApp.Commit()
	require.NoError(t, signer.IncrementSequence(testnode.DefaultValidatorAccountName))

	ctx = testApp.NewContext(true)
	getUpgradeResp, err := testApp.SignalKeeper.GetUpgrade(ctx, &signaltypes.QueryGetUpgradeRequest{})
	require.NoError(t, err)
	require.Equal(t, v3.Version, getUpgradeResp.Upgrade.AppVersion)

	initialHeight := int64(4)
	for height := initialHeight; height < initialHeight+appconsts.UpgradeHeightDelay(testApp.ChainID(), v2.Version); height++ {
		appVersion := v2.Version
		_ = testApp.BeginBlock(abci.RequestBeginBlock{
			Header: tmproto.Header{
				Height:  height,
				Version: tmversion.Consensus{App: appVersion},
			},
		})

		endBlockResp = testApp.EndBlock(abci.RequestEndBlock{
			Height: 3 + appconsts.UpgradeHeightDelay(testApp.ChainID(), v2.Version),
		})

		require.Equal(t, appconsts.GetTimeoutCommit(appVersion), endBlockResp.Timeouts.TimeoutCommit)
		require.Equal(t, appconsts.GetTimeoutPropose(appVersion), endBlockResp.Timeouts.TimeoutPropose)

		_ = testApp.Commit()
	}
	require.Equal(t, v3.Version, endBlockResp.ConsensusParamUpdates.Version.AppVersion)

	// confirm that an authored blob tx works
	blob, err := share.NewV1Blob(share.RandomBlobNamespace(), []byte("hello world"), accAddr.Bytes())
	require.NoError(t, err)
	blobTxBytes, _, err := signer.CreatePayForBlobs(
		testnode.DefaultValidatorAccountName,
		[]*share.Blob{blob},
		user.SetGasLimitAndGasPrice(200_000, appconsts.DefaultMinGasPrice),
	)
	require.NoError(t, err)
	blobTx, _, err := tx.UnmarshalBlobTx(blobTxBytes)
	require.NoError(t, err)

	_ = testApp.BeginBlock(abci.RequestBeginBlock{
		Header: tmproto.Header{
			ChainID: genesis.ChainID,
			Height:  initialHeight + appconsts.UpgradeHeightDelay(testApp.ChainID(), v3.Version),
			Version: tmversion.Consensus{App: 3},
		},
	})

	deliverTxResp = testApp.DeliverTx(abci.RequestDeliverTx{
		Tx: blobTx.Tx,
	})
	require.Equal(t, abci.CodeTypeOK, deliverTxResp.Code, deliverTxResp.Log)

	respEndBlock := testApp.EndBlock(abci.
		RequestEndBlock{Height: initialHeight + appconsts.UpgradeHeightDelay(testApp.ChainID(), v3.Version)})
	require.Equal(t, appconsts.GetTimeoutCommit(v3.Version), respEndBlock.Timeouts.TimeoutCommit)
	require.Equal(t, appconsts.GetTimeoutPropose(v3.Version), respEndBlock.Timeouts.TimeoutPropose)
}

// TestAppUpgradeV2 verifies that the all module's params are overridden during an
// upgrade from v1 -> v2 and the app version changes correctly.
func TestAppUpgradeV2(t *testing.T) {
	NetworkMinGasPriceDec, err := math.LegacyNewDecFromStr(fmt.Sprintf("%f", appconsts.DefaultNetworkMinGasPrice))
	require.NoError(t, err)

	tests := []struct {
		module        string
		subspace      string
		key           string
		expectedValue string
	}{
		{
			module:        "MinFee",
			subspace:      minfee.ModuleName,
			key:           string(minfee.KeyNetworkMinGasPrice),
			expectedValue: NetworkMinGasPriceDec.String(),
		},
		{
			module:        "ICA",
			subspace:      icahosttypes.SubModuleName,
			key:           string(icahosttypes.KeyHostEnabled),
			expectedValue: "true",
		},
		// {
		// 	module:        "PFM",
		// 	subspace:      packetforwardtypes.ModuleName,
		// 	key:           string(packetforwardtypes.KeyFeePercentage),
		// 	expectedValue: "0.000000000000000000",
		// },
	}
	for _, tt := range tests {
		t.Run(tt.module, func(t *testing.T) {
			testApp, _ := SetupTestAppWithUpgradeHeight(t, 3)

			ctx := testApp.NewContext(true, tmproto.Header{
				Version: tmversion.Consensus{
					App: 1,
				},
			})
			testApp.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{
				Height:  2,
				Version: tmversion.Consensus{App: 1},
			}})
			// app version should not have changed yet
			appVersion, err := testApp.AppVersion()
			require.NoError(t, err)

			require.EqualValues(t, 1, appVersion)

			// Query the module params
			gotBefore, err := testApp.ParamsKeeper.Params(ctx, &proposal.QueryParamsRequest{
				Subspace: tt.subspace,
				Key:      tt.key,
			})
			require.NoError(t, err)
			require.Equal(t, "", gotBefore.Param.Value)

			// Upgrade from v1 -> v2
			testApp.EndBlock(abci.RequestEndBlock{Height: 2})
			testApp.Commit()

			appVersion, err = testApp.AppVersion()
			require.NoError(t, err)
			require.EqualValues(t, 2, appVersion)

			newCtx := testApp.NewContext(true, tmproto.Header{Version: tmversion.Consensus{App: 2}})
			got, err := testApp.ParamsKeeper.Params(newCtx, &proposal.QueryParamsRequest{
				Subspace: tt.subspace,
				Key:      tt.key,
			})
			require.NoError(t, err)
			require.Equal(t, tt.expectedValue, strings.Trim(got.Param.Value, "\""))
		})
	}
}

// TestBlobstreamRemovedInV2 verifies that the blobstream params exist in v1 and
// do not exist in v2.
func TestBlobstreamRemovedInV2(t *testing.T) {
	testApp, _ := SetupTestAppWithUpgradeHeight(t, 3)
	ctx := testApp.NewContext(true)

	v, err := testApp.AppVersion(ctx)
	require.NoError(t, err)

	require.EqualValues(t, 1, v)
	got, err := testApp.ParamsKeeper.Params(ctx, &proposal.QueryParamsRequest{
		Subspace: blobstreamtypes.ModuleName,
		Key:      string(blobstreamtypes.ParamsStoreKeyDataCommitmentWindow),
	})
	require.NoError(t, err)
	require.Equal(t, "\"400\"", got.Param.Value)

	upgradeFromV1ToV2(t, testApp)

	v, err = testApp.AppVersion(ctx)
	require.NoError(t, err)

	require.EqualValues(t, 2, v)
	_, err = testApp.ParamsKeeper.Params(ctx, &proposal.QueryParamsRequest{
		Subspace: blobstreamtypes.ModuleName,
		Key:      string(blobstreamtypes.ParamsStoreKeyDataCommitmentWindow),
	})
	require.Error(t, err)
}

func SetupTestAppWithUpgradeHeight(t *testing.T, upgradeHeight int64) (*app.App, *genesis.Genesis) {
	t.Helper()

	db := coretesting.NewMemDB()
	encCfg := encoding.MakeConfig(app.ModuleEncodingRegisters...)
	testApp := app.New(log.NewNopLogger(), db, nil, 0, encCfg, upgradeHeight, 0, util.EmptyAppOptions{})
	genesis := genesis.NewDefaultGenesis().
		WithChainID(appconsts.TestChainID).
		WithValidators(genesis.NewDefaultValidator(testnode.DefaultValidatorAccountName)).
		WithConsensusParams(app.DefaultInitialConsensusParams())
	genDoc, err := genesis.Export()
	require.NoError(t, err)
	cp := genDoc.ConsensusParams
	abciParams := &abci.ConsensusParams{
		Block: &abci.BlockParams{
			MaxBytes: cp.Block.MaxBytes,
			MaxGas:   cp.Block.MaxGas,
		},
		Evidence:  &cp.Evidence,
		Validator: &cp.Validator,
		Version:   &cp.Version,
	}

	_ = testApp.InitChain(
		abci.RequestInitChain{
			Time:            genDoc.GenesisTime,
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: abciParams,
			AppStateBytes:   genDoc.AppState,
			ChainId:         genDoc.ChainID,
		},
	)

	// assert that the chain starts with version provided in genesis
	infoResp := testApp.Info(abci.RequestInfo{})
	appVersion := app.DefaultInitialConsensusParams().Version.AppVersion
	require.EqualValues(t, appVersion, infoResp.AppVersion)
	require.EqualValues(t, appconsts.GetTimeoutCommit(appVersion), infoResp.Timeouts.TimeoutCommit)
	require.EqualValues(t, appconsts.GetTimeoutPropose(appVersion), infoResp.Timeouts.TimeoutPropose)

	supportedVersions := []uint64{v1.Version, v2.Version, v3.Version}
	require.Equal(t, supportedVersions, testApp.SupportedVersions())

	_ = testApp.Commit()
	return testApp, genesis
}

func upgradeFromV1ToV2(t *testing.T, testApp *app.App) {
	t.Helper()
	testApp.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{
		Height:  2,
		Version: tmversion.Consensus{App: 1},
	}})
	endBlockResp := testApp.EndBlock(abci.RequestEndBlock{Height: 2})
	require.Equal(t, appconsts.GetTimeoutCommit(v1.Version),
		endBlockResp.Timeouts.TimeoutCommit)
	require.Equal(t, appconsts.GetTimeoutPropose(v1.Version),
		endBlockResp.Timeouts.TimeoutPropose)
	testApp.Commit()
	require.EqualValues(t, 2, testApp.AppVersion())
}
