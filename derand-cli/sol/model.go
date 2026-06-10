package sol

import (
	"encoding/json"
)

type solcInput struct {
	Language string                `json:"language"`
	Sources  map[string]sourceUnit `json:"sources"`
	Settings solcSettings          `json:"settings"`
}

type sourceUnit struct {
	Content string `json:"content"`
}

type solcSettings struct {
	Optimizer struct {
		Enabled bool `json:"enabled"`
		Runs    int  `json:"runs"`
	} `json:"optimizer"`

	ViaIR      bool   `json:"viaIR"`
	EVMVersion string `json:"evmVersion"`

	Metadata struct {
		BytecodeHash string `json:"bytecodeHash"`
	} `json:"metadata"`

	OutputSelection map[string]map[string][]string `json:"outputSelection"`
}

type SolcOutput struct {
	Contracts map[string]map[string]struct {
		ABI json.RawMessage `json:"abi"`
		EVM struct {
			Bytecode struct {
				Object string `json:"object"`
			} `json:"bytecode"`
			DeployedBytecode struct {
				Object string `json:"object"`
			} `json:"deployedBytecode"`
		} `json:"evm"`
		Metadata string `json:"metadata"`
	} `json:"contracts"`
	Errors []map[string]any `json:"errors"`
}
