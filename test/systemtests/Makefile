#!/usr/bin/make -f

WAIT_TIME ?= 45s

all: test format

test:
	go test -mod=readonly -failfast -timeout=15m -tags='system_test' ./... --wait-time=$(WAIT_TIME) --binary=celestia-appd --verbose

format:
	find . -name '*.go' -type f  | xargs gofumpt -w
	find . -name '*.go' -type f  | xargs misspell -w
	find . -name '*.go' -type f  | xargs gci write --skip-generated -s standard -s default -s "prefix(cosmossdk.io)" -s "prefix(github.com/cosmos/cosmos-sdk)" -s "github.com/celestiaorg/celestia-app/v4" --custom-order

.PHONY: all test format
