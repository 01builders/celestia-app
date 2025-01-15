package signal

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/math"
	stakingtypes "cosmossdk.io/x/staking/types"
	"github.com/celestiaorg/celestia-app/v3/pkg/appconsts"
	"github.com/celestiaorg/celestia-app/v3/x/signal/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper implements the MsgServer and QueryServer interfaces
var (
	_ types.MsgServer   = &Keeper{}
	_ types.QueryServer = Keeper{}

	// defaultSignalThreshold is 5/6 or approximately 83.33%
	defaultSignalThreshold = math.LegacyNewDec(5).Quo(math.LegacyNewDec(6))
)

// Threshold is the fraction of voting power that is required
// to signal for a version change. It is set to 5/6 as the middle point
// between 2/3 and 3/3 providing 1/6 fault tolerance to halting the
// network during an upgrade period. It can be modified through a
// hard fork change that modified the app version
func Threshold(_ uint64) math.LegacyDec {
	return defaultSignalThreshold
}

type Keeper struct {
	appmodule.Environment

	// binaryCodec is used to marshal and unmarshal data from the store.
	binaryCodec codec.BinaryCodec

	// stakingKeeper is used to fetch validators to calculate the total power
	// signalled to a version.
	stakingKeeper StakingKeeper

	// consensusKeeper is used to get the app version
	consensusKeeper ConsensusKeeper
}

// NewKeeper returns a signal keeper.
func NewKeeper(
	env appmodule.Environment,
	binaryCodec codec.BinaryCodec,
	stakingKeeper StakingKeeper,
	consensusKeeper ConsensusKeeper,
) Keeper {
	return Keeper{
		Environment:     env,
		binaryCodec:     binaryCodec,
		stakingKeeper:   stakingKeeper,
		consensusKeeper: consensusKeeper,
	}
}

// SignalVersion is a method required by the MsgServer interface.
func (k Keeper) SignalVersion(ctx context.Context, req *types.MsgSignalVersion) (*types.MsgSignalVersionResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if k.IsUpgradePending(sdkCtx) {
		return &types.MsgSignalVersionResponse{}, types.ErrUpgradePending.Wrapf("can not signal version")
	}

	valAddr, err := sdk.ValAddressFromBech32(req.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	// The signalled version can not be less than the current version.
	currentVersion, err := k.consensusKeeper.AppVersion(ctx)
	if err != nil {
		return nil, err
	}

	if req.Version < currentVersion {
		return nil, types.ErrInvalidSignalVersion.Wrapf("signalled version %d, current version %d", req.Version, currentVersion)
	}

	_, err = k.stakingKeeper.GetValidator(sdkCtx, valAddr)
	if err != nil {
		return nil, err
	}

	k.SetValidatorVersion(sdkCtx, valAddr, req.Version)
	return &types.MsgSignalVersionResponse{}, nil
}

// TryUpgrade is a method required by the MsgServer interface. It tallies the
// voting power that has voted on each version. If one version has reached a
// quorum, an upgrade is persisted to the store. The upgrade is used by the
// application later when it is time to upgrade to that version.
func (k *Keeper) TryUpgrade(ctx context.Context, _ *types.MsgTryUpgrade) (*types.MsgTryUpgradeResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if k.IsUpgradePending(sdkCtx) {
		return &types.MsgTryUpgradeResponse{}, types.ErrUpgradePending.Wrapf("can not try upgrade")
	}

	threshold := k.GetVotingPowerThreshold(sdkCtx)
	hasQuorum, version := k.TallyVotingPower(sdkCtx, threshold.Int64())
	if hasQuorum {
		appVersion, err := k.consensusKeeper.AppVersion(sdkCtx)
		if err != nil {
			return nil, err
		}

		if version <= appVersion {
			return &types.MsgTryUpgradeResponse{}, types.ErrInvalidUpgradeVersion.Wrapf("can not upgrade to version %v because it is less than or equal to current version %v", version, appVersion)
		}
		header := sdkCtx.HeaderInfo()
		upgrade := types.Upgrade{
			AppVersion:    version,
			UpgradeHeight: header.Height + appconsts.UpgradeHeightDelay(header.ChainID, appVersion),
		}
		k.setUpgrade(sdkCtx, upgrade)
	}
	return &types.MsgTryUpgradeResponse{}, nil
}

// VersionTally enables a client to query for the tally of voting power has
// signalled for a particular version.
func (k Keeper) VersionTally(ctx context.Context, req *types.QueryVersionTallyRequest) (*types.QueryVersionTallyResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	totalVotingPower, err := k.stakingKeeper.GetLastTotalPower(sdkCtx)
	if err != nil {
		return nil, err
	}
	currentVotingPower := math.NewInt(0)
	store := runtime.KVStoreAdapter(k.Environment.KVStoreService.OpenKVStore(ctx))
	iterator := store.Iterator(types.FirstSignalKey, nil)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		if bytes.Equal(iterator.Key(), types.UpgradeKey) {
			continue
		}
		valAddress := sdk.ValAddress(iterator.Key())
		power, err := k.stakingKeeper.GetLastValidatorPower(sdkCtx, valAddress)
		if err != nil {
			return nil, err
		}
		version := VersionFromBytes(iterator.Value())
		if version == req.Version {
			currentVotingPower = currentVotingPower.AddRaw(power)
		}
	}
	threshold := k.GetVotingPowerThreshold(sdkCtx)
	return &types.QueryVersionTallyResponse{
		VotingPower:      currentVotingPower.Uint64(),
		ThresholdPower:   threshold.Uint64(),
		TotalVotingPower: totalVotingPower.Uint64(),
	}, nil
}

