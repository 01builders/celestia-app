//go:build bench_abci_methods

package benchmarks_test

import (
	"fmt"
	"testing"
	"time"

	banktypes "cosmossdk.io/x/bank/types"
	types "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/celestia-app/v4/app"
	"github.com/celestiaorg/celestia-app/v4/app/encoding"
	"github.com/celestiaorg/celestia-app/v4/pkg/appconsts"
	"github.com/celestiaorg/celestia-app/v4/pkg/user"
	testutil "github.com/celestiaorg/celestia-app/v4/test/util"
	"github.com/celestiaorg/celestia-app/v4/test/util/testfactory"
	"github.com/celestiaorg/celestia-app/v4/test/util/testnode"
)

func BenchmarkCheckTx_MsgSend_1(b *testing.B) {
	testApp, rawTxs := generateMsgSendTransactions(b, 1)
	testApp.Commit()

	checkTxRequest := types.CheckTxRequest{
		Tx:   rawTxs[0],
		Type: types.CHECK_TX_TYPE_CHECK,
	}

	b.ResetTimer()
	resp, err := testApp.CheckTx(&checkTxRequest)
	require.NoError(b, err)
	b.StopTimer()
	require.Equal(b, uint32(0), resp.Code)
	require.Equal(b, "", resp.Codespace)
	b.ReportMetric(float64(resp.GasUsed), "gas_used")
}

func BenchmarkCheckTx_MsgSend_8MB(b *testing.B) {
	testApp, rawTxs := generateMsgSendTransactions(b, 31645)
	testApp.Commit()

	var totalGas int64
	b.ResetTimer()
	for _, tx := range rawTxs {
		checkTxRequest := types.CheckTxRequest{
			Tx:   tx,
			Type: types.CHECK_TX_TYPE_CHECK,
		}
		b.StartTimer()
		resp, err := testApp.CheckTx(&checkTxRequest)
		require.NoError(b, err)
		b.StopTimer()
		require.Equal(b, uint32(0), resp.Code)
		require.Equal(b, "", resp.Codespace)
		totalGas += resp.GasUsed
	}

	b.StopTimer()
	b.ReportMetric(float64(totalGas), "total_gas_used")
}

func BenchmarkDeliverTx_MsgSend_1(b *testing.B) {
	testApp, rawTxs := generateMsgSendTransactions(b, 1)

	deliverTxRequest := types.FinalizeBlockRequest{
		Txs: [][]byte{rawTxs[0]},
	}

	b.ResetTimer()
	resp, err := testApp.FinalizeBlock(&deliverTxRequest)
	require.NoError(b, err)
	b.StopTimer()
	require.Equal(b, uint32(0), resp.TxResults[0].Code)
	require.Equal(b, "", resp.TxResults[0].Codespace)
	b.ReportMetric(float64(resp.TxResults[0].GasUsed), "gas_used")
}

func BenchmarkDeliverTx_MsgSend_8MB(b *testing.B) {
	testApp, rawTxs := generateMsgSendTransactions(b, 31645)

	var totalGas int64
	b.ResetTimer()
	for _, tx := range rawTxs {
		deliverTxRequest := types.FinalizeBlockRequest{
			Txs: [][]byte{tx},
		}
		b.StartTimer()
		resp, err := testApp.FinalizeBlock(&deliverTxRequest)
		require.NoError(b, err)
		b.StopTimer()
		require.Equal(b, uint32(0), resp.TxResults[0].Code)
		require.Equal(b, "", resp.TxResults[0].Codespace)
		totalGas += resp.TxResults[0].GasUsed
	}
	b.StopTimer()
	b.ReportMetric(float64(totalGas), "total_gas_used")
}

func BenchmarkPrepareProposal_MsgSend_1(b *testing.B) {
	testApp, rawTxs := generateMsgSendTransactions(b, 1)

	prepareProposalRequest := types.PrepareProposalRequest{
		Txs:    rawTxs,
		Height: 10,
	}

	b.ResetTimer()
	resp, err := testApp.PrepareProposal(&prepareProposalRequest)
	require.NoError(b, err)
	b.StopTimer()
	require.GreaterOrEqual(b, len(resp.Txs), 1)
	b.ReportMetric(float64(calculateTotalGasUsed(testApp, resp.Txs)), "total_gas_used")
}

func BenchmarkPrepareProposal_MsgSend_8MB(b *testing.B) {
	// a full 8mb block equals to around 31645 msg send transactions.
	// using 31645 to let prepare proposal choose the maximum
	testApp, rawTxs := generateMsgSendTransactions(b, 31645)

	prepareProposalRequest := types.PrepareProposalRequest{
		Txs:    rawTxs,
		Height: 10,
	}

	b.ResetTimer()
	resp, err := testApp.PrepareProposal(&prepareProposalRequest)
	require.NoError(b, err)
	b.StopTimer()
	require.GreaterOrEqual(b, len(resp.Txs), 1)
	b.ReportMetric(float64(len(resp.Txs)), "number_of_transactions")
	b.ReportMetric(calculateBlockSizeInMb(resp.Txs), "block_size(mb)")
	b.ReportMetric(float64(calculateTotalGasUsed(testApp, resp.Txs)), "total_gas_used")
}

