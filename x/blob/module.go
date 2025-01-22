package blob

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/registry"

	"github.com/celestiaorg/celestia-app/v3/x/blob/client/cli"
	"github.com/celestiaorg/celestia-app/v3/x/blob/keeper"
	"github.com/celestiaorg/celestia-app/v3/x/blob/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
)

var (
	_ module.HasAminoCodec  = AppModule{}
	_ module.HasGRPCGateway = AppModule{}

	_ appmodule.AppModule             = AppModule{}
	_ appmodule.HasGenesis            = AppModule{}
	_ appmodule.HasRegisterInterfaces = AppModule{}
)

// AppModule implements the AppModule interface for the blob module.
type AppModule struct {
	cdc    codec.Codec
	keeper keeper.Keeper
}

func NewAppModule(cdc codec.Codec, keeper keeper.Keeper) AppModule {
	return AppModule{
		cdc:    cdc,
		keeper: keeper,
	}
}

// Name returns the blob module's name.
func (AppModule) Name() string {
	return types.ModuleName
}

func (AppModule) IsAppModule() {}

func (AppModule) IsOnePerModuleType() {}

// RegisterLegacyAminoCodec registers the blob module's types on the LegacyAmino codec.
func (AppModule) RegisterLegacyAminoCodec(registrar registry.AminoRegistrar) {
	types.RegisterLegacyAminoCodec(registrar)
}

// RegisterInterfaces registers interfaces and implementations of the blob module.
func (AppModule) RegisterInterfaces(registrar registry.InterfaceRegistrar) {
	types.RegisterInterfaces(registrar)
}

// DefaultGenesis returns the blob module's default genesis state.
func (am AppModule) DefaultGenesis() json.RawMessage {
	return am.cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis performs genesis state validation for the blob module.
func (am AppModule) ValidateGenesis(bz json.RawMessage) error {
	var genState types.GenesisState
	if err := am.cdc.UnmarshalJSON(bz, &genState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return genState.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the module.
func (am AppModule) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// GetTxCmd returns the blob module's root tx command.
func (AppModule) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// GetQueryCmd returns the blob module's root query command.
// TODO(@julienrbrt): Use AutoCLI
func (AppModule) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// RegisterServices registers a GRPC query service to respond to the
// module-specific GRPC queries.
func (am AppModule) RegisterServices(registrar grpc.ServiceRegistrar) {
	types.RegisterMsgServer(registrar, keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(registrar, am.keeper)
}

// InitGenesis performs the blob module's genesis initialization.
func (am AppModule) InitGenesis(ctx context.Context, gs json.RawMessage) error {
	var genState types.GenesisState
	if err := am.cdc.UnmarshalJSON(gs, &genState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return am.keeper.InitGenesis(ctx, genState)
}

// ExportGenesis returns the blob module's exported genesis state as raw JSON bytes.
func (am AppModule) ExportGenesis(ctx context.Context) (json.RawMessage, error) {
	genState := am.keeper.ExportGenesis(ctx)
	return am.cdc.MarshalJSON(genState)
}

// ConsensusVersion implements ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 2 }
