package module

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"google.golang.org/grpc"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/genesis"
	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	sdkmodule "github.com/cosmos/cosmos-sdk/types/module"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type hasServices interface {
	RegisterServices(grpc.ServiceRegistrar) error
}

// Manager defines a module manager that provides the high level utility for
// managing and executing operations for a group of modules. This implementation
// was originally inspired by the module manager defined in Cosmos SDK but this
// implementation maps the state machine version to different versions of the
// module. It also provides a way to run migrations between different versions
// of a module.
type Manager struct {
	// versionedModules is a map from app version -> module name -> module.
	versionedModules map[uint64]map[string]sdkmodule.AppModule
	// uniqueModuleVersions is a mapping of module name -> module consensus
	// version -> the range of app versions this particular module operates
	// over. The first element in the array represent the fromVersion and the
	// last the toVersion (this is inclusive).
	uniqueModuleVersions map[string]map[uint64][2]uint64
	allModules           []sdkmodule.AppModule
	// firstVersion is the lowest app version supported.
	firstVersion uint64
	// lastVersion is the highest app version supported.
	lastVersion        uint64
	OrderInitGenesis   []string
	OrderExportGenesis []string
	OrderBeginBlockers []string
	OrderEndBlockers   []string
	OrderMigrations    []string
}

// NewManager returns a new Manager object.
func NewManager(modules []VersionedModule) (*Manager, error) {
	versionedModules := make(map[uint64]map[string]sdkmodule.AppModule)
	allModules := make([]sdkmodule.AppModule, len(modules))
	modulesStr := make([]string, 0, len(modules))
	uniqueModuleVersions := make(map[string]map[uint64][2]uint64)
	for idx, module := range modules {
		name := module.Module.Name()
		moduleVersion, err := moduleConsensusVersion(module.Module)
		if err != nil {
			return nil, err
		}
		if module.FromVersion == 0 {
			return nil, sdkerrors.ErrInvalidVersion.Wrapf("v0 is not a valid version for module %s", module.Module.Name())
		}
		if module.FromVersion > module.ToVersion {
			return nil, sdkerrors.ErrLogic.Wrapf("FromVersion cannot be greater than ToVersion for module %s", module.Module.Name())
		}
		for version := module.FromVersion; version <= module.ToVersion; version++ {
			if versionedModules[version] == nil {
				versionedModules[version] = make(map[string]sdkmodule.AppModule)
			}
			if _, exists := versionedModules[version][name]; exists {
				return nil, sdkerrors.ErrLogic.Wrapf("Two different modules with domain %s are registered with the same version %d", name, version)
			}
			versionedModules[version][module.Module.Name()] = module.Module
		}
		allModules[idx] = module.Module
		modulesStr = append(modulesStr, name)
		if _, exists := uniqueModuleVersions[name]; !exists {
			uniqueModuleVersions[name] = make(map[uint64][2]uint64)
		}
		uniqueModuleVersions[name][moduleVersion] = [2]uint64{module.FromVersion, module.ToVersion}
	}
	firstVersion := slices.Min(getKeys(versionedModules))
	lastVersion := slices.Max(getKeys(versionedModules))

	m := &Manager{
		versionedModules:     versionedModules,
		uniqueModuleVersions: uniqueModuleVersions,
		allModules:           allModules,
		firstVersion:         firstVersion,
		lastVersion:          lastVersion,
		OrderInitGenesis:     modulesStr,
		OrderExportGenesis:   modulesStr,
		OrderBeginBlockers:   modulesStr,
		OrderEndBlockers:     modulesStr,
	}
	if err := m.checkUpgradeSchedule(); err != nil {
		return nil, err
	}
	return m, nil
}

// SetOrderInitGenesis sets the order of init genesis calls.
func (m *Manager) SetOrderInitGenesis(moduleNames ...string) {
	m.assertNoForgottenModules("SetOrderInitGenesis", moduleNames)
	m.OrderInitGenesis = moduleNames
}

