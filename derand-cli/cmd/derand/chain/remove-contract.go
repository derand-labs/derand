package chain

import (
	"derand-cli/config"
	"derand-cli/utils"
	"slices"

	"github.com/spf13/cobra"
)

var (
	flagRemoveContractHashToPrime128Index int
)

var removeContractCmd = &cobra.Command{
	Use:   "remove-contract",
	Short: "remove a hash to prime 128 contract",

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		currentChain, err := cfg.GetCurrentChain()
		if err != nil {
			return err
		}

		if flagRemoveContractHashToPrime128Index != -1 {
			currentChain.HashToPrime128 = slices.Delete(
				currentChain.HashToPrime128,
				flagRemoveContractHashToPrime128Index,
				flagRemoveContractHashToPrime128Index+1,
			)
		}

		if err := cfg.Save(); err != nil {
			return err
		}

		utils.PrintTitle("OK")
		return nil
	},
}

func init() {
	removeContractCmd.Flags().IntVar(&flagRemoveContractHashToPrime128Index, "hash-to-prime-128-index", -1, "index of HashToPrime128 contract")
	setupCmd.MarkFlagRequired("hash-to-prime-128-index")
}
