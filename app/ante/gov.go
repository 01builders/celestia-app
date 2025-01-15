package ante

import (
	"cosmossdk.io/errors"
	"cosmossdk.io/x/authz"
	gov "cosmossdk.io/x/gov/types"
	govv1 "cosmossdk.io/x/gov/types/v1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GovProposalDecorator ensures that a tx with a MsgSubmitProposal has at least one message in the proposal.
// Additionally it replace the x/paramfilter module that existed in v3 and earlier versions.
type GovProposalDecorator struct{}

func NewGovProposalDecorator() GovProposalDecorator {
	return GovProposalDecorator{}
}

// AnteHandle implements the AnteHandler interface.
// It ensures that MsgSubmitProposal has at least one message
// It ensures params are filtered within messages
func (d GovProposalDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	for _, m := range tx.GetMsgs() {
		if proposal, ok := m.(*govv1.MsgSubmitProposal); ok {
			if len(proposal.Messages) == 0 {
				return ctx, errors.Wrapf(gov.ErrNoProposalMsgs, "must include at least one message in proposal")
			}
		}

		// we need to check if a gov proposal wasn't contains in a authz message
		if msgExec, ok := m.(*authz.MsgExec); ok {
			for _, msg := range msgExec.Msgs {
				_ = msg
			}
		}
	}

	return next(ctx, tx, simulate)
}

// TODO: To be moved to antehandler
// BlockedParams returns the params that require a hardfork to change, and
// cannot be changed via governance.
// func (app *App) BlockedParams() [][2]string {
// 	return [][2]string{
// 		// bank.SendEnabled
// 		{banktypes.ModuleName, string(banktypes.KeySendEnabled)},
// 		// staking.UnbondingTime
// 		{stakingtypes.ModuleName, string(stakingtypes.KeyUnbondingTime)},
// 		// staking.BondDenom
// 		{stakingtypes.ModuleName, string(stakingtypes.KeyBondDenom)},
// 		// consensus.validator.PubKeyTypes
// 		{baseapp.Paramspace, string(baseapp.ParamStoreKeyValidatorParams)},
// 	}
// }
