package wallet

import (
	"derand-cli/config"
	"derand-cli/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "list wallets",

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if len(cfg.Wallets) == 0 {
			return fmt.Errorf("no wallet configured")
		}

		for i := range cfg.Wallets {
			utils.PrintTitle(utils.Bold(fmt.Sprintf("[%d]", i)), cfg.Wallets[i].Name)
		}

		return nil
	},
}
