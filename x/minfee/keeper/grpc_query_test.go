package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/celestia-app/v4/app"
	"github.com/celestiaorg/celestia-app/v4/pkg/appconsts"
	testutil "github.com/celestiaorg/celestia-app/v4/test/util"
	"github.com/celestiaorg/celestia-app/v4/x/minfee/types"
)

func TestQueryNetworkMinGasPrice(t *testing.T) {
	testApp, _, _ := testutil.NewTestAppWithGenesisSet(app.DefaultConsensusParams())
	queryServer := testApp.MinFeeKeeper
	sdkCtx := testApp.NewContext(false)

	// Perform a query for the network minimum gas price
	resp, err := queryServer.NetworkMinGasPrice(sdkCtx, &types.QueryNetworkMinGasPrice{})
	require.NoError(t, err)

	// Check the response
	require.Equal(t, appconsts.DefaultNetworkMinGasPrice, resp.NetworkMinGasPrice.MustFloat64())
}

func TestQueryParams(t *testing.T) {
	testApp, _, _ := testutil.NewTestAppWithGenesisSet(app.DefaultConsensusParams())
	queryServer := testApp.MinFeeKeeper
	sdkCtx := testApp.NewContext(false)

	// Perform a query for the params
	resp, err := queryServer.Params(sdkCtx, &types.QueryParamsRequest{})
	require.NoError(t, err)

	// Check the response
	require.NotNil(t, resp)
	require.Equal(t, testApp.MinFeeKeeper.GetParams(sdkCtx), resp.Params)
}
