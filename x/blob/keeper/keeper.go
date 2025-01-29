package keeper

import (
	"context"

	"cosmossdk.io/core/appmodule"
	paramtypes "cosmossdk.io/x/params/types"
	"github.com/celestiaorg/celestia-app/v4/pkg/appconsts"
	v2 "github.com/celestiaorg/celestia-app/v4/pkg/appconsts/v2"
	"github.com/celestiaorg/celestia-app/v4/x/blob/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	payForBlobGasDescriptor = "pay for blob"
)

// Keeper handles all the state changes for the blob module.
type Keeper struct {
	appmodule.Environment

	consensusKeeper interface {
		AppVersion(ctx context.Context) (uint64, error)
	}
	cdc        codec.Codec
	paramStore paramtypes.Subspace
}

func NewKeeper(
	env appmodule.Environment,
	cdc codec.Codec,
	ps paramtypes.Subspace,
) *Keeper {
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		Environment: env,
		cdc:         cdc,
		paramStore:  ps,
	}
}

// PayForBlobs consumes gas based on the blob sizes in the MsgPayForBlobs.
func (k Keeper) PayForBlobs(goCtx context.Context, msg *types.MsgPayForBlobs) (*types.MsgPayForBlobsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	appVersion, err := k.consensusKeeper.AppVersion(ctx)
	if err != nil {
		return &types.MsgPayForBlobsResponse{}, err
	}

	// GasPerBlobByte is a versioned param from version 3 onwards.
	var gasToConsume uint64
	if appVersion <= v2.Version {
		gasToConsume = types.GasToConsume(msg.BlobSizes, k.GasPerBlobByte(ctx))
	} else {
		gasToConsume = types.GasToConsume(msg.BlobSizes, appconsts.GasPerBlobByte(appVersion))
	}

	ctx.GasMeter().ConsumeGas(gasToConsume, payForBlobGasDescriptor)

	if err := ctx.EventManager().EmitTypedEvent(
		types.NewPayForBlobsEvent(msg.Signer, msg.BlobSizes, msg.Namespaces),
	); err != nil {
		return &types.MsgPayForBlobsResponse{}, err
	}

	return &types.MsgPayForBlobsResponse{}, nil
}
