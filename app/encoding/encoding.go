package encoding

import (
	txdecode "cosmossdk.io/x/tx/decode"
	"cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/gogoproto/proto"
)

// Config specifies the concrete encoding types to use for a given app.
// This is provided for compatibility between protobuf and amino implementations.
type Config struct {
	InterfaceRegistry codectypes.InterfaceRegistry
	Codec             codec.Codec
	TxConfig          client.TxConfig
	Amino             *codec.LegacyAmino
}

// MakeConfig returns an encoding config for the app.
func MakeConfig() Config {
	interfaceRegistry, _ := codectypes.NewInterfaceRegistryWithOptions(codectypes.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec:          address.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
			ValidatorAddressCodec: address.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		},
	})
	amino := codec.NewLegacyAmino()
	signingCtx := interfaceRegistry.SigningContext()

	// Register the standard types from the Cosmos SDK on interfaceRegistry and
	// amino.
	std.RegisterInterfaces(interfaceRegistry)
	std.RegisterLegacyAminoCodec(amino)

	protoCodec := codec.NewProtoCodec(interfaceRegistry)
	dec, err := txdecode.NewDecoder(txdecode.Options{
		SigningContext: signingCtx,
		ProtoCodec:     protoCodec,
	})
	if err != nil {
		panic(err)
	}
	txDecoder := authtx.DefaultTxDecoder(signingCtx.AddressCodec(), protoCodec, dec)
	txDecoder = indexWrapperDecoder(txDecoder)

	txConfig, err := authtx.NewTxConfigWithOptions(protoCodec, authtx.ConfigOptions{
		EnabledSignModes: authtx.DefaultSignModes,
		SigningOptions: &signing.Options{
			AddressCodec:          signingCtx.AddressCodec(),
			ValidatorAddressCodec: signingCtx.ValidatorAddressCodec(),
		},
		ProtoDecoder: txDecoder,
	})
	if err != nil {
		panic(err)
	}

	return Config{
		InterfaceRegistry: interfaceRegistry,
		Codec:             protoCodec,
		TxConfig:          txConfig,
		Amino:             amino,
	}
}
