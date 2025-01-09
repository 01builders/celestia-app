package app

import (
	"fmt"

	"cosmossdk.io/core/comet"
	"github.com/cosmos/cosmos-sdk/client"

	"cosmossdk.io/x/authz"
	authzkeeper "cosmossdk.io/x/authz/keeper"
	authzmodule "cosmossdk.io/x/authz/module"
	"cosmossdk.io/x/bank"
	banktypes "cosmossdk.io/x/bank/types"
	distr "cosmossdk.io/x/distribution"
	distrtypes "cosmossdk.io/x/distribution/types"
	"cosmossdk.io/x/evidence"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	feegrantmodule "cosmossdk.io/x/feegrant/module"
	"cosmossdk.io/x/gov"
	govtypes "cosmossdk.io/x/gov/types"
	"cosmossdk.io/x/params"
	paramstypes "cosmossdk.io/x/params/types"
	"cosmossdk.io/x/slashing"
	slashingtypes "cosmossdk.io/x/slashing/types"
	"cosmossdk.io/x/staking"
	stakingtypes "cosmossdk.io/x/staking/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/celestiaorg/celestia-app/v3/app/encoding"
	"github.com/celestiaorg/celestia-app/v3/app/module"
	"github.com/celestiaorg/celestia-app/v3/x/blob"
	blobtypes "github.com/celestiaorg/celestia-app/v3/x/blob/types"
	"github.com/celestiaorg/celestia-app/v3/x/blobstream"
	blobstreamtypes "github.com/celestiaorg/celestia-app/v3/x/blobstream/types"
	"github.com/celestiaorg/celestia-app/v3/x/minfee"
	"github.com/celestiaorg/celestia-app/v3/x/mint"
	minttypes "github.com/celestiaorg/celestia-app/v3/x/mint/types"
	"github.com/celestiaorg/celestia-app/v3/x/signal"
	signaltypes "github.com/celestiaorg/celestia-app/v3/x/signal/types"
	sdkmodule "github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	// "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v9/packetforward"
	// packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v9/packetforward/types"
	ica "github.com/cosmos/ibc-go/v9/modules/apps/27-interchain-accounts"
	icahosttypes "github.com/cosmos/ibc-go/v9/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v9/modules/apps/27-interchain-accounts/types"
	"github.com/cosmos/ibc-go/v9/modules/apps/transfer"
	ibctransfertypes "github.com/cosmos/ibc-go/v9/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v9/modules/core"
	ibchost "github.com/cosmos/ibc-go/v9/modules/core/24-host"
	ibcexported "github.com/cosmos/ibc-go/v9/modules/core/exported"
)

var (
	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics = sdkmodule.NewManager(
		auth.AppModule{},
		genutil.AppModule{},
		bankModule{},
		stakingModule{},
		mintModule{},
		distributionModule{},
		newGovModule(),
		params.AppModule{},
		slashingModule{},
		authzmodule.AppModule{},
		feegrantmodule.AppModule{},
		ibcModule{},
		evidence.AppModule{},
		transfer.AppModule{},
		vesting.AppModule{},
		blob.AppModule{},
		blobstream.AppModuleBasic{},
		signal.AppModuleBasic{},
		minfee.AppModuleBasic{},
		// packetforward.AppModuleBasic{},
		icaModule{},
	)

	// ModuleEncodingRegisters keeps track of all the module methods needed to
	// register interfaces and specific type to encoding config
	ModuleEncodingRegisters = extractRegisters(ModuleBasics)
)

