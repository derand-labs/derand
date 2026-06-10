package main

import (
	"derand-cli/config"
	"derand-cli/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init configuration",

	RunE: func(cmd *cobra.Command, args []string) error {
		initCfg := config.Config{
			CurrentChainIndex: 0,
			Chains: []config.ChainInfo{
				{
					Name:    "Sepolia Testnet",
					Symbol:  "ETH",
					ChainID: 11155111,
					RPCs: []string{
						"https://ethereum-sepolia-rpc.publicnode.com",
					},
					WSRPCs: []string{
						"wss://ethereum-sepolia-rpc.publicnode.com",
					},
					DeRand: &config.ContractInfo{
						Address:      common.HexToAddress("0x18e6cB395F2d7177701e92F0f602019e654dDc0F"),
						DeployTxHash: common.HexToHash("0x6d89ae39c2de06a2a662b11dcace95404b4f1203c5c584f101186a01141ebfa7"),
					},
					DeRandTreasuryAddress: common.HexToAddress("0x01e13f16AF7E6247d59A4d3395bCF0DA337AE717"),
					DeRandTreasuryTarget:  10000000000,
					HashToPrime128: []config.ContractInfo{{
						Address:      common.HexToAddress("0x05210d816F2E8B22D0249F31Bc46757A69a6f88e"),
						DeployTxHash: common.HexToHash("0xcb8041b8b378acf90a22448ce324e542287a26d059114813a3f952ae9d9d57da"),
					}},
				},
				{
					Name:           "Anvil Local Chain",
					Symbol:         "ETH",
					ChainID:        31337,
					RPCs:           []string{"http://127.0.0.1:8545"},
					WSRPCs:         []string{"ws://127.0.0.1:8545"},
					DeRand:         nil,
					HashToPrime128: nil,
				},
			},
		}

		if err := initCfg.Create(); err != nil {
			return err
		}

		utils.PrintTitle("OK")
		return nil
	},
}
