package app

import (
	"fmt"

	"cosmossdk.io/x/accounts"
	consensustypes "cosmossdk.io/x/consensus/types"
	pooltypes "cosmossdk.io/x/protocolpool/types"
	ibcfeetypes "github.com/cosmos/ibc-go/v9/modules/apps/29-fee/types"

	"cosmossdk.io/core/comet"

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
	"github.com/celestiaorg/celestia-app/v4/app/encoding"
	"github.com/celestiaorg/celestia-app/v4/app/module"
	"github.com/celestiaorg/celestia-app/v4/x/blob"
	blobtypes "github.com/celestiaorg/celestia-app/v4/x/blob/types"
	"github.com/celestiaorg/celestia-app/v4/x/blobstream"
	blobstreamtypes "github.com/celestiaorg/celestia-app/v4/x/blobstream/types"
	"github.com/celestiaorg/celestia-app/v4/x/minfee"
	"github.com/celestiaorg/celestia-app/v4/x/mint"
	minttypes "github.com/celestiaorg/celestia-app/v4/x/mint/types"
	"github.com/celestiaorg/celestia-app/v4/x/signal"
	signaltypes "github.com/celestiaorg/celestia-app/v4/x/signal/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	// "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v9/packetforward"
	// packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v9/packetforward/types"
	ica "github.com/cosmos/ibc-go/v9/modules/apps/27-interchain-accounts"
	icacontrollertypes "github.com/cosmos/ibc-go/v9/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v9/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v9/modules/apps/27-interchain-accounts/types"
	"github.com/cosmos/ibc-go/v9/modules/apps/transfer"
	ibctransfertypes "github.com/cosmos/ibc-go/v9/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v9/modules/core"
	ibcexported "github.com/cosmos/ibc-go/v9/modules/core/exported"
)

func (app *App) setupModuleManager(
	encodingConfig encoding.Config,
	cometService comet.Service,
) error {
	var err error
	app.ModuleManager, err = module.NewManager([]module.VersionedModule{
		{
			Module:      genutil.NewAppModule(encodingConfig.Codec, app.AuthKeeper, app.StakingKeeper, app, encodingConfig.TxConfig, genutiltypes.DefaultMessageValidator),
			FromVersion: v1, ToVersion: v4,
		},
		{
			Module:      auth.NewAppModule(encodingConfig.Codec, app.AuthKeeper, app.AccountsKeeper, nil, nil),
			FromVersion: v1, ToVersion: v4,
		},
		{
			Module:      vesting.NewAppModule(app.AuthKeeper, app.BankKeeper),
			FromVersion: v1, ToVersion: v4,
		},
		{
			Module:      bank.NewAppModule(encodingConfig.Codec, app.BankKeeper, app.AuthKeeper),
			FromVersion: v1, ToVersion: v4,
		},
		{
			Module:      feegrantmodule.NewAppModule(encodingConfig.Codec, app.FeeGrantKeeper, encodingConfig.InterfaceRegistry),
			FromVersion: v1, ToVersion: v4,
		},
		{
			Module:      gov.NewAppModule(encodingConfig.Codec, app.GovKeeper, app.AuthKeeper, app.BankKeeper, app.PoolKeeper),
			FromVersion: v1, ToVersion: v4,
		},
		{
			Module:      mint.NewAppModule(encodingConfig.Codec, app.MintKeeper, app.AuthKeeper),
			FromVersion: v1, ToVersion: v4,
		},
		{
			Module: slashing.NewAppModule(encodingConfig.Codec, app.SlashingKeeper, app.AuthKeeper,
				app.BankKeeper, app.StakingKeeper, encodingConfig.InterfaceRegistry, cometService),
			FromVersion: v1, ToVersion: v4,
		},
		{
			Module:      distr.NewAppModule(encodingConfig.Codec, app.DistrKeeper, app.StakingKeeper),
			FromVersion: v1, ToVersion: v4,
		},
		{
			Module:      staking.NewAppModule(encodingConfig.Codec, app.StakingKeeper),
			FromVersion: v1, ToVersion: v4,
		},
		{
			Module:      evidence.NewAppModule(encodingConfig.Codec, app.EvidenceKeeper, cometService),
			FromVersion: v1, ToVersion: v4,
		},
		{
			Module:      authzmodule.NewAppModule(encodingConfig.Codec, app.AuthzKeeper, encodingConfig.InterfaceRegistry),
			FromVersion: v1, ToVersion: v4,
		},
		{
			Module:      ibc.NewAppModule(encodingConfig.Codec, app.IBCKeeper),
			FromVersion: v1, ToVersion: v4,
		},
		{
			Module:      params.NewAppModule(app.ParamsKeeper),
			FromVersion: v1, ToVersion: v4,
		},
		{
			Module:      transfer.NewAppModule(encodingConfig.Codec, app.TransferKeeper),
			FromVersion: v1, ToVersion: v4,
		},
		{
			Module:      blob.NewAppModule(encodingConfig.Codec, app.BlobKeeper),
			FromVersion: v1, ToVersion: v4,
		},
		{
			Module:      blobstream.NewAppModule(encodingConfig.Codec, app.BlobstreamKeeper),
			FromVersion: v1, ToVersion: v1,
		},
		{
			Module:      signal.NewAppModule(app.SignalKeeper),
			FromVersion: v2, ToVersion: v4,
		},
		{
			Module:      minfee.NewAppModule(encodingConfig.Codec, app.ParamsKeeper),
			FromVersion: v2, ToVersion: v4,
		},
		// {
		// 	Module:      packetforward.NewAppModule(app.PacketForwardKeeper),
		// 	FromVersion: v2, ToVersion: v3,
		// },
		{
			Module:      ica.NewAppModule(encodingConfig.Codec, &app.ICAControllerKeeper, &app.ICAHostKeeper),
			FromVersion: v2, ToVersion: v4,
		},
	})
	return err
}

