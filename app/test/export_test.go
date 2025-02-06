package app_test

import (
	"testing"

	"github.com/celestiaorg/celestia-app/v4/test/util"
	v1 "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	"github.com/stretchr/testify/require"
)

func TestExportAppStateAndValidators(t *testing.T) {
	forZeroHeight := true
	jailAllowedAddrs := []string{}
	testApp, _ := util.SetupTestApp(t)

	// advance one block
	_, _ = testApp.FinalizeBlock(&v1.FinalizeBlockRequest{})
	_, _ = testApp.Commit()

	exported, err := testApp.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs)
	require.NoError(t, err)
	require.NotNil(t, exported)
	require.Equal(t, uint64(4), exported.ConsensusParams.Version.App)
}