func BenchmarkProcessProposal_MsgSend_1(b *testing.B) {
	testApp, rawTxs := generateMsgSendTransactions(b, 1)

	prepareProposalRequest := types.PrepareProposalRequest{
		Txs:    rawTxs,
		Height: 10,
	}
	resp, err := testApp.PrepareProposal(&prepareProposalRequest)
	require.NoError(b, err)
	require.GreaterOrEqual(b, len(resp.Txs), 1)

	processProposalRequest := types.ProcessProposalRequest{
		Txs:    resp.Txs,
		Height: 1,
	}

	b.ResetTimer()
	respProcessProposal, err := testApp.ProcessProposal(&processProposalRequest)
	require.NoError(b, err)
	b.StopTimer()
	require.Equal(b, types.PROCESS_PROPOSAL_STATUS_ACCEPT, respProcessProposal.Status)

	// b.ReportMetric(float64(calculateTotalGasUsed(testApp, respProcessProposal.Txs)), "total_gas_used") //TODO: do we need to return this
}

func BenchmarkProcessProposal_MsgSend_8MB(b *testing.B) {
	// a full 8mb block equals to around 31645 msg send transactions.
	// using 31645 to let prepare proposal choose the maximum
	testApp, rawTxs := generateMsgSendTransactions(b, 31645)

	prepareProposalRequest := types.PrepareProposalRequest{
		Txs:    rawTxs,
		Height: 10,
	}
	resp, err := testApp.PrepareProposal(&prepareProposalRequest)
	require.NoError(b, err)
	require.GreaterOrEqual(b, len(resp.Txs), 1)

	b.ReportMetric(float64(len(resp.Txs)), "number_of_transactions")
	b.ReportMetric(calculateBlockSizeInMb(resp.Txs), "block_size_(mb)")
	b.ReportMetric(float64(calculateTotalGasUsed(testApp, resp.Txs)), "total_gas_used")

	processProposalRequest := types.ProcessProposalRequest{
		Txs:    resp.Txs,
		Height: 10,
	}

	b.ResetTimer()
	respProcessProposal, err := testApp.ProcessProposal(&processProposalRequest)
	require.NoError(b, err)
	b.StopTimer()
	require.Equal(b, types.PROCESS_PROPOSAL_STATUS_ACCEPT, respProcessProposal.Status)

	// b.ReportMetric(float64(calculateTotalGasUsed(testApp, respProcessProposal.Txs)), "total_gas_used") //TODO: do we need to return this
}

func BenchmarkProcessProposal_MsgSend_8MB_Find_Half_Sec(b *testing.B) {
	targetTimeLowerBound := 0.499
	targetTimeUpperBound := 0.511
	numberOfTransaction := 5500
	testApp, rawTxs := generateMsgSendTransactions(b, numberOfTransaction)
	start := 0
	end := numberOfTransaction
	segment := end - start
	for {
		if segment == 1 {
			break
		}

		prepareProposalRequest := types.PrepareProposalRequest{
			Txs:    rawTxs[start:end],
			Height: 10,
		}
		resp, err := testApp.PrepareProposal(&prepareProposalRequest)
		require.NoError(b, err)
		require.GreaterOrEqual(b, len(resp.Txs), 1)

		processProposalRequest := types.ProcessProposalRequest{
			Txs:    resp.Txs,
			Height: 10,
		}

		startTime := time.Now()
		respProcessProposal, err := testApp.ProcessProposal(&processProposalRequest)
		require.NoError(b, err)
		endTime := time.Now()
		require.Equal(b, types.PROCESS_PROPOSAL_STATUS_ACCEPT, respProcessProposal.Status)

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
			b.ReportMetric(timeElapsed, fmt.Sprintf("elapsedTime(s)_%d", end-start))
		}
		break
	}
}

// generateMsgSendTransactions creates a test app then generates a number
// of valid msg send transactions.
func generateMsgSendTransactions(b *testing.B, count int) (*app.App, [][]byte) {
	account := "test"
	testApp, kr := testutil.SetupTestAppWithGenesisValSetAndMaxSquareSize(app.DefaultConsensusParams(), 128, account)
	addr := testfactory.GetAddress(kr, account)
	enc := encoding.MakeConfig()
	acc := testutil.DirectQueryAccount(testApp, addr)
	signer, err := user.NewSigner(kr, enc.TxConfig, testutil.ChainID, appconsts.LatestVersion, user.NewAccount(account, acc.GetAccountNumber(), acc.GetSequence()))
	require.NoError(b, err)
	rawTxs := make([][]byte, 0, count)
	for i := 0; i < count; i++ {
		msg := banktypes.NewMsgSend(
			addr.String(),
			testnode.RandomAddress().(sdk.AccAddress).String(),
			sdk.NewCoins(sdk.NewInt64Coin(appconsts.BondDenom, 10)),
		)
		rawTx, _, err := signer.CreateTx([]sdk.Msg{msg}, user.SetGasLimit(1000000), user.SetFee(10))
		require.NoError(b, err)
		rawTxs = append(rawTxs, rawTx)
		err = signer.IncrementSequence(account)
		require.NoError(b, err)
	}
	return testApp, rawTxs
}

// mebibyte the number of bytes in a mebibyte
const mebibyte = 1048576

// calculateBlockSizeInMb returns the block size in mb given a set
// of raw transactions.
func calculateBlockSizeInMb(txs [][]byte) float64 {
	numberOfBytes := 0
	for _, tx := range txs {
		numberOfBytes += len(tx)
	}
	mb := float64(numberOfBytes) / mebibyte
	return mb
}

// calculateTotalGasUsed simulates the provided transactions and returns the
// total gas used by all of them
func calculateTotalGasUsed(testApp *app.App, txs [][]byte) uint64 {
	var totalGas uint64
	for _, tx := range txs {
		gasInfo, _, _ := testApp.Simulate(tx)
		totalGas += gasInfo.GasUsed
	}
	return totalGas
}
