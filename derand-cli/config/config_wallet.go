package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethhexutil "github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/scrypt"
)

type WalletInfo struct {
	Name                string           `json:"name"`
	PublicKey           ethhexutil.Bytes `json:"public_key"`
	Salt                ethhexutil.Bytes `json:"salt"`
	Nonce               ethhexutil.Bytes `json:"nonce"`
	EncryptedPrivateKey ethhexutil.Bytes `json:"encrypted_private_key"`
}

func (w *WalletInfo) GetAddress() (common.Address, error) {
	pk, err := w.GetPubKey()
	if err != nil {
		return common.Address{}, err
	}

	return crypto.PubkeyToAddress(*pk), nil
}

func (w *WalletInfo) GetPubKey() (*ecdsa.PublicKey, error) {
	return crypto.UnmarshalPubkey(w.PublicKey)
}

func (w *WalletInfo) GetPrivKey(password string) (*ecdsa.PrivateKey, error) {
	key, err := scrypt.Key([]byte(password), w.Salt, 32768, 8, 1, 32)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	b, err := gcm.Open(nil, w.Nonce, w.EncryptedPrivateKey, nil)
	if err != nil {
		return nil, fmt.Errorf("incorrect password")
	}

	return crypto.ToECDSA(b)
}

func (c *WalletInfo) GetDefaultTxOpts(chainID int, password string) (*bind.TransactOpts, error) {
	privateKey, err := c.GetPrivKey(password)
	if err != nil {
		return nil, err
	}

	return bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(int64(chainID)))
}
