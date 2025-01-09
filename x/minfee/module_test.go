package minfee_test

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	paramkeeper "cosmossdk.io/x/params/keeper"
	paramtypes "cosmossdk.io/x/params/types"
	"github.com/celestiaorg/celestia-app/v3/x/minfee"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/stretchr/testify/require"
)

func TestNewModuleInitializesKeyTable(t *testing.T) {
	kvStoreKey := storetypes.NewKVStoreKey(paramtypes.StoreKey)
	tStoreKey := storetypes.NewTransientStoreKey(paramtypes.TStoreKey)
	_ = testutil.DefaultContextWithDB(t, kvStoreKey, tStoreKey)

	registry := codectypes.NewInterfaceRegistry()

	// Create a params keeper
	cdc := codec.NewProtoCodec(registry)
	paramsKeeper := paramkeeper.NewKeeper(codec.NewProtoCodec(registry), codec.NewLegacyAmino(), kvStoreKey, tStoreKey)
	subspace := paramsKeeper.Subspace(minfee.ModuleName)

	// Initialize the minfee module which registers the key table
	minfee.NewAppModule(cdc, paramsKeeper)

	// Require key table to be initialized
	hasKeyTable := subspace.HasKeyTable()
	require.True(t, hasKeyTable)
}
