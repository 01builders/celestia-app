package types

import (
	"cosmossdk.io/core/registry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

const URLMsgRegisterEVMAddress = "/celestia.blob.v1.MsgRegisterEVMAddress"

func RegisterLegacyAminoCodec(cdc registry.AminoRegistrar) {
	cdc.RegisterConcrete(&MsgRegisterEVMAddress{}, URLMsgRegisterEVMAddress)
}

func RegisterInterfaces(registry registry.InterfaceRegistrar) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRegisterEVMAddress{},
	)

	registry.RegisterInterface(
		"AttestationRequestI",
		(*AttestationRequestI)(nil),
		&DataCommitment{},
		&Valset{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
