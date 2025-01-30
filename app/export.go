package app

import (
	"encoding/json"
	"log"

	"cosmossdk.io/collections"
	slashingtypes "cosmossdk.io/x/slashing/types"
	"cosmossdk.io/x/staking"
	stakingtypes "cosmossdk.io/x/staking/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ExportAppStateAndValidators exports the state of the application for a
// genesis file.
func (app *App) ExportAppStateAndValidators(forZeroHeight bool, jailAllowedAddrs []string) (servertypes.ExportedApp, error) {
	ctx, err := app.CreateQueryContext(app.LastBlockHeight(), false)
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	appVersion, err := app.AppVersion(ctx)
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	if !app.IsSealed() {
		app.mountKeysAndInit(appVersion)
	}

	// Create a new context so that the commit multi-store reflects the store
	// key mounting performed above.
	ctx, err = app.CreateQueryContext(app.LastBlockHeight(), false)
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	if forZeroHeight {
		app.prepForZeroHeightGenesis(ctx, jailAllowedAddrs)
	}

	genState, err := app.ModuleManager.ExportGenesis(ctx, app.encodingConfig.Codec, appVersion)
	if err != nil {
		return servertypes.ExportedApp{}, err
	}
	appState, err := json.MarshalIndent(genState, "", "  ")
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	validators, err := staking.WriteValidators(ctx, app.StakingKeeper)
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	return servertypes.ExportedApp{
		AppState:        appState,
		Validators:      validators,
		Height:          app.getExportHeight(forZeroHeight),
		ConsensusParams: app.BaseApp.GetConsensusParams(ctx),
	}, nil
}

func (app *App) getExportHeight(forZeroHeight bool) int64 {
	if forZeroHeight {
		return 0
	}
	// We export at last height + 1, because that's the height at which
	// Tendermint will start InitChain.
	return app.LastBlockHeight() + 1
}

// prepForZeroHeightGenesis preps for fresh start at zero height. Zero height
// genesis is a temporary feature which will be deprecated in favour of export
// at a block height.
func (app *App) prepForZeroHeightGenesis(ctx sdk.Context, jailAllowedAddrs []string) {
	applyAllowedAddrs := false

	// check if there is an allowed address list
	if len(jailAllowedAddrs) > 0 {
		applyAllowedAddrs = true
	}

	allowedAddrsMap := make(map[string]bool)

	for _, addr := range jailAllowedAddrs {
		_, err := sdk.ValAddressFromBech32(addr)
		if err != nil {
			log.Fatal(err)
		}
		allowedAddrsMap[addr] = true
	}

	/* Handle fee distribution state. */

	// withdraw all validator commission
	app.StakingKeeper.IterateValidators(ctx, func(_ int64, val sdk.ValidatorI) (stop bool) {
		operatorAddress := val.GetOperator()
		_, _ = app.DistrKeeper.WithdrawValidatorCommission(ctx, []byte(operatorAddress))
		return false
	})

	// withdraw all delegator rewards
	dels, err := app.StakingKeeper.GetAllDelegations(ctx)
	if err != nil {
		panic(err)
	}
	for _, delegation := range dels {
		valAddr, err := sdk.ValAddressFromBech32(delegation.ValidatorAddress)
		if err != nil {
			panic(err)
		}

		delAddr, err := sdk.AccAddressFromBech32(delegation.DelegatorAddress)
		if err != nil {
			panic(err)
		}
		_, _ = app.DistrKeeper.WithdrawDelegationRewards(ctx, delAddr, valAddr)
	}

	// clear validator slash events
	err = app.DistrKeeper.ValidatorSlashEvents.Clear(ctx, nil)
	if err != nil {
		panic(err)
	}

	// clear validator historical rewards
	err = app.DistrKeeper.ValidatorHistoricalRewards.Clear(ctx, nil)
	if err != nil {
		panic(err)
	}

	// set context height to zero
	height := ctx.BlockHeight()
	ctx = ctx.WithBlockHeight(0)

	// reinitialize all validators
	app.StakingKeeper.IterateValidators(ctx, func(_ int64, val sdk.ValidatorI) (stop bool) {
		// donate any unwithdrawn outstanding reward fraction tokens to the community pool
		operatorAddress := []byte(val.GetOperator())
		scraps, err := app.DistrKeeper.ValidatorOutstandingRewards.Get(ctx, operatorAddress)
		if err != nil {
			panic(err)
		}
		feePool, err := app.DistrKeeper.FeePool.Get(ctx)
		if err != nil {
			panic(err)
		}
		feePool.CommunityPool = feePool.CommunityPool.Add(scraps.Rewards...)
		err = app.DistrKeeper.FeePool.Set(ctx, feePool)
		if err != nil {
			panic(err)
		}

		if err := app.DistrKeeper.Hooks().AfterValidatorCreated(ctx, operatorAddress); err != nil {
			panic(err)
		}
		return false
	})

	// reinitialize all delegations
	for _, del := range dels {
		valAddr, err := sdk.ValAddressFromBech32(del.ValidatorAddress)
		if err != nil {
			panic(err)
		}
		delAddr, err := sdk.AccAddressFromBech32(del.DelegatorAddress)
		if err != nil {
			panic(err)
		}
		_ = app.DistrKeeper.Hooks().BeforeDelegationCreated(ctx, delAddr, valAddr)
		_ = app.DistrKeeper.Hooks().AfterDelegationModified(ctx, delAddr, valAddr)
	}

	// reset context height
	ctx = ctx.WithBlockHeight(height)

	/* Handle staking state. */

	// iterate through redelegations, reset creation height
	err = app.StakingKeeper.IterateRedelegations(ctx, func(_ int64, red stakingtypes.Redelegation) (stop bool) {
		for i := range red.Entries {
			red.Entries[i].CreationHeight = 0
		}
		app.StakingKeeper.SetRedelegation(ctx, red)
		return false
	})
	if err != nil {
		panic(err)
	}

	// iterate through unbonding delegations, reset creation height

	err = app.StakingKeeper.UnbondingDelegations.Walk(ctx, nil,
		func(_ collections.Pair[[]byte, []byte], ubd stakingtypes.UnbondingDelegation) (stop bool, err error) {
			for i := range ubd.Entries {
				ubd.Entries[i].CreationHeight = 0
			}
			err = app.StakingKeeper.SetUnbondingDelegation(ctx, ubd)
			if err != nil {
				return true, err
			}
			return false, nil
		})
	if err != nil {
		panic(err)
	}

	// Iterate through validators by power descending, reset bond heights, and
	// update bond intra-tx counters.
	ranger := &collections.Range[[]byte]{}
	ranger = ranger.Prefix(stakingtypes.ValidatorsKey).Descending()
	counter := int16(0)
	err = app.StakingKeeper.Validators.Walk(ctx, ranger,
		func(k []byte, val stakingtypes.Validator) (stop bool, err error) {
			addr := sdk.ValAddress(stakingtypes.AddressFromValidatorsKey(k))
			validator, getErr := app.StakingKeeper.GetValidator(ctx, addr)
			if getErr != nil {
				return true, getErr
			}

			validator.UnbondingHeight = 0
			if applyAllowedAddrs && !allowedAddrsMap[addr.String()] {
				validator.Jailed = true
			}

			app.StakingKeeper.SetValidator(ctx, validator)
			counter++
			return false, nil
		},
	)
	if err != nil {
		panic(err)
	}

	_, err = app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	if err != nil {
		log.Fatal(err)
	}

	/* Handle slashing state. */

	err = app.SlashingKeeper.ValidatorSigningInfo.Walk(ctx, nil,
		func(addr sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) (stop bool, err error) {
			info.StartHeight = 0
			err = app.SlashingKeeper.ValidatorSigningInfo.Set(ctx, addr, info)
			if err != nil {
				return true, err
			}
			return false, nil
		},
	)
	if err != nil {
		panic(err)
	}
}
