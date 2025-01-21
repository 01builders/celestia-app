package server

import (
	"io"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/store/snapshots"
	storetypes "cosmossdk.io/store/types"
	cmtcrypto "github.com/cometbft/cometbft/crypto"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/gogoproto/grpc"
)

type Application interface {
	servertypes.ABCI

	RegisterAPIRoutes(*api.Server, config.APIConfig)

	// RegisterGRPCServerWithSkipCheckHeader registers gRPC services directly with the gRPC
	// server and bypass check header flag.
	RegisterGRPCServerWithSkipCheckHeader(grpc.Server, bool)

	// RegisterTxService registers the gRPC Query service for tx (such as tx
	// simulation, fetching txs by hash...).
	RegisterTxService(client.Context)

	// RegisterTendermintService registers the gRPC Query service for CometBFT queries.
	RegisterTendermintService(client.Context)

	// RegisterNodeService registers the node gRPC Query service.
	RegisterNodeService(client.Context, config.Config)

	// CommitMultiStore return the multistore instance
	CommitMultiStore() storetypes.CommitMultiStore

	// SnapshotManager return the snapshot manager
	SnapshotManager() *snapshots.Manager

	// ValidatorKeyProvider returns a function that generates a validator key
	ValidatorKeyProvider() func() (cmtcrypto.PrivKey, error)

	// Close is called in start cmd to gracefully cleanup resources.
	// Must be safe to be called multiple times.
	Close() error
}

// AppCreator is a function that allows us to lazily initialize an
// application using various configurations.
type AppCreator func(log.Logger, corestore.KVStoreWithBatch, io.Writer, servertypes.AppOptions) Application
