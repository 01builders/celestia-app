package random

import (
	"math/rand"
	"time"
)

// New creates a new random object with a random seed.
func New() *rand.Rand {
	seed := time.Now().UnixNano()
	return rand.New(rand.NewSource(seed))
}

// Bytes generates random bytes using math/rand.
func Bytes(r *rand.Rand, n int) []byte {
	bz := make([]byte, n)
	for i := range bz {
		bz[i] = byte(r.Intn(256)) // Random byte (0-255)
	}
	return bz
}
