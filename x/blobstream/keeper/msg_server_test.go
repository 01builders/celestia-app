package keeper_test

import (
	"testing"

	testutil "github.com/celestiaorg/celestia-app/v3/test/util"
	"github.com/celestiaorg/celestia-app/v3/x/blobstream/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestRegisterEVMAddress(t *testing.T) {
	input, sdkCtx := testutil.SetupFiveValChain(t)
	k := input.BlobstreamKeeper
	vals, err := input.StakingKeeper.GetValidators(sdkCtx, 100)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(vals), 1)
	val := vals[0]
	valAddr, err := input.StakingKeeper.ValidatorAddressCodec().StringToBytes(val.GetOperator())
	require.NoError(t, err)

	evmAddr, exists := k.GetEVMAddress(sdkCtx, valAddr)
	require.True(t, exists)

	// test again with an address that is not the validator
	valAddr2, err := sdk.ValAddressFromBech32("celestiavaloper1xcy3els9ua75kdm783c3qu0rfa2eplestc6sqc")
	require.NoError(t, err)
	msg := types.NewMsgRegisterEVMAddress(valAddr2, evmAddr)

	_, err = k.RegisterEVMAddress(sdkCtx, msg)
	require.Error(t, err)

	// override the previous EVM address with a new one
	evmAddr = common.BytesToAddress([]byte("evm_address"))
	msg = types.NewMsgRegisterEVMAddress(valAddr, evmAddr)
	_, err = k.RegisterEVMAddress(sdkCtx, msg)
	require.NoError(t, err)

	addr, _ := k.GetEVMAddress(sdkCtx, valAddr)
	require.Equal(t, evmAddr, addr)
}
