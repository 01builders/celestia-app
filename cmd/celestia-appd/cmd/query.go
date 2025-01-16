package cmd

import (
	"github.com/celestiaorg/celestia-app/v3/app"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/spf13/cobra"
)

func queryCommand() *cobra.Command {
	command := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	command.AddCommand(
		// FIXME: not sure what these should map to in 0.52
		// authcmd.GetAccountCmd(),
		// rpc.ValidatorCommand(),
		// rpc.BlockCommand(),
		authcmd.QueryTxsByEventsCmd(),
		authcmd.QueryTxCmd(),
	)

	app.ModuleBasics.AddQueryCommands(command)
	command.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return command
}
