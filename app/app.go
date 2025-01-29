package app

import (
	"context"
	"fmt"
	"io"
	"slices"
	"time"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	coreheader "cosmossdk.io/core/header"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/accounts"
	"cosmossdk.io/x/accounts/accountstd"
	baseaccount "cosmossdk.io/x/accounts/defaults/base"
	lockup "cosmossdk.io/x/accounts/defaults/lockup"
	"cosmossdk.io/x/accounts/defaults/multisig"
	authzkeeper "cosmossdk.io/x/authz/keeper"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktypes "cosmossdk.io/x/bank/types"
	"cosmossdk.io/x/consensus"
	consensuskeeper "cosmossdk.io/x/consensus/keeper"
	consensustypes "cosmossdk.io/x/consensus/types"
	distrkeeper "cosmossdk.io/x/distribution/keeper"
	distrtypes "cosmossdk.io/x/distribution/types"
	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	govkeeper "cosmossdk.io/x/gov/keeper"
	govtypes "cosmossdk.io/x/gov/types"
	govv1beta1 "cosmossdk.io/x/gov/types/v1beta1"
	"cosmossdk.io/x/params"
	paramskeeper "cosmossdk.io/x/params/keeper"
	paramstypes "cosmossdk.io/x/params/types"
	paramproposal "cosmossdk.io/x/params/types/proposal"
	poolkeeper "cosmossdk.io/x/protocolpool/keeper"
	pooltypes "cosmossdk.io/x/protocolpool/types"
	slashingkeeper "cosmossdk.io/x/slashing/keeper"
	slashingtypes "cosmossdk.io/x/slashing/types"
	stakingkeeper "cosmossdk.io/x/staking/keeper"
	stakingtypes "cosmossdk.io/x/staking/types"
	txdecode "cosmossdk.io/x/tx/decode"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/celestiaorg/celestia-app/v4/app/ante"
	"github.com/celestiaorg/celestia-app/v4/app/encoding"
	celestiatx "github.com/celestiaorg/celestia-app/v4/app/grpc/tx"
	"github.com/celestiaorg/celestia-app/v4/app/module"
	"github.com/celestiaorg/celestia-app/v4/app/posthandler"
	"github.com/celestiaorg/celestia-app/v4/pkg/appconsts"
	appv1 "github.com/celestiaorg/celestia-app/v4/pkg/appconsts/v1"
	appv2 "github.com/celestiaorg/celestia-app/v4/pkg/appconsts/v2"
	appv3 "github.com/celestiaorg/celestia-app/v4/pkg/appconsts/v3"
	celestiaserver "github.com/celestiaorg/celestia-app/v4/server"
	blobkeeper "github.com/celestiaorg/celestia-app/v4/x/blob/keeper"
	blobtypes "github.com/celestiaorg/celestia-app/v4/x/blob/types"
	blobstreamkeeper "github.com/celestiaorg/celestia-app/v4/x/blobstream/keeper"
	blobstreamtypes "github.com/celestiaorg/celestia-app/v4/x/blobstream/types"
	"github.com/celestiaorg/celestia-app/v4/x/minfee"
	mintkeeper "github.com/celestiaorg/celestia-app/v4/x/mint/keeper"
	minttypes "github.com/celestiaorg/celestia-app/v4/x/mint/types"
	"github.com/celestiaorg/celestia-app/v4/x/signal"
	signaltypes "github.com/celestiaorg/celestia-app/v4/x/signal/types"
	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	cmtcrypto "github.com/cometbft/cometbft/crypto"
	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/gogoproto/grpc"
	gogoproto "github.com/cosmos/gogoproto/proto"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v9/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v9/modules/apps/27-interchain-accounts/controller/types"
	ibcfeekeeper "github.com/cosmos/ibc-go/v9/modules/apps/29-fee/keeper"
	ibcfeetypes "github.com/cosmos/ibc-go/v9/modules/apps/29-fee/types"
	ibcexported "github.com/cosmos/ibc-go/v9/modules/core/exported"

	// "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v9/packetforward"
	// packetforwardkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v9/packetforward/keeper"
	// packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v9/packetforward/types"
	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	tmjson "github.com/cometbft/cometbft/libs/json"
	icahost "github.com/cosmos/ibc-go/v9/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v9/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v9/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v9/modules/apps/27-interchain-accounts/types"
	ibctransferkeeper "github.com/cosmos/ibc-go/v9/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v9/modules/apps/transfer/types"
	ibcporttypes "github.com/cosmos/ibc-go/v9/modules/core/05-port/types"
	ibckeeper "github.com/cosmos/ibc-go/v9/modules/core/keeper"
)

