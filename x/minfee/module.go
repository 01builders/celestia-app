package minfee

import (
	"context"
	"encoding/json"
	"fmt"

	grpc "google.golang.org/grpc"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/math"
	params "cosmossdk.io/x/params/keeper"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ appmodule.AppModule  = AppModule{}
	_ appmodule.HasGenesis = AppModule{}
)

// AppModule implements the AppModule interface for the minfee module.
type AppModule struct {
	cdc          codec.Codec
	paramsKeeper params.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, k params.Keeper) AppModule {
	// Register the parameter key table in its associated subspace.
	subspace, exists := k.GetSubspace(ModuleName)
	if !exists {
		panic("minfee subspace not set")
	}
	RegisterMinFeeParamTable(subspace)

	return AppModule{
		cdc:          cdc,
		paramsKeeper: k,
	}
}

func (AppModule) IsAppModule() {}

func (AppModule) IsOnePerModuleType() {}

// Name returns the minfee module's name.
func (AppModule) Name() string {
	return ModuleName
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(registrar grpc.ServiceRegistrar) {
	RegisterQueryServer(registrar, NewQueryServerImpl(am.paramsKeeper))
}

// DefaultGenesis returns default genesis state as raw bytes for the minfee module.
func (am AppModule) DefaultGenesis() json.RawMessage {
	return am.cdc.MustMarshalJSON(DefaultGenesis())
}

// ValidateGenesis performs genesis state validation for the minfee module.
func (am AppModule) ValidateGenesis(bz json.RawMessage) error {
	var data GenesisState
	if err := am.cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", ModuleName, err)
	}

	return ValidateGenesis(&data)
}

// InitGenesis performs genesis initialization for the minfee module.
func (am AppModule) InitGenesis(ctx context.Context, gs json.RawMessage) error {
	var genesisState GenesisState
	if err := am.cdc.UnmarshalJSON(gs, &genesisState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", ModuleName, err)
	}

	subspace, exists := am.paramsKeeper.GetSubspace(ModuleName)
	if !exists {
		return fmt.Errorf("minfee subspace not set")
	}

	subspace = RegisterMinFeeParamTable(subspace)

	// Set the network min gas price initial value
	networkMinGasPriceDec, err := math.LegacyNewDecFromStr(fmt.Sprintf("%f", genesisState.NetworkMinGasPrice))
	if err != nil {
		return fmt.Errorf("failed to convert NetworkMinGasPrice to ")
	}
	subspace.SetParamSet(sdk.UnwrapSDKContext(ctx), &Params{NetworkMinGasPrice: networkMinGasPriceDec})

	return nil
}

// ExportGenesis returns the exported genesis state as raw bytes for the minfee module.
func (am AppModule) ExportGenesis(ctx context.Context) (json.RawMessage, error) {
	gs := ExportGenesis(sdk.UnwrapSDKContext(ctx), am.paramsKeeper)
	return am.cdc.MarshalJSON(gs)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }
