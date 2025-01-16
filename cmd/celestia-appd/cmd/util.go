package cmd

import (
	"context"
	"fmt"
	"io"
	"reflect"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/store/snapshots"
	"cosmossdk.io/store/types"
	"github.com/celestiaorg/celestia-app/v3/server"
	v1 "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/libs/bytes"
	"github.com/cometbft/cometbft/rpc/client"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	comettypes "github.com/cometbft/cometbft/types"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/gogoproto/grpc"
	tmlog "github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/rpc/client/local"
)

// deepClone deep clones the given object using reflection.
// If the structs differ in any way an error is returned.
func DeepClone(src, dst interface{}) error {
	return deepClone(reflect.ValueOf(src), reflect.ValueOf(dst))
}

func deepClone(src, dst reflect.Value) error {
	switch {
	case src.Kind() == reflect.Ptr && dst.Kind() == reflect.Ptr:
		return deepClone(src.Elem(), dst.Elem())
	case src.Kind() == reflect.Struct && dst.Kind() == reflect.Struct:
		if !dst.IsValid() {
			dst = reflect.New(src.Type())
		}
	default:
		return fmt.Errorf("kind mismatch: %s != %s", src.Kind(), dst.Kind())
	}

	for i := 0; i < src.NumField(); i++ {
		srcField := src.Field(i)
		srcFieldType := src.Type().Field(i)
		dstField := dst.Field(i)
		dstFieldType := dst.Type().Field(i)
		if srcFieldType.Name != dstFieldType.Name {
			return fmt.Errorf("field name mismatch: %s != %s", srcFieldType.Name, dstFieldType.Name)
		}
		if srcFieldType.Type != dstFieldType.Type {
			return fmt.Errorf("field type mismatch: %s != %s", srcFieldType.Type, dstFieldType.Type)
		}

		if srcField.Kind() == reflect.Struct || srcField.Kind() == reflect.Ptr {
			if err := deepClone(srcField, dstField); err != nil {
				return err
			}
		} else {
			dstField.Set(srcField)
		}
	}

	return nil
}

// logWrapperCoreToTM wraps cosmossdk.io/log.Logger to implement tendermint/libs/log.Logger.
type logWrapperCoreToTM struct {
	log.Logger
}

func (w *logWrapperCoreToTM) With(keyvals ...interface{}) tmlog.Logger {
	return &logWrapperCoreToTM{Logger: w.Logger.With(keyvals...)}
}

var _ log.Logger = (*logWrapperTmToCore)(nil)

type logWrapperTmToCore struct {
	tmlog.Logger
}

func (l *logWrapperTmToCore) Impl() any {
	panic("unimplemented")
}

func (l *logWrapperTmToCore) Warn(msg string, keyVals ...any) {
	panic("unimplemented")
}

func (w *logWrapperTmToCore) With(keyvals ...interface{}) log.Logger {
	return &logWrapperTmToCore{Logger: w.Logger.With(keyvals...)}
}

var _ sdkclient.CometRPC = (*tmLocalWrapper)(nil)

// tmLocalWrapper wraps a tendermint/rpc/client/local.Local to implement github.com/cosmos/cosmos-sdk/client.CometRPC.
type tmLocalWrapper struct {
	*local.Local
}

// ABCIInfo implements client.CometRPC.
// Subtle: this method shadows the method (*Local).ABCIInfo of tmLocalWrapper.Local.
func (t *tmLocalWrapper) ABCIInfo(ctx context.Context) (*coretypes.ResultABCIInfo, error) {
	panic("unimplemented")
}

// ABCIQuery implements client.CometRPC.
// Subtle: this method shadows the method (*Local).ABCIQuery of tmLocalWrapper.Local.
func (t *tmLocalWrapper) ABCIQuery(ctx context.Context, path string, data bytes.HexBytes) (*coretypes.ResultABCIQuery, error) {
	panic("unimplemented")
}

// ABCIQueryWithOptions implements client.CometRPC.
// Subtle: this method shadows the method (*Local).ABCIQueryWithOptions of tmLocalWrapper.Local.
func (t *tmLocalWrapper) ABCIQueryWithOptions(ctx context.Context, path string, data bytes.HexBytes, opts client.ABCIQueryOptions) (*coretypes.ResultABCIQuery, error) {
	panic("unimplemented")
}

