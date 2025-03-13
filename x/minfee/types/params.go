package types

import (
	"fmt"

	"cosmossdk.io/math"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/celestiaorg/celestia-app/v4/pkg/appconsts"
)

var _ paramtypes.ParamSet = (*LegacyParams)(nil)

var (
	KeyNetworkMinGasPrice     = []byte("NetworkMinGasPrice")
	DefaultNetworkMinGasPrice math.LegacyDec
)

func init() {
	DefaultNetworkMinGasPriceDec, err := math.LegacyNewDecFromStr(fmt.Sprintf("%f", appconsts.DefaultNetworkMinGasPrice))
	if err != nil {
		panic(err)
	}
	DefaultNetworkMinGasPrice = DefaultNetworkMinGasPriceDec
}

// RegisterMinFeeParamTable returns a subspace with a key table attached.
func RegisterMinFeeParamTable(subspace paramtypes.Subspace) paramtypes.Subspace {
	if subspace.HasKeyTable() {
		return subspace
	}
	return subspace.WithKeyTable(ParamKeyTable())
}

type LegacyParams struct {
	NetworkMinGasPrice math.LegacyDec `json:"network_min_gas_price" yaml:"network_min_gas_price"`
}

// Validate validates the set of params
func (p LegacyParams) Validate() error {
	return ValidateMinGasPrice(p.NetworkMinGasPrice)
}

// ParamKeyTable returns the param key table for the minfee module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&LegacyParams{})
}

// ParamSetPairs gets the param key-value pair
func (p *LegacyParams) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyNetworkMinGasPrice, &p.NetworkMinGasPrice, ValidateMinGasPrice),
	}
}

// Validate validates the param type
func ValidateMinGasPrice(i interface{}) error {
	_, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

// Validate validates the set of params
func (p Params) Validate() error {
	return ValidateMinGasPrice(p.NetworkMinGasPrice)
}

// DefaultParams returns the default parameters for the module.
func DefaultParams() Params {
	return Params{
		NetworkMinGasPrice: DefaultNetworkMinGasPrice,
	}
}

// NewParams creates a new instance of Params with the provided NetworkMinGasPrice.
func NewParams(networkMinGasPrice math.LegacyDec) Params {
	return Params{
		NetworkMinGasPrice: networkMinGasPrice,
	}
}
