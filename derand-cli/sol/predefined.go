package sol

var fixedSetting = solcSettings{
	Optimizer: struct {
		Enabled bool `json:"enabled"`
		Runs    int  `json:"runs"`
	}{
		Enabled: true,
		Runs:    1000,
	},
	ViaIR:      true,
	EVMVersion: "prague",
	Metadata: struct {
		BytecodeHash string `json:"bytecodeHash"`
	}{
		BytecodeHash: "ipfs",
	},
	OutputSelection: map[string]map[string][]string{
		"*": {
			"*": {
				"abi",
				"evm.bytecode",
				"evm.deployedBytecode",
				"metadata",
			},
		},
	},
}
