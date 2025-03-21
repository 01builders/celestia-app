//go:build multiplexer

package cmd

import (
	"github.com/01builders/nova/abci"
	"github.com/01builders/nova/appd"

	embedding "github.com/celestiaorg/celestia-app/v4/internal/embedding"
)

func Versions() abci.Versions {
	v3AppBinary, err := embedding.CelestiaAppV3()
	if err != nil {
		panic(err)
	}
	v3, err := appd.New("v3", v3AppBinary)
	if err != nil {
		panic(err)
	}

	v4AppBinary, err := embedding.CelestiaAppV4()
	if err != nil {
		panic(err)
	}

	v4, err := appd.New("v4", v4AppBinary)
	if err != nil {
		panic(err)
	}

	return abci.Versions{
		{
			ABCIVersion: abci.ABCIClientVersion1,
			Appd:        v3,
			AppVersion:  3,
		},
		{
			ABCIVersion: abci.ABCIClientVersion2,
			Appd:        v4,
			AppVersion:  4,
		},
	}
}
