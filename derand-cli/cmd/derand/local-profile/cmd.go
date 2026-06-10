package localprofile

import (
	"derand-cli/config"
	"derand-cli/profile"
	"derand-cli/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	flagName string
)

var Cmd = &cobra.Command{
	Use:   "local-profile",
	Short: "list all local profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		currentChain, err := cfg.GetCurrentChain()
		if err != nil {
			return err
		}

		if len(cfg.LocalProfiles) == 0 {
			return fmt.Errorf("no local profiles added, please run derand local-profile create or import first!")
		}

		if flagName != "" {
			printSingleLocalProfile(currentChain.ChainID, flagName, cfg.LocalProfiles[flagName])
		} else {
			for name, profile := range cfg.LocalProfiles {
				printSingleLocalProfile(currentChain.ChainID, name, profile)
			}
		}

		return nil
	},
}

func init() {
	Cmd.Flags().StringVar(&flagName, "name", "", "local profile name (empty for displaying all local profiles)")

	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(importCmd)
	Cmd.AddCommand(deleteCmd)
	Cmd.AddCommand(installCmd)
	Cmd.AddCommand(exportCmd)
	Cmd.AddCommand(deployCmd)
	Cmd.AddCommand(verifyCmd)
	Cmd.AddCommand(benchmarkCmd)
}

func printSingleLocalProfile(chainID int, name string, p config.LocalProfileInfo) {
	utils.PrintTitle(utils.Bold(fmt.Sprintf("%s", name)))
	utils.PrintSubtitle("path:", p.Path)
	utils.PrintSubtitle(utils.Bold("type:"), utils.Bold(p.Data.Type))
	if verifier, ok := p.Data.VerifierByChainID[chainID]; !ok {
		utils.PrintSubtitle("verifier: not yet deployed")
	} else {
		utils.PrintSubtitle("verifier:", verifier.Address)
	}
	if p.Data.Type == "standard_classgroup_zk_plonk_bn254" {
		utils.PrintSubtitle("system:", p.Data.StandardClassgroupZKPlonkBn254.GetSystemID())
		utils.PrintSubtitle("seed:", p.Data.StandardClassgroupZKPlonkBn254.Seed)
		utils.PrintSubtitle("d-bits:", p.Data.StandardClassgroupZKPlonkBn254.DBits)
		utils.PrintSubtitle("srs:", p.Data.StandardClassgroupZKPlonkBn254.SRSSource)
		utils.PrintSubtitle("hash-to-form entropy space:", fmt.Sprintf("2^%d",
			int(profile.Log2CombRepeat(
				int64(p.Data.StandardClassgroupZKPlonkBn254.HashToFormGenerators),
				int64(p.Data.StandardClassgroupZKPlonkBn254.HashToFormSteps)),
			),
		))

		if err := p.Data.StandardClassgroupZKPlonkBn254.Validate(config.GetVDFDir()); err != nil {
			utils.PrintSubtitle("installation status:", utils.Red(err.Error()))
		} else {
			utils.PrintSubtitle("installation status:", utils.Green("OK"))
		}

		warnings := p.Data.StandardClassgroupZKPlonkBn254.GetWarnings()
		for _, w := range warnings {
			if w.Level == profile.ProfileWarningLevelLow {
				utils.PrintSubtitle(utils.Yellow(fmt.Sprintf("[WARNING] %s", w.Message)))
			} else {
				utils.PrintSubtitle(utils.Red(fmt.Sprintf("[WARNING-%d] %s", w.Level, w.Message)))
			}
		}
	}
}
