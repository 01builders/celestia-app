package app

import (
	"cosmossdk.io/x/accounts"
	consensustypes "cosmossdk.io/x/consensus/types"
	pooltypes "cosmossdk.io/x/protocolpool/types"
	ibcfeetypes "github.com/cosmos/ibc-go/v9/modules/apps/29-fee/types"

	"cosmossdk.io/x/authz"
	authzkeeper "cosmossdk.io/x/authz/keeper"
	banktypes "cosmossdk.io/x/bank/types"
	distrtypes "cosmossdk.io/x/distribution/types"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	govtypes "cosmossdk.io/x/gov/types"
	paramstypes "cosmossdk.io/x/params/types"
	slashingtypes "cosmossdk.io/x/slashing/types"
	stakingtypes "cosmossdk.io/x/staking/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	blobtypes "github.com/celestiaorg/celestia-app/v4/x/blob/types"
	"github.com/celestiaorg/celestia-app/v4/x/minfee"
	minttypes "github.com/celestiaorg/celestia-app/v4/x/mint/types"
	signaltypes "github.com/celestiaorg/celestia-app/v4/x/signal/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	// "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v9/packetforward"
	// packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v9/packetforward/types"

	icacontrollertypes "github.com/cosmos/ibc-go/v9/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v9/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v9/modules/apps/27-interchain-accounts/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v9/modules/apps/transfer/types"
	ibcexported "github.com/cosmos/ibc-go/v9/modules/core/exported"
)

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
		consensustypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		accounts.ModuleName,
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
	return []string{
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
	}
}