// SetOrderExportGenesis sets the order of export genesis calls.
func (m *Manager) SetOrderExportGenesis(moduleNames ...string) {
	m.assertNoForgottenModules("SetOrderExportGenesis", moduleNames)
	m.OrderExportGenesis = moduleNames
}

// SetOrderBeginBlockers sets the order of begin-blocker calls.
func (m *Manager) SetOrderBeginBlockers(moduleNames ...string) {
	m.assertNoForgottenModules("SetOrderBeginBlockers", moduleNames)
	m.OrderBeginBlockers = moduleNames
}

// SetOrderEndBlockers sets the order of end-blocker calls.
func (m *Manager) SetOrderEndBlockers(moduleNames ...string) {
	m.assertNoForgottenModules("SetOrderEndBlockers", moduleNames)
	m.OrderEndBlockers = moduleNames
}

// SetOrderMigrations sets the order of migrations to be run. If not set
// then migrations will be run with an order defined in `defaultMigrationsOrder`.
func (m *Manager) SetOrderMigrations(moduleNames ...string) {
	m.assertNoForgottenModules("SetOrderMigrations", moduleNames)
	m.OrderMigrations = moduleNames
}

// RegisterServices registers all module services.
func (m *Manager) RegisterServices(cfg Configurator) error {
	for _, module := range m.allModules {
		moduleVersion, err := moduleConsensusVersion(module)
		if err != nil {
			return err
		}
		fromVersion, toVersion := m.getAppVersionsForModule(module.Name(), moduleVersion)
		c := cfg.WithVersions(fromVersion, toVersion)

		if module, ok := module.(hasServices); ok {
			err := module.RegisterServices(c)
			if err != nil {
				return err
			}
		}

		if module, ok := module.(appmodule.HasMigrations); ok {
			err := module.RegisterMigrations(c)
			if err != nil {
				return err
			}
		}

		if cfg.Error() != nil {
			return cfg.Error()
		}
	}
	return nil
}

func (m *Manager) getAppVersionsForModule(moduleName string, moduleVersion uint64) (uint64, uint64) {
	return m.uniqueModuleVersions[moduleName][moduleVersion][0], m.uniqueModuleVersions[moduleName][moduleVersion][1]
}

// InitGenesis performs init genesis functionality for modules. Exactly one
// module must return a non-empty validator set update to correctly initialize
// the chain.
func (m *Manager) InitGenesis(ctx sdk.Context, genesisData map[string]json.RawMessage, appVersion uint64) (*abci.InitChainResponse, error) {
	var validatorUpdates []sdkmodule.ValidatorUpdate
	ctx.Logger().Info("initializing blockchain state from genesis.json")
	modules, versionSupported := m.versionedModules[appVersion]
	if !versionSupported {
		panic(fmt.Sprintf("version %d not supported", appVersion))
	}
	for _, moduleName := range m.OrderInitGenesis {
		if genesisData[moduleName] == nil {
			continue
		}

		mod := modules[moduleName]
		// we might get an adapted module, a native core API module or a legacy module
		if module, ok := mod.(appmodule.HasGenesisAuto); ok {
			ctx.Logger().Debug("running initialization for module", "module", moduleName)
			// core API genesis
			source, err := genesis.SourceFromRawJSON(genesisData[moduleName])
			if err != nil {
				return &abci.InitChainResponse{}, err
			}

			err = module.InitGenesis(ctx, source)
			if err != nil {
				return &abci.InitChainResponse{}, err
			}
		} else if module, ok := mod.(appmodule.HasGenesis); ok {
			ctx.Logger().Debug("running initialization for module", "module", moduleName)
			if err := module.InitGenesis(ctx, genesisData[moduleName]); err != nil {
				return &abci.InitChainResponse{}, err
			}
		} else if module, ok := mod.(appmodule.HasABCIGenesis); ok {
			ctx.Logger().Debug("running initialization for module", "module", moduleName)
			moduleValUpdates, err := module.InitGenesis(ctx, genesisData[moduleName])
			if err != nil {
				return &abci.InitChainResponse{}, err
			}

			// use these validator updates if provided, the module manager assumes
			// only one module will update the validator set
			if len(moduleValUpdates) > 0 {
				if len(validatorUpdates) > 0 {
					return &abci.InitChainResponse{}, errors.New("validator InitGenesis updates already set by a previous module")
				}
				validatorUpdates = moduleValUpdates
			}
		}
	}

	// a chain must initialize with a non-empty validator set
	if len(validatorUpdates) == 0 {
		panic(fmt.Sprintf("validator set is empty after InitGenesis, please ensure at least one validator is initialized with a delegation greater than or equal to the DefaultPowerReduction (%d)", sdk.DefaultPowerReduction))
	}

	cometValidatorUpdates := make([]abci.ValidatorUpdate, len(validatorUpdates))
	for i, v := range validatorUpdates {
		cometValidatorUpdates[i] = abci.ValidatorUpdate{
			PubKeyBytes: v.PubKey,
			Power:       v.Power,
			PubKeyType:  v.PubKeyType,
		}
	}

	return &abci.InitChainResponse{
		Validators: cometValidatorUpdates,
	}, nil
}