// maccPerms is short for module account permissions. It is a map from module
// account name to a list of permissions for that module account.
var maccPerms = map[string][]string{
	authtypes.FeeCollectorName:     nil,
	distrtypes.ModuleName:          nil,
	govtypes.ModuleName:            {authtypes.Burner},
	minttypes.ModuleName:           {authtypes.Minter},
	stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
	stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
	ibctransfertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
	icatypes.ModuleName:            nil,
}

const (
	v1                    = appv1.Version
	v2                    = appv2.Version
	v3                    = appv3.Version
	DefaultInitialVersion = v1
)

var (
	_ celestiaserver.Application = (*App)(nil)
	// TODO: removed pending full IBC integration
	// _ ibctesting.TestingApp      = (*App)(nil)
)

// App extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type App struct {
	*baseapp.BaseApp

	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	interfaceRegistry types.InterfaceRegistry
	txConfig          client.TxConfig

	invCheckPeriod uint

	// keys to access the substores
	keyVersions map[uint64][]string
	keys        map[string]*storetypes.KVStoreKey
	tkeys       map[string]*storetypes.TransientStoreKey
	memKeys     map[string]*storetypes.MemoryStoreKey

	// keepers
	AccountsKeeper      accounts.Keeper
	AuthKeeper          authkeeper.AccountKeeper
	BankKeeper          bankkeeper.Keeper
	AuthzKeeper         authzkeeper.Keeper
	ConsensusKeeper     consensuskeeper.Keeper
	StakingKeeper       *stakingkeeper.Keeper
	SlashingKeeper      slashingkeeper.Keeper
	MintKeeper          mintkeeper.Keeper
	DistrKeeper         distrkeeper.Keeper
	GovKeeper           *govkeeper.Keeper
	ParamsKeeper        paramskeeper.Keeper // To be removed after successful migration
	PoolKeeper          poolkeeper.Keeper
	UpgradeKeeper       *upgradekeeper.Keeper // This is included purely for the IBC Keeper. It is not used for upgrading
	SignalKeeper        signal.Keeper
	IBCKeeper           *ibckeeper.Keeper // IBCKeeper must be a pointer in the app, so we can SetRouter on it correctly
	IBCFeeKeeper        ibcfeekeeper.Keeper
	ICAControllerKeeper icacontrollerkeeper.Keeper
	EvidenceKeeper      evidencekeeper.Keeper
	TransferKeeper      ibctransferkeeper.Keeper
	FeeGrantKeeper      feegrantkeeper.Keeper
	ICAHostKeeper       icahostkeeper.Keeper
	// PacketForwardKeeper *packetforwardkeeper.Keeper // TODO: deprecated by ibc-go/v10 ?
	BlobKeeper       blobkeeper.Keeper
	BlobstreamKeeper blobstreamkeeper.Keeper

	manager      *module.Manager
	configurator module.Configurator
	// upgradeHeightV2 is used as a coordination mechanism for the height-based
	// upgrade from v1 to v2.
	upgradeHeightV2 int64
	// timeoutCommit is used to override the default timeoutCommit. This is
	// useful for testing purposes and should not be used on public networks
	// (Arabica, Mocha, or Mainnet Beta).
	timeoutCommit time.Duration
	// MsgGateKeeper is used to define which messages are accepted for a given
	// app version.
	MsgGateKeeper *ante.MsgVersioningGateKeeper
}

// RegisterGRPCServerWithSkipCheckHeader implements server.Application.
func (app *App) RegisterGRPCServerWithSkipCheckHeader(srv grpc.Server, skip bool) {
	app.RegisterGRPCServerWithSkipCheckHeader(srv, skip)
}

