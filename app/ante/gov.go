package ante

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

type ParamFilter func(sdk.Msg) error

// GovProposalDecorator ensures that a tx with a MsgSubmitProposal has at least one message in the proposal.
// Additionally it replace the x/paramfilter module that existed in v3 and earlier versions.
type GovProposalDecorator struct {
	// paramFilters is a map of type_url to a list of parameter fields that cannot be changed via governance.
	paramFilters map[string]ParamFilter
}

func NewGovProposalDecorator(paramFilters map[string]ParamFilter) GovProposalDecorator {
	return GovProposalDecorator{
		paramFilters: paramFilters,
	}
}

// AnteHandle implements the AnteHandler interface.
// It ensures that MsgSubmitProposal has at least one message
// It ensures params are filtered within messages
func (d GovProposalDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	for _, m := range tx.GetMsgs() {
		if msgSubmitProp, ok := m.(*govv1.MsgSubmitProposal); ok {
			msgs, err := msgSubmitProp.GetMsgs()
			if err != nil {
				return ctx, err
			}

			if len(msgs) == 0 {
				return ctx, errors.Wrapf(gov.ErrNoProposalMsgs, "must include at least one message in proposal")
			}

			if err := d.checkNestedMsgs(msgs); err != nil {
				return ctx, err
			}
		}

		// we need to check if a gov proposal wasn't contained in a authz message
		if msgExec, ok := m.(*authz.MsgExec); ok {
			msgs, err := msgExec.GetMessages()
			if err != nil {
				return ctx, err
			}

			if err := d.checkNestedMsgs(msgs); err != nil {
				return ctx, err
			}
		}
	}

	return next(ctx, tx, simulate)
}

func (d GovProposalDecorator) checkNestedMsgs(msgs []sdk.Msg) error {
	for _, msg := range msgs {
		switch m := msg.(type) {
		case *authz.MsgExec:
			nested, err := m.GetMessages()
			if err != nil {
				return err
			}

			if err := d.checkNestedMsgs(nested); err != nil {
				return err
			}
		case *govv1.MsgSubmitProposal:
			nested, err := m.GetMsgs()
			if err != nil {
				return err
			}

			if err := d.checkNestedMsgs(nested); err != nil {
				return err
			}
		default:
			if paramFilter, found := d.paramFilters[sdk.MsgTypeURL(m)]; found {
				if err := paramFilter(m); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
