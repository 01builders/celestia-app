package ante_test

import (
	"context"
	"testing"

	"cosmossdk.io/x/authz"
	banktypes "cosmossdk.io/x/bank/types"
	"github.com/celestiaorg/celestia-app/v4/app/ante"
	"github.com/celestiaorg/celestia-app/v4/app/encoding"
	"github.com/celestiaorg/celestia-app/v4/test/util/testnode"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

type mockConsensusKeeper struct {
	ante.ConsensusKeeper
	appVersion uint64
}

func (m mockConsensusKeeper) AppVersion(ctx context.Context) (uint64, error) {
	return m.appVersion, nil
}

func TestMsgGateKeeperAnteHandler(t *testing.T) {
	nestedBankSend := authz.NewMsgExec(testnode.RandomAddress().String(), []sdk.Msg{&banktypes.MsgSend{
		FromAddress: testnode.RandomAddress().String(),
	}})
	nestedMultiSend := authz.NewMsgExec(testnode.RandomAddress().String(), []sdk.Msg{&banktypes.MsgMultiSend{}})
	cdc := encoding.MakeConfig()
	banktypes.RegisterInterfaces(cdc.InterfaceRegistry)

	// Define test cases
	tests := []struct {
		name      string
		msg       sdk.Msg
		acceptMsg bool
		version   uint64
	}{
		{
			name: "Accept MsgSend",
			msg: &banktypes.MsgSend{
				FromAddress: testnode.RandomAddress().String(),
			},
			acceptMsg: true,
			version:   1,
		},
		{
			name:      "Accept nested MsgSend",
			msg:       &nestedBankSend,
			acceptMsg: true,
			version:   1,
		},
		{
			name:      "Reject MsgMultiSend",
			msg:       &banktypes.MsgMultiSend{},
			acceptMsg: false,
			version:   1,
		},
		{
			name:      "Reject nested MsgMultiSend",
			msg:       &nestedMultiSend,
			acceptMsg: false,
			version:   1,
		},
		{
			name: "Reject MsgSend with version 2",
			msg: &banktypes.MsgSend{
				FromAddress: testnode.RandomAddress().String(),
			},
			acceptMsg: false,
			version:   2,
		},
		{
			name:      "Reject nested MsgSend with version 2",
			msg:       &nestedBankSend,
			acceptMsg: false,
			version:   2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msgGateKeeper := ante.NewMsgVersioningGateKeeper(map[uint64]map[string]struct{}{
				1: {
					"/cosmos.bank.v1beta1.MsgSend":  {},
					"/cosmos.authz.v1beta1.MsgExec": {},
				},
				2: {},
			}, mockConsensusKeeper{appVersion: tc.version})
			anteHandler := sdk.ChainAnteDecorators(msgGateKeeper)

			ctx := sdk.NewContext(nil, false, nil)
			txBuilder := cdc.TxConfig.NewTxBuilder()
			require.NoError(t, txBuilder.SetMsgs(tc.msg))
			_, err := anteHandler(ctx, txBuilder.GetTx(), false)

			msg := tc.msg
			if sdk.MsgTypeURL(msg) == "/cosmos.authz.v1beta1.MsgExec" {
				execMsg, ok := msg.(*authz.MsgExec)
				require.True(t, ok)

				nestedMsgs, err := execMsg.GetMessages()
				require.NoError(t, err)
				msg = nestedMsgs[0]
			}

			allowed, err2 := msgGateKeeper.IsAllowed(ctx, sdk.MsgTypeURL(msg))

			require.NoError(t, err2)
			if tc.acceptMsg {
				require.NoError(t, err, "expected message to be accepted")
				require.True(t, allowed)
			} else {
				require.Error(t, err, "expected message to be rejected")
				require.False(t, allowed)
			}
		})
	}
}