// ExportGenesis performs export genesis functionality for the modules supported
// in a particular version.
func (m *Manager) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec, version uint64) (map[string]json.RawMessage, error) {
	genesisData := make(map[string]json.RawMessage)
	modules := m.versionedModules[version]
	moduleNamesForVersion := m.ModuleNames(version)
	moduleNamesToExport := filter(m.OrderExportGenesis, func(moduleName string) bool {
		// filter out modules that are not supported by this version
		return slices.Contains(moduleNamesForVersion, moduleName)
	})
	for _, moduleName := range moduleNamesToExport {
		mod := modules[moduleName]
		if module, ok := mod.(appmodule.HasGenesisAuto); ok {
			target := genesis.RawJSONTarget{}
			err := module.ExportGenesis(ctx, target.Target())
			if err != nil {
				return nil, err
			}
			rawJSON, err := target.JSON()
			if err != nil {
				return nil, err
			}
			genesisData[moduleName] = rawJSON
		} else if module, ok := mod.(appmodule.HasGenesis); ok {
			jm, err := module.ExportGenesis(ctx)
			if err != nil {
				return nil, err
			}
			genesisData[moduleName] = jm
		} else if module, ok := mod.(appmodule.HasABCIGenesis); ok {
			jm, err := module.ExportGenesis(ctx)
			if err != nil {
				return nil, err
			}
			genesisData[moduleName] = jm
		}
	}

	return genesisData, nil
}

// assertNoForgottenModules checks that we didn't forget any modules in the
// SetOrder* functions.
func (m *Manager) assertNoForgottenModules(setOrderFnName string, moduleNames []string) {
	ms := make(map[string]bool)
	for _, m := range moduleNames {
		ms[m] = true
	}
	var missing []string
	for _, m := range m.allModules {
		if _, ok := ms[m.Name()]; !ok {
			missing = append(missing, m.Name())
		}
	}
	if len(missing) != 0 {
		panic(fmt.Sprintf(
			"%s: all modules must be defined when setting %s, missing: %v", setOrderFnName, setOrderFnName, missing))
	}
}

