package cmd

import (
	"io"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/celestiaorg/celestia-app/v3/app"
	"github.com/celestiaorg/celestia-app/v3/app/encoding"
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
	config := encoding.MakeConfig(app.ModuleEncodingRegisters...)
	application := app.New(logger, db, traceStore, uint(1), config, 0, 0, appOptions)
	if height != -1 {
		if err := application.LoadHeight(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	}
	return application.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
}
