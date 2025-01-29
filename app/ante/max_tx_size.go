package ante

import (
	"fmt"

	errors "cosmossdk.io/errors"
	"github.com/celestiaorg/celestia-app/v4/pkg/appconsts"
	v3 "github.com/celestiaorg/celestia-app/v4/pkg/appconsts/v3"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerror "github.com/cosmos/cosmos-sdk/types/errors"
)

// MaxTxSizeDecorator ensures that a tx can not be larger than
// application's configured versioned constant.
type MaxTxSizeDecorator struct {
	consensusKeeper ConsensusKeeper
}

func NewMaxTxSizeDecorator(consensusKeeper ConsensusKeeper) MaxTxSizeDecorator {
	return MaxTxSizeDecorator{
		consensusKeeper: consensusKeeper,
	}
}

// AnteHandle implements the AnteHandler interface. It ensures that tx size is under application's configured threshold.
func (d MaxTxSizeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	appVersion, err := d.consensusKeeper.AppVersion(ctx)
	if err != nil {
		return ctx, errors.Wrap(sdkerror.ErrLogic, "failed to get app version")
	}

	// Tx size rule applies to app versions v3 and onwards.
	if appVersion < v3.Version {
		return next(ctx, tx, simulate)
	}

	currentTxSize := len(ctx.TxBytes())
	maxTxSize := appconsts.MaxTxSize(appVersion)
	if currentTxSize > maxTxSize {
		bytesOverLimit := currentTxSize - maxTxSize
		return ctx, fmt.Errorf("tx size %d bytes is larger than the application's configured threshold of %d bytes. Please reduce the size by %d bytes", currentTxSize, maxTxSize, bytesOverLimit)
	}
	return next(ctx, tx, simulate)
}
