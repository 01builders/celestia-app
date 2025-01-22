# Upgrade to SDK 0.52

## Setup

```text
.
├── celestia-app
├── cosmos-sdk @ 44d09b5f4dbf398c69fe0eb4661abcee397f9823 (branch: kocu/cdev)
├── ibc-apps @ d8473b7e9e39b5d35cd1024920c0878aec8775e6
└── ibc-go @ decc8ec9ae8eeda9cf3791d45d3005a6e929a990
```

- ibc-apps d8473b7e9e39b5d35cd1024920c0878aec8775e6 is at <https://github.com/01builders/ibc-apps/tree/v9>
- ibc-go is main @ decc8ec9ae8eeda9cf3791d45d3005a6e929a990

Replace directives to local copies of ibc-apps, until PFM is ugpraded.

## Goals

- [x] Goal #1: fix import paths and go.mod until `go mod tidy` will run without error.
- [ ] Goal #2: build celestia-app
- [x] Goal #3: regen modules using cosmos/gogoproto fork
- [ ] Goal #3: Create necessary migrations
  - crisis state (kv storekey) should be purged
  - capability state (kv / mem storekey) should be purged
- [ ] Goal #4: Re-enable rosetta
- [ ] Goal #5: Re-enable upgrades and ante handlers in test/pfm
- [ ] Goal #6: Upgrade modules to core v1
- [ ] Goal #7: Upgrade proposals to gov v1 and relevant tests
- [ ] Goal #8: Create params migration
- [ ] Goal #9: Simplify code & hygiene
  - [ ] integration autocli
  - [ ] runtime x depinject
  - [ ] possibly collections for indexer support
  - [ ] address codec instead of sdk config
  - [ ] Cleanp sdk.Context to context.Context
  - [ ] Use environment services
  - [ ] Remove panics for errors
- [ ] Goal #10: Remove core logic of custom mint module to use x/mint.
  - [ ] Wrap x/mint within celestia mint and set minting function.
  - [ ] Wrap x/mint within celestia mint for extending queries and keeping query path identical
- [x] Goal #11: Replace x/paramfilter by ante handler
- [ ] Cache consensus keeper AppVersion calls as each addition is a state read
- [ ] It looks like we could totally delete their custom `start.go`, and add the checkBBR flag using the start options, as well as, wrapping the RunE method to inject their `checkBBR` function first

### app.go checklist

- [ ] Validate IBC wiring against [this simapp](https://github.com/cosmos/ibc-go/blob/main/simapp/app.go#L575). Dependent on PFM upgrade clarity.

### Progress

### 2025-01-07

- Started <https://github.com/01builders/celestia-app/tree/sdk-v0.52.x>, also stuck on `go mod tidy` still.
- Created <https://github.com/01builders/ibc-apps/tree/v9/middleware/packet-forward-middleware> for PFM, pretty rough so far.  `go mod tidy` in middleware/packet-forward-middleware will show the progress so far.
- Created <https://github.com/01builders/ibc-go/tree/kocu/capability/v2> to try to remove some legacy module import paths so thhat PFM can build. NOTE(@julienrbrt): We should not need to migrate capability.

### 2025-01-08

- Do not use local copies of components, so easier to pick up changes cold.
- Removed simapp (cosmos-sdk and ibc-go) dependency from celestia-app
- Deleted crisis module import from celestia-app
- Delete capability module imports and scoped keepers from celestia-app
- Remove legacy proposal handlers
- Delete rosetta (should be re-enabled later)
- Update proto-builder to latest in makefile + re-generate proto files.
  buf.yaml was updated to v1 to give the correct fully qualified name
  buf.yaml was moved to proto/ because there is no excluding folder in v1 (and specs contains invalid protos)
- Re-gen mocks manually (`mockgen -source=modules/core/05-port/types/module.go -package mock_types -destination ../../01builders/celestia-app/app/module/mocks/versioned_ibc.go` )

### 2025-01-09

- Comment out pfm in celestia-app for unblocking progress
- Migrate all modules to 0.52
- Remove x/paramfilter (should be replaced by ~~the circuit breaker~~ ante handler)
Made the following changes to app/app.go:

- Remove `x/capability` module
- Add `x/consensus` module
- Add `x/accounts` module
- Add `ibc-29/fee` module
- Add `x/protocolpool` module
- Update API breaks in existing keeper constructions
- Remove in memory keys (no longer required with removal of x/capability)
- Modify EndBlocker to reflect removal of `x/params`

In app/modules.go:

- Remove ModuleBasics references
- Remove `x/capability` module
- Remove `x/crisis` module

In the root command, began to reason about and fix genesis commands including DefaultGenesis.

### 2025-01-13

- [Fix build fixes in app.go and module manager](https://github.com/01builders/celestia-app/pull/1)

### 2025-01-14

- Continue fixing build issues

### 2025-01-15

- Fix build issue ante handlers
- TxSizeGas ante handler doesn't check for accounts anymore, in accordance with v0.52 ante handler behavior
- Migrate `BlockHeader().Version.App` to consensus keeper appversion
- Replace x/paramfilter by custom ante handler

### 2025-01-17

- Remove test/pfm
- More build issue fix

### 2025-01-21/22

- Upgrade to cometbft/cometbft
- Fix more build issues
- Remove custom start.go in favor of SDK's

## Problems

- SDK 0.52 has modules with `cosmossdk.io/*` import paths
- celestia-app needs ibc-go v9 (checked out at decc8ec9ae8eeda9cf3791d45d3005a6e929a990 locally) for `cosmossdk.io/*` import paths
- celestia-app also depends on `github.com/cosmos/ibc-apps/middleware/packet-forward-middleware`
- `packet-forward-middleware` depends on ibc-go.  the latest version available of PFM is v8, which uses `github.com/cosmos/cosmos-sdk/*` import paths.  therefore a PFM v9 which depends on cosmos-sdk @ 0.52 is needed.
- PFM depends on [github.com/cosmos/ibc-go/module/capability](https://github.com/cosmos/ibc-go/blob/v9.0.2/modules/capability/go.mod), from `testing/simapp`. which depends on SDK 0.50.  This module is absent in the `ibc-go @ decc8ec9ae8eeda9cf3791d45d3005a6e929a990` tree. PFM needs to be refactored to work without capability. It is unclear from IBC documentation what is the future of this module. PFM tests have been removed.

## Upstream

- <https://github.com/cosmos/cosmos-sdk/pull/23318> - Fixes encoding by providing support for indexWrapperDecoder
