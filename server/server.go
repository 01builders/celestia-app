package server

import (
	"io"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/store/snapshots"
	store "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	gogogrpc "github.com/cosmos/gogoproto/grpc"
	abci "github.com/tendermint/tendermint/abci/types"
)

type Application interface {
	abci.Application

	RegisterAPIRoutes(*api.Server, config.APIConfig)

	// RegisterGRPCServer registers gRPC services directly with the gRPC
	// server.
	RegisterGRPCServer(gogogrpc.Server)

	RegisterGRPCServerWithSkipCheckHeader(gogogrpc.Server, bool)

	// RegisterTxService registers the gRPC Query service for tx (such as tx
	// simulation, fetching txs by hash...).
	RegisterTxService(client.Context)

	RegisterNodeService(client.Context, config.Config)

	// RegisterTendermintService registers the gRPC Query service for tendermint queries.
	RegisterTendermintService(client.Context)

	// Return the multistore instance
	CommitMultiStore() store.CommitMultiStore

	// Return the snapshot manager
	SnapshotManager() *snapshots.Manager

	// Close is called in start cmd to gracefully cleanup resources.
	Close() error
}

// AppCreator is a function that allows us to lazily initialize an
// application using various configurations.
type AppCreator func(log.Logger, corestore.KVStoreWithBatch, io.Writer, servertypes.AppOptions) Application
