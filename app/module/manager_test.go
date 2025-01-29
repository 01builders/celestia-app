package module_test

import (
	"encoding/json"
	"testing"

	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/celestiaorg/celestia-app/v4/app/module"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestManagerOrderSetters(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)
	mockAppModule1 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
	mockAppModule2 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)

	mockAppModule1.EXPECT().Name().Times(6).Return("module1")
	mockAppModule1.EXPECT().ConsensusVersion().Times(1).Return(uint64(1))
	mockAppModule2.EXPECT().Name().Times(6).Return("module2")
	mockAppModule2.EXPECT().ConsensusVersion().Times(1).Return(uint64(1))
	mm, err := module.NewManager([]module.VersionedModule{
		{Module: mockAppModule1, FromVersion: 1, ToVersion: 1},
		{Module: mockAppModule2, FromVersion: 1, ToVersion: 1},
	})
	require.NoError(t, err)
	require.NotNil(t, mm)
	require.Equal(t, 2, len(mm.ModuleNames(1)))

	require.Equal(t, []string{"module1", "module2"}, mm.OrderInitGenesis)
	mm.SetOrderInitGenesis("module2", "module1")
	require.Equal(t, []string{"module2", "module1"}, mm.OrderInitGenesis)

	require.Equal(t, []string{"module1", "module2"}, mm.OrderExportGenesis)
	mm.SetOrderExportGenesis("module2", "module1")
	require.Equal(t, []string{"module2", "module1"}, mm.OrderExportGenesis)

	require.Equal(t, []string{"module1", "module2"}, mm.OrderBeginBlockers)
	mm.SetOrderBeginBlockers("module2", "module1")
	require.Equal(t, []string{"module2", "module1"}, mm.OrderBeginBlockers)

	require.Equal(t, []string{"module1", "module2"}, mm.OrderEndBlockers)
	mm.SetOrderEndBlockers("module2", "module1")
	require.Equal(t, []string{"module2", "module1"}, mm.OrderEndBlockers)
}

func TestManager_RegisterInvariants(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAppModule1 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
	mockAppModule2 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
	mockAppModule1.EXPECT().Name().Times(2).Return("module1")
	mockAppModule1.EXPECT().ConsensusVersion().Times(1).Return(uint64(1))
	mockAppModule2.EXPECT().Name().Times(2).Return("module2")
	mockAppModule2.EXPECT().ConsensusVersion().Times(1).Return(uint64(1))
	mm, err := module.NewManager([]module.VersionedModule{
		{Module: mockAppModule1, FromVersion: 1, ToVersion: 1},
		{Module: mockAppModule2, FromVersion: 1, ToVersion: 1},
	})
	require.NoError(t, err)
	require.NotNil(t, mm)
	require.Equal(t, 2, len(mm.ModuleNames(1)))

	// test RegisterInvariants
	// TODO: Can add these mocks to the generator if its really necessary
	// mockInvariantRegistry := mock.NewMockInvariantRegistry(mockCtrl)
	// mockAppModule1.EXPECT().RegisterInvariants(gomock.Eq(mockInvariantRegistry)).Times(1)
	// mockAppModule2.EXPECT().RegisterInvariants(gomock.Eq(mockInvariantRegistry)).Times(1)
	// mm.RegisterInvariants(mockInvariantRegistry)
}