// New returns a reference to an uninitialized app. Callers must subsequently
// call app.Info or app.InitChain to initialize the baseapp.
//
// NOTE: upgradeHeightV2 refers specifically to the height that a node will
// upgrade from v1 to v2. It will be deprecated in v3 in place for a dynamically
// signaling scheme
func New(
	logger log.Logger,
	db corestore.KVStoreWithBatch,
	traceStore io.Writer,
	invCheckPeriod uint,
	encodingConfig encoding.Config,
	upgradeHeightV2 int64,
	timeoutCommit time.Duration,
	baseAppOptions ...func(*baseapp.BaseApp),
) *App {
	appCodec := encodingConfig.Codec
	interfaceRegistry := encodingConfig.InterfaceRegistry
	signingCtx := interfaceRegistry.SigningContext()
	txDecoder, err := txdecode.NewDecoder(txdecode.Options{
		SigningContext: signingCtx,
		ProtoCodec:     appCodec,
	})
	if err != nil {
		panic(fmt.Errorf("failed to create tx decoder: %w", err))
	}
	txConfig := authtx.NewTxConfig(
		appCodec,
		signingCtx.AddressCodec(),
		signingCtx.ValidatorAddressCodec(),
		authtx.DefaultSignModes,
	)
	govModuleAddr, err := signingCtx.AddressCodec().BytesToString(authtypes.NewModuleAddress(govtypes.ModuleName))
	if err != nil {
		panic(fmt.Errorf("failed to create gov authority: %w", err))
	}
	cometService := runtime.NewContextAwareCometInfoService()
	legacyAmino := codec.NewLegacyAmino()

	baseApp := baseapp.NewBaseApp(Name, logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	baseApp.SetCommitMultiStoreTracer(traceStore)
	baseApp.SetVersion(version.Version)
	baseApp.SetInterfaceRegistry(interfaceRegistry)

	keys := storetypes.NewKVStoreKeys(allStoreKeys()...)
	envFactory := &moduleEnvFactory{logger: logger, keys: keys, routerProvider: baseApp}
	tkeys := storetypes.NewTransientStoreKeys(paramstypes.TStoreKey)

	app := &App{
		BaseApp:           baseApp,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		txConfig:          encodingConfig.TxConfig,
		invCheckPeriod:    invCheckPeriod,
		keyVersions:       versionedStoreKeys(),
		keys:              keys,
		tkeys:             tkeys,
		// memKeys now nil, was only in use by x/capability
		memKeys:         nil,
		upgradeHeightV2: upgradeHeightV2,
		timeoutCommit:   timeoutCommit,
	}

	// needed for migration from x/params -> module's ownership of own params
	app.ParamsKeeper = initParamsKeeper(appCodec, encodingConfig.Amino, keys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey])
	// only consensus keeper is global scope
	app.ConsensusKeeper = consensuskeeper.NewKeeper(
		appCodec,
		envFactory.make(consensustypes.ModuleName, consensustypes.StoreKey),
		govModuleAddr,
	)

	baseApp.SetParamStore(app.ConsensusKeeper.ParamsStore)
	// set the version modifier
	baseApp.SetVersionModifier(consensus.ProvideAppVersionModifier(app.ConsensusKeeper))

	app.AccountsKeeper, err = accounts.NewKeeper(
		appCodec,
		envFactory.makeWithRouters(accounts.ModuleName, accounts.StoreKey),
		signingCtx.AddressCodec(),
		appCodec.InterfaceRegistry(),
		txDecoder,
		// Lockup account
		accountstd.AddAccount(lockup.CONTINUOUS_LOCKING_ACCOUNT, lockup.NewContinuousLockingAccount),
		accountstd.AddAccount(lockup.PERIODIC_LOCKING_ACCOUNT, lockup.NewPeriodicLockingAccount),
		accountstd.AddAccount(lockup.DELAYED_LOCKING_ACCOUNT, lockup.NewDelayedLockingAccount),
		accountstd.AddAccount(lockup.PERMANENT_LOCKING_ACCOUNT, lockup.NewPermanentLockingAccount),
		accountstd.AddAccount("multisig", multisig.NewAccount),
		// PRODUCTION: add
		baseaccount.NewAccount("base", txConfig.SignModeHandler(), baseaccount.WithSecp256K1PubKey()),
	)
	app.AuthKeeper = authkeeper.NewAccountKeeper(
		envFactory.make(authtypes.ModuleName, authtypes.StoreKey),
		appCodec,
		authtypes.ProtoBaseAccount,
		app.AccountsKeeper,
		maccPerms,
		signingCtx.AddressCodec(),
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
		govModuleAddr,
	)
	app.BankKeeper = bankkeeper.NewBaseKeeper(
		envFactory.make(banktypes.ModuleName, banktypes.StoreKey),
		appCodec,
		app.AuthKeeper,
		app.ModuleAccountAddrs(),
		govModuleAddr,
	)
	app.AuthzKeeper = authzkeeper.NewKeeper(
		envFactory.makeWithRouters(authtypes.ModuleName, authzkeeper.StoreKey),
		appCodec,
		app.AuthKeeper,
	)
	app.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec,
		envFactory.makeWithRouters(stakingtypes.ModuleName, stakingtypes.StoreKey),
		app.AuthKeeper,
		app.BankKeeper,
		app.ConsensusKeeper,
		govModuleAddr,
		signingCtx.ValidatorAddressCodec(),
		signingCtx.AddressCodec(),
		cometService,
	)
	app.MintKeeper = mintkeeper.NewKeeper(
		envFactory.make(minttypes.ModuleName, minttypes.StoreKey),
		appCodec,
		app.StakingKeeper,
		app.AuthKeeper,
		app.BankKeeper,
		authtypes.FeeCollectorName,
	)
	app.PoolKeeper = poolkeeper.NewKeeper(appCodec,
		envFactory.make(pooltypes.ModuleName, pooltypes.StoreKey),
		app.AuthKeeper,
		app.BankKeeper,
		govModuleAddr,
	)
	app.DistrKeeper = distrkeeper.NewKeeper(
		appCodec,
		envFactory.make(distrtypes.ModuleName, distrtypes.StoreKey),
		app.AuthKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		cometService,
		authtypes.FeeCollectorName,
		govModuleAddr,
	)
	app.SlashingKeeper = slashingkeeper.NewKeeper(
		envFactory.make(slashingtypes.ModuleName, slashingtypes.StoreKey),
		appCodec,
		legacyAmino,
		app.StakingKeeper,
		govModuleAddr,
	)

	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(
		envFactory.make(feegrant.ModuleName, feegrant.StoreKey),
		appCodec,
		app.AuthKeeper,
	)
	// The upgrade keeper is initialised solely for the ibc keeper which depends on it to know what the next validator hash is for after the
	// upgrade. This keeper is not used for the actual upgrades but merely for compatibility reasons. Ideally IBC has their own upgrade module
	// for performing IBC based upgrades. Note, as we use rolling upgrades, IBC technically never needs this functionality.
	app.UpgradeKeeper = upgradekeeper.NewKeeper(
		envFactory.makeWithRouters(upgradetypes.ModuleName, upgradetypes.StoreKey),
		nil, appCodec, "", app.BaseApp, govModuleAddr, app.ConsensusKeeper)

	app.BlobstreamKeeper = *blobstreamkeeper.NewKeeper(
		envFactory.make(blobstreamtypes.ModuleName, blobstreamtypes.StoreKey),
		appCodec,
		app.GetSubspace(blobstreamtypes.ModuleName),
		app.StakingKeeper,
		app.ConsensusKeeper,
	)

	// Register the staking hooks. NOTE: stakingKeeper is passed by reference
	// above so that it will contain these hooks.
	app.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(
			app.DistrKeeper.Hooks(),
			app.SlashingKeeper.Hooks(),
			app.BlobstreamKeeper.Hooks(),
		),
	)

	app.SignalKeeper = signal.NewKeeper(
		envFactory.make(signaltypes.ModuleName, signaltypes.StoreKey),
		appCodec,
		&signalStakingWrapper{app.StakingKeeper},
		app.ConsensusKeeper,
	)

	app.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		signingCtx.AddressCodec(),
		envFactory.make(ibcexported.ModuleName, ibcexported.StoreKey),
		app.GetSubspace(ibcexported.ModuleName),
		app.UpgradeKeeper,
		govModuleAddr,
	)

	// ICA Controller keeper
	app.ICAControllerKeeper = icacontrollerkeeper.NewKeeper(
		appCodec,
		runtime.NewEnvironment(
			runtime.NewKVStoreService(keys[icacontrollertypes.StoreKey]), logger.With(log.ModuleKey, "x/icacontroller"),
			runtime.EnvWithMsgRouterService(app.MsgServiceRouter())),
		app.GetSubspace(icacontrollertypes.SubModuleName),
		app.IBCFeeKeeper, // use ics29 fee as ics4Wrapper in middleware stack
		app.IBCKeeper.ChannelKeeper,
		govModuleAddr,
	)

	app.ICAHostKeeper = icahostkeeper.NewKeeper(
		appCodec,
		envFactory.makeWithRouters(icahosttypes.SubModuleName, icahosttypes.StoreKey),
		app.GetSubspace(icahosttypes.SubModuleName),
		app.IBCKeeper.ChannelKeeper, // ICS4Wrapper
		app.IBCKeeper.ChannelKeeper,
		app.AuthKeeper,
		govModuleAddr,
	)

	govRouter := govv1beta1.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.ParamsKeeper))

	govConfig := govkeeper.DefaultConfig()
	/*
		Example of setting gov params:
		govConfig.MaxMetadataLen = 10000
	*/
	app.GovKeeper = govkeeper.NewKeeper(appCodec,
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[govtypes.StoreKey]), logger.With(log.ModuleKey, "x/gov"),
			runtime.EnvWithMsgRouterService(app.MsgServiceRouter()),
			runtime.EnvWithQueryRouterService(app.GRPCQueryRouter())),
		app.AuthKeeper, app.BankKeeper, app.StakingKeeper, app.PoolKeeper, govConfig, govModuleAddr)

	// Set legacy router for backwards compatibility with gov v1beta1
	app.GovKeeper.SetLegacyRouter(govRouter)

	app.GovKeeper = app.GovKeeper.SetHooks(
		govtypes.NewMultiGovHooks(
		// register the governance hooks
		),
	)

	// IBC Fee Module keeper
	app.IBCFeeKeeper = ibcfeekeeper.NewKeeper(
		appCodec,
		signingCtx.AddressCodec(),
		envFactory.make(ibcfeetypes.ModuleName, ibcfeetypes.StoreKey),
		app.IBCKeeper.ChannelKeeper, // may be replaced with IBC middleware
		app.IBCKeeper.ChannelKeeper,
		app.AuthKeeper,
		app.BankKeeper,
	)
	app.TransferKeeper = ibctransferkeeper.NewKeeper(
		appCodec,
		signingCtx.AddressCodec(),
		envFactory.make(ibctransfertypes.ModuleName, ibctransfertypes.StoreKey),
		app.GetSubspace(ibctransfertypes.ModuleName),
		app.IBCFeeKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.AuthKeeper,
		app.BankKeeper,
		govModuleAddr,
	)
	// Transfer stack contains (from top to bottom):
	// - Token Filter
	// - Packet Forwarding Middleware
	// - Transfer
	// var transferStack ibcporttypes.IBCModule
	// transferStack = transfer.NewIBCModule(app.TransferKeeper)
	// packetForwardMiddleware := packetforward.NewIBCMiddleware(
	// 	transferStack,
	// 	app.PacketForwardKeeper,
	// 	0, // retries on timeout
	// 	packetforwardkeeper.DefaultForwardTransferPacketTimeoutTimestamp, // forward timeout
	// 	packetforwardkeeper.DefaultRefundTransferPacketTimeoutTimestamp,  // refund timeout
	// )
	// // PacketForwardMiddleware is used only for version >= 2.
	// transferStack = module.NewVersionedIBCModule(packetForwardMiddleware, transferStack, v2, v3)
	// Token filter wraps packet forward middleware and is thus the first module in the transfer stack.
	//tokenFilterMiddelware := tokenfilter.NewIBCMiddleware(transferStack)
	//transferStack = module.NewVersionedIBCModule(tokenFilterMiddelware, transferStack, v1, v3)

	app.EvidenceKeeper = *evidencekeeper.NewKeeper(
		appCodec,
		envFactory.make(evidencetypes.ModuleName, evidencetypes.StoreKey),
		app.StakingKeeper,
		app.SlashingKeeper,
		app.ConsensusKeeper,
		signingCtx.AddressCodec(),
	)

	app.GovKeeper = govkeeper.NewKeeper(
		appCodec,
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[govtypes.StoreKey]), logger.With(log.ModuleKey, "x/gov"),
			runtime.EnvWithMsgRouterService(app.MsgServiceRouter()),
			runtime.EnvWithQueryRouterService(app.GRPCQueryRouter())),
		app.AuthKeeper, app.BankKeeper, app.StakingKeeper, app.PoolKeeper, govConfig, govModuleAddr)

	app.BlobKeeper = *blobkeeper.NewKeeper(
		envFactory.make(blobtypes.ModuleName, blobtypes.StoreKey),
		appCodec,
		app.GetSubspace(blobtypes.ModuleName),
	)

	// app.PacketForwardKeeper.SetTransferKeeper(app.TransferKeeper)
	ibcRouter := ibcporttypes.NewRouter() // Create static IBC router
	// ibcRouter.AddRoute(ibctransfertypes.ModuleName, transferStack)                          // Add transfer route
	ibcRouter.AddRoute(icahosttypes.SubModuleName, icahost.NewIBCModule(app.ICAHostKeeper)) // Add ICA route
	app.IBCKeeper.SetRouter(ibcRouter)

	/****  Module Options ****/

	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.

	// NOTE: Modules can't be modified or else must be passed by reference to the module manager
	err = app.setupModuleManager(txConfig, cometService, true)
	if err != nil {
		panic(err)
	}

	// order begin block, end block and init genesis
	app.setModuleOrder()

	// TODO: How to support this in 0.52?
	// app.QueryRouter().AddRoute(proof.TxInclusionQueryPath, proof.QueryTxInclusionProof)
	// app.QueryRouter().AddRoute(proof.ShareInclusionQueryPath, proof.QueryShareInclusionProof)

	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.manager.RegisterServices(app.configurator)

	// extract the accepted message list from the configurator and create a gatekeeper
	// which will be used both as the antehandler and as part of the circuit breaker in
	// the msg service router
	app.MsgGateKeeper = ante.NewMsgVersioningGateKeeper(app.configurator.GetAcceptedMessages(), app.ConsensusKeeper)
	app.MsgServiceRouter().SetCircuit(app.MsgGateKeeper)

	// Initialize the KV stores for the base modules (e.g. params). The base modules will be included in every app version.
	app.MountKVStores(app.baseKeys())
	app.MountTransientStores(tkeys)

	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)
	app.SetPrepareProposal(app.PrepareProposalHandler)
	app.SetProcessProposal(app.ProcessProposalHandler)

	app.SetAnteHandler(ante.NewAnteHandler(
		app.AuthKeeper,
		app.AccountsKeeper,
		app.BankKeeper,
		app.BlobKeeper,
		app.ConsensusKeeper,
		app.FeeGrantKeeper,
		encodingConfig.TxConfig.SignModeHandler(),
		ante.DefaultSigVerificationGasConsumer,
		app.IBCKeeper,
		app.ParamsKeeper,
		app.MsgGateKeeper,
		app.BlockedParamsGovernance(),
	))
	app.SetPostHandler(posthandler.New())

	// TODO: migration related, delaying implemenation for now
	// app.SetMigrateStoreFn(app.migrateCommitStore)
	// app.SetMigrateModuleFn(app.migrateModules)

	// assert that keys are present for all supported versions
	app.assertAllKeysArePresent()

	// we don't seal the store until the app version has been initialised
	// this will just initialise the base keys (i.e. the param store)
	if err := app.CommitMultiStore().LoadLatestVersion(); err != nil {
		panic(err)
	}

	return app
}

