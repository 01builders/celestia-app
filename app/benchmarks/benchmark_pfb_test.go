//go:build benchmarks

package benchmarks_test

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/log"
	"github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/go-square/v2/share"
	blobtx "github.com/celestiaorg/go-square/v2/tx"

	"github.com/celestiaorg/celestia-app/v4/app"
	"github.com/celestiaorg/celestia-app/v4/app/encoding"
	"github.com/celestiaorg/celestia-app/v4/pkg/user"
	testutil "github.com/celestiaorg/celestia-app/v4/test/util"
	"github.com/celestiaorg/celestia-app/v4/test/util/testfactory"
)

func init() {
	testutil.TestAppLogger = log.NewNopLogger()
}

func BenchmarkCheckTx_PFB_Multi(b *testing.B) {
	testCases := []struct {
		blobSize int
	}{
		{blobSize: 300},
		{blobSize: 500},
		{blobSize: 1000},
		{blobSize: 5000},
		{blobSize: 10_000},
		{blobSize: 50_000},
		{blobSize: 100_000},
		{blobSize: 200_000},
		{blobSize: 300_000},
		{blobSize: 400_000},
		{blobSize: 500_000},
		{blobSize: 1_000_000},
		{blobSize: 2_000_000},
		{blobSize: 3_000_000}, // once you get here you're over appconsts.MaxTxSize
		{blobSize: 4_000_000},
		{blobSize: 5_000_000},
		{blobSize: 6_000_000},
	}
	for _, testCase := range testCases {
		b.Run(fmt.Sprintf("%d bytes", testCase.blobSize), func(b *testing.B) {
			benchmarkCheckTxPFB(b, testCase.blobSize)
		})
	}
}

func benchmarkCheckTxPFB(b *testing.B, size int) {
	testApp, rawTxs := generatePayForBlobTransactions(b, 1, size)

	finalizeBlockResp, err := testApp.FinalizeBlock(&types.RequestFinalizeBlock{
		Time:   testutil.GenesisTime.Add(blockTime),
		Height: testApp.LastBlockHeight() + 1,
		Hash:   testApp.LastCommitID().Hash,
	})
	require.NotNil(b, finalizeBlockResp)
	require.NoError(b, err)

	commitResp, err := testApp.Commit()
	require.NotNil(b, commitResp)
	require.NoError(b, err)

	checkTxRequest := types.RequestCheckTx{
		Tx:   rawTxs[0],
		Type: types.CheckTxType_New,
	}

	b.ResetTimer()
	resp, err := testApp.CheckTx(&checkTxRequest)
	require.NoError(b, err)
	b.StopTimer()
	require.Equal(b, uint32(0), resp.Code)
	require.Equal(b, "", resp.Codespace)
	b.ReportMetric(float64(resp.GasUsed), "gas_used")
	b.ReportMetric(float64(len(rawTxs[0])), "transaction_size(byte)")
}

func BenchmarkFinalizeBlock_PFB_Multi(b *testing.B) {
	testCases := []struct {
		blobSize int
	}{
		{blobSize: 300},
		{blobSize: 500},
		{blobSize: 1000},
		{blobSize: 5000},
		{blobSize: 10_000},
		{blobSize: 50_000},
		{blobSize: 100_000},
		{blobSize: 200_000},
		{blobSize: 300_000},
		{blobSize: 400_000},
		{blobSize: 500_000},
		{blobSize: 1_000_000},
		{blobSize: 2_000_000},
		{blobSize: 3_000_000},
		{blobSize: 4_000_000},
		{blobSize: 5_000_000},
		{blobSize: 6_000_000},
	}
	for _, testCase := range testCases {
		b.Run(fmt.Sprintf("%d bytes", testCase.blobSize), func(b *testing.B) {
			benchmarkFinalizeBlockPFB(b, testCase.blobSize)
		})
	}
}

