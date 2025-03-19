package cmd

import (
	"fmt"
	"path/filepath"

	"cosmossdk.io/log"
	"cosmossdk.io/x/upgrade/types"
	"github.com/celestiaorg/celestia-app/v4/app"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
)

func ForceUpgradeCmd(appCreator servertypes.AppCreator) *cobra.Command {
	return &cobra.Command{
		Use:     "force-upgrade [upgrade-name]",
		Short:   "Force upgrade the node to a specific version",
		Long:    "Force upgrade the node to a specific version. This command is used to force the node to upgrade to a specific version, even if the upgrade is not scheduled. The chain must be stopped before running this command. The name of the upgrade must be the name of the upgrade handler in the codebase.",
		Example: `celestia-appd force-upgrade v4.0.0`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := server.GetServerContextFromCmd(cmd)

			version := args[0]
			if version == "" {
				return fmt.Errorf("version cannot be empty")
			}

			home := ctx.Config.RootDir
			db, err := openDB(home, server.GetAppDBBackend(ctx.Viper))
			if err != nil {
				return err
			}
			logger := log.NewLogger(cmd.OutOrStdout())
			app, ok := appCreator(logger, db, nil, ctx.Viper).(*app.App)
			if !ok {
				return fmt.Errorf("failed to create app, expected *app.App, got %T", app)
			}

			if err := app.UpgradeKeeper.ScheduleUpgrade(
				app.GetBaseApp().NewContext(false),
				types.Plan{
					Name:   version,
					Height: app.CommitMultiStore().LatestVersion() + 1,
				},
			); err != nil {
				return fmt.Errorf("failed to schedule upgrade: %w", err)
			}

			return nil
		},
	}
}

func openDB(rootDir string, backendType dbm.BackendType) (dbm.DB, error) {
	dataDir := filepath.Join(rootDir, "data")
	return dbm.NewDB("application", backendType, dataDir)
}
