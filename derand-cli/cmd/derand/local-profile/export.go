package localprofile

import (
	"derand-cli/config"
	"derand-cli/utils"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	flagExportName string
	flagExportOut  string
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "export a local profile to a separated json file",

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		localProfile, ok := cfg.LocalProfiles[flagExportName]
		if !ok {
			return fmt.Errorf("not found local profile")
		}

		if flagExportOut == "" {
			flagExportOut = "local-profile-" + flagExportName + ".json"
		}

		f, err := os.Create(flagExportOut)
		if err != nil {
			return err
		}

		if err := json.NewEncoder(f).Encode(localProfile.Data); err != nil {
			return err
		}

		utils.PrintTitle("Export successfully")
		utils.PrintSubtitle(flagExportOut)
		return nil
	},
}

func init() {
	exportCmd.Flags().StringVar(&flagExportName, "name", "", "local profile name")
	exportCmd.Flags().StringVar(&flagExportOut, "out", "", "out file")

	exportCmd.MarkFlagRequired("name")
}