func benchmarkFinalizeBlockPFB(b *testing.B, size int) {
	testApp, rawTxs := generatePayForBlobTransactions(b, 1, size)

	blobTx, ok, err := blobtx.UnmarshalBlobTx(rawTxs[0])
	require.NoError(b, err)
	require.True(b, ok)

	finalizeBlockReq := types.RequestFinalizeBlock{
		Time:   testutil.GenesisTime.Add(blockTime),
		Height: testApp.LastBlockHeight() + 1,
		Hash:   testApp.LastCommitID().Hash,
		Txs:    [][]byte{blobTx.Tx},
	}

	b.ResetTimer()
	resp, err := testApp.FinalizeBlock(&finalizeBlockReq)
	require.NoError(b, err)
	b.StopTimer()
	require.Equal(b, uint32(0), resp.TxResults[0].Code)
	require.Equal(b, "", resp.TxResults[0].Codespace)
	b.ReportMetric(float64(resp.TxResults[0].GasUsed), "gas_used")
	b.ReportMetric(float64(len(rawTxs[0])), "transaction_size(byte)")
}

func BenchmarkPrepareProposal_PFB_Multi(b *testing.B) {
	testCases := []struct {
		numberOfTransactions, blobSize int
	}{
		{numberOfTransactions: 15_000, blobSize: 300},
		{numberOfTransactions: 10_000, blobSize: 500},
		{numberOfTransactions: 6_000, blobSize: 1000},
		{numberOfTransactions: 3_000, blobSize: 5000},
		{numberOfTransactions: 1_000, blobSize: 10_000},
		{numberOfTransactions: 500, blobSize: 50_000},
		{numberOfTransactions: 100, blobSize: 100_000},
		{numberOfTransactions: 100, blobSize: 200_000},
		{numberOfTransactions: 50, blobSize: 300_000},
		{numberOfTransactions: 50, blobSize: 400_000},
		{numberOfTransactions: 30, blobSize: 500_000},
		{numberOfTransactions: 10, blobSize: 1_000_000},
		{numberOfTransactions: 5, blobSize: 2_000_000},
		{numberOfTransactions: 3, blobSize: 3_000_000},
		{numberOfTransactions: 3, blobSize: 4_000_000},
		{numberOfTransactions: 2, blobSize: 5_000_000},
		{numberOfTransactions: 2, blobSize: 6_000_000},
	}
	for _, testCase := range testCases {
		b.Run(fmt.Sprintf("%d transactions of %d bytes", testCase.numberOfTransactions, testCase.blobSize), func(b *testing.B) {
			benchmarkPrepareProposalPFB(b, testCase.numberOfTransactions, testCase.blobSize)
		})
	}
}

func benchmarkPrepareProposalPFB(b *testing.B, count, size int) {
	testApp, rawTxs := generatePayForBlobTransactions(b, count, size)

	prepareProposalReq := types.RequestPrepareProposal{
		Txs:    rawTxs,
		Height: testApp.LastBlockHeight() + 1,
	}

	b.ResetTimer()
	prepareProposalResp, err := testApp.PrepareProposal(&prepareProposalReq)
	require.NoError(b, err)
	b.StopTimer()
	require.GreaterOrEqual(b, len(prepareProposalResp.Txs), 1)
	b.ReportMetric(float64(b.Elapsed().Nanoseconds()), "prepare_proposal_time(ns)")
	b.ReportMetric(float64(len(prepareProposalResp.Txs)), "number_of_transactions")
	b.ReportMetric(float64(len(rawTxs[0])), "transactions_size(byte)")
	b.ReportMetric(calculateBlockSizeInMb(prepareProposalResp.Txs), "block_size(mb)")
	b.ReportMetric(float64(calculateTotalGasUsed(testApp, rawTxs)), "total_gas_used")
}