func TestManager_RegisterQueryServices(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAppModule1 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
	mockAppModule2 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
	mockAppModule1.EXPECT().Name().Times(3).Return("module1")
	mockAppModule1.EXPECT().ConsensusVersion().Times(2).Return(uint64(1))
	mockAppModule2.EXPECT().Name().Times(3).Return("module2")
	mockAppModule2.EXPECT().ConsensusVersion().Times(2).Return(uint64(1))
	mm, err := module.NewManager([]module.VersionedModule{
		{Module: mockAppModule1, FromVersion: 1, ToVersion: 1},
		{Module: mockAppModule2, FromVersion: 1, ToVersion: 1},
	})
	require.NoError(t, err)
	require.NotNil(t, mm)
	require.Equal(t, 2, len(mm.ModuleNames(1)))

	msgRouter := mock.NewMockServer(mockCtrl)
	queryRouter := mock.NewMockServer(mockCtrl)
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	cfg := module.NewConfigurator(cdc, msgRouter, queryRouter)
	mockAppModule1.EXPECT().RegisterServices(gomock.Any()).Times(1)
	mockAppModule2.EXPECT().RegisterServices(gomock.Any()).Times(1)

	mm.RegisterServices(cfg)
}

func TestManager_InitGenesis(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAppModule1 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
	mockAppModule2 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
	mockAppModule1.EXPECT().Name().Times(2).Return("module1")
	mockAppModule1.EXPECT().ConsensusVersion().Times(1).Return(uint64(1))
	mockAppModule2.EXPECT().Name().Times(2).Return("module2")
	mockAppModule2.EXPECT().ConsensusVersion().Times(1).Return(uint64(1))
	mm, err := module.NewManager([]module.VersionedModule{
		{Module: mockAppModule1, FromVersion: 1, ToVersion: 1},
		{Module: mockAppModule2, FromVersion: 1, ToVersion: 1},
	})
	require.NoError(t, err)
	require.NotNil(t, mm)
	require.Equal(t, 2, len(mm.ModuleNames(1)))

	ctx := sdk.NewContext(nil, false, log.NewNopLogger())
	genesisData := map[string]json.RawMessage{"module1": json.RawMessage(`{"key": "value"}`)}

	// this should panic since the validator set is empty even after init genesis
	mockAppModule1.EXPECT().InitGenesis(gomock.Eq(ctx), gomock.Eq(genesisData["module1"])).Times(1).Return(nil)
	require.Panics(t, func() { mm.InitGenesis(ctx, genesisData, 1) })

	// test panic
	genesisData = map[string]json.RawMessage{
		"module1": json.RawMessage(`{"key": "value"}`),
		"module2": json.RawMessage(`{"key": "value"}`),
	}
	mockAppModule1.EXPECT().InitGenesis(gomock.Eq(ctx), gomock.Eq(genesisData["module1"])).Times(1).Return([]abci.ValidatorUpdate{{}})
	mockAppModule2.EXPECT().InitGenesis(gomock.Eq(ctx), gomock.Eq(genesisData["module2"])).Times(1).Return([]abci.ValidatorUpdate{{}})
	require.Panics(t, func() { mm.InitGenesis(ctx, genesisData, 1) })
}

