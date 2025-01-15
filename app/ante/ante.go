package ante

import (
	"context"

	paramkeeper "cosmossdk.io/x/params/keeper"
	"cosmossdk.io/x/tx/signing"
	blobante "github.com/celestiaorg/celestia-app/v3/x/blob/ante"
	blob "github.com/celestiaorg/celestia-app/v3/x/blob/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	ibcante "github.com/cosmos/ibc-go/v9/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v9/modules/core/keeper"
)

func NewAnteHandler(
	authkeeper ante.AccountKeeper,
	accountAbstractionKeeper ante.AccountAbstractionKeeper,
	bankKeeper authtypes.BankKeeper,
	blobKeeper blob.Keeper,
	consensusKeeper ConsensusKeeper,
	feegrantKeeper ante.FeegrantKeeper,
	signModeHandler *signing.HandlerMap,
	sigGasConsumer ante.SignatureVerificationGasConsumer,
	channelKeeper *ibckeeper.Keeper,
	paramKeeper paramkeeper.Keeper,
	msgVersioningGateKeeper *MsgVersioningGateKeeper,
	forbiddenGovUpdateParams map[string][]string,
) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		// Wraps the panic with the string format of the transaction
		NewHandlePanicDecorator(),
		// Prevents messages that don't belong to the correct app version
		// from being executed
		msgVersioningGateKeeper,
		// Set up the context with a gas meter.
		// Must be called before gas consumption occurs in any other decorator.
		ante.NewSetUpContextDecorator(authkeeper.GetEnvironment(), consensusKeeper),
		// Ensure the tx is not larger than the configured threshold.
		NewMaxTxSizeDecorator(consensusKeeper),
		// Ensure the tx does not contain any extension options.
		ante.NewExtensionOptionsDecorator(nil),
		// Ensure the tx passes ValidateBasic.
		ante.NewValidateBasicDecorator(authkeeper.GetEnvironment()),
		// Ensure the tx has not reached a height timeout.
		ante.NewTxTimeoutHeightDecorator(authkeeper.GetEnvironment()),
		// Ensure the tx memo <= max memo characters.
		ante.NewValidateMemoDecorator(authkeeper),
		// Ensure the tx's gas limit is > the gas consumed based on the tx size.
		// Side effect: consumes gas from the gas meter.
		NewConsumeGasForTxSizeDecorator(authkeeper, consensusKeeper),
		// Ensure the feepayer (fee granter or first signer) has enough funds to pay for the tx.
		// Ensure the gas price >= network min gas price if app version >= 2.
		// Side effect: deducts fees from the fee payer. Sets the tx priority in context.
		ante.NewDeductFeeDecorator(authkeeper, bankKeeper, feegrantKeeper, ValidateTxFeeWrapper(paramKeeper, consensusKeeper)),
		// Ensure that the tx's count of signatures is <= the tx signature limit.
		ante.NewValidateSigCountDecorator(authkeeper),
		// Ensure that the tx's signatures are valid. For each signature, ensure
		// that the signature's sequence number (a.k.a nonce) matches the
		// account sequence number of the signer.
		// Ensure that the tx's gas limit is > the gas consumed based on signature verification.
		// Set public keys in the context for fee-payer and all signers.
		// Side effect: consumes gas from the gas meter.
		// Side effect: increment the nonce for all tx signers.
		ante.NewSigVerificationDecorator(authkeeper, signModeHandler, sigGasConsumer, accountAbstractionKeeper),
		// Ensure that the tx's gas limit is > the gas consumed based on the blob size(s).
		// Contract: must be called after all decorators that consume gas.
		// Note: does not consume gas from the gas meter.
		blobante.NewMinGasPFBDecorator(blobKeeper, consensusKeeper),
		// Ensure that the tx's total blob size is <= the max blob size.
		// Only applies to app version == 1.
		blobante.NewMaxTotalBlobSizeDecorator(blobKeeper, consensusKeeper),
		// Ensure that the blob shares occupied by the tx <= the max shares
		// available to blob data in a data square. Only applies to app version
		// >= 2.
		blobante.NewBlobShareDecorator(blobKeeper, consensusKeeper),
		// Ensure that tx's with a MsgSubmitProposal have at least one proposal
		// message. Additionally ensure that the proposals do not contain any
		NewGovProposalDecorator(forbiddenGovUpdateParams),
		// Ensure that the tx is not an IBC packet or update message that has already been processed.
		ibcante.NewRedundantRelayDecorator(channelKeeper),
	)
}

var DefaultSigVerificationGasConsumer = ante.DefaultSigVerificationGasConsumer

// ConsensusKeeper is the expected interface of the consensus keeper
type ConsensusKeeper interface {
	ante.ConsensusKeeper
	AppVersion(context.Context) (uint64, error)
}
