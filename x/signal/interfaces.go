package signal

import (
	"context"

	"cosmossdk.io/math"
	stakingtypes "cosmossdk.io/x/staking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type StakingKeeper interface {
	GetLastValidatorPower(ctx context.Context, addr sdk.ValAddress) (int64, error)
	GetLastTotalPower(ctx context.Context) (math.Int, error)
	GetValidator(ctx context.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, err error)
}

type ConsensusKeeper interface {
	AppVersion(ctx context.Context) (uint64, error)
}
