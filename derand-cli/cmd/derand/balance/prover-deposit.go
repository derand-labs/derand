package balance

import (
	"derand-cli/config"
	"derand-cli/gen"
	"derand-cli/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var depositProverCmd = &cobra.Command{
	Use:   "prover-deposit",
	Short: "deposit to derand prover account",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if len(args) == 0 {
			return fmt.Errorf("require amount")
		}

		if len(args) > 1 {
			return fmt.Errorf("Usage: derand deposit-prover [amount]")
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
		currentTxOpts.Value = amount

		tx, err := derand.ProverDeposit(currentTxOpts)
		if err != nil {
			return err
		}

		utils.PrintTitle("Sending", args[0], "ETH into DeRand Prover account")
		utils.PrintSubtitle("Transaction hash:", tx.Hash())

		return nil
	},
}
