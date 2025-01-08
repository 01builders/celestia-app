# Upgrade to SDK 0.52

## Setup

Replace directives to local copies are used for now until the app builds.

```text
.
├── celestia-app
├── cosmos-sdk @ release/v0.52.x
├── ibc-apps @ d8473b7e9e39b5d35cd1024920c0878aec8775e6
└── ibc-go @ decc8ec9ae8eeda9cf3791d45d3005a6e929a990
└── ibc-go-capability @ 324e3d853ad5a88aeb3a2d9b972b9cba88d894ff
```

- ibc-apps d8473b7e9e39b5d35cd1024920c0878aec8775e6 is at <https://github.com/01builders/ibc-apps/tree/v9>
- ibc-go is main @ decc8ec9ae8eeda9cf3791d45d3005a6e929a990
- ibc-go-capability branches off `modules/capability/v1.0.1` at 324e3d853ad5a88aeb3a2d9b972b9cba88d894ff on `01builders/kocu/capability/v2`

## Goals

- Goal #1: fix import paths and go.mod until `go mod tidy` will run without error.
- Goal #2: build celestia-app
- Goal #3: regen modules using cosmos/gogoproto fork
- Goal #3: Create necessary migrations
  - crisis state (kv storekey) should be purged
  - capability state (kv / mem storekey) should be purged
- Goal #4: Re-enable rosetta

- Status: neither goal yet reached.

### Progress

### 2025-01-07

- Started <https://github.com/01builders/celestia-app/tree/sdk-v0.52.x>, also stuck on `go mod tidy` still.
- Created <https://github.com/01builders/ibc-apps/tree/v9/middleware/packet-forward-middleware> for PFM, pretty rough so far.  `go mod tidy` in middleware/packet-forward-middleware will show the progress so far.
- Created <https://github.com/01builders/ibc-go/tree/kocu/capability/v2> to try to remove some legacy module import paths so thhat PFM can build.

### 2025-01-08

- Do not use local copies of components, so easier to pick up changes cold.
- Removed simapp dependency from celestia-app
- Deleted crisis module import from celestia-app
- Delete capability module imports and scoped keepers from celestia-app
- Remove legacy proposal handlers
- Delete rosetta (should be re-enabled later)
- Update proto-builder to latest in makefile + re-generate proto files.
  buf.yaml was updated to v1 to give the correct fully qualified name
  buf.yaml was moved to proto/ because there is no excluding folder in v1 (and specs contains invalid protos)

## Problems

- SDK 0.52 has modules with `cosmossdk.io/*` import paths
- celestia-app needs ibc-go v9 (checked out at decc8ec9ae8eeda9cf3791d45d3005a6e929a990 locally) for `cosmossdk.io/*` import paths
- celestia-app also depends on `github.com/cosmos/ibc-apps/middleware/packet-forward-middleware`
- `packet-forward-middleware` depends on ibc-go.  the latest version available of PFM is v8, which uses `github.com/cosmos/cosmos-sdk/*` import paths.  therefore a PFM v9 which depends on cosmos-sdk @ 0.52 is needed.
- PFM depends on [github.com/cosmos/ibc-go/module/capability](https://github.com/cosmos/ibc-go/blob/v9.0.2/modules/capability/go.mod), from `testing/simapp`. which depends on SDK 0.50.  This module is absent in the `ibc-go @ decc8ec9ae8eeda9cf3791d45d3005a6e929a990` tree
- crisis module doesn't exist in v0.52, which is fine, but need to be thought about for the multiplexer (if in process)
