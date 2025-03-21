//go:build multiplexer

package cmd

import (
	"github.com/01builders/nova/abci"
	"github.com/01builders/nova/appd"

	embedding "github.com/celestiaorg/celestia-app/v4/internal/embedding"
)

func Versions() abci.Versions {
	v3AppBinary, err := embedding.CelestiaAppV3()
	_ = err // TODO: handle this error, explicitly ignoring this for now as ledger tests fail due to not having the binary

	v3, err := appd.New("v3", v3AppBinary)
	_ = err // TODO: handle this error, explicitly ignoring this for now as ledger tests fail due to not having the binary

	v4AppBinary, err := embedding.CelestiaAppV4()
	_ = err // TODO: handle this error, explicitly ignoring this for now as ledger tests fail due to not having the binary

	v4, err := appd.New("v4", v4AppBinary)
	_ = err // TODO: handle this error, explicitly ignoring this for now as ledger tests fail due to not having the binary

	return abci.Versions{
		{
			ABCIVersion: abci.ABCIClientVersion1,
			Appd:        v3,
			AppVersion:  3,
			StartArgs: []string{
				"--grpc.enable=true",
				"--api.enable=true",
				"--api.swagger=false",
				"--with-tendermint=false",
				"--transport=grpc",
				"--v2-upgrade-height=5",
			},
		},
		{
			ABCIVersion: abci.ABCIClientVersion2,
			Appd:        v4,
			AppVersion:  4,
		},
	}
}
