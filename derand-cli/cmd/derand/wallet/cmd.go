package wallet

import (
	"derand-cli/config"
	"derand-cli/utils"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "wallet",
	Short: "get current wallet info",

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

		addr, err := currentWallet.GetAddress()
		if err != nil {
			return err
		}

		utils.PrintTitle(utils.Bold("Name:"), currentWallet.Name)
		utils.PrintTitle(utils.Bold("Address:"), addr)

		backend, err := cfg.GetCurrentChainBackend()
		if err != nil {
			return err
		}

		balance, err := backend.BalanceAt(cmd.Context(), addr, nil)
		if err != nil {
			utils.PrintSubtitle("Balance:", err)
		} else {
			utils.PrintTitle(utils.Bold("Balance:"), utils.WeiToETHString(balance), currentChain.Symbol)
		}

		return nil
	},
}

func init() {
	Cmd.AddCommand(addCmd)
	Cmd.AddCommand(lsCmd)
	Cmd.AddCommand(switchCmd)
	Cmd.AddCommand(deleteCmd)
	Cmd.AddCommand(getPrivateCmd)
}
