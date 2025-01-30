package cmd

import (
	"io"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/store"
	snapshottypes "cosmossdk.io/store/snapshots/types"
	storetypes "cosmossdk.io/store/types"
	"github.com/celestiaorg/celestia-app/v4/app"
	celestiaserver "github.com/celestiaorg/celestia-app/v4/server"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/spf13/cast"
)

func NewAppServer(
	logger log.Logger,
	db corestore.KVStoreWithBatch,
	traceStore io.Writer,
	appOptions servertypes.AppOptions,
) celestiaserver.Application {
	var cache storetypes.MultiStorePersistentCache

	if cast.ToBool(appOptions.Get(server.FlagInterBlockCache)) {
		cache = store.NewCommitKVStoreCacheManager()
	}

	pruningOpts, err := server.GetPruningOptionsFromFlags(appOptions)
	if err != nil {
		panic(err)
	}

	snapshotStore, err := server.GetSnapshotStore(appOptions)
	if err != nil {
		panic(err)
	}

	return app.New(
		logger,
		db,
		traceStore,
		cast.ToInt64(appOptions.Get(UpgradeHeightFlag)),
		cast.ToDuration(appOptions.Get(TimeoutCommitFlag)),
		baseapp.SetPruning(pruningOpts),
		baseapp.SetMinGasPrices(cast.ToString(appOptions.Get(server.FlagMinGasPrices))),
		baseapp.SetMinRetainBlocks(cast.ToUint64(appOptions.Get(server.FlagMinRetainBlocks))),
		baseapp.SetHaltHeight(cast.ToUint64(appOptions.Get(server.FlagHaltHeight))),
		baseapp.SetHaltTime(cast.ToUint64(appOptions.Get(server.FlagHaltTime))),
		baseapp.SetMinRetainBlocks(cast.ToUint64(appOptions.Get(server.FlagMinRetainBlocks))),
		baseapp.SetInterBlockCache(cache),
		baseapp.SetTrace(cast.ToBool(appOptions.Get(server.FlagTrace))),
		baseapp.SetIndexEvents(cast.ToStringSlice(appOptions.Get(server.FlagIndexEvents))),
		baseapp.SetSnapshot(snapshotStore, snapshottypes.NewSnapshotOptions(cast.ToUint64(appOptions.Get(server.FlagStateSyncSnapshotInterval)), cast.ToUint32(appOptions.Get(server.FlagStateSyncSnapshotKeepRecent)))),
	)
}
