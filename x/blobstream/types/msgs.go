package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
)

var _ sdk.Msg = &MsgRegisterEVMAddress{}

func NewMsgRegisterEVMAddress(valAddress sdk.ValAddress, evmAddress common.Address) *MsgRegisterEVMAddress {
	msg := &MsgRegisterEVMAddress{
		ValidatorAddress: valAddress.String(),
		EvmAddress:       evmAddress.Hex(),
	}
	return msg
}

// ValidateBasic verifies that the EVM address and val address are of a valid type
func (msg MsgRegisterEVMAddress) ValidateBasic() error {
	_, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return err
	}

	if !common.IsHexAddress(msg.EvmAddress) {
		return ErrEVMAddressNotHex
	}

	return nil
}