// RunMigrations performs in-place store migrations for all modules. This
// function MUST be called when the state machine changes appVersion
func (m Manager) RunMigrations(ctx sdk.Context, cfg Configurator, fromVersion, toVersion uint64) error {
	modules := m.OrderMigrations
	if modules == nil {
		modules = defaultMigrationsOrder(m.ModuleNames(toVersion))
	}
	currentVersionModules, exists := m.versionedModules[fromVersion]
	if !exists {
		return sdkerrors.ErrInvalidVersion.Wrapf("fromVersion %d not supported", fromVersion)
	}
	nextVersionModules, exists := m.versionedModules[toVersion]
	if !exists {
		return sdkerrors.ErrInvalidVersion.Wrapf("toVersion %d not supported", toVersion)
	}

	for _, moduleName := range modules {
		currentModule, currentModuleExists := currentVersionModules[moduleName]
		nextModule, nextModuleExists := nextVersionModules[moduleName]

		// if the module exists for both upgrades
		if currentModuleExists && nextModuleExists {
			// by using consensus version instead of app version we support the SDK's legacy method
			// of migrating modules which were made of several versions and consisted of a mapping of
			// app version to module version. Now, using go.mod, each module will have only a single
			// consensus version and each breaking upgrade will result in a new module and a new consensus
			// version.
			fromModuleVersion, err := moduleConsensusVersion(currentModule)
			if err != nil {
				return err
			}
			toModuleVersion, err := moduleConsensusVersion(nextModule)
			if err != nil {
				return err
			}
			err = cfg.runModuleMigrations(ctx, moduleName, fromModuleVersion, toModuleVersion)
			if err != nil {
				return err
			}
		} else if !currentModuleExists && nextModuleExists {
			ctx.Logger().Info(fmt.Sprintf("adding a new module: %s", moduleName))
			if mod, ok := nextModule.(appmodule.HasGenesis); ok {
				if err := mod.InitGenesis(ctx, mod.DefaultGenesis()); err != nil {
					return err
				}
			}
			if mod, ok := nextModule.(appmodule.HasABCIGenesis); ok {
				moduleValUpdates, err := mod.InitGenesis(ctx, mod.DefaultGenesis())
				if err != nil {
					return err
				}
				// The module manager assumes only one module will update the
				// validator set, and it can't be a new module.
				if len(moduleValUpdates) > 0 {
					return sdkerrors.ErrLogic.Wrap("validator InitGenesis update is already set by another module")
				}
			}
		}
		// TODO: handle the case where a module is no longer supported (i.e. removed from the state machine)
	}

	return nil
}

// BeginBlock performs begin block functionality for all modules. It creates a
// child context with an event manager to aggregate events emitted from all
// modules.
func (m *Manager) BeginBlock(ctx sdk.Context) (sdk.BeginBlock, error) {
	ctx = ctx.WithEventManager(sdk.NewEventManager())

	modules := m.versionedModules[ctx.BlockHeader().Version.App] // TODO: get from consensus params
	if modules == nil {
		panic(fmt.Sprintf("no modules for version %d", ctx.BlockHeader().Version.App)) // TODO: get from consensus params
	}
	for _, moduleName := range m.OrderBeginBlockers {
		module, ok := modules[moduleName].(appmodule.HasBeginBlocker)
		if ok {
			if err := module.BeginBlock(ctx); err != nil {

			}
		}
	}

	return sdk.BeginBlock{Events: ctx.EventManager().ABCIEvents()}, nil
}

