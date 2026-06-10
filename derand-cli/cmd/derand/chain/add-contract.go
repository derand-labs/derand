package chain

import (
	"derand-cli/config"
	"derand-cli/utils"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

var (
	flagSetupDeRandAddress         string
	flagSetupDeRandTxHash          string
	flagSetupDeRandTreasuryAddress string
	flagSetupDeRandTreasuryTarget  string
	flagSetupHashToPrime128Address string
	flagSetupHashToPrime128TxHash  string
)

var setupCmd = &cobra.Command{
	Use:   "add-contract",
	Short: "add derand address of chain",

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		currentChain, err := cfg.GetCurrentChain()
		if err != nil {
			return err
		}

		var derand, hashToPrime128 *config.ContractInfo
		if flagSetupDeRandAddress != "" {
			derand = new(config.ContractInfo)
			if err := derand.Address.UnmarshalText([]byte(flagSetupDeRandAddress)); err != nil {
				return fmt.Errorf("invalid DeRand address: %w", err)
			}
			if err := derand.DeployTxHash.UnmarshalText([]byte(flagSetupDeRandTxHash)); err != nil {
				return fmt.Errorf("invalid DeRand tx hash: %w", err)
			}
		}

		if flagSetupHashToPrime128Address != "" {
			hashToPrime128 = new(config.ContractInfo)
			if err := hashToPrime128.Address.UnmarshalText([]byte(flagSetupHashToPrime128Address)); err != nil {
				return fmt.Errorf("invalid HashToPrime128 address: %w", err)
			}
			if err := hashToPrime128.Address.UnmarshalText([]byte(flagSetupHashToPrime128TxHash)); err != nil {
				return fmt.Errorf("invalid HashToPrime128 tx hash: %w", err)
			}
		}

		if derand != nil {
			targetInGwei, err := utils.ETHStringToGwei(flagSetupDeRandTreasuryTarget)
			if err != nil {
				return err
			}

			currentChain.DeRand = derand
			currentChain.DeRandTreasuryAddress = common.HexToAddress(flagSetupDeRandTreasuryAddress)
			currentChain.DeRandTreasuryTarget = targetInGwei.Uint64()

			// Reset remote profile map if deploy new smart contract
			currentChain.RemoteProfileMap = make(map[int]string)
		}

		if hashToPrime128 != nil {
			currentChain.HashToPrime128 = append(currentChain.HashToPrime128, *hashToPrime128)
		}

		if err := cfg.Save(); err != nil {
			return err
		}

		utils.PrintTitle("OK")
		return nil
	},
}

func init() {
	setupCmd.Flags().StringVar(&flagSetupDeRandAddress, "derand", "", "DeRand smart contract address")
	setupCmd.Flags().StringVar(&flagSetupDeRandTxHash, "derand-tx", "", "DeRand smart contract deployment transaction")
	setupCmd.Flags().StringVar(&flagSetupDeRandTreasuryAddress, "derand-treasury-address", "", "DeRand treasury address")
	setupCmd.Flags().StringVar(&flagSetupDeRandTreasuryTarget, "derand-treasury-target", "0", "DeRand treasury target")
	setupCmd.Flags().StringVar(&flagAddDeRandTxHash, "hash-to-prime-128", "", "HashToPrime128 contract address")
	setupCmd.Flags().StringVar(&flagSetupHashToPrime128TxHash, "hash-to-prime-128-tx", "", "HashToPrime128 contract deployment transaction")

	setupCmd.MarkFlagsOneRequired("derand", "hash-to-prime-128")
	setupCmd.MarkFlagsRequiredTogether("derand", "derand-tx", "derand-treasury-address", "derand-treasury-target")
	setupCmd.MarkFlagsRequiredTogether("hash-to-prime-128", "hash-to-prime-128-tx")
}
