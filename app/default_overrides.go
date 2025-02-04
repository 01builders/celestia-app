package app

import (
	"encoding/json"
	"fmt"
	"time"

	"cosmossdk.io/math"
	"cosmossdk.io/x/bank"
	banktypes "cosmossdk.io/x/bank/types"
	"cosmossdk.io/x/gov"
	govtypes "cosmossdk.io/x/gov/types/v1"
	"cosmossdk.io/x/slashing"
	slashingtypes "cosmossdk.io/x/slashing/types"
	"cosmossdk.io/x/staking"
	stakingtypes "cosmossdk.io/x/staking/types"
	"github.com/celestiaorg/celestia-app/v4/pkg/appconsts"
	"github.com/celestiaorg/celestia-app/v4/x/mint"
	minttypes "github.com/celestiaorg/celestia-app/v4/x/mint/types"
	tmproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	tmcfg "github.com/cometbft/cometbft/config"
	coretypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/codec"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ica "github.com/cosmos/ibc-go/v9/modules/apps/27-interchain-accounts"
	icagenesistypes "github.com/cosmos/ibc-go/v9/modules/apps/27-interchain-accounts/genesis/types"
	ibc "github.com/cosmos/ibc-go/v9/modules/core"
	ibctypes "github.com/cosmos/ibc-go/v9/modules/core/types"
)

// bankModule defines a custom wrapper around the x/bank module's AppModuleBasic
// implementation to provide custom default genesis state.
type bankModule struct {
	bank.AppModule
	codec.Codec
}

// DefaultGenesis returns custom x/bank module genesis state.
func (m bankModule) DefaultGenesis() json.RawMessage {
	metadata := banktypes.Metadata{
		Description: "The native token of the Celestia network.",
		Base:        BondDenom,
		Name:        DisplayDenom,
		Display:     DisplayDenom,
		Symbol:      DisplayDenom,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    BondDenom,
				Exponent: 0,
				Aliases: []string{
					BondDenomAlias,
				},
			},
			{
				Denom:    DisplayDenom,
				Exponent: 6,
				Aliases:  []string{},
			},
		},
	}

	genState := banktypes.DefaultGenesisState()
	genState.DenomMetadata = append(genState.DenomMetadata, metadata)

	return m.Codec.MustMarshalJSON(genState)
}

// stakingModule wraps the x/staking module in order to overwrite specific
// ModuleManager APIs.
type stakingModule struct {
	staking.AppModule
	codec.Codec
}

// DefaultGenesis returns custom x/staking module genesis state.
func (m stakingModule) DefaultGenesis() json.RawMessage {
	genesis := stakingtypes.DefaultGenesisState()
	genesis.Params.UnbondingTime = appconsts.DefaultUnbondingTime
	genesis.Params.BondDenom = BondDenom
	genesis.Params.MinCommissionRate = math.LegacyNewDecWithPrec(5, 2) // 5%

	return m.Codec.MustMarshalJSON(genesis)
}

// stakingModule wraps the x/staking module in order to overwrite specific
// ModuleManager APIs.
type slashingModule struct {
	slashing.AppModule
	codec.Codec
}

// DefaultGenesis returns custom x/staking module genesis state.
func (m slashingModule) DefaultGenesis() json.RawMessage {
	genesis := slashingtypes.DefaultGenesisState()
	genesis.Params.MinSignedPerWindow = math.LegacyNewDecWithPrec(75, 2) // 75%
	genesis.Params.SignedBlocksWindow = 5000
	genesis.Params.DowntimeJailDuration = time.Minute * 1
	genesis.Params.SlashFractionDoubleSign = math.LegacyNewDecWithPrec(2, 2) // 2%
	genesis.Params.SlashFractionDowntime = math.LegacyZeroDec()              // 0%

	return m.Codec.MustMarshalJSON(genesis)
}

type ibcModule struct {
	ibc.AppModule
	codec.Codec
}

// DefaultGenesis returns custom x/ibc module genesis state.
func (m ibcModule) DefaultGenesis() json.RawMessage {
	// per ibc documentation, this value should be 3-5 times the expected block
	// time. The expected block time is 15 seconds, therefore this value is 75
	// seconds.
	maxBlockTime := appconsts.GoalBlockTime * 5
	gs := ibctypes.DefaultGenesisState()
	gs.ClientGenesis.Params.AllowedClients = []string{"06-solomachine", "07-tendermint"}
	gs.ConnectionGenesis.Params.MaxExpectedTimePerBlock = uint64(maxBlockTime.Nanoseconds())

	return m.Codec.MustMarshalJSON(gs)
}

// icaModule defines a custom wrapper around the ica module to provide custom
// default genesis state.
type icaModule struct {
	ica.AppModule
	codec.Codec
}

// DefaultGenesis returns custom ica module genesis state.
func (m icaModule) DefaultGenesis() json.RawMessage {
	gs := icagenesistypes.DefaultGenesis()
	gs.HostGenesisState.Params.AllowMessages = icaAllowMessages()
	gs.HostGenesisState.Params.HostEnabled = true
	gs.ControllerGenesisState.Params.ControllerEnabled = false
	return m.Codec.MustMarshalJSON(gs)
}

