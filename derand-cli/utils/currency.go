package utils

import (
	"fmt"
	"math/big"
)

func WeiToETHString(b *big.Int) string {
	return new(big.Float).Quo(new(big.Float).SetInt(b), big.NewFloat(1e18)).Text('f', 9)
}

func ETHStringToWei(s string) (*big.Int, error) {
	amountF, ok := new(big.Float).SetString(s)
	if !ok {
		return nil, fmt.Errorf("invalid amount argument")
	}

	amount, _ := new(big.Float).Mul(amountF, big.NewFloat(1e18)).Int(new(big.Int))
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return nil, fmt.Errorf("require a positive amount")
	}

	return amount, nil
}

func ETHStringToGwei(s string) (*big.Int, error) {
	wei, err := ETHStringToWei(s)
	if err != nil {
		return nil, err
	}

	return WeiToGwei(wei), nil
}

func WeiToGwei(e *big.Int) *big.Int {
	return new(big.Int).Div(e, big.NewInt(1e9))
}