// SetValidatorVersion saves a signalled version for a validator.
func (k Keeper) SetValidatorVersion(ctx sdk.Context, valAddress sdk.ValAddress, version uint64) {
	store := k.Environment.KVStoreService.OpenKVStore(ctx)
	store.Set(valAddress, VersionToBytes(version))
}

// DeleteValidatorVersion deletes a signalled version for a validator.
func (k Keeper) DeleteValidatorVersion(ctx sdk.Context, valAddress sdk.ValAddress) {
	store := k.Environment.KVStoreService.OpenKVStore(ctx)
	store.Delete(valAddress)
}

// TallyVotingPower tallies the voting power for each version and returns true
// and the version if any version has reached the quorum in voting power.
// Returns false and 0 otherwise.
func (k Keeper) TallyVotingPower(ctx sdk.Context, threshold int64) (bool, uint64) {
	versionToPower := make(map[uint64]int64)
	store := runtime.KVStoreAdapter(k.Environment.KVStoreService.OpenKVStore(ctx))
	iterator := store.Iterator(types.FirstSignalKey, nil)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		if bytes.Equal(iterator.Key(), types.UpgradeKey) {
			continue
		}
		valAddress := sdk.ValAddress(iterator.Key())
		// check that the validator is still part of the bonded set
		found := true
		val, err := k.stakingKeeper.GetValidator(ctx, valAddress)
		if err != nil {
			if errors.Is(err, stakingtypes.ErrNoValidatorFound) {
				k.DeleteValidatorVersion(ctx, valAddress)
				found = false
			} else {
				panic(err)
			}
		}
		// if the validator is not bonded, skip it's voting power
		if !found || !val.IsBonded() {
			continue
		}
		power, err := k.stakingKeeper.GetLastValidatorPower(ctx, valAddress)
		if err != nil {
			panic(err)
		}
		version := VersionFromBytes(iterator.Value())
		if _, ok := versionToPower[version]; !ok {
			versionToPower[version] = power
		} else {
			versionToPower[version] += power
		}
		if versionToPower[version] >= threshold {
			return true, version
		}
	}
	return false, 0
}

// GetVotingPowerThreshold returns the voting power threshold required to
// upgrade to a new version.
func (k Keeper) GetVotingPowerThreshold(ctx sdk.Context) math.Int {
	totalVotingPower, err := k.stakingKeeper.GetLastTotalPower(ctx)
	if err != nil {
		panic(err)
	}

	appVersion, err := k.consensusKeeper.AppVersion(ctx)
	if err != nil {
		panic(err)
	}

	thresholdFraction := Threshold(appVersion)
	return thresholdFraction.MulInt(totalVotingPower).Ceil().TruncateInt()
}

// ShouldUpgrade returns whether the signalling mechanism has concluded that the
// network is ready to upgrade and the version to upgrade to. It returns false
// and 0 if no version has reached quorum.
func (k *Keeper) ShouldUpgrade(ctx sdk.Context) (isQuorumVersion bool, version uint64) {
	upgrade, ok := k.getUpgrade(ctx)
	if !ok {
		return false, 0
	}

	hasUpgradeHeightBeenReached := ctx.BlockHeight() >= upgrade.UpgradeHeight
	if hasUpgradeHeightBeenReached {
		return true, upgrade.AppVersion
	}
	return false, 0
}

// ResetTally resets the tally after a version change. It iterates over the
// store and deletes all versions. It also resets the quorumVersion and
// upgradeHeight to 0.
func (k *Keeper) ResetTally(ctx sdk.Context) {
	store := runtime.KVStoreAdapter(k.Environment.KVStoreService.OpenKVStore(ctx))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()
	// delete the value in the upgrade key and all signals.
	for ; iterator.Valid(); iterator.Next() {
		store.Delete(iterator.Key())
	}
}

func VersionToBytes(version uint64) []byte {
	return binary.BigEndian.AppendUint64(nil, version)
}

func VersionFromBytes(version []byte) uint64 {
	return binary.BigEndian.Uint64(version)
}

// GetUpgrade returns the current upgrade information.
func (k Keeper) GetUpgrade(ctx context.Context, _ *types.QueryGetUpgradeRequest) (*types.QueryGetUpgradeResponse, error) {
	upgrade, ok := k.getUpgrade(sdk.UnwrapSDKContext(ctx))
	if !ok {
		return &types.QueryGetUpgradeResponse{}, nil
	}
	return &types.QueryGetUpgradeResponse{Upgrade: &upgrade}, nil
}

// IsUpgradePending returns true if an app version has reached quorum and the
// chain should upgrade to the app version at the upgrade height. While the
// keeper has an upgrade pending the SignalVersion and TryUpgrade messages will
// be rejected.
func (k *Keeper) IsUpgradePending(ctx sdk.Context) bool {
	_, ok := k.getUpgrade(ctx)
	return ok
}

// getUpgrade returns the current upgrade information from the store.
// If an upgrade is found, it returns the upgrade object and true.
// If no upgrade is found, it returns an empty upgrade object and false.
func (k *Keeper) getUpgrade(ctx sdk.Context) (upgrade types.Upgrade, ok bool) {
	store := k.Environment.KVStoreService.OpenKVStore(ctx)
	value, err := store.Get(types.UpgradeKey)
	if value == nil || err != nil {
		return types.Upgrade{}, false
	}
	k.binaryCodec.MustUnmarshal(value, &upgrade)
	return upgrade, true
}

// setUpgrade sets the upgrade in the store.
func (k *Keeper) setUpgrade(ctx sdk.Context, upgrade types.Upgrade) {
	store := k.Environment.KVStoreService.OpenKVStore(ctx)
	value := k.binaryCodec.MustMarshal(&upgrade)
	store.Set(types.UpgradeKey, value)
}