func (app *App) setupModuleManager(
	txConfig client.TxConfig,
	cometService comet.Service,
	skipGenesisInvariants bool,
) error {
	var err error
	app.manager, err = module.NewManager([]module.VersionedModule{
		{
			Module:      genutil.NewAppModule(app.appCodec, app.AuthKeeper, app.StakingKeeper, app, txConfig, genutiltypes.DefaultMessageValidator),
			FromVersion: v1, ToVersion: v3,
		},
		{
			Module:      auth.NewAppModule(app.appCodec, app.AuthKeeper, app.AccountsKeeper, nil, nil),
			FromVersion: v1, ToVersion: v3,
		},
		{
			Module:      vesting.NewAppModule(app.AuthKeeper, app.BankKeeper),
			FromVersion: v1, ToVersion: v3,
		},
		{
			Module:      bank.NewAppModule(app.appCodec, app.BankKeeper, app.AuthKeeper),
			FromVersion: v1, ToVersion: v3,
		},
		// {
		// 	Module:      capability.NewAppModule(app.appCodec, *app.CapabilityKeeper),
		// 	FromVersion: v1, ToVersion: v3,
		// },
		{
			Module:      feegrantmodule.NewAppModule(app.appCodec, app.FeeGrantKeeper, app.interfaceRegistry),
			FromVersion: v1, ToVersion: v3,
		},
		// {
		// 	Module:      crisis.NewAppModule(&app.CrisisKeeper, skipGenesisInvariants),
		// 	FromVersion: v1, ToVersion: v3,
		// },
		{
			Module:      gov.NewAppModule(app.appCodec, app.GovKeeper, app.AuthKeeper, app.BankKeeper, app.PoolKeeper),
			FromVersion: v1, ToVersion: v3,
		},
		{
			Module:      mint.NewAppModule(app.appCodec, app.MintKeeper, app.AuthKeeper),
			FromVersion: v1, ToVersion: v3,
		},
		{
			Module: slashing.NewAppModule(app.appCodec, app.SlashingKeeper, app.AuthKeeper,
				app.BankKeeper, app.StakingKeeper, app.interfaceRegistry, cometService),
			FromVersion: v1, ToVersion: v3,
		},
		{
			Module:      distr.NewAppModule(app.appCodec, app.DistrKeeper, app.StakingKeeper),
			FromVersion: v1, ToVersion: v3,
		},
		{
			Module:      staking.NewAppModule(app.appCodec, app.StakingKeeper),
			FromVersion: v1, ToVersion: v3,
		},
		{
			Module:      evidence.NewAppModule(app.appCodec, app.EvidenceKeeper, cometService),
			FromVersion: v1, ToVersion: v3,
		},
		{
			Module:      authzmodule.NewAppModule(app.appCodec, app.AuthzKeeper, app.interfaceRegistry),
			FromVersion: v1, ToVersion: v3,
		},
		{
			Module:      ibc.NewAppModule(app.appCodec, app.IBCKeeper),
			FromVersion: v1, ToVersion: v3,
		},
		{
			Module:      params.NewAppModule(app.ParamsKeeper),
			FromVersion: v1, ToVersion: v3,
		},
		{
			Module:      transfer.NewAppModule(app.appCodec, app.TransferKeeper),
			FromVersion: v1, ToVersion: v3,
		},
		{
			Module:      blob.NewAppModule(app.appCodec, app.BlobKeeper),
			FromVersion: v1, ToVersion: v3,
		},
		{
			Module:      blobstream.NewAppModule(app.appCodec, app.BlobstreamKeeper),
			FromVersion: v1, ToVersion: v1,
		},
		{
			Module:      signal.NewAppModule(app.SignalKeeper),
			FromVersion: v2, ToVersion: v3,
		},
		{
			Module:      minfee.NewAppModule(app.ParamsKeeper),
			FromVersion: v2, ToVersion: v3,
		},
		// {
		// 	Module:      packetforward.NewAppModule(app.PacketForwardKeeper),
		// 	FromVersion: v2, ToVersion: v3,
		// },
		{
			Module:      ica.NewAppModule(nil, &app.ICAHostKeeper),
			FromVersion: v2, ToVersion: v3,
		},
	})
	return err
}