// Block implements client.CometRPC.
// Subtle: this method shadows the method (*Local).Block of tmLocalWrapper.Local.
func (t *tmLocalWrapper) Block(ctx context.Context, height *int64) (*coretypes.ResultBlock, error) {
	panic("unimplemented")
}

// BlockByHash implements client.CometRPC.
// Subtle: this method shadows the method (*Local).BlockByHash of tmLocalWrapper.Local.
func (t *tmLocalWrapper) BlockByHash(ctx context.Context, hash []byte) (*coretypes.ResultBlock, error) {
	panic("unimplemented")
}

// BlockResults implements client.CometRPC.
// Subtle: this method shadows the method (*Local).BlockResults of tmLocalWrapper.Local.
func (t *tmLocalWrapper) BlockResults(ctx context.Context, height *int64) (*coretypes.ResultBlockResults, error) {
	panic("unimplemented")
}

// BlockSearch implements client.CometRPC.
// Subtle: this method shadows the method (*Local).BlockSearch of tmLocalWrapper.Local.
func (t *tmLocalWrapper) BlockSearch(ctx context.Context, query string, page *int, perPage *int, orderBy string) (*coretypes.ResultBlockSearch, error) {
	panic("unimplemented")
}

// BlockchainInfo implements client.CometRPC.
// Subtle: this method shadows the method (*Local).BlockchainInfo of tmLocalWrapper.Local.
func (t *tmLocalWrapper) BlockchainInfo(ctx context.Context, minHeight int64, maxHeight int64) (*coretypes.ResultBlockchainInfo, error) {
	panic("unimplemented")
}

// BroadcastTxAsync implements client.CometRPC.
// Subtle: this method shadows the method (*Local).BroadcastTxAsync of tmLocalWrapper.Local.
func (t *tmLocalWrapper) BroadcastTxAsync(ctx context.Context, tx comettypes.Tx) (*coretypes.ResultBroadcastTx, error) {
	panic("unimplemented")
}

// BroadcastTxCommit implements client.CometRPC.
// Subtle: this method shadows the method (*Local).BroadcastTxCommit of tmLocalWrapper.Local.
func (t *tmLocalWrapper) BroadcastTxCommit(ctx context.Context, tx comettypes.Tx) (*coretypes.ResultBroadcastTxCommit, error) {
	panic("unimplemented")
}

// BroadcastTxSync implements client.CometRPC.
// Subtle: this method shadows the method (*Local).BroadcastTxSync of tmLocalWrapper.Local.
func (t *tmLocalWrapper) BroadcastTxSync(ctx context.Context, tx comettypes.Tx) (*coretypes.ResultBroadcastTx, error) {
	panic("unimplemented")
}

// Commit implements client.CometRPC.
// Subtle: this method shadows the method (*Local).Commit of tmLocalWrapper.Local.
func (t *tmLocalWrapper) Commit(ctx context.Context, height *int64) (*coretypes.ResultCommit, error) {
	panic("unimplemented")
}

// Status implements client.CometRPC.
// Subtle: this method shadows the method (*Local).Status of tmLocalWrapper.Local.
func (t *tmLocalWrapper) Status(context.Context) (*coretypes.ResultStatus, error) {
	panic("unimplemented")
}

// Tx implements client.CometRPC.
// Subtle: this method shadows the method (*Local).Tx of tmLocalWrapper.Local.
func (t *tmLocalWrapper) Tx(ctx context.Context, hash []byte, prove bool) (*coretypes.ResultTx, error) {
	panic("unimplemented")
}

// TxSearch implements client.CometRPC.
// Subtle: this method shadows the method (*Local).TxSearch of tmLocalWrapper.Local.
func (t *tmLocalWrapper) TxSearch(ctx context.Context, query string, prove bool, page *int, perPage *int, orderBy string) (*coretypes.ResultTxSearch, error) {
	panic("unimplemented")
}

// Validators implements client.CometRPC.
// Subtle: this method shadows the method (*Local).Validators of tmLocalWrapper.Local.
func (t *tmLocalWrapper) Validators(ctx context.Context, height *int64, page *int, perPage *int) (*coretypes.ResultValidators, error) {
	panic("unimplemented")
}

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