type mintModule struct {
	mint.AppModule
	codec.Codec
}

// DefaultGenesis returns custom x/mint module genesis state.
func (m mintModule) DefaultGenesis() json.RawMessage {
	genState := minttypes.DefaultGenesisState()
	genState.BondDenom = BondDenom

	return m.Codec.MustMarshalJSON(genState)
}

// govModule is a custom wrapper around the x/gov module's AppModuleBasic
// implementation to provide custom default genesis state.
type govModule struct {
	gov.AppModule
	codec.Codec
}

// DefaultGenesis returns custom x/gov module genesis state.
func (m govModule) DefaultGenesis() json.RawMessage {
	genState := govtypes.DefaultGenesisState()
	day := time.Hour * 24
	oneWeek := day * 7

	genState.Params.MinDeposit = sdk.NewCoins(sdk.NewCoin(BondDenom, math.NewInt(10_000_000_000))) // 10,000 TIA
	genState.Params.MaxDepositPeriod = &oneWeek
	genState.Params.VotingPeriod = &oneWeek

	return m.Codec.MustMarshalJSON(genState)
}

// DefaultConsensusParams returns a ConsensusParams with a MaxBytes
// determined using a goal square size.
func DefaultConsensusParams() *tmproto.ConsensusParams {
	return &tmproto.ConsensusParams{
		Block:    DefaultBlockParams(),
		Evidence: DefaultEvidenceParams(),
		Validator: &tmproto.ValidatorParams{
			PubKeyTypes: coretypes.DefaultValidatorParams().PubKeyTypes,
		}, Version: &tmproto.VersionParams{
			App: appconsts.LatestVersion,
		},
	}
}

func DefaultInitialConsensusParams() *tmproto.ConsensusParams {
	return &tmproto.ConsensusParams{
		Block:    DefaultBlockParams(),
		Evidence: DefaultEvidenceParams(),
		Validator: &tmproto.ValidatorParams{
			PubKeyTypes: coretypes.DefaultValidatorParams().PubKeyTypes,
		},
		Version: &tmproto.VersionParams{
			App: DefaultInitialVersion,
		},
	}
}

// DefaultBlockParams returns a default BlockParams with a MaxBytes determined
// using a goal square size.
func DefaultBlockParams() *tmproto.BlockParams {
	return &tmproto.BlockParams{
		MaxBytes: appconsts.DefaultMaxBytes,
		MaxGas:   -1,
	}
}

// DefaultEvidenceParams returns a default EvidenceParams with a MaxAge
// determined using a goal block time.
func DefaultEvidenceParams() *tmproto.EvidenceParams {
	evdParams := coretypes.DefaultEvidenceParams()
	evdParams.MaxAgeDuration = appconsts.DefaultUnbondingTime
	evdParams.MaxAgeNumBlocks = int64(appconsts.DefaultUnbondingTime.Seconds())/int64(appconsts.GoalBlockTime.Seconds()) + 1
	return &tmproto.EvidenceParams{
		MaxAgeNumBlocks: evdParams.MaxAgeNumBlocks,
		MaxAgeDuration:  evdParams.MaxAgeDuration,
		MaxBytes:        evdParams.MaxBytes,
	}
}

func DefaultConsensusConfig() *tmcfg.Config {
	cfg := tmcfg.DefaultConfig()
	// Set broadcast timeout to be 50 seconds in order to avoid timeouts for long block times
	cfg.RPC.TimeoutBroadcastTxCommit = 50 * time.Second
	cfg.RPC.MaxBodyBytes = int64(8388608) // 8 MiB

	// TODO: check if priority mempool needed
	// cfg.Mempool.TTLNumBlocks = 12
	// cfg.Mempool.TTLDuration = 75 * time.Second
	cfg.Mempool.MaxTxBytes = 7_897_088
	cfg.Mempool.MaxTxsBytes = 39_485_440
	cfg.Mempool.Type = "flood" // flood mempool

	cfg.Consensus.TimeoutPropose = appconsts.GetTimeoutPropose(appconsts.LatestVersion)
	cfg.Consensus.TimeoutCommit = appconsts.GetTimeoutCommit(appconsts.LatestVersion)

	cfg.TxIndex.Indexer = "null"
	cfg.Storage.DiscardABCIResponses = true

	const mebibyte = 1048576
	cfg.P2P.SendRate = 10 * mebibyte
	cfg.P2P.RecvRate = 10 * mebibyte

	return cfg
}

func DefaultAppConfig() *serverconfig.Config {
	cfg := serverconfig.DefaultConfig()
	cfg.API.Enable = false
	cfg.GRPC.Enable = false

	// the default snapshot interval was determined by picking a large enough
	// value as to not dramatically increase resource requirements while also
	// being greater than zero so that there are more nodes that will serve
	// snapshots to nodes that state sync
	cfg.StateSync.SnapshotInterval = 1500
	cfg.StateSync.SnapshotKeepRecent = 2
	cfg.MinGasPrices = fmt.Sprintf("%v%s", appconsts.DefaultMinGasPrice, BondDenom)
	return cfg
}