// Name returns the name of the App
func (app *App) Name() string { return app.BaseApp.Name() }

// BeginBlocker application updates every begin block
func (app *App) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	if ctx.HeaderInfo().Height == app.upgradeHeightV2 {
		app.BaseApp.Logger().Info("upgraded from app version 1 to 2")
	}
	return app.manager.BeginBlock(ctx)
}

// EndBlocker executes application updates at the end of every block.
func (app *App) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	res, err := app.manager.EndBlock(ctx)
	if err != nil {
		return sdk.EndBlock{}, err
	}
	currentVersion, err := app.AppVersion(ctx)
	if err != nil {
		panic(err)
	}
	// For v1 only we upgrade using an agreed upon height known ahead of time
	if currentVersion == v1 {
		// check that we are at the height before the upgrade
		if ctx.HeaderInfo().Height == app.upgradeHeightV2-1 {
			app.BaseApp.Logger().Info(fmt.Sprintf("upgrading from app version %v to 2", currentVersion))
			if err = app.SetAppVersion(ctx, v2); err != nil {
				panic(err)
			}
		}
		// from v2 to v3 and onwards we use a signaling mechanism
	} else if shouldUpgrade, newVersion := app.SignalKeeper.ShouldUpgrade(ctx); shouldUpgrade {
		// Version changes must be increasing. Downgrades are not permitted
		if newVersion > currentVersion {
			app.BaseApp.Logger().Info("upgrading app version", "current version", currentVersion, "new version", newVersion)
			if err = app.SetAppVersion(ctx, newVersion); err != nil {
				panic(err)
			}
			app.SignalKeeper.ResetTally(ctx)
		}
	}
	// REVIEW: these are moved to YAML, ok to delete?
	// res.Timeouts.TimeoutCommit = app.getTimeoutCommit(currentVersion)
	// res.Timeouts.TimeoutPropose = appconsts.GetTimeoutPropose(currentVersion)
	return res, nil
}

