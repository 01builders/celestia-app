package app_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/celestiaorg/celestia-app/v4/app"
	"github.com/celestiaorg/celestia-app/v4/app/encoding"
	"github.com/celestiaorg/celestia-app/v4/pkg/appconsts"
	"github.com/celestiaorg/celestia-app/v4/pkg/user"
	"github.com/celestiaorg/celestia-app/v4/test/txsim"
	"github.com/celestiaorg/celestia-app/v4/test/util/blobfactory"
	"github.com/celestiaorg/celestia-app/v4/test/util/genesis"
	"github.com/celestiaorg/celestia-app/v4/test/util/testfactory"
	"github.com/celestiaorg/celestia-app/v4/test/util/testnode"
	blobtypes "github.com/celestiaorg/celestia-app/v4/x/blob/types"
)

func TestSquareSizeIntegrationTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping square size integration test in short mode.")
	}
	suite.Run(t, new(SquareSizeIntegrationTest))
}

type SquareSizeIntegrationTest struct {
	suite.Suite

	cctx              testnode.Context
	rpcAddr, grpcAddr string
	ecfg              encoding.Config
}

func (s *SquareSizeIntegrationTest) SetupSuite() {
	t := s.T()
	t.Log("setting up square size integration test")
	s.ecfg = encoding.MakeTestConfig(app.ModuleEncodingRegisters...)

	cfg := testnode.DefaultConfig().
		WithModifiers(genesis.ImmediateProposals(s.ecfg.Codec)).WithSuppressLogs(false)

	cctx, rpcAddr, grpcAddr := testnode.NewNetwork(t, cfg)

	s.cctx = cctx
	s.rpcAddr = rpcAddr
	s.grpcAddr = grpcAddr
	err := s.cctx.WaitForNextBlock()
	require.NoError(t, err)
}

// TestSquareSizeUpperBound sets the app's params to specific sizes, then fills the
// block with spam txs to measure that the desired max is getting hit
func (s *SquareSizeIntegrationTest) TestSquareSizeUpperBound() {
	t := s.T()

	const numBlocks = 10

	type test struct {
		name                  string
		govMaxSquareSize      int
		maxBytes              int
		expectedMaxSquareSize int
	}

	tests := []test{
		{
			name:                  "default",
			govMaxSquareSize:      appconsts.DefaultGovMaxSquareSize,
			maxBytes:              appconsts.DefaultMaxBytes,
			expectedMaxSquareSize: appconsts.DefaultGovMaxSquareSize,
		},
		{
			name:                  "max bytes constrains square size",
			govMaxSquareSize:      appconsts.DefaultGovMaxSquareSize,
			maxBytes:              appconsts.DefaultMaxBytes,
			expectedMaxSquareSize: appconsts.DefaultGovMaxSquareSize,
		},
		{
			name:                  "gov square size == hardcoded max",
			govMaxSquareSize:      appconsts.DefaultSquareSizeUpperBound,
			maxBytes:              appconsts.DefaultUpperBoundMaxBytes,
			expectedMaxSquareSize: appconsts.DefaultSquareSizeUpperBound,
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error)
	go func() {
		seqs := txsim.NewBlobSequence(
			txsim.NewRange(100_000, 100_000),
			txsim.NewRange(1, 1),
		).Clone(100)
		err := txsim.Run(
			ctx,
			s.grpcAddr,
			s.cctx.Keyring,
			encoding.MakeTestConfig(app.ModuleEncodingRegisters...),
			txsim.DefaultOptions().
				WithSeed(rand.Int63()).
				WithPollTime(time.Second).
				SuppressLogs(),
			seqs...,
		)
		errCh <- err
	}()

	require.NoError(t, s.cctx.WaitForBlocks(2))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.setBlockSizeParams(t, tt.govMaxSquareSize, tt.maxBytes)
			require.NoError(t, s.cctx.WaitForBlocks(numBlocks))

			// check that we're not going above the specified size and that we hit the specified size
			actualMaxSize := 0
			end, err := s.cctx.LatestHeight()
			require.NoError(t, err)
			for i := end - numBlocks; i < end; i++ {
				block, err := s.cctx.Client.Block(s.cctx.GoContext(), &i)
				require.NoError(t, err)
				require.LessOrEqual(t, block.Block.Data.SquareSize, uint64(tt.govMaxSquareSize))

				if block.Block.Data.SquareSize > uint64(actualMaxSize) {
					actualMaxSize = int(block.Block.Data.SquareSize)
				}
			}

			require.Equal(t, tt.expectedMaxSquareSize, actualMaxSize)
		})
	}
	cancel()
	err := <-errCh
	require.Contains(t, err.Error(), context.Canceled.Error())
}

