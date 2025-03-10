package ante_test

import (
	"testing"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cometbft/cometbft/proto/tendermint/version"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/go-square/v2/share"

	"github.com/celestiaorg/celestia-app/v4/app"
	"github.com/celestiaorg/celestia-app/v4/app/encoding"
	"github.com/celestiaorg/celestia-app/v4/pkg/appconsts"
	appv4 "github.com/celestiaorg/celestia-app/v4/pkg/appconsts/v4"
	ante "github.com/celestiaorg/celestia-app/v4/x/blob/ante"
	blob "github.com/celestiaorg/celestia-app/v4/x/blob/types"
)

const (
	testGasPerBlobByte   = 10
	testGovMaxSquareSize = 64
)

func TestPFBAnteHandler(t *testing.T) {
	enc := encoding.MakeTestConfig(app.ModuleEncodingRegisters...)

	testCases := []struct {
		name        string
		pfb         *blob.MsgPayForBlobs
		txGas       func(uint32) uint32
		gasConsumed uint64
		wantErr     bool
	}{
		{
			name: "valid pfb single blob",
			pfb: &blob.MsgPayForBlobs{
				// 1 share = 512 bytes = 5120 gas
				BlobSizes: []uint32{uint32(share.AvailableBytesFromSparseShares(1))},
			},
			txGas: func(testGasPerBlobByte uint32) uint32 {
				return share.ShareSize * testGasPerBlobByte
			},
			gasConsumed: 0,
			wantErr:     false,
		},
		{
			name: "valid pfb multi blob",
			pfb: &blob.MsgPayForBlobs{
				BlobSizes: []uint32{uint32(share.AvailableBytesFromSparseShares(1)), uint32(share.AvailableBytesFromSparseShares(2))},
			},
			txGas: func(testGasPerBlobByte uint32) uint32 {
				return 3 * share.ShareSize * testGasPerBlobByte
			},
			gasConsumed: 0,
			wantErr:     false,
		},
		{
			name: "pfb single blob not enough gas",
			pfb: &blob.MsgPayForBlobs{
				// 2 share = 1024 bytes = 10240 gas
				BlobSizes: []uint32{uint32(share.AvailableBytesFromSparseShares(1) + 1)},
			},
			txGas: func(testGasPerBlobByte uint32) uint32 {
				return 2*share.ShareSize*testGasPerBlobByte - 1
			},
			gasConsumed: 0,
			wantErr:     true,
		},
		{
			name: "pfb multi blob not enough gas",
			pfb: &blob.MsgPayForBlobs{
				BlobSizes: []uint32{uint32(share.AvailableBytesFromSparseShares(1)), uint32(share.AvailableBytesFromSparseShares(2))},
			},
			txGas: func(testGasPerBlobByte uint32) uint32 {
				return 3*share.ShareSize*testGasPerBlobByte - 1
			},
			gasConsumed: 0,
			wantErr:     true,
		},
		{
			name: "pfb with existing gas consumed",
			pfb: &blob.MsgPayForBlobs{
				// 1 share = 512 bytes = 5120 gas
				BlobSizes: []uint32{uint32(share.AvailableBytesFromSparseShares(1))},
			},
			txGas: func(testGasPerBlobByte uint32) uint32 {
				return share.ShareSize*testGasPerBlobByte + 10000 - 1
			},
			gasConsumed: 10000,
			wantErr:     true,
		},
		{
			name: "valid pfb with existing gas consumed",
			pfb: &blob.MsgPayForBlobs{
				// 1 share = 512 bytes = 5120 gas
				BlobSizes: []uint32{uint32(share.AvailableBytesFromSparseShares(10))},
			},
			txGas: func(_ uint32) uint32 {
				return 1000000
			},
			gasConsumed: 10000,
			wantErr:     false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			anteHandler := ante.NewMinGasPFBDecorator(mockBlobKeeper{})
			header := tmproto.Header{
				Version: version.Consensus{
					App: appconsts.LatestVersion,
				},
			}
			ctx := sdk.NewContext(nil, header, true, log.NewNopLogger()).
				WithGasMeter(storetypes.NewGasMeter(uint64(tc.txGas(appv4.GasPerBlobByte)))).
				WithIsCheckTx(true)

			ctx.GasMeter().ConsumeGas(tc.gasConsumed, "test")
			txBuilder := enc.TxConfig.NewTxBuilder()
			require.NoError(t, txBuilder.SetMsgs(tc.pfb))
			tx := txBuilder.GetTx()
			_, err := anteHandler.AnteHandle(ctx, tx, false, func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) { return ctx, nil })
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

type mockBlobKeeper struct{}

func (mockBlobKeeper) GasPerBlobByte(_ sdk.Context) uint32 {
	return testGasPerBlobByte
}

func (mockBlobKeeper) GovMaxSquareSize(_ sdk.Context) uint64 {
	return testGovMaxSquareSize
}
