package localprofile

import (
	"derand-cli/config"
	"derand-cli/profile"
	"derand-cli/utils"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	flagImportName string
	flagImportPath string
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "import a new local profile from file/url",

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if _, ok := cfg.LocalProfiles[flagImportName]; ok {
			return fmt.Errorf("duplicate local profile name")
		}

		f, err := utils.DownloadOrOpenFile(cmd.Context(), flagImportPath)
		if err != nil {
			return err
		}
		defer f.Close()

		var localProfile profile.LocalProfile
		if err := json.NewDecoder(f).Decode(&localProfile); err != nil {
			return fmt.Errorf("invalid profile data: %w", err)
		}

		if cfg.LocalProfiles == nil {
			cfg.LocalProfiles = make(map[string]config.LocalProfileInfo)
		}

		cfg.LocalProfiles[flagImportName] = config.LocalProfileInfo{
			Path: flagImportPath,
			Data: localProfile,
		}

		if err := cfg.Save(); err != nil {
			return err
		}

		utils.PrintTitle("OK")
		return nil
	},
}

func init() {
	importCmd.Flags().StringVar(&flagImportName, "name", "", "profile name")
	importCmd.Flags().StringVar(&flagImportPath, "path", "", "path or url to profile")

	importCmd.MarkFlagRequired("name")
	importCmd.MarkFlagRequired("path")
}
