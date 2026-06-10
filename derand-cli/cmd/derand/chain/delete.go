package chain

import (
	"derand-cli/config"
	"derand-cli/utils"
	"fmt"
	"slices"
	"strconv"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete chain",

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if len(args) < 1 {
			return fmt.Errorf("must specify chain index, run `derand chain ls` to get more detail")
		}

		if len(args) > 1 {
			return fmt.Errorf("derand chain delete [index]")
		}

		index, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid chain index: %w", err)
		}

		currentChain, err := cfg.GetChain(index)
		if err != nil {
			return err
		}

		if !utils.Confirm("Delete chain '%s'? [y/N]: ", currentChain.Name) {
			return fmt.Errorf("cancelled")
		}

		cfg.Chains = slices.Delete(cfg.Chains, index, index+1)
		if cfg.CurrentChainIndex == index {
			fmt.Println("You deleted the current chain, switched to chain 0")
			cfg.CurrentChainIndex = 0
		}

		if err := cfg.Save(); err != nil {
			return err
		}

		utils.PrintTitle("OK")
		return nil
	},
}
