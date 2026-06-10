package prover

import (
	"derand-cli/config"
	"derand-cli/gen"
	"derand-cli/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "prover",
	Short: "get prover info",

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		currentChain, err := cfg.GetCurrentChain()
		if err != nil {
			return err
		}

		currentWallet, err := cfg.GetCurrentWallet()
		if err != nil {
			return err
		}

		backend, err := cfg.GetCurrentChainBackend()
		if err != nil {
			return err
		}

		derand, err := gen.NewDeRand(currentChain.DeRand.Address, backend)
		if err != nil {
			return err
		}

		addr, err := currentWallet.GetAddress()
		if err != nil {
			return err
		}

		registeredProfiles, err := derand.RegisteredProfilesOf(nil, addr)
		if err != nil {
			return err
		}

		utils.PrintTitle("Registered profiles")
		if len(registeredProfiles) == 0 {
			utils.PrintSubtitle("(empty)")
		} else {
			for _, p := range registeredProfiles {
				utils.PrintSubtitle(fmt.Sprintf("Profile %d[%d]", p.Id, p.Version))
			}
		}

		return nil
	},
}

func init() {
	Cmd.AddCommand(proveCmd)
	Cmd.AddCommand(registerCmd)
	Cmd.AddCommand(unregisterCmd)
	Cmd.AddCommand(runCmd)
}