// migrateCommitStore tells the baseapp during a version upgrade, which stores to add and which
// stores to remove
func (app *App) migrateCommitStore(fromVersion, toVersion uint64) (*storetypes.StoreUpgrades, error) {
	oldStoreKeys := app.keyVersions[fromVersion]
	newStoreKeys := app.keyVersions[toVersion]
	result := &storetypes.StoreUpgrades{}
	for _, oldKey := range oldStoreKeys {
		if !slices.Contains(newStoreKeys, oldKey) {
			result.Deleted = append(result.Deleted, oldKey)
		}
	}
	for _, newKey := range newStoreKeys {
		if !slices.Contains(oldStoreKeys, newKey) {
			result.Added = append(result.Added, newKey)
		}
	}
	return result, nil
}

// migrateModules performs migrations on existing modules that have registered migrations
// between versions and initializes the state of new modules for the specified app version.
func (app *App) migrateModules(ctx sdk.Context, fromVersion, toVersion uint64) error {
	return app.manager.RunMigrations(ctx, app.configurator, fromVersion, toVersion)
}

// Info implements the ABCI interface. This method is a wrapper around baseapp's
// Info command so that it can take the app version and setup the multicommit
// store.
//
// Side-effect: calls baseapp.Init()
func (app *App) InfoV1(req *abci.InfoRequest) (*abci.InfoResponse, error) {
	if height := app.LastBlockHeight(); height > 0 {
		ctx, err := app.CreateQueryContext(height, false)
		if err != nil {
			return nil, err
		}
		appVersion, err := app.AppVersion(ctx)
		if err != nil {
			return nil, err
		}
		if appVersion == 0 {
			app.SetAppVersion(ctx, v1)
		}
	}

	resp, err := app.BaseApp.Info(req)
	if err != nil {
		return nil, err
	}
	// mount the stores for the provided app version
	if resp.AppVersion > 0 && !app.IsSealed() {
		app.mountKeysAndInit(resp.AppVersion)
	}

	// REVIEW: these were moved to YAML, okay to delete?
	// resp.Timeouts.TimeoutCommit = app.getTimeoutCommit(resp.AppVersion)
	// resp.Timeouts.TimeoutPropose = appconsts.GetTimeoutPropose(resp.AppVersion)

	return resp, nil
}

