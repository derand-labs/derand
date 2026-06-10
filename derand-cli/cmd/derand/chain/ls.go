package chain

import (
	"derand-cli/config"
	"derand-cli/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "list all chains",

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if len(cfg.Chains) == 0 {
			return fmt.Errorf("no chain configured")
		}

		for i := range cfg.Chains {
			utils.PrintTitle(utils.Bold(fmt.Sprintf("[%d]", i)), cfg.Chains[i].Name)
		}

		return nil
	},
}