// EndBlock performs end block functionality for all modules. It creates a
// child context with an event manager to aggregate events emitted from all
// modules.
func (m *Manager) EndBlock(ctx sdk.Context) (sdk.EndBlock, error) {
	ctx = ctx.WithEventManager(sdk.NewEventManager())
	var validatorUpdates []sdkmodule.ValidatorUpdate

	modules := m.versionedModules[ctx.BlockHeader().Version.App] // TODO: get from consensus params
	if modules == nil {
		panic(fmt.Sprintf("no modules for version %d", ctx.BlockHeader().Version.App)) // TODO: get from consensus params
	}
	for _, moduleName := range m.OrderEndBlockers {
		if module, ok := modules[moduleName].(appmodule.HasEndBlocker); ok {
			err := module.EndBlock(ctx)
			if err != nil {
				return sdk.EndBlock{}, err
			}
		} else if module, ok := modules[moduleName].(sdkmodule.HasABCIEndBlock); ok {
			moduleValUpdates, err := module.EndBlock(ctx)
			if err != nil {
				return sdk.EndBlock{}, err
			}
			// use these validator updates if provided, the module manager assumes
			// only one module will update the validator set
			if len(moduleValUpdates) > 0 {
				if len(validatorUpdates) > 0 {
					return sdk.EndBlock{}, errors.New("validator EndBlock updates already set by a previous module")
				}

				validatorUpdates = append(validatorUpdates, moduleValUpdates...)
			}
		}
	}

	cometValidatorUpdates := make([]abci.ValidatorUpdate, len(validatorUpdates))
	for i, v := range validatorUpdates {
		cometValidatorUpdates[i] = abci.ValidatorUpdate{
			PubKeyBytes: v.PubKey,
			PubKeyType:  v.PubKeyType,
			Power:       v.Power,
		}
	}

	return sdk.EndBlock{
		ValidatorUpdates: cometValidatorUpdates,
		Events:           ctx.EventManager().ABCIEvents(),
	}, nil
}

// GetVersionMap gets consensus version from all modules
func (m *Manager) GetVersionMap(version uint64) (appmodule.VersionMap, error) {
	vermap := make(appmodule.VersionMap)
	if version > m.lastVersion || version < m.firstVersion {
		return vermap, nil
	}

	for _, v := range m.versionedModules[version] {
		version, err := moduleConsensusVersion(v)
		if err != nil {
			return nil, err
		}
		name := v.Name()
		vermap[name] = version
	}

	return vermap, nil
}

// ModuleNames returns the list of module names that are supported for a
// particular version in no particular order.
func (m *Manager) ModuleNames(version uint64) []string {
	modules, ok := m.versionedModules[version]
	if !ok {
		return []string{}
	}

	names := make([]string, 0, len(modules))
	for name := range modules {
		names = append(names, name)
	}
	return names
}

// SupportedVersions returns all the supported versions for the module manager
func (m *Manager) SupportedVersions() []uint64 {
	return getKeys(m.versionedModules)
}

// checkUpgradeSchedule performs a dry run of all the upgrades in all versions and asserts that the consensus version
// for a module domain i.e. auth, always increments for each module that uses the auth domain name
func (m *Manager) checkUpgradeSchedule() error {
	if m.firstVersion == m.lastVersion {
		// there are no upgrades to check
		return nil
	}
	for _, moduleName := range m.OrderInitGenesis {
		lastConsensusVersion := uint64(0)
		for appVersion := m.firstVersion; appVersion <= m.lastVersion; appVersion++ {
			module, exists := m.versionedModules[appVersion][moduleName]
			if !exists {
				continue
			}
			moduleVersion, err := moduleConsensusVersion(module)
			if err != nil {
				return err
			}
			if moduleVersion < lastConsensusVersion {
				return fmt.Errorf("error: module %s in appVersion %d goes from moduleVersion %d to %d", moduleName, appVersion, lastConsensusVersion, moduleVersion)
			}
			lastConsensusVersion = moduleVersion
		}
	}
	return nil
}

// assertMatchingModules performs a sanity check that the SDK module manager
// contains all the same modules present in the versioned module manager
func (m *Manager) AssertMatchingModules(mm sdkmodule.Manager) error {
	for _, module := range m.allModules {
		if _, exists := mm.Modules[module.Name()]; !exists {
			return fmt.Errorf("module %s not found in basic module manager", module.Name())
		}
	}
	return nil
}

func moduleConsensusVersion(mod sdkmodule.AppModule) (uint64, error) {
	cv, ok := mod.(appmodule.HasConsensusVersion)
	if !ok {
		return 0, fmt.Errorf("module %s does not implement HasConsensusVersion", mod.Name())
	}
	return cv.ConsensusVersion(), nil
}
