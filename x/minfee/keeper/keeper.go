package keeper

//
//import (
//	storetypes "cosmossdk.io/store/types"
//	"github.com/celestiaorg/celestia-app/v4/x/minfee"
//	"github.com/cosmos/cosmos-sdk/codec"
//	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
//)
//
//type Keeper struct {
//	cdc            codec.Codec
//	storeKey       storetypes.StoreKey
//	legacySubspace paramtypes.Subspace
//	authority      string
//}
//
//func NewKeeper(
//	cdc codec.Codec,
//	storeKey storetypes.StoreKey,
//	legacySubspace paramtypes.Subspace,
//	authority string,
//) *Keeper {
//	if !legacySubspace.HasKeyTable() {
//		legacySubspace = legacySubspace.WithKeyTable(minfee.ParamKeyTable())
//	}
//
//	return &Keeper{
//		cdc:            cdc,
//		storeKey:       storeKey,
//		legacySubspace: legacySubspace,
//		authority:      authority,
//	}
//}
//
//// GetAuthority returns the client submodule's authority.
//func (k Keeper) GetAuthority() string {
//	return k.authority
//}
