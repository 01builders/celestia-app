package keeper_test

import (
	"github.com/celestiaorg/celestia-app/v4/pkg/appconsts"
	"github.com/celestiaorg/celestia-app/v4/x/blob/keeper"
	"testing"

	blobtypes "github.com/celestiaorg/celestia-app/v4/x/blob/types"
	"github.com/stretchr/testify/require"
)

func TestMigrateParams(t *testing.T) {
	tests := []struct {
		name           string
		expectedParams blobtypes.Params
	}{
		{
			name:           "success",
			expectedParams: blobtypes.DefaultParams(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, _, ctx := CreateKeeper(t, appconsts.LatestVersion)

			err := keeper.MigrateParams(ctx, *k)
			require.NoError(t, err)
			require.Equal(t, tt.expectedParams, k.GetParams(ctx))
		})
	}
}