func BenchmarkProcessProposal_PFB_Multi(b *testing.B) {
	testCases := []struct {
		numberOfTransactions, blobSize int
	}{
		{numberOfTransactions: 15_000, blobSize: 300},
		{numberOfTransactions: 10_000, blobSize: 500},
		{numberOfTransactions: 6_000, blobSize: 1000},
		{numberOfTransactions: 3_000, blobSize: 5000},
		{numberOfTransactions: 1_000, blobSize: 10_000},
		{numberOfTransactions: 500, blobSize: 50_000},
		{numberOfTransactions: 100, blobSize: 100_000},
		{numberOfTransactions: 100, blobSize: 200_000},
		{numberOfTransactions: 50, blobSize: 300_000},
		{numberOfTransactions: 50, blobSize: 400_000},
		{numberOfTransactions: 30, blobSize: 500_000},
		{numberOfTransactions: 10, blobSize: 1_000_000},
		{numberOfTransactions: 5, blobSize: 2_000_000},
		{numberOfTransactions: 3, blobSize: 3_000_000},
		{numberOfTransactions: 3, blobSize: 4_000_000},
		{numberOfTransactions: 2, blobSize: 5_000_000},
		{numberOfTransactions: 2, blobSize: 6_000_000},
	}
	for _, testCase := range testCases {
		b.Run(fmt.Sprintf("%d transactions of %d bytes", testCase.numberOfTransactions, testCase.blobSize), func(b *testing.B) {
			benchmarkProcessProposalPFB(b, testCase.numberOfTransactions, testCase.blobSize)
		})
	}
}

func benchmarkProcessProposalPFB(b *testing.B, count, size int) {
	testApp, rawTxs := generatePayForBlobTransactions(b, count, size)

	prepareProposalReq := types.RequestPrepareProposal{
		Txs:    rawTxs,
		Height: testApp.LastBlockHeight() + 1,
	}

	prepareProposalResp, err := testApp.PrepareProposal(&prepareProposalReq)
	require.NoError(b, err)
	require.GreaterOrEqual(b, len(prepareProposalResp.Txs), 1)

	processProposalReq := types.RequestProcessProposal{
		Txs:          prepareProposalResp.Txs,
		Height:       testApp.LastBlockHeight() + 1,
		DataRootHash: prepareProposalResp.DataRootHash,
		SquareSize:   prepareProposalResp.SquareSize,
	}

	b.ResetTimer()
	resp, err := testApp.ProcessProposal(&processProposalReq)
	require.NoError(b, err)
	b.StopTimer()
	require.Equal(b, types.ResponseProcessProposal_ACCEPT, resp.Status)

	b.ReportMetric(float64(b.Elapsed().Nanoseconds()), "process_proposal_time(ns)")
	b.ReportMetric(float64(len(prepareProposalResp.Txs)), "number_of_transactions")
	b.ReportMetric(float64(len(rawTxs[0])), "transactions_size(byte)")
	b.ReportMetric(calculateBlockSizeInMb(prepareProposalResp.Txs), "block_size(mb)")
	b.ReportMetric(float64(calculateTotalGasUsed(testApp, rawTxs)), "total_gas_used")
}

func BenchmarkProcessProposal_PFB_Half_Second(b *testing.B) {
	testCases := []struct {
		numberOfTransactions, blobSize int
	}{
		{numberOfTransactions: 11_000, blobSize: 50},
		{numberOfTransactions: 11_000, blobSize: 100},
		{numberOfTransactions: 11_000, blobSize: 200},
		{numberOfTransactions: 11_000, blobSize: 300},
		{numberOfTransactions: 11_000, blobSize: 400},
		{numberOfTransactions: 7000, blobSize: 500},
		{numberOfTransactions: 7000, blobSize: 600},
		{numberOfTransactions: 5000, blobSize: 1_000},
		{numberOfTransactions: 5000, blobSize: 1200},
		{numberOfTransactions: 5000, blobSize: 1500},
		{numberOfTransactions: 5000, blobSize: 1800},
		{numberOfTransactions: 5000, blobSize: 2000},
	}
	for _, testCase := range testCases {
		b.Run(fmt.Sprintf("%d transactions of %d bytes", testCase.numberOfTransactions, testCase.blobSize), func(b *testing.B) {
			benchmarkProcessProposalPFBHalfSecond(b, testCase.numberOfTransactions, testCase.blobSize)
		})
	}
}

