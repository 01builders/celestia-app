package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
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

func TestMsgUpdateParams(t *testing.T) {
	tests := []struct {
		name                string
		authority           string
		newParams           types.Params
		expectedError       bool
		expectedMinGasPrice sdkmath.LegacyDec
	}{
		{
			name:      "valid update with default authority (gov)",
			authority: "",
			newParams: types.Params{
				NetworkMinGasPrice: sdkmath.LegacyMustNewDecFromStr("0.0005"),
			},
			expectedError:       false,
			expectedMinGasPrice: sdkmath.LegacyMustNewDecFromStr("0.0005"),
		},
		{
			name:      "valid update with correct authority",
			authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
			newParams: types.Params{
				NetworkMinGasPrice: sdkmath.LegacyMustNewDecFromStr("0.0005"),
			},
			expectedError:       false,
			expectedMinGasPrice: sdkmath.LegacyMustNewDecFromStr("0.0005"),
		},
		{
			name:      "invalid update with incorrect authority",
			authority: "invalid-authority",
			newParams: types.Params{
				NetworkMinGasPrice: sdkmath.LegacyMustNewDecFromStr("0.0005"),
			},
			expectedError:       true,
			expectedMinGasPrice: types.DefaultNetworkMinGasPrice, // should remain unchanged in case of error
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testApp, _, _ := testutil.NewTestAppWithGenesisSet(app.DefaultConsensusParams())
			msgServer := testApp.MinFeeKeeper
			sdkCtx := testApp.NewContext(false)
			if tc.authority == "" {
				tc.authority = testApp.MinFeeKeeper.GetAuthority()
			}

			// Create a message to update params
			msg := &types.MsgUpdateMinfeeParams{
				Authority: tc.authority,
				Params:    tc.newParams,
			}

			// Perform the update
			_, err := msgServer.UpdateMinfeeParams(sdkCtx, msg)

			if tc.expectedError {
				require.Error(t, err)
				params := testApp.MinFeeKeeper.GetParams(sdkCtx)
				require.Equal(t, tc.expectedMinGasPrice, params.NetworkMinGasPrice)
			} else {
				require.NoError(t, err)
				updatedParams := testApp.MinFeeKeeper.GetParams(sdkCtx)
				require.Equal(t, tc.expectedMinGasPrice, updatedParams.NetworkMinGasPrice)
			}
		})
	}
}
