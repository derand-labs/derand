package localprofile

import (
	"derand-cli/config"
	"derand-cli/utils"
	"fmt"
	"maps"
	"slices"

	"github.com/spf13/cobra"
)

var (
	flagDeleteName string
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete a local profile",

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		currentChain, err := cfg.GetCurrentChain()
		if err != nil {
			return err
		}

		if _, ok := cfg.LocalProfiles[flagDeleteName]; !ok {
			return fmt.Errorf("not found local profile")
		}

		if !utils.Confirm("Delete local profile '%s'? [y/N]: ", flagDeleteName) {
			return fmt.Errorf("cancelled")
		}

		delete(cfg.LocalProfiles, flagDeleteName)
		for k := range slices.Collect(maps.Keys(currentChain.RemoteProfileMap)) {
			if currentChain.RemoteProfileMap[k] == flagDeleteName {
				delete(currentChain.RemoteProfileMap, k)
			}
		}

		if err := cfg.Save(); err != nil {
			return err
		}

		utils.PrintTitle("OK")
		return nil
	},
}

func init() {
	deleteCmd.Flags().StringVar(&flagDeleteName, "name", "", "local profile name")
	deleteCmd.MarkFlagRequired("name")
}