// InitChain implements the ABCI interface. This method is a wrapper around
// baseapp's InitChain so that we can take the app version and setup the multicommit
// store.
//
// Side-effect: calls baseapp.Init()
func (app *App) InitChainV1(req *abci.InitChainRequest) (res *abci.InitChainResponse, err error) {
	req = setDefaultAppVersion(req)
	appVersion := req.ConsensusParams.Version.App
	ctx := context.Background()
	av, err := app.AppVersion(ctx)
	if err != nil {
		return nil, err
	}
	if av == 0 && !app.IsSealed() {
		app.mountKeysAndInit(appVersion)
	}

	res, err = app.BaseApp.InitChain(req)
	if err != nil {
		return nil, err
	}

	if appVersion != v1 {
		app.SetAppVersion(ctx, appVersion)
	}
	return res, nil
}

// setDefaultAppVersion sets the default app version in the consensus params if
// it was 0. This is needed because chains (e.x. mocha-4) did not explicitly set
// an app version in genesis.json.
func setDefaultAppVersion(req *abci.InitChainRequest) *abci.InitChainRequest {
	if req.ConsensusParams == nil {
		panic("no consensus params set")
	}
	if req.ConsensusParams.Version == nil {
		panic("no version set in consensus params")
	}
	if req.ConsensusParams.Version.App == 0 {
		req.ConsensusParams.Version.App = v1
	}
	return req
}

