package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/celestiaorg/celestia-app/v4/x/minfee/types"
)

var _ types.QueryServer = &Keeper{}

// NetworkMinGasPrice returns the network minimum gas price.
func (k *Keeper) NetworkMinGasPrice(ctx context.Context, _ *types.QueryNetworkMinGasPrice) (*types.QueryNetworkMinGasPriceResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	var params types.Params
	subspace, found := k.GetParamsKeeper().GetSubspace(types.ModuleName)
	if !found {
		return nil, status.Errorf(codes.NotFound, "subspace not found for minfee. Minfee is only active in app version 2 and onwards")
	}
	subspace.GetParamSet(sdkCtx, &params)
	return &types.QueryNetworkMinGasPriceResponse{NetworkMinGasPrice: params.NetworkMinGasPrice}, nil
}

func (q *Keeper) Params(ctx context.Context, request *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	panic("implement me")
}
