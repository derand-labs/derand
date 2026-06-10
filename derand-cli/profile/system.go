package profile

import (
	"crypto/sha256"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rlp"
)

type Form struct {
	A string `json:"a"`
	B string `json:"b"`
	C string `json:"c"`
}

type StandardSystem struct {
	D                    string `json:"D"`
	DBits                uint16 `json:"d_bits"`
	DSeed                string `json:"d_seed"`
	HashToFormGenerators []Form `json:"hash_to_form_generators"`
	HashToFormSteps      uint16 `json:"hash_to_form_steps"`
	LBits                uint16 `json:"l_bits"`
	LimbBits             uint16 `json:"limb_bits"`
	SplitExp             uint16 `json:"split_exp"`
}

func (s *StandardSystem) GetHash() hexutil.Bytes {
	b, err := rlp.EncodeToBytes(*s)
	if err != nil {
		panic(err)
	}

	h := sha256.New()
	h.Write(b)

	return h.Sum(nil)
}
