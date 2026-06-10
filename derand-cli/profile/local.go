package profile

import (
	"errors"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

var ErrNotFound = errors.New("not found")

type ProfileWarningLevel int

const (
	ProfileWarningLevelLow ProfileWarningLevel = iota
	ProfileWarningLevelMedium
	ProfileWarningLevelHigh
	ProfileWarningLevelCritical
)

type ProfileWarning struct {
	Level   ProfileWarningLevel
	Message string
}

type ContractInfo struct {
	Address      ethcommon.Address `json:"address"`
	DeployTxHash ethcommon.Hash    `json:"deploy_tx_hash"`
}

type LocalProfile struct {
	Type                           string                                      `json:"type"`
	StandardClassgroupZKPlonkBn254 *StandardClassgroupZKPlonkBn254LocalProfile `json:"standard_classgroup_zk_plonk_bn254"`
	VerifierByChainID              map[int]ContractInfo                        `json:"verifier_by_chain_id"`
}
