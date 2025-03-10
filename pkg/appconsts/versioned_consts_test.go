package appconsts_test

import (
	appv4 "github.com/celestiaorg/celestia-app/v4/pkg/appconsts/v4"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/celestia-app/v4/pkg/appconsts"
	v3 "github.com/celestiaorg/celestia-app/v4/pkg/appconsts/v3"
)

func TestUpgradeHeightDelay(t *testing.T) {
	tests := []struct {
		name                       string
		chainID                    string
		expectedUpgradeHeightDelay int64
	}{
		{
			name:                       "v2 upgrade delay on arabica",
			chainID:                    "arabica-11",
			expectedUpgradeHeightDelay: v3.UpgradeHeightDelay, // falls back to v3 because of arabica bug
		},
		{
			name:                       "the upgrade delay for chainID 'test' should be 3",
			chainID:                    appconsts.TestChainID,
			expectedUpgradeHeightDelay: 3,
		},

		{
			name:                       "the upgrade delay should be latest value",
			chainID:                    "arabica-11",
			expectedUpgradeHeightDelay: appv4.UpgradeHeightDelay,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := appconsts.UpgradeHeightDelay(tc.chainID)
			require.Equal(t, tc.expectedUpgradeHeightDelay, actual)
		})
	}
}
