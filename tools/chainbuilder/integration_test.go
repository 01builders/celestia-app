package main

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"cosmossdk.io/log"
	tmrand "cosmossdk.io/math/unsafe"
	"github.com/celestiaorg/celestia-app/v4/app"
	"github.com/celestiaorg/celestia-app/v4/app/encoding"
	"github.com/celestiaorg/celestia-app/v4/pkg/appconsts"
	"github.com/celestiaorg/celestia-app/v4/test/util/testnode"
	tmconfig "github.com/cometbft/cometbft/config"
	tmlog "github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	"github.com/cometbft/cometbft/rpc/client/local"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/server"

	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping chainbuilder tool test")
	}

	numBlocks := 10

	cfg := BuilderConfig{
		NumBlocks:     numBlocks,
		BlockSize:     appconsts.DefaultMaxBytes,
		BlockInterval: time.Second,
		ChainID:       tmrand.Str(6),
		Namespace:     defaultNamespace,
	}

	dir := t.TempDir()

	// First run
	err := Run(context.Background(), cfg, dir)
	require.NoError(t, err)

	// Second run with existing directory
	cfg.ExistingDir = filepath.Join(dir, fmt.Sprintf("testnode-%s", cfg.ChainID))
	err = Run(context.Background(), cfg, dir)
	require.NoError(t, err)

	tmCfg := testnode.DefaultTendermintConfig()
	tmCfg.SetRoot(cfg.ExistingDir)

	appDB, err := dbm.NewDB("application", dbm.GoLevelDBBackend, tmCfg.DBDir())
	require.NoError(t, err)

	encCfg := encoding.MakeConfig(app.ModuleBasics)

	app := app.New(
		log.NewNopLogger(),
		appDB,
		nil,
		0,
		encCfg,
		0, // upgrade height v2
		0, // timeout commit
		baseapp.SetMinGasPrices(fmt.Sprintf("%f%s", appconsts.DefaultMinGasPrice, appconsts.BondDenom)),
	)

	nodeKey, err := p2p.LoadNodeKey(tmCfg.NodeKeyFile())
	require.NoError(t, err)

	prival, err := privval.LoadOrGenFilePV(tmCfg.PrivValidatorKeyFile(), tmCfg.PrivValidatorStateFile(), app.ValidatorKeyProvider())
	require.NoError(t, err)

	cmtApp := server.NewCometABCIWrapper(app)
	cometNode, err := node.NewNode(
		context.TODO(),
		tmCfg,
		prival,
		nodeKey,
		proxy.NewLocalClientCreator(cmtApp),
		node.DefaultGenesisDocProviderFunc(tmCfg),
		tmconfig.DefaultDBProvider,
		node.DefaultMetricsProvider(tmCfg.Instrumentation),
		tmlog.NewNopLogger(),
	)
	require.NoError(t, err)

	require.NoError(t, cometNode.Start())
	defer func() { _ = cometNode.Stop() }()

	client := local.New(cometNode)
	status, err := client.Status(context.Background())
	require.NoError(t, err)
	require.NotNil(t, status)
	// assert that the new node eventually makes progress in the chain
	require.Eventually(t, func() bool {
		status, err := client.Status(context.Background())
		require.NoError(t, err)
		return status.SyncInfo.LatestBlockHeight >= int64(numBlocks*2)
	}, time.Second*10, time.Millisecond*100)
	require.NoError(t, cometNode.Stop())
	cometNode.Wait()
}
