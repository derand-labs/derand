package main

import (
	"derand-cli/config"
	"derand-cli/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "get current derand deploy info",

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		currentChain, err := cfg.GetCurrentChain()
		if err != nil {
			return err
		}

		utils.PrintTitle(utils.Bold("DeRand:"))
		if currentChain.DeRand != nil {
			utils.PrintSubtitle("Address:", currentChain.DeRand.Address)
			utils.PrintSubtitle("Transaction:", currentChain.DeRand.DeployTxHash)
		} else {
			utils.PrintSubtitle("**not yet deployed**")
		}

		if len(currentChain.HashToPrime128) > 0 {
			for i, info := range currentChain.HashToPrime128 {
				utils.PrintTitle(utils.Bold(fmt.Sprintf("[%d] HashToPrime128:", i)))
				utils.PrintSubtitle("Address:", info.Address)
				utils.PrintSubtitle("Transaction:", info.DeployTxHash)
			}
		} else {
			utils.PrintTitle(utils.Bold("HashToPrime128:"))
			utils.PrintSubtitle("**not yet deployed**")
		}

		return nil
	},
}
