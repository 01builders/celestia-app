//go:build multiplexer

package cmd

import (
	"github.com/01builders/nova"
	"github.com/01builders/nova/abci"
	"github.com/celestiaorg/celestia-app/v4/app"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/spf13/cobra"
)

// modifyRootCommand enhances the root command with the pass through and multiplexer.
func modifyRootCommand(rootCommand *cobra.Command) {
	versions, err := abci.NewVersions(Versions()...)
	if err != nil {
		panic(err)
	}
	passthroughCmd := nova.NewPassthroughCmd(versions)
	rootCommand.AddCommand(passthroughCmd)
	// Add the following commands to the rootCommand: start, tendermint, export, version, and rollback.
	server.AddCommandsWithStartCmdOptions(rootCommand, app.DefaultNodeHome, NewAppServer, appExporter, server.StartCmdOptions{
		AddFlags:            addStartFlags,
		StartCommandHandler: nova.New(versions), // multiplexer
	})
}
