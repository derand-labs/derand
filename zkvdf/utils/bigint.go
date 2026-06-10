package utils

import (
	"math/big"
)

func BigIntFromString(s string, base int) *big.Int {
	a, ok := new(big.Int).SetString(s, base)
	if !ok {
		panic("invalid bigint hex: " + s)
	}

	return a
}

func MaxValue(bits int) *big.Int {
	m := Modulus(bits)
	m.Sub(m, big.NewInt(1))
	return m
}

func Modulus(bits int) *big.Int {
	return new(big.Int).Lsh(big.NewInt(1), uint(bits))
}
