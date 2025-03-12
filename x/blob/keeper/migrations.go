package keeper

import (
	blobtypes "github.com/celestiaorg/celestia-app/v4/x/blob/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MigrateParams handles the migration of blob module parameters stored in the legacy subspace to the new parameter store.
// It validates the existing parameters and sets them in the updated format using the Keeper's parameter store.
func MigrateParams(ctx sdk.Context, blobKeeper Keeper) error {
	var params blobtypes.Params
	blobKeeper.legacySubspace.GetParamSet(ctx, &params)

	if err := params.Validate(); err != nil {
		return err
	}

	blobKeeper.SetParams(ctx, params)
	return nil
}
