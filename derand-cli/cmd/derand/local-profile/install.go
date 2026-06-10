package localprofile

import (
	"derand-cli/config"
	"derand-cli/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	flagInstallName      string
	flagInstallReinstall bool
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "install a local profile",

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		localProfile, ok := cfg.LocalProfiles[flagInstallName]
		if !ok {
			return fmt.Errorf("not found local profile")
		}

		if localProfile.Data.Type != "standard_classgroup_zk_plonk_bn254" {
			return fmt.Errorf("not support install non-standard profile")
		}

		if err := localProfile.Data.StandardClassgroupZKPlonkBn254.Install(config.GetBuildDir(), config.GetVDFDir(), flagInstallReinstall); err != nil {
			return err
		}

		cfg.LocalProfiles[flagInstallName] = localProfile
		if err := cfg.Save(); err != nil {
			return err
		}

		utils.PrintTitle("OK")
		return nil
	},
}

func init() {
	installCmd.Flags().StringVar(&flagInstallName, "name", "", "local profile name")
	installCmd.Flags().BoolVar(&flagInstallReinstall, "reinstall", false, "require re-install even if file already existed")

	installCmd.MarkFlagRequired("name")
}
