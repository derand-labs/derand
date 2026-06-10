package proverlogic

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
)

var proofABI abi.Type
var yABI abi.Type

func init() {
	var err error
	proofABI, err = abi.NewType(
		"tuple",
		"",
		[]abi.ArgumentMarshaling{
			{
				Name: "y",
				Type: "tuple",
				Components: []abi.ArgumentMarshaling{
					{Name: "asign", Type: "int8"},
					{Name: "bsign", Type: "int8"},
					{Name: "csign", Type: "int8"},
					{Name: "a", Type: "uint128[]"},
					{Name: "b", Type: "uint128[]"},
					{Name: "c", Type: "uint128[]"},
				},
			},
			{
				Name: "pi",
				Type: "tuple",
				Components: []abi.ArgumentMarshaling{
					{Name: "asign", Type: "int8"},
					{Name: "bsign", Type: "int8"},
					{Name: "csign", Type: "int8"},
					{Name: "a", Type: "uint128[]"},
					{Name: "b", Type: "uint128[]"},
					{Name: "c", Type: "uint128[]"},
				},
			},
			{
				Name: "deriveChallengeTranscript",
				Type: "uint128[]",
			},
			{
				Name: "zkProof",
				Type: "bytes",
			},
		},
	)
	if err != nil {
		panic(err)
	}

	yABI, err = abi.NewType(
		"tuple",
		"",
		[]abi.ArgumentMarshaling{
			{Name: "asign", Type: "int8"},
			{Name: "bsign", Type: "int8"},
			{Name: "csign", Type: "int8"},
			{Name: "a", Type: "uint128[]"},
			{Name: "b", Type: "uint128[]"},
			{Name: "c", Type: "uint128[]"},
		},
	)
	if err != nil {
		panic(err)
	}
}
