package posthandler

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// New returns a new posthandler chain.
func New() sdk.PostHandler {
	postDecorators := []sdk.PostDecorator{}
	return sdk.ChainPostDecorators(postDecorators...)
}