// mountKeysAndInit mounts the keys for the provided app version and then
// invokes baseapp.Init().
func (app *App) mountKeysAndInit(appVersion uint64) {
	app.Logger().Info(fmt.Sprintf("mounting KV stores for app version %v", appVersion))
	app.MountKVStores(app.versionedKeys(appVersion))

	// Invoke load latest version for its side-effect of invoking baseapp.Init()
	if err := app.LoadLatestVersion(); err != nil {
		panic(fmt.Sprintf("loading latest version: %s", err.Error()))
	}
}

// InitChainer is middleware that gets invoked part-way through the baseapp's InitChain invocation.
func (app *App) InitChainer(ctx sdk.Context, req *abci.InitChainRequest) (*abci.InitChainResponse, error) {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		return nil, err
	}
	appVersion := req.ConsensusParams.Version.App
	versionMap, err := app.manager.GetVersionMap(appVersion)
	if err != nil {
		return nil, err
	}
	app.UpgradeKeeper.SetModuleVersionMap(ctx, versionMap)
	return app.manager.InitGenesis(ctx, genesisState, appVersion)
}

// LoadHeight loads a particular height
func (app *App) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// SupportedVersions returns all the state machines that the
// application supports
func (app *App) SupportedVersions() []uint64 {
	return app.manager.SupportedVersions()
}

// versionedKeys returns a map from moduleName to KV store key for the given app
// version.
func (app *App) versionedKeys(appVersion uint64) map[string]*storetypes.KVStoreKey {
	output := make(map[string]*storetypes.KVStoreKey)
	if keys, exists := app.keyVersions[appVersion]; exists {
		for _, moduleName := range keys {
			if key, exists := app.keys[moduleName]; exists {
				output[moduleName] = key
			}
		}
	}
	return output
}

// baseKeys returns the base keys that are mounted to every version
func (app *App) baseKeys() map[string]*storetypes.KVStoreKey {
	return map[string]*storetypes.KVStoreKey{
		// we need to know the app version to know what stores to mount
		// thus the paramstore must always be a store that is mounted
		paramstypes.StoreKey: app.keys[paramstypes.StoreKey],
	}
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *App) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// GetBaseApp implements the TestingApp interface.
func (app *App) GetBaseApp() *baseapp.BaseApp {
	return app.BaseApp
}

// GetIBCKeeper implements the TestingApp interface.
func (app *App) GetIBCKeeper() *ibckeeper.Keeper {
	return app.IBCKeeper
}

// GetTxConfig implements the TestingApp interface.
func (app *App) GetTxConfig() client.TxConfig {
	return app.txConfig
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// AppCodec returns the app's appCodec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns the app's InterfaceRegistry
func (app *App) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetKey(storeKey string) *storetypes.KVStoreKey {
	return app.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetTKey(storeKey string) *storetypes.TransientStoreKey {
	return app.tkeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (app *App) GetMemKey(storeKey string) *storetypes.MemoryStoreKey {
	return app.memKeys[storeKey]
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *App) RegisterAPIRoutes(apiSvr *api.Server, _ config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	app.RegisterTendermintService(clientCtx)
	// Register node gRPC service for grpc-gateway.
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	celestiatx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *App) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
	celestiatx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *App) RegisterTendermintService(clientCtx client.Context) {
	cmtservice.RegisterTendermintService(
		clientCtx,
		app.BaseApp.GRPCQueryRouter(),
		app.interfaceRegistry,
		func(ctx context.Context, req *abci.QueryRequest) (*abci.QueryResponse, error) {
			return app.BaseApp.Query(ctx, req)
		},
	)
}

func (app *App) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg)
}

