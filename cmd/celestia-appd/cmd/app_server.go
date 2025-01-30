package cmd

import (
	"io"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/celestiaorg/celestia-app/v4/app"
	celestiaserver "github.com/celestiaorg/celestia-app/v4/server"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/spf13/cast"
)

func NewAppServer(
	logger log.Logger,
	db corestore.KVStoreWithBatch,
	traceStore io.Writer,
	appOpts servertypes.AppOptions,
) celestiaserver.Application {
	baseappOptions := server.DefaultBaseappOptions(appOpts)
	return app.New(
		logger,
		db,
		traceStore,
		cast.ToInt64(appOpts.Get(UpgradeHeightFlag)),
		cast.ToDuration(appOpts.Get(TimeoutCommitFlag)),
		baseappOptions...,
	)
}
