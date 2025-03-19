package appconsts

import (
	"time"

	v4 "github.com/celestiaorg/celestia-app/v4/pkg/appconsts/v4"
)

const (
	LatestVersion = v4.Version
)

var (
	DefaultSubtreeRootThreshold = v4.SubtreeRootThreshold
	DefaultSquareSizeUpperBound = v4.SquareSizeUpperBound
	DefaultTxSizeCostPerByte    = v4.TxSizeCostPerByte
	DefaultGasPerBlobByte       = v4.GasPerBlobByte
	DefaultVersion              = v4.Version
	DefaultTimeoutCommit        = v4.TimeoutCommit
	DefaultUpgradeHeightDelay   = v4.UpgradeHeightDelay
	DefaultMaxTxSize            = v4.MaxTxSize
)

func GetTimeoutCommit(_ uint64) time.Duration {
	return v4.TimeoutCommit // TODO: remove this fn currently just used in tests, those tests should fail with this currently logic.
}

// UpgradeHeightDelay returns the delay in blocks after a quorum has been reached that the chain should upgrade to the new version.
func UpgradeHeightDelay(chainID string) int64 {
	if chainID == TestChainID || chainID == Test2ChainID {
		return 3
	}
	return v4.UpgradeHeightDelay
}
