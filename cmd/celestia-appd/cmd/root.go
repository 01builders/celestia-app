package cmd

import (
	"fmt"
	"os"

	kitlog "github.com/go-kit/log"

	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/log"
	confixcmd "cosmossdk.io/tools/confix/cmd"
	"github.com/celestiaorg/celestia-app/v4/app"
	blobstreamclient "github.com/celestiaorg/celestia-app/v4/x/blobstream/client"
	"github.com/cometbft/cometbft/cmd/cometbft/commands"
	tmcli "github.com/cometbft/cometbft/libs/cli"
	"github.com/cosmos/cosmos-sdk/client"
	clientconfig "github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/snapshot"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"
)

const (
	// FlagLogToFile specifies whether to log to file or not.
	FlagLogToFile = "log-to-file"

	// UpgradeHeightFlag is the flag to specify the upgrade height for v1 to v2
	// application upgrade.
	UpgradeHeightFlag = "v2-upgrade-height"

	// TimeoutCommit is a flag that can be used to override the timeout_commit.
	TimeoutCommitFlag = "timeout-commit"
)

// NewRootCmd creates a new root command for celestia-appd.
func NewRootCmd() *cobra.Command {
	// we "pre"-instantiate the application for getting the injected/configured encoding configuration
	// note, this is not necessary when using app wiring, as depinject can be directly used (see root_v2.go)
	tempApp := app.New(log.NewNopLogger(), coretesting.NewMemDB(), nil, 0, 0)

	initClientContext := client.Context{}.
		WithCodec(tempApp.AppCodec()).
		WithInterfaceRegistry(tempApp.InterfaceRegistry()).
		WithTxConfig(tempApp.GetTxConfig()).
		WithLegacyAmino(tempApp.LegacyAmino()).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastSync).
		WithHomeDir(app.DefaultNodeHome).
		WithViper(app.EnvPrefix)

	rootCommand := &cobra.Command{
		Use: "celestia-appd",
		PersistentPreRunE: func(command *cobra.Command, _ []string) error {
			command.SetOut(command.OutOrStdout())
			command.SetErr(command.ErrOrStderr())

			clientContext, err := client.ReadPersistentCommandFlags(initClientContext, command.Flags())
			if err != nil {
				return err
			}
			clientContext, err = clientconfig.ReadFromClientConfig(clientContext)
			if err != nil {
				return err
			}
			if err := client.SetCmdClientContextHandler(clientContext, command); err != nil {
				return err
			}

			appTemplate := serverconfig.DefaultConfigTemplate
			appConfig := app.DefaultAppConfig()
			tmConfig := app.DefaultConsensusConfig()

			// Override the default tendermint config and app config for celestia-app
			err = server.InterceptConfigsPreRunHandler(command, appTemplate, appConfig, tmConfig)
			if err != nil {
				return err
			}

			if command.Flags().Changed(FlagLogToFile) {
				// optionally log to file by replacing the default logger with a file logger
				err = replaceLogger(command)
				if err != nil {
					return err
				}
			}

			return nil
		},
		SilenceUsage: true,
	}

	rootCommand.PersistentFlags().String(FlagLogToFile, "", "Write logs directly to a file. If empty, logs are written to stderr")
	initRootCommand(rootCommand, tempApp)

	return rootCommand
}

// initRootCommand performs a bunch of side-effects on the root command.
func initRootCommand(rootCommand *cobra.Command, app *app.App) {
	var genesisModule genutil.AppModule
	if gm, err := app.ModuleManager.Module(genutiltypes.ModuleName); err == nil {
		genesisModule = gm.(genutil.AppModule)
	}

	genesisCmd := genutilcli.Commands(genesisModule, app.ModuleManager, appExporter)
	rootCommand.AddCommand(
		genesisCmd,
		tmcli.NewCompletionCmd(rootCommand, true),
		debug.Cmd(),
		confixcmd.ConfigCommand(),
		commands.CompactGoLevelDBCmd,
		addrbookCommand(),
		downloadGenesisCommand(),
		addrConversionCmd(),
		server.StatusCommand(),
		queryCommand(app.ModuleManager),
		txCommand(app.ModuleManager),
		keys.Commands(),
		blobstreamclient.VerifyCmd(),
		snapshot.Cmd(newCmdApplication),
	)

	// Add the following commands to the rootCommand: start, tendermint, export, version, and rollback.
	server.AddCommands(rootCommand, newCmdApplication, server.StartCmdOptions[servertypes.Application]{
		AddFlags: addStartFlags,
	})

	// find start command
	startCmd, _, err := rootCommand.Find([]string{"start"})
	if err != nil {
		panic(fmt.Errorf("failed to find start command: %w", err))
	}
	startCmdRunE := startCmd.RunE

	// Add the BBR check to the start command
	startCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := checkBBR(cmd); err != nil {
			return err
		}

		return startCmdRunE(cmd, args)
	}
}

// addStartFlags adds flags to the start command.
func addStartFlags(startCmd *cobra.Command) {
	startCmd.Flags().Int64(UpgradeHeightFlag, 0, "Upgrade height to switch from v1 to v2. Must be coordinated amongst all validators")
	startCmd.Flags().Duration(TimeoutCommitFlag, 0, "Override the application configured timeout_commit. Note: only for testing purposes.")
	startCmd.Flags().Bool(FlagForceNoBBR, false, "bypass the requirement to use bbr locally")
}

// replaceLogger optionally replaces the logger with a file logger if the flag
// is set to something other than the default.
func replaceLogger(cmd *cobra.Command) error {
	logFilePath, err := cmd.Flags().GetString(FlagLogToFile)
	if err != nil {
		return err
	}

	if logFilePath == "" {
		return nil
	}

	file, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}

	sctx := server.GetServerContextFromCmd(cmd)
	sctx.Logger = log.NewLogger(kitlog.NewSyncWriter(file))
	return server.SetCmdServerContext(cmd, sctx)
}
