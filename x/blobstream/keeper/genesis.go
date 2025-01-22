package keeper

import (
	"context"

	"github.com/celestiaorg/celestia-app/v3/x/blobstream/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// InitialLatestAttestationNonce the initial value set in genesis of the latest attestation
	// nonce value in store.
	InitialLatestAttestationNonce = uint64(0)
	// InitialEarliestAvailableAttestationNonce the initial value set in genesis of the earliest
	/// available attestation nonce in store.
	InitialEarliestAvailableAttestationNonce = uint64(1)
)

// InitGenesis initializes the cap(ability module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx context.Context, genState types.GenesisState) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	k.SetLatestAttestationNonce(sdkCtx, InitialLatestAttestationNonce)
	// The reason we're setting the earliest available nonce to 1 is because at
	// chain startup, a new valset will always be created. Also, it's easier to
	// set it once here rather than conditionally setting it in abci.EndBlocker
	// which is executed on every block.
	k.SetEarliestAvailableAttestationNonce(sdkCtx, InitialEarliestAvailableAttestationNonce)
	k.SetParams(sdkCtx, *genState.Params)

	return nil
}

// ExportGenesis returns the capability module's exported genesis.
func (k Keeper) ExportGenesis(ctx context.Context) *types.GenesisState {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	genesis := types.DefaultGenesis()
	genesis.Params.DataCommitmentWindow = k.GetDataCommitmentWindowParam(sdkCtx)
	return genesis
}
