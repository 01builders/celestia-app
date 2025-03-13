package keeper

import (
	"context"

	"cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	params "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/celestiaorg/celestia-app/v4/x/minfee/types"
)

type Keeper struct {
	cdc            codec.Codec
	storeKey       storetypes.StoreKey
	paramsKeeper   params.Keeper
	legacySubspace paramtypes.Subspace
	authority      string
}

func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	paramsKeeper params.Keeper,
	legacySubspace paramtypes.Subspace,
	authority string,
) *Keeper {
	if !legacySubspace.HasKeyTable() {
		legacySubspace = legacySubspace.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:            cdc,
		storeKey:       storeKey,
		paramsKeeper:   paramsKeeper,
		legacySubspace: legacySubspace,
		authority:      authority,
	}
}

// GetParamsKeeper returns the params keeper.
func (k Keeper) GetParamsKeeper() params.Keeper {
	return k.paramsKeeper
}

// GetAuthority returns the client submodule's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// UpdateMinfeeParams updates minfee module parameters.
func (k Keeper) UpdateMinfeeParams(goCtx context.Context, msg *types.MsgUpdateMinfeeParams) (*types.MsgUpdateMinfeeParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// ensure that the sender has the authority to update the parameters.
	if msg.Authority != k.GetAuthority() {
		return nil, errors.Wrapf(sdkerrors.ErrUnauthorized, "invalid authority: expected: %s, got: %s", k.authority, msg.Authority)
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid parameters: %s", err)
	}

	k.SetParams(ctx, msg.Params)

	// Emit an event indicating successful parameter update.
	if err := ctx.EventManager().EmitTypedEvent(
		types.NewUpdateMinfeeParamsEvent(msg.Authority, msg.Params),
	); err != nil {
		return nil, err
	}

	return &types.MsgUpdateMinfeeParamsResponse{}, nil
}
