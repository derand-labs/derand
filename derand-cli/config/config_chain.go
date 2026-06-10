package config

import (
	"derand-cli/backend"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

type ChainInfo struct {
	Name                   string            `json:"name"`
	ChainID                int               `json:"chain_id"`
	Symbol                 string            `json:"symbol"`
	RPCs                   []string          `json:"rpcs"`
	WSRPCs                 []string          `json:"ws_rpcs"`
	DeRand                 *ContractInfo     `json:"derand"`
	DeRandTreasuryAddress  ethcommon.Address `json:"derand_treasury_address"`
	DeRandTreasuryTarget   uint64            `json:"derand_treasury_target"`
	HashToPrime128         []ContractInfo    `json:"hash_to_prime_128"`
	RemoteProfileMap       map[int]string    `json:"remote_profile_map"`
	LastProverWatchedBlock uint64            `json:"last_prover_watched_block"`
}

func (c *ChainInfo) GetBackend() (*backend.BackendPool, error) {
	return backend.NewPool(c.RPCs, c.WSRPCs)
}

type ContractInfo struct {
	Address      ethcommon.Address `json:"address"`
	DeployTxHash ethcommon.Hash    `json:"deploy_tx_hash"`
}
