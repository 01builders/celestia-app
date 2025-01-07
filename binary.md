# Upgrade to SDK 0.52

## Setup

Replace directives to local copies are used for now until the app builds.

```text
├── celestia-app
├── cosmos-sdk @ release/v0.52.x
├── ibc-apps @ d8473b7e9e39b5d35cd1024920c0878aec8775e6
└── ibc-go @ decc8ec9ae8eeda9cf3791d45d3005a6e929a990
```

ibc-apps d8473b7e9e39b5d35cd1024920c0878aec8775e6 is at <https://github.com/01builders/ibc-apps/tree/v9>

## Progress

- Goal #1: fix import paths and go.mod until `go mod tidy` will run without error.
- Goal #2: build celestia-app
- Status: neither goal yet reached.

### WIP

- Created <https://github.com/01builders/ibc-apps/tree/v9> for PFM, pretty rough so far.  `go mod tidy` in middleware/packet-forward-middleware will show the progress so far.
- Started <https://github.com/01builders/celestia-app/tree/sdk-v0.52.x>, also stuck on `go mod tidy` still.

## Problems

- SDK 0.52 has modules with `cosmossdk.io/*` import paths
- celestia-app needs ibc-go v9 (checked out at decc8ec9ae8eeda9cf3791d45d3005a6e929a990 locally) for `cosmossdk.io/*` import paths
- celestia-app also depends on `github.com/cosmos/ibc-apps/middleware/packet-forward-middleware`
- `packet-forward-middleware` depends on ibc-go.  the latest version available of PFM is v8, which uses `github.com/cosmos/cosmos-sdk/*` import paths.  therefore a PFM v9 which depends on cosmos-sdk @ 0.52 is needed.
- PFM depends on [github.com/cosmos/ibc-go/module/capability](https://github.com/cosmos/ibc-go/blob/v9.0.2/modules/capability/go.mod), from `testing/simapp`. which depends on SDK 0.50.  This module is absent in the `ibc-go @ decc8ec9ae8eeda9cf3791d45d3005a6e929a990` tree