func TestManager_ExportGenesis(t *testing.T) {
	t.Run("export genesis with two modules at version 1", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		t.Cleanup(mockCtrl.Finish)

		mockAppModule1 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
		mockAppModule2 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
		mockAppModule1.EXPECT().Name().Times(2).Return("module1")
		mockAppModule1.EXPECT().ConsensusVersion().Times(1).Return(uint64(1))
		mockAppModule2.EXPECT().Name().Times(2).Return("module2")
		mockAppModule2.EXPECT().ConsensusVersion().Times(1).Return(uint64(1))
		mm, err := module.NewManager([]module.VersionedModule{
			{Module: mockAppModule1, FromVersion: 1, ToVersion: 1},
			{Module: mockAppModule2, FromVersion: 1, ToVersion: 1},
		})
		require.NoError(t, err)
		require.NotNil(t, mm)
		require.Equal(t, 2, len(mm.ModuleNames(1)))

		ctx := sdk.Context{}
		interfaceRegistry := types.NewInterfaceRegistry()
		cdc := codec.NewProtoCodec(interfaceRegistry)
		mockAppModule1.EXPECT().ExportGenesis(gomock.Eq(ctx)).Times(1).Return(json.RawMessage(`{"key1": "value1"}`))
		mockAppModule2.EXPECT().ExportGenesis(gomock.Eq(ctx)).Times(1).Return(json.RawMessage(`{"key2": "value2"}`))

		want := map[string]json.RawMessage{
			"module1": json.RawMessage(`{"key1": "value1"}`),
			"module2": json.RawMessage(`{"key2": "value2"}`),
		}
		exported, err := mm.ExportGenesis(ctx, cdc, 1)
		require.NoError(t, err)
		require.Equal(t, want, exported)
	})
	t.Run("export genesis with one modules at version 1, one modules at version 2", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		t.Cleanup(mockCtrl.Finish)

		mockAppModule1 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
		mockAppModule2 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
		mockAppModule1.EXPECT().Name().Times(2).Return("module1")
		mockAppModule1.EXPECT().ConsensusVersion().Times(2).Return(uint64(1))
		mockAppModule2.EXPECT().Name().Times(2).Return("module2")
		mockAppModule2.EXPECT().ConsensusVersion().Times(2).Return(uint64(1))
		mm, err := module.NewManager([]module.VersionedModule{
			{Module: mockAppModule1, FromVersion: 1, ToVersion: 1},
			{Module: mockAppModule2, FromVersion: 2, ToVersion: 2},
		})
		require.NoError(t, err)
		require.NotNil(t, mm)
		require.Equal(t, 1, len(mm.ModuleNames(1)))
		require.Equal(t, 1, len(mm.ModuleNames(2)))

		ctx := sdk.Context{}
		interfaceRegistry := types.NewInterfaceRegistry()
		cdc := codec.NewProtoCodec(interfaceRegistry)
		mockAppModule1.EXPECT().ExportGenesis(gomock.Eq(ctx)).Times(1).Return(json.RawMessage(`{"key1": "value1"}`))
		mockAppModule2.EXPECT().ExportGenesis(gomock.Eq(ctx)).Times(1).Return(json.RawMessage(`{"key2": "value2"}`))

		want := map[string]json.RawMessage{
			"module1": json.RawMessage(`{"key1": "value1"}`),
		}
		exported, err := mm.ExportGenesis(ctx, cdc, 1)
		require.NoError(t, err)
		assert.Equal(t, want, exported)

		want2 := map[string]json.RawMessage{
			"module2": json.RawMessage(`{"key2": "value2"}`),
		}
		exported, err = mm.ExportGenesis(ctx, cdc, 2)
		require.NoError(t, err)
		assert.Equal(t, want2, exported)
	})
}

func TestManager_BeginBlock(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAppModule1 := mock.NewMockAppModuleWithAllExtensionsABCI(mockCtrl)
	mockAppModule2 := mock.NewMockAppModuleWithAllExtensionsABCI(mockCtrl)
	mockAppModule1.EXPECT().Name().Times(2).Return("module1")
	mockAppModule1.EXPECT().ConsensusVersion().Times(1).Return(uint64(1))
	mockAppModule2.EXPECT().Name().Times(2).Return("module2")
	mockAppModule2.EXPECT().ConsensusVersion().Times(1).Return(uint64(1))
	mm, err := module.NewManager([]module.VersionedModule{
		{Module: mockAppModule1, FromVersion: 1, ToVersion: 1},
		{Module: mockAppModule2, FromVersion: 1, ToVersion: 1},
	})
	require.NoError(t, err)
	require.NotNil(t, mm)
	require.Equal(t, 2, len(mm.ModuleNames(1)))

	mockAppModule1.EXPECT().BeginBlock(gomock.Any()).Times(1)
	mockAppModule2.EXPECT().BeginBlock(gomock.Any()).Times(1)
	ctx := sdk.NewContext(nil, false, log.NewNopLogger())
	_, err = mm.BeginBlock(ctx)
	require.NoError(t, err)
}

