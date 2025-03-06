//go:build nova

package nova

import (
	"bytes"
	"github.com/01builders/nova/appd"
	"github.com/stretchr/testify/require"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCelestiaAppBinaryIsAvailable(t *testing.T) {
	bz, err := appd.CelestiaApp()
	require.NoError(t, err)
	require.NotNil(t, bz)
	t.Logf("bz len: %d", len(bz))
}

func TestMultiplexerSetup(t *testing.T) {
	// Find the celestia-appd binary
	celestiaBin := "celestia-appd"
	// Get the Celestia home directory
	celestiaHome := execCommand(t, celestiaBin, "config", "home")

	// Set up Celestia config
	execCommand(t, celestiaBin, "config", "set", "client", "chain-id", "local_devnet")
	execCommand(t, celestiaBin, "config", "set", "client", "keyring-backend", "test")
	execCommand(t, celestiaBin, "config", "set", "app", "api.enable", "true")

	// Add Alice's key
	execCommand(t, celestiaBin, "keys", "add", "alice")

	genesisPath := getTestFilePath("multi-plexer-genesis.json")

	targetGenesisPath := filepath.Join(celestiaHome, "config", "genesis.json")
	require.NoError(t, copyFile(genesisPath, targetGenesisPath), "failed to copy genesis file")

	execCommand(t, celestiaBin, "passthrough", "v3", "add-genesis-account", "alice", "5000000000utia", "--keyring-backend", "test")
	execCommand(t, celestiaBin, "passthrough", "v3", "gentx", "alice", "1000000utia", "--chain-id", "local_devnet")
	execCommand(t, celestiaBin, "passthrough", "v3", "collect-gentxs")
}

// execCommand runs a command and returns stdout/stderr.
func execCommand(t *testing.T, cmd string, args ...string) string {
	t.Helper()
	var out bytes.Buffer
	command := exec.Command(cmd, args...)
	command.Stdout = &out
	command.Stderr = &out
	err := command.Run()
	require.NoError(t, err, "command failed: %s\nOutput: %s", cmd, out.String())
	return out.String()
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0644)
}

// getTestFilePath constructs the absolute path for test data.
func getTestFilePath(filename string) string {
	return filepath.Join("testdata", filename)
}
