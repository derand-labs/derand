package chain

import (
	"derand-cli/config"
	"derand-cli/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "chain",
	Short: "get current chain info",

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		currentChain, err := cfg.GetCurrentChain()
		if err != nil {
			return err
		}

		utils.PrintTitle(utils.Bold("Name:"), currentChain.Name)
		utils.PrintTitle(utils.Bold("Chain ID:"), currentChain.ChainID)

		utils.PrintTitle(utils.Bold("RPCs:"))
		for i := range currentChain.RPCs {
			utils.PrintSubtitle(fmt.Sprintf("[%d]", i), currentChain.RPCs[i])
		}

		utils.PrintTitle(utils.Bold("WS RPCs:"))
		for i := range currentChain.WSRPCs {
			utils.PrintSubtitle(fmt.Sprintf("[%d]", i), currentChain.WSRPCs[i])
		}

		backend, err := currentChain.GetBackend()
		if err != nil {
			utils.PrintTitle(utils.Bold("Gas price:"), err)
		} else {
			gasPrice, err := backend.SuggestGasPrice(cmd.Context())
			if err != nil {
				utils.PrintTitle(utils.Bold("Gas price:"), err)
			} else {
				utils.PrintTitle(utils.Bold("Gas price:"), gasPrice)
			}
		}

		return nil
	},
}

func init() {
	Cmd.AddCommand(addCmd)
	Cmd.AddCommand(setupCmd)
	Cmd.AddCommand(lsCmd)
	Cmd.AddCommand(switchCmd)
	Cmd.AddCommand(deleteCmd)
	Cmd.AddCommand(addRPCCmd)
	Cmd.AddCommand(removeRPCCmd)
	Cmd.AddCommand(removeContractCmd)
}