// setBlockSizeParams will use the validator account to set the square size and
// max bytes parameters. It assumes that the governance params have been set to
// allow for fast acceptance of proposals, and will fail the test if the
// parameters are not set as expected.
func (s *SquareSizeIntegrationTest) setBlockSizeParams(t *testing.T, squareSize, maxBytes int) {
	account := "validator"

	consQueryClient := consensustypes.NewQueryClient(s.cctx.GRPCClient)
	consParamsResp, err := consQueryClient.Params(s.cctx.GoContext(), &consensustypes.QueryParamsRequest{})
	require.NoError(t, err)

	updatedParams := consParamsResp.Params
	updatedParams.Block.MaxBytes = int64(maxBytes)

	content := proposal.NewParameterChangeProposal(
		"x/blob max square size param",
		"param update proposal",
		[]proposal.ParamChange{GovMaxSquareSizeParamChange(squareSize)},
	)

	contentAny, err := codectypes.NewAnyWithValue(content)
	require.NoError(t, err)

	govAuthority := authtypes.NewModuleAddress("gov").String()
	msgExecLegacyContext := govv1.NewMsgExecLegacyContent(contentAny, govAuthority)
	msgUpdateConsensusParams := &consensustypes.MsgUpdateParams{
		Authority: govAuthority,
		Abci:      updatedParams.Abci,
		Block:     updatedParams.Block,
		Evidence:  updatedParams.Evidence,
		Validator: updatedParams.Validator,
	}

	proposerAddr := testfactory.GetAddress(s.cctx.Keyring, account)
	msgSubmitProp, err := govv1.NewMsgSubmitProposal([]sdk.Msg{msgExecLegacyContext, msgUpdateConsensusParams}, sdk.NewCoins(
		sdk.NewCoin(appconsts.BondDenom, math.NewInt(1000000000))), proposerAddr.String(), "meta", "title", "summary", false)
	require.NoError(t, err)

	txClient, err := user.SetupTxClient(s.cctx.GoContext(), s.cctx.Keyring, s.cctx.GRPCClient, s.ecfg)
	require.NoError(t, err)

	res, err := txClient.SubmitTx(s.cctx.GoContext(), []sdk.Msg{msgSubmitProp}, blobfactory.DefaultTxOpts()...)
	require.NoError(t, err)

	txService := sdktx.NewServiceClient(s.cctx.GRPCClient)
	getTxResp, err := txService.GetTx(s.cctx.GoContext(), &sdktx.GetTxRequest{Hash: res.TxHash})
	require.NoError(t, err)
	require.Equal(t, res.Code, abci.CodeTypeOK, getTxResp.TxResponse.RawLog)

	// require.NoError(t, s.cctx.WaitForNextBlock())

	// query the proposal to get the id
	govQueryClient := govv1.NewQueryClient(s.cctx.GRPCClient)
	propResp, err := govQueryClient.Proposals(s.cctx.GoContext(), &govv1.QueryProposalsRequest{ProposalStatus: govv1.StatusVotingPeriod})
	require.NoError(t, err)
	require.Len(t, propResp.Proposals, 1)

	// create and submit a new msgVote
	msgVote := govv1.NewMsgVote(testfactory.GetAddress(s.cctx.Keyring, account), propResp.Proposals[0].Id, govv1.OptionYes, "")
	res, err = txClient.SubmitTx(s.cctx.GoContext(), []sdk.Msg{msgVote}, blobfactory.DefaultTxOpts()...)
	require.NoError(t, err)
	require.Equal(t, abci.CodeTypeOK, res.Code)

	// wait for the voting period to complete
	require.NoError(t, s.cctx.WaitForBlocks(1))

	// check that the parameters got updated as expected
	blobQueryClient := blobtypes.NewQueryClient(s.cctx.GRPCClient)
	blobParamsResp, err := blobQueryClient.Params(s.cctx.GoContext(), &blobtypes.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, uint64(squareSize), blobParamsResp.Params.GovMaxSquareSize)

	// consQueryClient := consensustypes.NewQueryClient(s.cctx.GRPCClient)
	consParamsResp, err = consQueryClient.Params(s.cctx.GoContext(), &consensustypes.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, int64(maxBytes), consParamsResp.Params.Block.MaxBytes)
}

// MaxBlockBytesParamChange returns a param change for the max block bytes
// consensus params.
func MaxBlockBytesParamChange(codec codec.Codec, maxBytes int) proposal.ParamChange {
	bparams := &cmtproto.BlockParams{
		MaxBytes: int64(maxBytes),
		MaxGas:   -1,
	}
	return proposal.NewParamChange(
		baseapp.Paramspace,
		string(baseapp.ParamStoreKeyBlockParams),
		string(codec.MustMarshalJSON(bparams)),
	)
}

// GovMaxSquareSizeParamChange returns a param change for the blob module's max
// square size.
func GovMaxSquareSizeParamChange(squareSize int) proposal.ParamChange {
	return proposal.NewParamChange(
		blobtypes.ModuleName,
		string(blobtypes.KeyGovMaxSquareSize),
		fmt.Sprintf("\"%d\"", squareSize),
	)
}
