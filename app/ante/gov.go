package ante

import (
	"cosmossdk.io/errors"
	"cosmossdk.io/x/authz"
	gov "cosmossdk.io/x/gov/types"
	govv1 "cosmossdk.io/x/gov/types/v1"
	"cosmossdk.io/x/params/types/proposal"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoproto "github.com/cosmos/gogoproto/proto"
	gogoany "github.com/cosmos/gogoproto/types/any"
)

// GovProposalDecorator ensures that a tx with a MsgSubmitProposal has at least one message in the proposal.
// Additionally it replace the x/paramfilter module that existed in v3 and earlier versions.
type GovProposalDecorator struct {
	// forbiddenGovUpdateParams is a map of type_url to a list of parameter fiels that cannot be changed via governance.
	forbiddenGovUpdateParams map[string][]string
}

func NewGovProposalDecorator(forbiddenParams map[string][]string) GovProposalDecorator {
	return GovProposalDecorator{forbiddenParams}
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

			if err := d.checkNestedMsgs(proposal.Messages); err != nil {
				return ctx, err
			}
		}

		// we need to check if a gov proposal wasn't contained in a authz message
		if msgExec, ok := m.(*authz.MsgExec); ok {
			if err := d.checkNestedMsgs(msgExec.Msgs); err != nil {
				return ctx, err
			}
		}
	}

	return next(ctx, tx, simulate)
}

func (d GovProposalDecorator) checkNestedMsgs(msgs []*gogoany.Any) error {
	for _, msg := range msgs {
		if msg.TypeUrl == "/"+gogoproto.MessageName((*authz.MsgExec)(nil)) {
			// unmarshal the authz.MsgExec and check nested messages
			exec := &authz.MsgExec{}
			// todo unmarshal

			if err := d.checkNestedMsgs(exec.Msgs); err != nil {
				return err
			}
		}

		if msg.TypeUrl == "/"+gogoproto.MessageName((*govv1.MsgSubmitProposal)(nil)) {
			// unmarshal the gov.MsgSubmitProposal and check nested messages
			proposal := &govv1.MsgSubmitProposal{}
			// todo unmarshal

			if len(proposal.Messages) == 0 {
				return errors.Wrapf(gov.ErrNoProposalMsgs, "must include at least one message in proposal")
			}

			if err := d.checkNestedMsgs(proposal.Messages); err != nil {
				return err
			}
		}

		forbiddenParams, ok := d.forbiddenGovUpdateParams[msg.TypeUrl]
		if !ok {
			continue
		}

		if hasForbiddenParams(msg, msg.TypeUrl, forbiddenParams) {
			return errors.Wrapf(proposal.ErrSettingParameter, "cannot update %s parameters via governance", msg.TypeUrl)
		}
	}

	return nil
}

func hasForbiddenParams(msg sdk.Msg, typeURL string, forbiddenParams []string) bool {
	// unmarshal msg to go struct
	// check if any forbidden param is present and different from the default value

	return false
}
