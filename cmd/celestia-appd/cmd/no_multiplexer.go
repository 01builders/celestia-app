//go:build disable_multiplexer
// +build disable_multiplexer

package cmd

import (
	"fmt"

	confixcmd "cosmossdk.io/tools/confix/cmd"
	"github.com/cometbft/cometbft/cmd/cometbft/commands"
	tmcli "github.com/cometbft/cometbft/libs/cli"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/snapshot"
	"github.com/cosmos/cosmos-sdk/server"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/spf13/cobra"

	"github.com/celestiaorg/celestia-app/v4/app"
)

// initRootCommand performs a bunch of side-effects on the root command.
func initRootCommand(rootCommand *cobra.Command, capp *app.App) {
	debugCmd := debug.Cmd()
	debugCmd.AddCommand(
		NewInPlaceTestnetCmd(),
		AppGenesisToCometGenesisConverterCmd(),
	)

	rootCommand.AddCommand(
		genutilcli.InitCmd(capp.BasicManager, app.DefaultNodeHome),
		genutilcli.Commands(capp.GetTxConfig(), capp.BasicManager, app.DefaultNodeHome),
		tmcli.NewCompletionCmd(rootCommand, true),
		debugCmd,
		confixcmd.ConfigCommand(),
		commands.CompactGoLevelDBCmd,
		addrbookCommand(),
		downloadGenesisCommand(),
		addrConversionCmd(),
		server.StatusCommand(),
		queryCommand(capp.BasicManager),
		txCommand(capp.BasicManager),
		keys.Commands(),
		snapshot.Cmd(NewAppServer),
	)

	// Add the following commands to the rootCommand: start, tendermint, export, version, and rollback.
	server.AddCommandsWithStartCmdOptions(rootCommand, app.DefaultNodeHome, NewAppServer, appExporter, server.StartCmdOptions{
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
