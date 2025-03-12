package types

import sdk "github.com/cosmos/cosmos-sdk/types"

var (
	_ sdk.Msg = (*MsgPayForBlobs)(nil)
	_ sdk.Msg = (*MsgUpdateBlobParams)(nil)
)
