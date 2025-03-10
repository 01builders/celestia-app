package v1

const (
	Version uint64 = 1
	// UpgradeHeightDelay is deprecated because v1 does not contain the signal
	// module so this constant should not be used.
	UpgradeHeightDelay = int64(7 * 24 * 60 * 60 / 12) // 7 days * 24 hours * 60 minutes * 60 seconds / 12 seconds per block = 50,400 blocks.
)