func (app *App) setModuleOrder() {
	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	app.ModuleManager.SetOrderBeginBlockers(
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
		ibcfeetypes.ModuleName,
		// packetforwardtypes.ModuleName,
	)

	app.ModuleManager.SetOrderEndBlockers(
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		minttypes.ModuleName,
		distrtypes.ModuleName,
		pooltypes.ModuleName,
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
		ibcfeetypes.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: The minfee module must occur before genutil so DeliverTx can
	// successfully pass the fee checking logic
	app.ModuleManager.SetOrderInitGenesis(
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		pooltypes.ModuleName,
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
		ibcfeetypes.ModuleName,
	)
}

func allStoreKeys() []string {
	return versionedStoreKeys()[DefaultInitialVersion]
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
			ibcexported.StoreKey,
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
			ibcexported.StoreKey,
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
			ibcexported.StoreKey,
			ibctransfertypes.StoreKey,
			icahosttypes.StoreKey,
			minttypes.StoreKey,
			// packetforwardtypes.StoreKey,
			signaltypes.StoreKey,
			slashingtypes.StoreKey,
			stakingtypes.StoreKey,
			upgradetypes.StoreKey,
		},
		v4: {
			authtypes.StoreKey,
			authzkeeper.StoreKey,
			banktypes.StoreKey,
			stakingtypes.StoreKey,
			minttypes.StoreKey,
			distrtypes.StoreKey,
			slashingtypes.StoreKey,
			govtypes.StoreKey,
			paramstypes.StoreKey,
			upgradetypes.StoreKey,
			feegrant.StoreKey,
			pooltypes.StoreKey, // added in v4
			evidencetypes.StoreKey,
			blobstreamtypes.StoreKey,
			ibctransfertypes.StoreKey,
			ibcexported.StoreKey,
			// packetforwardtypes.StoreKey,
			icahosttypes.StoreKey,
			ibcfeetypes.StoreKey,
			signaltypes.StoreKey,
			blobtypes.StoreKey,
			consensustypes.StoreKey,     // added in v4
			accounts.StoreKey,           // added in v4
			icacontrollertypes.StoreKey, // added in v4
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
