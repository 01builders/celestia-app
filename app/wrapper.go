package app

import (
	"context"

	cmtabci "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	tmabci "github.com/tendermint/tendermint/abci/types"
	tmcrypto "github.com/tendermint/tendermint/proto/tendermint/crypto"
)

func (app *App) ApplySnapshotChunk(req tmabci.RequestApplySnapshotChunk) tmabci.ResponseApplySnapshotChunk {
	v1Req := &cmtabci.ApplySnapshotChunkRequest{
		Index:  req.Index,
		Chunk:  req.Chunk,
		Sender: req.Sender,
	}
	v1Res, err := app.BaseApp.ApplySnapshotChunk(v1Req)
	if err != nil {
		panic(err)
	}
	return tmabci.ResponseApplySnapshotChunk{
		Result:        tmabci.ResponseApplySnapshotChunk_Result(v1Res.Result),
		RefetchChunks: v1Res.RefetchChunks,
		RejectSenders: v1Res.RejectSenders,
	}
}

func (app *App) BeginBlock(req tmabci.RequestBeginBlock) tmabci.ResponseBeginBlock {
	// TODO: remove after migration to FinalizeBlock
	return tmabci.ResponseBeginBlock{}
}

func (app *App) DeliverTx(tmabci.RequestDeliverTx) tmabci.ResponseDeliverTx {
	// TODO: remove after migration to FinalizeBlock
	return tmabci.ResponseDeliverTx{}
}

func (app *App) EndBlock(req tmabci.RequestEndBlock) tmabci.ResponseEndBlock {
	// TODO: remove after migration to FinalizeBlock
	return tmabci.ResponseEndBlock{}
}

func (app *App) CheckTx(req tmabci.RequestCheckTx) tmabci.ResponseCheckTx {
	v1Req := &cmtabci.CheckTxRequest{
		Tx: req.Tx,
	}
	v1Res, err := app.CheckTxV1(v1Req)
	if err != nil {
		panic(err)
	}
	events := make([]tmabci.Event, 0, len(v1Res.Events))
	for _, event := range v1Res.Events {
		attrs := make([]tmabci.EventAttribute, 0, len(event.Attributes))
		for _, attr := range event.Attributes {
			attrs = append(attrs, tmabci.EventAttribute{
				Key:   []byte(attr.Key),
				Value: []byte(attr.Value),
			})
		}
		events = append(events, tmabci.Event{
			Type:       event.Type,
			Attributes: attrs,
		})
	}
	return tmabci.ResponseCheckTx{
		Code:      v1Res.Code,
		Data:      v1Res.Data,
		Log:       v1Res.Log,
		Info:      v1Res.Info,
		GasWanted: v1Res.GasWanted,
		GasUsed:   v1Res.GasUsed,
		Events:    events,
		Codespace: v1Res.Codespace,
	}
}

func (app *App) Commit() tmabci.ResponseCommit {
	v1Res, err := app.BaseApp.Commit()
	if err != nil {
		panic(err)
	}
	return tmabci.ResponseCommit{
		RetainHeight: v1Res.RetainHeight,
	}
}

func (app *App) Info(req tmabci.RequestInfo) tmabci.ResponseInfo {
	v1Req := &cmtabci.InfoRequest{
		Version:      req.Version,
		BlockVersion: req.BlockVersion,
		P2PVersion:   req.P2PVersion,
	}
	v1Res, err := app.InfoV1(v1Req)
	if err != nil {
		panic(err)
	}
	return tmabci.ResponseInfo{
		Data:             v1Res.Data,
		Version:          v1Res.Version,
		AppVersion:       v1Res.AppVersion,
		LastBlockHeight:  v1Res.LastBlockHeight,
		LastBlockAppHash: v1Res.LastBlockAppHash,
	}
}

func (app *App) InitChain(req tmabci.RequestInitChain) tmabci.ResponseInitChain {
	v1Req := &cmtabci.InitChainRequest{
		Time:    req.Time,
		ChainId: req.ChainId,
	}
	// TODO map the rest of the fields in request and response
	v1Res, err := app.InitChainV1(v1Req)
	if err != nil {
		panic(err)
	}
	return tmabci.ResponseInitChain{
		AppHash: v1Res.AppHash,
	}
}

func (app *App) ListSnapshots(req tmabci.RequestListSnapshots) tmabci.ResponseListSnapshots {
	v1Req := &cmtabci.ListSnapshotsRequest{}
	v1Res, err := app.BaseApp.ListSnapshots(v1Req)
	if err != nil {
		panic(err)
	}
	snapshots := make([]*tmabci.Snapshot, 0, len(v1Res.Snapshots))
	for _, snapshot := range v1Res.Snapshots {
		snapshots = append(snapshots, &tmabci.Snapshot{
			Height:   snapshot.Height,
			Format:   snapshot.Format,
			Chunks:   snapshot.Chunks,
			Hash:     snapshot.Hash,
			Metadata: snapshot.Metadata,
		})
	}
	return tmabci.ResponseListSnapshots{
		Snapshots: snapshots,
	}
}

