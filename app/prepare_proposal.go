package app

import (
	"fmt"
	"time"

	"github.com/celestiaorg/celestia-app/v4/app/ante"
	"github.com/celestiaorg/celestia-app/v4/pkg/appconsts"
	"github.com/celestiaorg/celestia-app/v4/pkg/da"
	shares "github.com/celestiaorg/go-square/shares"
	square "github.com/celestiaorg/go-square/square"
	squarev2 "github.com/celestiaorg/go-square/v2"
	sharev2 "github.com/celestiaorg/go-square/v2/share"
	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PrepareProposal fulfills the celestia-core version of the ABCI interface by
// preparing the proposal block data. This method generates the data root for
// the proposal block and passes it back to tendermint via the BlockData. Panics
// indicate a developer error and should immediately halt the node for
// visibility and so they can be quickly resolved.
func (app *App) PrepareProposalHandler(ctx sdk.Context, req *abci.PrepareProposalRequest) (*abci.PrepareProposalResponse, error) {
	defer telemetry.MeasureSince(time.Now(), "prepare_proposal")
	// Create a context using a branch of the state.

	appVersion, err := app.ConsensusKeeper.AppVersion(ctx)
	if err != nil {
		logInvalidPropBlockError(app.Logger(), ctx.BlockHeader(), "failure to get app version", err)
	}

	handler := ante.NewAnteHandler(
		app.AuthKeeper,
		app.AccountsKeeper,
		app.BankKeeper,
		app.BlobKeeper,
		app.ConsensusKeeper,
		app.FeeGrantKeeper,
		app.GetTxConfig().SignModeHandler(),
		ante.DefaultSigVerificationGasConsumer,
		app.IBCKeeper,
		app.ParamsKeeper,
		app.BlockedParamsGovernance(),
	)

	// Filter out invalid transactions.
	txs := FilterTxs(app.Logger(), ctx, handler, app.encodingConfig.TxConfig, req.Txs)

	// Build the square from the set of valid and prioritised transactions.
	// The txs returned are the ones used in the square and block.
	var (
		dataSquareBytes [][]byte
		size            uint64
	)

	switch appVersion {
	case v4, v3:
		var dataSquare squarev2.Square
		dataSquare, txs, err = squarev2.Build(txs,
			app.MaxEffectiveSquareSize(ctx),
			appconsts.SubtreeRootThreshold(appVersion),
		)
		dataSquareBytes = sharev2.ToBytes(dataSquare)
		size = uint64(dataSquare.Size())
	case v2, v1:
		var dataSquare square.Square
		dataSquare, txs, err = square.Build(txs,
			app.MaxEffectiveSquareSize(ctx),
			appconsts.SubtreeRootThreshold(appVersion),
		)
		dataSquareBytes = shares.ToBytes(dataSquare)
		size = uint64(dataSquare.Size())
	default:
		err = fmt.Errorf("unsupported app version: %d", appVersion)
	}
	if err != nil {
		panic(err)
	}

	// Erasure encode the data square to create the extended data square (eds).
	// Note: uses the nmt wrapper to construct the tree. See
	// pkg/wrapper/nmt_wrapper.go for more information.
	eds, err := da.ExtendShares(dataSquareBytes)
	if err != nil {
		app.Logger().Error(
			"failure to erasure the data square while creating a proposal block",
			"error",
			err.Error(),
		)
		panic(err)
	}

	dah, err := da.NewDataAvailabilityHeader(eds)
	if err != nil {
		app.Logger().Error(
			"failure to create new data availability header",
			"error",
			err.Error(),
		)
		panic(err)
	}

	// Tendermint doesn't need to use any of the erasure data because only the
	// protobuf encoded version of the block data is gossiped. Therefore, the
	// eds is not returned here.
	return &abci.PrepareProposalResponse{
		Txs:          txs,
		SquareSize:   size,
		DataRootHash: dah.Hash(), // also known as the data root
	}, nil
}
