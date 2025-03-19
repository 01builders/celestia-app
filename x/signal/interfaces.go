package signal

import (
	"context"

	"cosmossdk.io/math"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type StakingKeeper interface {
	GetLastValidatorPower(ctx context.Context, addr sdk.ValAddress) (int64, error)
	GetLastTotalPower(ctx context.Context) (math.Int, error)
	GetValidator(ctx context.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, err error)
}

type UpgradeKeeper interface {
	ScheduleUpgrade(ctx context.Context, plan upgradetypes.Plan) error
	ApplyUpgrade(ctx context.Context, plan upgradetypes.Plan) error
}
