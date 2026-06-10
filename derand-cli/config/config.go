package config

import (
	"derand-cli/backend"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path"
	"path/filepath"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

var dir = ""
var vdfdir = ""
var builddir = "./build"

func GetVDFDir() string {
	if dir == "" {
		var err error
		dir, err = os.UserConfigDir()
		if err != nil {
			panic(err)
		}

		dir = path.Join(dir, "derand")
	}

	if vdfdir == "" {
		vdfdir = path.Join(dir, ".vdf")
	}

	return vdfdir
}

func GetBuildDir() string {
	return builddir
}

type Config struct {
	CurrentChainIndex  int                         `json:"current_chain_index"`
	CurrentWalletIndex int                         `json:"current_wallet_index"`
	Chains             []ChainInfo                 `json:"chains"`
	Wallets            []WalletInfo                `json:"wallets"`
	LocalProfiles      map[string]LocalProfileInfo `json:"local_profiles"`
}

func Load() (*Config, error) {
	data, err := os.ReadFile(getConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("not found configuration file, please run `derand init` first")
		}

		return nil, err
	}

	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *Config) Save() error {
	f, err := os.OpenFile(getConfigPath(), os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(c); err != nil {
		return err
	}

	return nil
}

func (c *Config) Create() error {
	fdata, err := os.OpenFile(getConfigPath(), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("configuration existed, please run `derand cleanup` first")
		}

		return err
	}

	if err := json.NewEncoder(fdata).Encode(c); err != nil {
		return err
	}

	return nil
}

func (c *Config) GetCurrentChainBackend() (*backend.BackendPool, error) {
	currentChain, err := c.GetCurrentChain()
	if err != nil {
		return nil, err
	}

	return currentChain.GetBackend()
}

func (c *Config) GetCurrentDefaultTxOpts(password string) (*bind.TransactOpts, error) {
	currentChain, err := c.GetCurrentChain()
	if err != nil {
		return nil, err
	}

	currentWallet, err := c.GetCurrentWallet()
	if err != nil {
		return nil, err
	}

	privateKey, err := currentWallet.GetPrivKey(password)
	if err != nil {
		return nil, err
	}

	return bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(int64(currentChain.ChainID)))
}

func (c *Config) GetCurrentChain() (*ChainInfo, error) {
	if c.CurrentChainIndex == 0 && len(c.Chains) == 0 {
		return nil, fmt.Errorf("no chain configured")
	}

	if c.CurrentChainIndex < 0 || c.CurrentChainIndex >= len(c.Chains) {
		return nil, fmt.Errorf("invalid configuration: chain index is out of bound")
	}

	return &c.Chains[c.CurrentChainIndex], nil
}

func (c *Config) GetChain(index int) (*ChainInfo, error) {
	if len(c.Chains) == 0 {
		return nil, fmt.Errorf("no chain configured")
	}

	if index < 0 || index >= len(c.Chains) {
		return nil, fmt.Errorf("invalid index: out of bound")
	}

	return &c.Chains[index], nil
}

func (c *Config) GetCurrentWallet() (*WalletInfo, error) {
	if c.CurrentWalletIndex == 0 && len(c.Wallets) == 0 {
		return nil, fmt.Errorf("no wallet configured")
	}

	if c.CurrentWalletIndex < 0 || c.CurrentWalletIndex >= len(c.Wallets) {
		return nil, fmt.Errorf("invalid configuration: wallet index is out of bound")
	}

	return &c.Wallets[c.CurrentWalletIndex], nil
}

func (c *Config) GetWallet(index int) (*WalletInfo, error) {
	if len(c.Wallets) == 0 {
		return nil, fmt.Errorf("no wallet configured")
	}

	if index < 0 || index >= len(c.Wallets) {
		return nil, fmt.Errorf("invalid index: out of bound")
	}

	return &c.Wallets[index], nil
}

func Cleanup() error {
	if err := os.Remove(getConfigPath()); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("not found configuration file")
		}

		return err
	}

	return nil
}

func SetConfigDir(d string) {
	dir = d
}

func SetVDFDir(d string) {
	vdfdir = d
}

func SetBuildDir(d string) {
	builddir = d
}

func getConfigPath() string {
	if dir == "" {
		var err error
		dir, err = os.UserConfigDir()
		if err != nil {
			panic(err)
		}

		dir = path.Join(dir, "derand")
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		panic(err)
	}

	return filepath.Join(dir, "data.json")
}
