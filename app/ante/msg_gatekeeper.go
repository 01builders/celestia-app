package ante

import (
	"context"

	"cosmossdk.io/x/authz"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.AnteDecorator      = MsgVersioningGateKeeper{}
	_ baseapp.CircuitBreaker = MsgVersioningGateKeeper{}
)

// MsgVersioningGateKeeper dictates which transactions are accepted for an app version
type MsgVersioningGateKeeper struct {
	// acceptedMsgs is a map from appVersion -> msgTypeURL -> struct{}.
	// If a msgTypeURL is present in the map it should be accepted for that appVersion.
	acceptedMsgs    map[uint64]map[string]struct{}
	consensusKeeper ConsensusKeeper
}

func NewMsgVersioningGateKeeper(acceptedList map[uint64]map[string]struct{}, consensusKeeper ConsensusKeeper) *MsgVersioningGateKeeper {
	return &MsgVersioningGateKeeper{
		acceptedMsgs:    acceptedList,
		consensusKeeper: consensusKeeper,
	}
}

// AnteHandle implements the ante.Decorator interface
func (mgk MsgVersioningGateKeeper) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	appVersion, err := mgk.consensusKeeper.AppVersion(ctx)
	if err != nil {
		return ctx, err
	}

	acceptedMsgs, exists := mgk.acceptedMsgs[appVersion]
	if !exists {
		return ctx, sdkerrors.ErrNotSupported.Wrapf("app version %d is not supported", appVersion)
	}

	if err := mgk.hasInvalidMsg(ctx, acceptedMsgs, tx.GetMsgs(), appVersion); err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

func (mgk MsgVersioningGateKeeper) hasInvalidMsg(ctx sdk.Context, acceptedMsgs map[string]struct{}, msgs []sdk.Msg, appVersion uint64) error {
	for _, msg := range msgs {
		// Recursively check for invalid messages in nested authz messages.
		if execMsg, ok := msg.(*authz.MsgExec); ok {
			nestedMsgs, err := execMsg.GetMessages()
			if err != nil {
				return err
			}
			if err = mgk.hasInvalidMsg(ctx, acceptedMsgs, nestedMsgs, appVersion); err != nil {
				return err
			}
		}

		msgTypeURL := sdk.MsgTypeURL(msg)
		_, exists := acceptedMsgs[msgTypeURL]
		if !exists {
			return sdkerrors.ErrNotSupported.Wrapf("message type %s is not supported in version %d", msgTypeURL, appVersion)
		}
	}

	return nil
}

func (mgk MsgVersioningGateKeeper) IsAllowed(ctx context.Context, msgName string) (bool, error) {
	appVersion, err := mgk.consensusKeeper.AppVersion(ctx)
	if err != nil {
		return false, sdkerrors.ErrLogic.Wrap("failed to get app version")
	}

	acceptedMsgs, exists := mgk.acceptedMsgs[appVersion]
	if !exists {
		return false, sdkerrors.ErrNotSupported.Wrapf("app version %d is not supported", appVersion)
	}
	_, exists = acceptedMsgs[msgName]
	if !exists {
		return false, nil
	}
	return true, nil
}
