package main

import (
	"derand-cli/config"
	"derand-cli/gen"
	"derand-cli/utils"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

var (
	flagDeployDeRand               bool
	flagDeployDeRandTreasury       string
	flagDeployDeRandTreasuryTarget string
	flagDeployHashToPrime128       bool
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "deploy smart contract",

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		currentChain, err := cfg.GetCurrentChain()
		if err != nil {
			return err
		}

		currentWallet, err := cfg.GetCurrentWallet()
		if err != nil {
			return err
		}

		password := utils.AskPassword("Enter password: ")
		backend, err := cfg.GetCurrentChainBackend()
		if err != nil {
			return err
		}

		txOpts, err := cfg.GetCurrentDefaultTxOpts(password)
		if err != nil {
			return err
		}

		if flagDeployDeRand {
			if currentChain.DeRand != nil {
				if !utils.Confirm("Contract DeRand was already deployed. Continue? [y/N]: ") {
					return fmt.Errorf("cancelled")
				}
			}

			treasuryAddress, err := currentWallet.GetAddress()
			if err != nil {
				return err
			}
			if flagDeployDeRandTreasury != "" {
				treasuryAddress = common.HexToAddress(flagDeployDeRandTreasury)
			}

			treasuryTargetInWei, err := utils.ETHStringToWei(flagDeployDeRandTreasuryTarget)
			if err != nil {
				return err
			}

			utils.PrintTitle("DeRand deploy constructor:")
			utils.PrintSubtitle("Treasury address:", treasuryAddress)
			utils.PrintSubtitle("Treasury target:", utils.WeiToETHString(treasuryTargetInWei), currentChain.Symbol)

			if utils.Confirm("Deploy DeRand? [y/N]: ") {
				txOpts.GasLimit = 5000000

				treasuryTargetETHInGwei := utils.WeiToGwei(treasuryTargetInWei)
				addr, tx, _, err := gen.DeployDeRand(txOpts, backend, treasuryAddress, treasuryTargetETHInGwei.Uint64())
				if err != nil {
					return err
				}

				utils.PrintTitle(utils.Bold("Deployed Derand:"), utils.Bold(utils.Green("OK")))
				utils.PrintSubtitle("Contract Address", addr)
				utils.PrintSubtitle("Transaction Hash:", tx.Hash())
				utils.PrintSubtitle("Cost:", utils.WeiToETHString(tx.Cost()))
				utils.PrintSubtitle("Size:", len(tx.Data()))

				currentChain.DeRand = &config.ContractInfo{
					Address:      addr,
					DeployTxHash: tx.Hash(),
				}
				currentChain.DeRandTreasuryAddress = treasuryAddress
				currentChain.DeRandTreasuryTarget = treasuryTargetETHInGwei.Uint64()

				// Reset remote profile map if deploy new smart contract
				currentChain.RemoteProfileMap = make(map[int]string)
			}
		}

		if flagDeployHashToPrime128 {
			if currentChain.HashToPrime128 != nil {
				if !utils.Confirm("Contract HashToPrime128 was already deployed. Continue? [y/N]: ") {
					return fmt.Errorf("cancelled")
				}
			}

			if utils.Confirm("Deploy HashToPrime128? [y/N]: ") {
				txOpts.GasLimit = 1000000
				addr, tx, _, err := gen.DeployHashToPrime128(txOpts, backend)
				if err != nil {
					return err
				}

				utils.PrintTitle(utils.Bold("Deployed HashToPrime128:"), utils.Bold(utils.Green("OK")))
				utils.PrintSubtitle("Contract Address", addr)
				utils.PrintSubtitle("Transaction Hash:", tx.Hash())
				utils.PrintSubtitle("Cost:", utils.WeiToETHString(tx.Cost()))
				utils.PrintSubtitle("Size:", len(tx.Data()))

				currentChain.HashToPrime128 = append(currentChain.HashToPrime128, config.ContractInfo{
					Address:      addr,
					DeployTxHash: tx.Hash(),
				})
			}

		}

		if err := cfg.Save(); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	deployCmd.Flags().BoolVar(&flagDeployDeRand, "derand", false, "deploy DeRand")
	deployCmd.Flags().StringVar(&flagDeployDeRandTreasury, "derand-treasury-address", "", "derand treasury address")
	deployCmd.Flags().StringVar(&flagDeployDeRandTreasuryTarget, "derand-treasury-target", "0", "derand treasury target")
	deployCmd.Flags().BoolVar(&flagDeployHashToPrime128, "hash-to-prime-128", false, "deploy HashToPrime128")

	deployCmd.MarkFlagsOneRequired("derand", "hash-to-prime-128")
}
