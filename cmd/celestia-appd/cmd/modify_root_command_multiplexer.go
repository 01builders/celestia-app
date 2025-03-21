//go:build multiplexer

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/01builders/nova"
	"github.com/01builders/nova/abci"
	"github.com/01builders/nova/appd"
	"github.com/celestiaorg/celestia-app/v4/app"
	"github.com/cosmos/cosmos-sdk/server"

	embedding "github.com/celestiaorg/celestia-app/v4/internal/embedding"
)

// modifyRootCommand enhances the root command with the pass through and multiplexer.
func modifyRootCommand(rootCommand *cobra.Command) {
	v3AppBinary, err := embedding.CelestiaAppV3()
	_ = err // TODO: handle the error in this case.

	v3, err := appd.New("v3", v3AppBinary)
	_ = err // TODO: handle the error in this case.

	v4AppBinary, err := embedding.CelestiaAppV4()
	_ = err // TODO: handle the error in this case.

	v4, err := appd.New("v4", v4AppBinary)
	_ = err // TODO: handle the error in this case.

	versions, err := abci.NewVersions(abci.Version{
		Appd:        v3,
		ABCIVersion: abci.ABCIClientVersion1,
		AppVersion:  3,
		StartArgs: []string{
			"--grpc.enable=true",
			"--api.enable=true",
			"--api.swagger=false",
			"--with-tendermint=false",
			"--transport=grpc",
			// "--v2-upgrade-height=0",
		},
		abci.Version{
			Appd:        v4,
			ABCIVersion: abci.ABCIClientVersion2,
			AppVersion:  4,
		},
	})
	_ = err // TODO: handle the error in this case.

	rootCommand.AddCommand(
		nova.NewPassthroughCmd(versions),
	)

	// Add the following commands to the rootCommand: start, tendermint, export, version, and rollback and wire multiplexer.
	server.AddCommandsWithStartCmdOptions(
		rootCommand,
		app.DefaultNodeHome,
		NewAppServer,
		appExporter,
		server.StartCmdOptions{
			AddFlags:            addStartFlags,
			StartCommandHandler: nova.New(versions),
		},
	)
}
