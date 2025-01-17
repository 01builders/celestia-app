package testnode

import (
	"time"

	"github.com/celestiaorg/celestia-app/v3/app"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// wrapEndBlocker overrides the app's endblocker to set the timeout commit to a
// different value for testnode.
func wrapEndBlocker(app *app.App, timeoutCommit time.Duration) func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	endBlocker := func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
		resp, err := app.EndBlocker(ctx, req)
		if err != nil {
			panic(err)
		}

		resp.Timeouts.TimeoutCommit = timeoutCommit
		return resp
	}

	return endBlocker
}
