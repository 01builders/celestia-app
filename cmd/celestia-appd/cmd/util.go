package cmd

import (
	"context"
	"io"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/store/snapshots"
	"cosmossdk.io/store/types"
	"github.com/celestiaorg/celestia-app/v3/server"
	v1 "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	"github.com/cometbft/cometbft/crypto"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/gogoproto/grpc"
)

var _ servertypes.Application = (*cmdApplication)(nil)

type cmdApplication struct {
	server.Application
}

// RegisterGRPCServerWithSkipCheckHeader implements types.Application.
func (c *cmdApplication) RegisterGRPCServerWithSkipCheckHeader(grpc.Server, bool) {
	panic("unimplemented")
}

func newCmdApplication(log log.Logger, db corestore.KVStoreWithBatch, w io.Writer, opts servertypes.AppOptions) servertypes.Application {
	return &cmdApplication{Application: NewAppServer(log, db, w, opts)}
}

// ApplySnapshotChunk implements types.Application.
// Subtle: this method shadows the method (Application).ApplySnapshotChunk of cmdApplication.Application.
func (c *cmdApplication) ApplySnapshotChunk(*v1.ApplySnapshotChunkRequest) (*v1.ApplySnapshotChunkResponse, error) {
	panic("unimplemented")
}

// CheckTx implements types.Application.
// Subtle: this method shadows the method (Application).CheckTx of cmdApplication.Application.
func (c *cmdApplication) CheckTx(*v1.CheckTxRequest) (*v1.CheckTxResponse, error) {
	panic("unimplemented")
}

// Close implements types.Application.
// Subtle: this method shadows the method (Application).Close of cmdApplication.Application.
func (c *cmdApplication) Close() error {
	return c.Close()
}

// Commit implements types.Application.
// Subtle: this method shadows the method (Application).Commit of cmdApplication.Application.
func (c *cmdApplication) Commit() (*v1.CommitResponse, error) {
	panic("unimplemented")
}

// CommitMultiStore implements types.Application.
// Subtle: this method shadows the method (Application).CommitMultiStore of cmdApplication.Application.
func (c *cmdApplication) CommitMultiStore() types.CommitMultiStore {
	return c.CommitMultiStore()
}

// ExtendVote implements types.Application.
func (c *cmdApplication) ExtendVote(context.Context, *v1.ExtendVoteRequest) (*v1.ExtendVoteResponse, error) {
	panic("unimplemented")
}

// FinalizeBlock implements types.Application.
func (c *cmdApplication) FinalizeBlock(*v1.FinalizeBlockRequest) (*v1.FinalizeBlockResponse, error) {
	panic("unimplemented")
}

// Info implements types.Application.
// Subtle: this method shadows the method (Application).Info of cmdApplication.Application.
func (c *cmdApplication) Info(*v1.InfoRequest) (*v1.InfoResponse, error) {
	panic("unimplemented")
}

// InitChain implements types.Application.
// Subtle: this method shadows the method (Application).InitChain of cmdApplication.Application.
func (c *cmdApplication) InitChain(*v1.InitChainRequest) (*v1.InitChainResponse, error) {
	panic("unimplemented")
}

// ListSnapshots implements types.Application.
// Subtle: this method shadows the method (Application).ListSnapshots of cmdApplication.Application.
func (c *cmdApplication) ListSnapshots(*v1.ListSnapshotsRequest) (*v1.ListSnapshotsResponse, error) {
	panic("unimplemented")
}

// LoadSnapshotChunk implements types.Application.
// Subtle: this method shadows the method (Application).LoadSnapshotChunk of cmdApplication.Application.
func (c *cmdApplication) LoadSnapshotChunk(*v1.LoadSnapshotChunkRequest) (*v1.LoadSnapshotChunkResponse, error) {
	panic("unimplemented")
}

// OfferSnapshot implements types.Application.
// Subtle: this method shadows the method (Application).OfferSnapshot of cmdApplication.Application.
func (c *cmdApplication) OfferSnapshot(*v1.OfferSnapshotRequest) (*v1.OfferSnapshotResponse, error) {
	panic("unimplemented")
}

// PrepareProposal implements types.Application.
// Subtle: this method shadows the method (Application).PrepareProposal of cmdApplication.Application.
func (c *cmdApplication) PrepareProposal(*v1.PrepareProposalRequest) (*v1.PrepareProposalResponse, error) {
	panic("unimplemented")
}

// ProcessProposal implements types.Application.
// Subtle: this method shadows the method (Application).ProcessProposal of cmdApplication.Application.
func (c *cmdApplication) ProcessProposal(*v1.ProcessProposalRequest) (*v1.ProcessProposalResponse, error) {
	panic("unimplemented")
}

// Query implements types.Application.
// Subtle: this method shadows the method (Application).Query of cmdApplication.Application.
func (c *cmdApplication) Query(context.Context, *v1.QueryRequest) (*v1.QueryResponse, error) {
	panic("unimplemented")
}

// RegisterAPIRoutes implements types.Application.
// Subtle: this method shadows the method (Application).RegisterAPIRoutes of cmdApplication.Application.
func (c *cmdApplication) RegisterAPIRoutes(*api.Server, config.APIConfig) {
	panic("unimplemented")
}

// RegisterGRPCServer implements types.Application.
// Subtle: this method shadows the method (Application).RegisterGRPCServer of cmdApplication.Application.
func (c *cmdApplication) RegisterGRPCServer(grpc.Server) {
	panic("unimplemented")
}

// RegisterNodeService implements types.Application.
// Subtle: this method shadows the method (Application).RegisterNodeService of cmdApplication.Application.
func (c *cmdApplication) RegisterNodeService(sdkclient.Context, config.Config) {
	panic("unimplemented")
}

// RegisterTendermintService implements types.Application.
// Subtle: this method shadows the method (Application).RegisterTendermintService of cmdApplication.Application.
func (c *cmdApplication) RegisterTendermintService(sdkclient.Context) {
	panic("unimplemented")
}

// RegisterTxService implements types.Application.
// Subtle: this method shadows the method (Application).RegisterTxService of cmdApplication.Application.
func (c *cmdApplication) RegisterTxService(sdkclient.Context) {
	panic("unimplemented")
}

// SnapshotManager implements types.Application.
// Subtle: this method shadows the method (Application).SnapshotManager of cmdApplication.Application.
func (c *cmdApplication) SnapshotManager() *snapshots.Manager {
	return c.SnapshotManager()
}

// ValidatorKeyProvider implements types.Application.
func (c *cmdApplication) ValidatorKeyProvider() func() (crypto.PrivKey, error) {
	panic("unimplemented")
}

// VerifyVoteExtension implements types.Application.
func (c *cmdApplication) VerifyVoteExtension(*v1.VerifyVoteExtensionRequest) (*v1.VerifyVoteExtensionResponse, error) {
	panic("unimplemented")
}
