package wallet

import (
	"derand-cli/config"
	"derand-cli/utils"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch",
	Short: "switch wallet",

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if len(args) < 1 {
			return fmt.Errorf("must specify wallet index, run `derand wallet ls` to get more detail")
		}

		if len(args) > 1 {
			return fmt.Errorf("derand wallet switch [index]")
		}

		index, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid wallet index: %w", err)
		}

		if index >= len(cfg.Wallets) {
			return fmt.Errorf("invalid wallet index, out of bound")
		}

		cfg.CurrentWalletIndex = index
		if err := cfg.Save(); err != nil {
			return err
		}

		utils.PrintTitle("OK")
		return nil
	},
}
