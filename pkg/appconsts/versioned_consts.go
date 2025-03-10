package appconsts

import (
	"strconv"
	"time"

	v4 "github.com/celestiaorg/celestia-app/v4/pkg/appconsts/v4"
)

const (
	LatestVersion = v4.Version
)

// SubtreeRootThreshold works as a target upper bound for the number of subtree
// roots in the share commitment. If a blob contains more shares than this
// number, then the height of the subtree roots will increase by one so that the
// number of subtree roots in the share commitment decreases by a factor of two.
// This step is repeated until the number of subtree roots is less than the
// SubtreeRootThreshold.
//
// The rationale for this value is described in more detail in ADR-013.
func SubtreeRootThreshold(_ uint64) int {
	return v4.SubtreeRootThreshold
}

// SquareSizeUpperBound imposes an upper bound on the max effective square size.
func SquareSizeUpperBound(_ uint64) int {
	if OverrideSquareSizeUpperBoundStr != "" {
		parsedValue, err := strconv.Atoi(OverrideSquareSizeUpperBoundStr)
		if err != nil {
			panic("Invalid OverrideSquareSizeUpperBoundStr value")
		}
		return parsedValue
	}
	return v4.SquareSizeUpperBound
}

func TxSizeCostPerByte(_ uint64) uint64 {
	return v4.TxSizeCostPerByte
}

func GasPerBlobByte(_ uint64) uint32 {
	return v4.GasPerBlobByte
}

func MaxTxSize(_ uint64) int {
	return v4.MaxTxSize
}

var (
	DefaultSubtreeRootThreshold = SubtreeRootThreshold(LatestVersion)
	DefaultSquareSizeUpperBound = SquareSizeUpperBound(LatestVersion)
	DefaultTxSizeCostPerByte    = TxSizeCostPerByte(LatestVersion)
	DefaultGasPerBlobByte       = GasPerBlobByte(LatestVersion)
)

func GetTimeoutCommit(v uint64) time.Duration {
	return v4.TimeoutCommit // TODO: remove this fn currently just used in tests, those tests should fail with this currently logic.
}

// UpgradeHeightDelay returns the delay in blocks after a quorum has been reached that the chain should upgrade to the new version.
func UpgradeHeightDelay(chainID string) int64 {
	if chainID == TestChainID {
		return 3
	}
	// TODO: this check is the same as v4, does it need to be special cased still or can we just return v4.UpgradeHeightDelay
	if chainID == ArabicaChainID {
		// ONLY ON ARABICA: don't return the v2 value even when the app version is
		// v2 on arabica. This is due to a bug that was shipped on arabica, where
		// the next version was used.
		return v4.UpgradeHeightDelay
	}
	return v4.UpgradeHeightDelay
}
