package keeper

import (
	"context"
	"encoding/binary"

	addresscodec "cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	paramtypes "cosmossdk.io/x/params/types"
	stakingtypes "cosmossdk.io/x/staking/types"
	"github.com/celestiaorg/celestia-app/v3/x/blobstream/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keeper struct {
	appmodule.Environment

	cdc        codec.BinaryCodec
	paramSpace paramtypes.Subspace

	StakingKeeper  StakingKeeper
	ConsenusKeeper ConsenusKeeper
}

func NewKeeper(env appmodule.Environment, cdc codec.BinaryCodec, paramSpace paramtypes.Subspace, stakingKeeper StakingKeeper, consensusKeeper ConsenusKeeper) *Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		Environment:    env,
		cdc:            cdc,
		StakingKeeper:  stakingKeeper,
		ConsenusKeeper: consensusKeeper,
		paramSpace:     paramSpace,
	}
}

// GetParams returns the parameters from the store
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the parameters in the store
func (k Keeper) SetParams(ctx sdk.Context, ps types.Params) {
	k.paramSpace.SetParamSet(ctx, &ps)
}

// DeserializeValidatorIterator returns validators from the validator iterator.
// Adding here in Blobstream keeper as cdc is not available inside endblocker.
func (k Keeper) DeserializeValidatorIterator(vals []byte) stakingtypes.ValAddresses {
	validators := stakingtypes.ValAddresses{
		Addresses: []string{},
	}
	k.cdc.MustUnmarshal(vals, &validators)
	return validators
}

// StakingKeeper restricts the functionality of the bank keeper used in the blobstream
// keeper
type StakingKeeper interface {
	GetValidator(ctx context.Context, addr sdk.ValAddress) (stakingtypes.Validator, error)
	GetBondedValidatorsByPower(ctx context.Context) ([]stakingtypes.Validator, error)
	GetLastValidatorPower(ctx context.Context, valAddr sdk.ValAddress) (int64, error)
	ValidatorAddressCodec() addresscodec.Codec
}

type ConsenusKeeper interface {
	AppVersion(ctx context.Context) (uint64, error)
}

// UInt64FromBytes create uint from binary big endian representation.
func UInt64FromBytes(s []byte) uint64 {
	return binary.BigEndian.Uint64(s)
}
