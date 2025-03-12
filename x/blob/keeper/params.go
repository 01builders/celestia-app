package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/celestiaorg/celestia-app/v4/x/blob/types"
)

// GetParams returns the total set blob parameters.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(types.ParamsKey))
	if len(bz) == 0 {
		return types.Params{}
	}

	var params types.Params
	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// SetParams sets the total set of blob parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&params)
	store.Set([]byte(types.ParamsKey), bz)
}

// GasPerBlobByte returns the GasPerBlobByte param
func (k Keeper) GasPerBlobByte(ctx sdk.Context) (res uint32) {
	k.legacySubspace.Get(ctx, types.KeyGasPerBlobByte, &res)
	return res
}

// GovMaxSquareSize returns the GovMaxSquareSize param
func (k Keeper) GovMaxSquareSize(ctx sdk.Context) (res uint64) {
	k.legacySubspace.Get(ctx, types.KeyGovMaxSquareSize, &res)
	return res
}