func TestManager_EndBlock(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAppModule1 := mock.NewMockAppModuleWithAllExtensionsABCI(mockCtrl)
	mockAppModule2 := mock.NewMockAppModuleWithAllExtensionsABCI(mockCtrl)
	mockAppModule1.EXPECT().Name().Times(2).Return("module1")
	mockAppModule1.EXPECT().ConsensusVersion().Times(1).Return(uint64(1))
	mockAppModule2.EXPECT().Name().Times(2).Return("module2")
	mockAppModule2.EXPECT().ConsensusVersion().Times(1).Return(uint64(1))
	mm, err := module.NewManager([]module.VersionedModule{
		{Module: mockAppModule1, FromVersion: 1, ToVersion: 1},
		{Module: mockAppModule2, FromVersion: 1, ToVersion: 1},
	})
	require.NoError(t, err)
	require.NotNil(t, mm)
	require.Equal(t, 2, len(mm.ModuleNames(1)))

	mockAppModule1.EXPECT().EndBlock(gomock.Any()).Times(1).Return([]abci.ValidatorUpdate{{}})
	mockAppModule2.EXPECT().EndBlock(gomock.Any()).Times(1)
	ctx := sdk.NewContext(nil, false, log.NewNopLogger())
	ret, err := mm.EndBlock(ctx)
	require.NoError(t, err)
	require.Equal(t, []abci.ValidatorUpdate{{}}, ret.ValidatorUpdates)

	// test panic
	mockAppModule1.EXPECT().EndBlock(gomock.Any()).Times(1).Return([]abci.ValidatorUpdate{{}})
	mockAppModule2.EXPECT().EndBlock(gomock.Any()).Times(1).Return([]abci.ValidatorUpdate{{}})
	require.Panics(t, func() { mm.EndBlock(ctx) })
}

func TestManager_UpgradeSchedule(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAppModule1 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
	mockAppModule2 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
	mockAppModule1.EXPECT().Name().Times(2).Return("blob")
	mockAppModule2.EXPECT().Name().Times(2).Return("blob")
	mockAppModule1.EXPECT().ConsensusVersion().Times(2).Return(uint64(3))
	mockAppModule2.EXPECT().ConsensusVersion().Times(2).Return(uint64(2))
	_, err := module.NewManager([]module.VersionedModule{
		{Module: mockAppModule1, FromVersion: 1, ToVersion: 1},
		{Module: mockAppModule2, FromVersion: 2, ToVersion: 2},
	})
	require.Error(t, err)
}

func TestManager_ModuleNames(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAppModule1 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
	mockAppModule2 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)

	mockAppModule1.EXPECT().Name().Times(2).Return("module1")
	mockAppModule1.EXPECT().ConsensusVersion().Return(uint64(1))

	mockAppModule2.EXPECT().Name().Times(2).Return("module2")
	mockAppModule2.EXPECT().ConsensusVersion().Return(uint64(1))

	mm, err := module.NewManager([]module.VersionedModule{
		{Module: mockAppModule1, FromVersion: 1, ToVersion: 1},
		{Module: mockAppModule2, FromVersion: 1, ToVersion: 1},
	})
	require.NoError(t, err)

	got := mm.ModuleNames(1)
	want := []string{"module1", "module2"}
	assert.ElementsMatch(t, want, got)
}

func TestManager_SupportedVersions(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAppModule1 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
	mockAppModule2 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)

	mockAppModule1.EXPECT().Name().Times(2).Return("module1")
	mockAppModule1.EXPECT().ConsensusVersion().Times(2).Return(uint64(10))

	mockAppModule2.EXPECT().Name().Times(3).Return("module2")
	mockAppModule2.EXPECT().ConsensusVersion().Times(3).Return(uint64(10))

	mm, err := module.NewManager([]module.VersionedModule{
		{Module: mockAppModule1, FromVersion: 1, ToVersion: 1},
		{Module: mockAppModule2, FromVersion: 3, ToVersion: 4},
	})
	require.NoError(t, err)

	got := mm.SupportedVersions()
	assert.Equal(t, []uint64{1, 3, 4}, got)
}