func (app *App) setModuleOrder() {
	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	app.manager.SetOrderBeginBlockers(
		minttypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		ibcexported.ModuleName,
		ibctransfertypes.ModuleName,
		feegrant.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		govtypes.ModuleName,
		genutiltypes.ModuleName,
		blobtypes.ModuleName,
		blobstreamtypes.ModuleName,
		paramstypes.ModuleName,
		authz.ModuleName,
		vestingtypes.ModuleName,
		signaltypes.ModuleName,
		minfee.ModuleName,
		icatypes.ModuleName,
		// packetforwardtypes.ModuleName,
	)

	app.manager.SetOrderEndBlockers(
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		minttypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		ibcexported.ModuleName,
		ibctransfertypes.ModuleName,
		feegrant.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		genutiltypes.ModuleName,
		blobtypes.ModuleName,
		blobstreamtypes.ModuleName,
		paramstypes.ModuleName,
		authz.ModuleName,
		vestingtypes.ModuleName,
		signaltypes.ModuleName,
		minfee.ModuleName,
		// packetforwardtypes.ModuleName,
		icatypes.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: The minfee module must occur before genutil so DeliverTx can
	// successfully pass the fee checking logic
	app.manager.SetOrderInitGenesis(
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		ibcexported.ModuleName,
		minfee.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		ibctransfertypes.ModuleName,
		blobtypes.ModuleName,
		blobstreamtypes.ModuleName,
		vestingtypes.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		authz.ModuleName,
		signaltypes.ModuleName,
		// packetforwardtypes.ModuleName,
		icatypes.ModuleName,
	)
}

func allStoreKeys() []string {
	return []string{
		authtypes.StoreKey, authzkeeper.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, paramstypes.StoreKey, upgradetypes.StoreKey, feegrant.StoreKey,
		evidencetypes.StoreKey,
		blobstreamtypes.StoreKey,
		ibctransfertypes.StoreKey,
		ibcexported.StoreKey,
		// packetforwardtypes.StoreKey,
		icahosttypes.StoreKey,
		signaltypes.StoreKey,
		blobtypes.StoreKey,
	}
}

// versionedStoreKeys returns the store keys for each app version.
func versionedStoreKeys() map[uint64][]string {
	return map[uint64][]string{
		v1: {
			authtypes.StoreKey,
			authzkeeper.StoreKey,
			banktypes.StoreKey,
			blobstreamtypes.StoreKey,
			blobtypes.StoreKey,
			distrtypes.StoreKey,
			evidencetypes.StoreKey,
			feegrant.StoreKey,
			govtypes.StoreKey,
			ibchost.StoreKey,
			ibctransfertypes.StoreKey,
			minttypes.StoreKey,
			slashingtypes.StoreKey,
			stakingtypes.StoreKey,
			upgradetypes.StoreKey,
		},
		v2: {
			authtypes.StoreKey,
			authzkeeper.StoreKey,
			banktypes.StoreKey,
			blobtypes.StoreKey,
			distrtypes.StoreKey,
			evidencetypes.StoreKey,
			feegrant.StoreKey,
			govtypes.StoreKey,
			ibchost.StoreKey,
			ibctransfertypes.StoreKey,
			icahosttypes.StoreKey, // added in v2
			minttypes.StoreKey,
			// packetforwardtypes.StoreKey, // added in v2
			signaltypes.StoreKey, // added in v2
			slashingtypes.StoreKey,
			stakingtypes.StoreKey,
			upgradetypes.StoreKey,
		},
		v3: { // same as v2
			authtypes.StoreKey,
			authzkeeper.StoreKey,
			banktypes.StoreKey,
			blobtypes.StoreKey,
			distrtypes.StoreKey,
			evidencetypes.StoreKey,
			feegrant.StoreKey,
			govtypes.StoreKey,
			ibchost.StoreKey,
			ibctransfertypes.StoreKey,
			icahosttypes.StoreKey,
			minttypes.StoreKey,
			// packetforwardtypes.StoreKey,
			signaltypes.StoreKey,
			slashingtypes.StoreKey,
			stakingtypes.StoreKey,
			upgradetypes.StoreKey,
		},
	}
}

// assertAllKeysArePresent performs a couple sanity checks on startup to ensure each versions key names have
// a key and that all versions supported by the module manager have a respective versioned key
func (app *App) assertAllKeysArePresent() {
	supportedAppVersions := app.SupportedVersions()
	supportedVersionsMap := make(map[uint64]bool, len(supportedAppVersions))
	for _, version := range supportedAppVersions {
		supportedVersionsMap[version] = false
	}

	for appVersion, keys := range app.keyVersions {
		if _, exists := supportedVersionsMap[appVersion]; exists {
			supportedVersionsMap[appVersion] = true
		} else {
			panic(fmt.Sprintf("keys %v for app version %d are not supported by the module manager", keys, appVersion))
		}
		for _, key := range keys {
			if _, ok := app.keys[key]; !ok {
				panic(fmt.Sprintf("key %s is not present", key))
			}
		}
	}
	for appVersion, supported := range supportedVersionsMap {
		if !supported {
			panic(fmt.Sprintf("app version %d is supported by the module manager but has no keys", appVersion))
		}
	}
}

// extractRegisters returns the encoding module registers from the basic
// manager.
func extractRegisters(manager sdkmodule.BasicManager) (modules []encoding.ModuleRegister) {
	for _, module := range manager {
		modules = append(modules, module)
	}
	return modules
}
