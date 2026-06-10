package chain

import (
	"derand-cli/config"
	"derand-cli/utils"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch",
	Short: "switch chain",

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if len(args) < 1 {
			return fmt.Errorf("must specify chain index, run `derand chain ls` to get more detail")
		}

		if len(args) > 1 {
			return fmt.Errorf("derand chain switch [index]")
		}

		index, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid chain index: %w", err)
		}

		if index < 0 || index >= len(cfg.Chains) {
			return fmt.Errorf("invalid chain index: out of bound")
		}

		cfg.CurrentChainIndex = index
		if err := cfg.Save(); err != nil {
			return err
		}

		utils.PrintTitle("OK")
		return nil
	},
}
