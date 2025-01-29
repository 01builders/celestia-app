package ante_test

import (
	"testing"

	"github.com/celestiaorg/celestia-app/v4/app/ante"
	v2 "github.com/celestiaorg/celestia-app/v4/pkg/appconsts/v2"
	v3 "github.com/celestiaorg/celestia-app/v4/pkg/appconsts/v3"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestMaxTxSizeDecorator(t *testing.T) {
	testCases := []struct {
		name        string
		txSize      int
		expectError bool
		appVersion  uint64
		isCheckTx   []bool
	}{
		{
			name:        "good tx; under max tx size threshold",
			txSize:      v3.MaxTxSize - 1,
			appVersion:  v3.Version,
			expectError: false,
			isCheckTx:   []bool{true, false},
		},
		{
			name:        "bad tx; over max tx size threshold",
			txSize:      v3.MaxTxSize + 1,
			appVersion:  v3.Version,
			expectError: true,
			isCheckTx:   []bool{true, false},
		},
		{
			name:        "good tx; equal to max tx size threshold",
			txSize:      v3.MaxTxSize,
			appVersion:  v3.Version,
			expectError: false,
			isCheckTx:   []bool{true, false},
		},
		{
			name:        "good tx; limit only applies to v3 and above",
			txSize:      v3.MaxTxSize + 10,
			appVersion:  v2.Version,
			expectError: false,
			isCheckTx:   []bool{true, false},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, isCheckTx := range tc.isCheckTx {
				decorator := ante.NewMaxTxSizeDecorator(mockConsensusKeeper{appVersion: tc.appVersion})
				anteHandler := sdk.ChainAnteDecorators(decorator)

				ctx := sdk.NewContext(nil, isCheckTx, nil)

				txBytes := make([]byte, tc.txSize)

				ctx = ctx.WithTxBytes(txBytes)
				_, err := anteHandler(ctx, nil, false)
				if tc.expectError {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			}
		})
	}
}
