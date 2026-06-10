package balance

import (
	"derand-cli/config"
	"derand-cli/gen"
	"derand-cli/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var withdrawProverCmd = &cobra.Command{
	Use:   "prover-withdraw",
	Short: "withdraw from derand prover account to current balance",

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if len(args) == 0 {
			return fmt.Errorf("require amount")
		}

		if len(args) > 1 {
			return fmt.Errorf("Usage: derand withdraw-prover [amount]")
		}

		amount, err := utils.ETHStringToWei(args[0])
		if err != nil {
			return err
		}

		currentChain, err := cfg.GetCurrentChain()
		if err != nil {
			return err
		}

		if currentChain.DeRand == nil {
			return fmt.Errorf("derand has not been deployed yet, please run `derand deploy` or `derand chain setup` first!")
		}

		backend, err := cfg.GetCurrentChainBackend()
		if err != nil {
			return err
		}

		derand, err := gen.NewDeRand(currentChain.DeRand.Address, backend)
		if err != nil {
			return fmt.Errorf("failed to initialize derand: %w", err)
		}

		currentTxOpts, err := cfg.GetCurrentDefaultTxOpts(utils.AskPassword("Enter password: "))
		if err != nil {
			return err
		}

		tx, err := derand.ProverWithdraw(currentTxOpts, amount)
		if err != nil {
			errName, errArgs, err2 := utils.DecodeCustomError(gen.DeRandMetaData.ABI, err)
			if err2 != nil {
				return err2
			}

			return fmt.Errorf("%s%v", errName, errArgs)
		}

		utils.PrintTitle("Withdrawing", args[0], "ETH from DeRand Prover account")
		utils.PrintSubtitle("Transaction hash:", tx.Hash())

		return nil
	},
}
