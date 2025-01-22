package types

import (
	"cosmossdk.io/core/registry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers the upgrade types on the provided
// LegacyAmino codec.
func RegisterLegacyAminoCodec(registrar registry.AminoRegistrar) {
	registrar.RegisterConcrete(&MsgTryUpgrade{}, URLMsgTryUpgrade)
	registrar.RegisterConcrete(&MsgSignalVersion{}, URLMsgSignalVersion)
}

// RegisterInterfaces registers the upgrade module types on the provided
// registry.
func RegisterInterfaces(registrar registry.InterfaceRegistrar) {
	registrar.RegisterImplementations((*sdk.Msg)(nil), &MsgTryUpgrade{})
	registrar.RegisterImplementations((*sdk.Msg)(nil), &MsgSignalVersion{})
	msgservice.RegisterMsgServiceDesc(registrar, &_Msg_serviceDesc)
}
