package cmd

import (
	"io"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/celestiaorg/celestia-app/v4/app"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
)

func appExporter(
	logger log.Logger,
	db corestore.KVStoreWithBatch,
	traceStore io.Writer,
	height int64,
	forZeroHeight bool,
	jailWhiteList []string,
	appOptions servertypes.AppOptions,
	_ []string,
) (servertypes.ExportedApp, error) {
	application := app.New(logger, db, traceStore, 0, 0)
	if height != -1 {
		if err := application.LoadHeight(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	}
	return application.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
}