func benchmarkProcessProposalPFBHalfSecond(b *testing.B, count, size int) {
	testApp, rawTxs := generatePayForBlobTransactions(b, count, size)

	targetTimeLowerBound := 0.499
	targetTimeUpperBound := 0.511

	start := 0
	end := count
	segment := end - start
	maxIterations := 100000
	iterations := 0
	for {
		iterations++
		if iterations >= maxIterations {
			b.Errorf("Maximum iterations reached without achieving target processing time")
			break
		}
		if segment == 1 {
			break
		}

		prepareProposalReq := types.RequestPrepareProposal{
			Txs:    rawTxs[start:end],
			Height: testApp.LastBlockHeight() + 1,
		}

		prepareProposalResp, err := testApp.PrepareProposal(&prepareProposalReq)
		require.NoError(b, err)
		require.GreaterOrEqual(b, len(prepareProposalResp.Txs), 1)

		processProposalReq := types.RequestProcessProposal{
			Txs:          prepareProposalResp.Txs,
			Height:       testApp.LastBlockHeight() + 1,
			DataRootHash: prepareProposalResp.DataRootHash,
			SquareSize:   prepareProposalResp.SquareSize,
		}

		startTime := time.Now()
		resp, err := testApp.ProcessProposal(&processProposalReq)
		require.NoError(b, err)
		endTime := time.Now()
		require.Equal(b, types.ResponseProcessProposal_ACCEPT, resp.Status)

		timeElapsed := float64(endTime.Sub(startTime).Nanoseconds()) / 1e9

		switch {
		case timeElapsed < targetTimeLowerBound:
			newEnd := end + segment/2
			if newEnd > len(rawTxs) {
				newEnd = len(rawTxs)
			}
			end = newEnd
			segment = end - start
			if segment <= 1 {
				break
			}
			continue
		case timeElapsed > targetTimeUpperBound:
			newEnd := end / 2
			if newEnd <= start {
				break
			}
			end = newEnd
			segment = end - start
			continue
		default:
			b.ReportMetric(
				timeElapsed,
				fmt.Sprintf(
					"processProposalTime(s)_%d_%d_%f",
					end-start,
					size,
					calculateBlockSizeInMb(prepareProposalResp.Txs[start:end]),
				),
			)
		}
		break
	}
}

// generatePayForBlobTransactions creates a test app then generates a number
// of valid PFB transactions.
func generatePayForBlobTransactions(b *testing.B, count, size int) (*app.App, [][]byte) {
	account := "test"
	testApp, kr := testutil.SetupTestAppWithGenesisValSetAndMaxSquareSize(app.DefaultConsensusParams(), 128, account)
	addr := testfactory.GetAddress(kr, account)
	enc := encoding.MakeTestConfig(app.ModuleEncodingRegisters...)
	acc := testutil.DirectQueryAccount(testApp, addr)
	accountSequence := acc.GetSequence()
	signer, err := user.NewSigner(kr, enc.TxConfig, testutil.ChainID, user.NewAccount(account, acc.GetAccountNumber(), acc.GetSequence()))
	require.NoError(b, err)

	rawTxs := make([][]byte, 0, count)
	randomBytes := crypto.CRandBytes(size)
	blob, err := share.NewBlob(share.RandomNamespace(), randomBytes, 1, acc.GetAddress().Bytes())
	require.NoError(b, err)
	for i := 0; i < count; i++ {
		tx, _, err := signer.CreatePayForBlobs(account, []*share.Blob{blob}, user.SetGasLimit(2549760000), user.SetFee(10000))
		require.NoError(b, err)
		rawTxs = append(rawTxs, tx)
		accountSequence++
		err = signer.SetSequence(account, accountSequence)
		require.NoError(b, err)
	}
	return testApp, rawTxs
}
