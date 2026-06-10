package main

import (
	"bytes"
	"derand-cli/config"
	"derand-cli/gen"
	"derand-cli/utils"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "verify smart contract",

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		currentChain, err := cfg.GetCurrentChain()
		if err != nil {
			return err
		}

		backend, err := cfg.GetCurrentChainBackend()
		if err != nil {
			return err
		}

		if currentChain.DeRand == nil {
			utils.PrintTitle(utils.Bold("DeRand:"), utils.Yellow("NOT INSTALLED YET"))
		} else {
			abi, err := abi.JSON(strings.NewReader(gen.DeRandABI))
			if err != nil {
				return err
			}

			constructorData, err := abi.Pack("", currentChain.DeRandTreasuryAddress, currentChain.DeRandTreasuryTarget)
			if err != nil {
				return err
			}

			fullByteCode := append(common.Hex2Bytes(gen.DeRandBin[2:]), constructorData...)

			tx, _, err := backend.TransactionByHash(cmd.Context(), currentChain.DeRand.DeployTxHash)
			if err != nil {
				return err
			}

			code, err := backend.CodeAt(cmd.Context(), currentChain.DeRand.Address, nil)
			if err != nil {
				return err
			}

			if len(code) == 0 {
				utils.PrintTitle(utils.Bold("Verify DeRand:"), utils.Yellow("NOT FOUND"))
			} else if !bytes.Equal(fullByteCode, tx.Data()) {
				utils.PrintTitle(utils.Bold("Verify DeRand:"), utils.Red("MISMATCH"))
			} else {
				utils.PrintTitle(utils.Bold("Verify DeRand:"), utils.Green("MATCH"))
			}
		}

		if len(currentChain.HashToPrime128) == 0 {
			utils.PrintTitle(utils.Bold("HashToPrime128:"), utils.Yellow("NOT INSTALLED YET"))
		} else {
			for i, info := range currentChain.HashToPrime128 {
				code, err := backend.CodeAt(cmd.Context(), info.Address, nil)
				if err != nil {
					return err
				}

				if len(code) == 0 {
					utils.PrintTitle(utils.Bold("Verify HashToPrime128:"), utils.Yellow("NOT FOUND"))
				} else if !bytes.Equal(code, common.Hex2Bytes(gen.HashToPrime128DeployedBin[2:])) {
					utils.PrintTitle(utils.Bold(fmt.Sprintf("Verify HashToPrime128 [%d]:", i)), utils.Red("MISMATCH"))
				} else {
					utils.PrintTitle(utils.Bold(fmt.Sprintf("Verify HashToPrime128 [%d]:", i)), utils.Green("MATCH"))
				}
			}
		}

		return nil
	},
}
