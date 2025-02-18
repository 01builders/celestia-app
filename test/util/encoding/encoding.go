package encoding

import (
	"github.com/celestiaorg/celestia-app/v4/app"
	"github.com/celestiaorg/celestia-app/v4/app/encoding"
	"github.com/celestiaorg/celestia-app/v4/app/params"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

func MakeTestConfig() encoding.Config {
	codecOpts := codectestutil.CodecOptions{AccAddressPrefix: params.Bech32PrefixAccAddr, ValAddressPrefix: params.Bech32PrefixValAddr}
	enc := moduletestutil.MakeTestEncodingConfigWithOpts(codecOpts, app.ModuleEncodingRegisters...)
	addressPrefix, validatorPrefix := sdk.GetConfig().GetBech32AccountAddrPrefix(), sdk.GetConfig().GetBech32ValidatorAddrPrefix()
	addressCodec := address.NewBech32Codec(addressPrefix)
	validatorAddressCodec := address.NewBech32Codec(validatorPrefix)
	consensusAddressCodec := address.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix())

	return encoding.Config{
		InterfaceRegistry:     enc.InterfaceRegistry,
		Codec:                 enc.Codec,
		TxConfig:              enc.TxConfig,
		Amino:                 enc.Amino,
		AddressPrefix:         addressPrefix,
		AddressCodec:          addressCodec,
		ValidatorAddressCodec: validatorAddressCodec,
		ConsensusAddressCodec: consensusAddressCodec,
	}
}
