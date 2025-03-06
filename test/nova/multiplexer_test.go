//go:build nova

package nova

import (
	"github.com/01builders/nova/appd"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNova(t *testing.T) {
	bz, err := appd.CelestiaApp()
	require.NoError(t, err)
	require.NotNil(t, bz)
	t.Logf("bz len: %d", len(bz))
}