// initParamsKeeper initializes the params keeper and its subspaces.
func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(minttypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	// paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govv1beta2.ParamKeyTable())
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(ibcexported.ModuleName)
	paramsKeeper.Subspace(icahosttypes.SubModuleName)
	paramsKeeper.Subspace(blobtypes.ModuleName)
	paramsKeeper.Subspace(blobstreamtypes.ModuleName)
	paramsKeeper.Subspace(minfee.ModuleName)
	// paramsKeeper.Subspace(packetforwardtypes.ModuleName)

	return paramsKeeper
}

func isSupportedAppVersion(appVersion uint64) bool {
	return appVersion == v1 || appVersion == v2 || appVersion == v3
}

// getTimeoutCommit returns the timeoutCommit if a user has overridden it via the
// --timeout-commit flag. Otherwise, it returns the default timeout commit based
// on the app version.
func (app *App) getTimeoutCommit(appVersion uint64) time.Duration {
	if app.timeoutCommit != 0 {
		return app.timeoutCommit
	}
	return appconsts.GetTimeoutCommit(appVersion)
}

type moduleEnvFactory struct {
	keys           map[string]*storetypes.KVStoreKey
	logger         log.Logger
	routerProvider interface {
		MsgServiceRouter() *baseapp.MsgServiceRouter
		GRPCQueryRouter() *baseapp.GRPCQueryRouter
	}
}

type signalStakingWrapper struct {
	*stakingkeeper.Keeper
}

func (s *signalStakingWrapper) GetLastTotalPower(ctx context.Context) (math.Int, error) {
	return s.Keeper.LastTotalPower.Get(ctx)
}

func (f *moduleEnvFactory) make(
	name string,
	storeKey string,
) appmodulev2.Environment {
	kvKey, ok := f.keys[storeKey]
	if !ok {
		panic(fmt.Sprintf("store key %s not found", storeKey))
	}
	return runtime.NewEnvironment(
		runtime.NewKVStoreService(kvKey),
		f.logger.With(log.ModuleKey, name),
	)
}

func (f *moduleEnvFactory) makeWithRouters(
	name string,
	storeKey string,
) appmodulev2.Environment {
	kvKey, ok := f.keys[storeKey]
	if !ok {
		panic(fmt.Sprintf("store key %s not found", storeKey))
	}
	return runtime.NewEnvironment(
		runtime.NewKVStoreService(kvKey),
		f.logger.With(log.ModuleKey, name),
		runtime.EnvWithMsgRouterService(f.routerProvider.MsgServiceRouter()),
		runtime.EnvWithQueryRouterService(f.routerProvider.GRPCQueryRouter()),
	)
}

// ValidatorKeyProvider returns a function that generates a validator key
// Supported key types are those supported by Comet: ed25519, secp256k1, bls12-381
func (app *App) ValidatorKeyProvider() runtime.KeyGenF {
	return func() (cmtcrypto.PrivKey, error) {
		return cmted25519.GenPrivKey(), nil
	}
}

// BlockedParamsGovernance returns the params that require a hardfork to change, and
// cannot be changed via governance.
func (app *App) BlockedParamsGovernance() map[string][]string {
	return map[string][]string{
		gogoproto.MessageName(&banktypes.MsgUpdateParams{}):      []string{"send_enabled"},
		gogoproto.MessageName(&stakingtypes.MsgUpdateParams{}):   []string{"params.bond_denom", "params.unbonding_time"},
		gogoproto.MessageName(&consensustypes.MsgUpdateParams{}): []string{"validator"},
	}

}

// NewProposalContext returns a context with a branched version of the state
// that is safe to query during ProcessProposal.
func (app *App) NewProposalContext(header cmtproto.Header) sdk.Context {
	// use custom query multistore if provided
	ms := app.CommitMultiStore().CacheMultiStore()
	ctx := sdk.NewContext(ms, false, app.Logger()).
		WithBlockGasMeter(storetypes.NewInfiniteGasMeter()).
		WithBlockHeader(header).
		WithHeaderInfo(coreheader.Info{
			Height:  header.Height,
			AppHash: header.AppHash,
			Hash:    header.ConsensusHash,
			Time:    header.Time,
			ChainID: app.ChainID(),
		})
	ctx = ctx.WithConsensusParams(app.GetConsensusParams(ctx))

	return ctx
}