func (app *App) LoadSnapshotChunk(req tmabci.RequestLoadSnapshotChunk) tmabci.ResponseLoadSnapshotChunk {
	v1Req := &cmtabci.LoadSnapshotChunkRequest{
		Height: req.Height,
		Format: req.Format,
		Chunk:  req.Chunk,
	}
	v1Res, err := app.BaseApp.LoadSnapshotChunk(v1Req)
	if err != nil {
		panic(err)
	}
	return tmabci.ResponseLoadSnapshotChunk{
		Chunk: v1Res.Chunk,
	}
}

func (app *App) OfferSnapshot(req tmabci.RequestOfferSnapshot) tmabci.ResponseOfferSnapshot {
	if app.IsSealed() {
		// If the app is sealed, keys have already been mounted so this can
		// delegate to the baseapp's OfferSnapshot.
		return app.offerSnapshot(req)
	}

	app.Logger().Info("offering snapshot", "height", req.Snapshot.Height, "app_version", req.AppVersion)
	if req.AppVersion != 0 {
		if !isSupportedAppVersion(req.AppVersion) {
			app.Logger().Info("rejecting snapshot because unsupported app version", "app_version", req.AppVersion)
			return tmabci.ResponseOfferSnapshot{
				Result: tmabci.ResponseOfferSnapshot_REJECT,
			}
		}

		app.Logger().Info("mounting keys for snapshot", "app_version", req.AppVersion)
		app.mountKeysAndInit(req.AppVersion)
		return app.offerSnapshot(req)
	}

	// If the app version is not set in the snapshot, this falls back to inferring the app version based on the upgrade height.
	if app.upgradeHeightV2 == 0 {
		app.Logger().Info("v2 upgrade height not set, assuming app version 2")
		app.mountKeysAndInit(v2)
		return app.offerSnapshot(req)
	}

	if req.Snapshot.Height >= uint64(app.upgradeHeightV2) {
		app.Logger().Info("snapshot height is greater than or equal to upgrade height, assuming app version 2")
		app.mountKeysAndInit(v2)
		return app.offerSnapshot(req)
	}

	app.Logger().Info("snapshot height is less than upgrade height, assuming app version 1")
	app.mountKeysAndInit(v1)
	return app.offerSnapshot(req)
}

func (app *App) offerSnapshot(req tmabci.RequestOfferSnapshot) tmabci.ResponseOfferSnapshot {
	v1Req := &cmtabci.OfferSnapshotRequest{
		Snapshot: &cmtabci.Snapshot{
			Height:   req.Snapshot.Height,
			Format:   req.Snapshot.Format,
			Chunks:   req.Snapshot.Chunks,
			Hash:     req.Snapshot.Hash,
			Metadata: req.Snapshot.Metadata,
		},
		AppHash: req.AppHash,
	}
	v1Res, err := app.BaseApp.OfferSnapshot(v1Req)
	if err != nil {
		panic(err)
	}
	return tmabci.ResponseOfferSnapshot{
		Result: tmabci.ResponseOfferSnapshot_Result(v1Res.Result),
	}
}

func (app *App) Query(req tmabci.RequestQuery) tmabci.ResponseQuery {
	v1Req := &cmtabci.QueryRequest{
		Data:   req.Data,
		Path:   req.Path,
		Height: req.Height,
		Prove:  req.Prove,
	}
	// context is unused in a v0.52 Query
	ctx := context.Background()
	v1Res, err := app.BaseApp.Query(ctx, v1Req)
	if err != nil {
		panic(err)
	}
	proofOps := make([]tmcrypto.ProofOp, 0, len(v1Res.ProofOps.Ops))
	for _, proofOp := range v1Res.ProofOps.Ops {
		proofOps = append(proofOps, tmcrypto.ProofOp{
			Type: proofOp.Type,
			Key:  proofOp.Key,
			Data: proofOp.Data,
		})
	}
	return tmabci.ResponseQuery{
		Code:      v1Res.Code,
		Log:       v1Res.Log,
		Info:      v1Res.Info,
		Index:     v1Res.Index,
		Key:       v1Res.Key,
		Value:     v1Res.Value,
		Height:    v1Res.Height,
		Codespace: v1Res.Codespace,
		ProofOps:  &tmcrypto.ProofOps{Ops: proofOps},
	}
}

func (app *App) SetOption(req tmabci.RequestSetOption) tmabci.ResponseSetOption {
	panic("not implemented")
}
