package sdkutil

import (
	"fmt"

	"cosmossdk.io/x/params/types/proposal"
	blobtypes "github.com/celestiaorg/celestia-app/v4/x/blob/types"
	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
)

// MaxBlockBytesParamChange returns a param change for the max block bytes
// consensus params.
func MaxBlockBytesParamChange(codec codec.Codec, maxBytes int) proposal.ParamChange {
	bparams := &abci.BlockParams{
		MaxBytes: int64(maxBytes),
		MaxGas:   -1,
	}
	return proposal.NewParamChange(
		baseapp.Paramspace,
		string(baseapp.ParamStoreKeyBlockParams),
		string(codec.MustMarshalJSON(bparams)),
	)
}

// GovMaxSquareSizeParamChange returns a param change for the blob module's max
// square size.
func GovMaxSquareSizeParamChange(squareSize int) proposal.ParamChange {
	return proposal.NewParamChange(
		blobtypes.ModuleName,
		string(blobtypes.KeyGovMaxSquareSize),
		fmt.Sprintf("\"%d\"", squareSize),
	)
}
