package cmd

import (
	"fmt"
	"os"

	kitlog "github.com/go-kit/log"

	"cosmossdk.io/log"
	confixcmd "cosmossdk.io/tools/confix/cmd"
	"github.com/celestiaorg/celestia-app/v3/app"
	"github.com/celestiaorg/celestia-app/v3/app/encoding"
	blobstreamclient "github.com/celestiaorg/celestia-app/v3/x/blobstream/client"
	"github.com/cometbft/cometbft/cmd/cometbft/commands"
	cmtcfg "github.com/cometbft/cometbft/config"
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
	// EnvPrefix is the environment variable prefix for celestia-appd.
	// Environment variables that Cobra reads must be prefixed with this value.
	EnvPrefix = "CELESTIA"

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
	encodingConfig := encoding.MakeConfig(app.ModuleEncodingRegisters...)
	initClientContext := client.Context{}.
		WithCodec(encodingConfig.Codec).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastSync).
		WithHomeDir(app.DefaultNodeHome).
		WithViper(EnvPrefix)

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
			cometConfig := &cmtcfg.Config{}
			if err := DeepClone(tmConfig, cometConfig); err != nil {
				return err
			}

			// Override the default tendermint config and app config for celestia-app
			err = server.InterceptConfigsPreRunHandler(command, appTemplate, appConfig, cometConfig)
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
	initRootCommand(rootCommand, encodingConfig)

	return rootCommand
}

// initRootCommand performs a bunch of side-effects on the root command.
func initRootCommand(rootCommand *cobra.Command, encodingConfig encoding.Config) {
	genesisModule := app.ModuleBasics.Modules[genutiltypes.ModuleName].(genutil.AppModule)
	genesisCmd := genutilcli.Commands(genesisModule, app.ModuleBasics, appExporter)
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
		queryCommand(),
		txCommand(),
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
