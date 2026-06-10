package prover

import (
	"derand-cli/config"
	"derand-cli/gen"
	"derand-cli/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	flagUnregisterProfileId      uint64
	flagUnregisterProfileVersion uint32
)

var unregisterCmd = &cobra.Command{
	Use:   "unregister",
	Short: "unregister from a profile",

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		currentChain, err := cfg.GetCurrentChain()
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

		auth, err := cfg.GetCurrentDefaultTxOpts(utils.AskPassword("Enter password: "))
		if err != nil {
			return err
		}

		utils.PrintTitle("Unregister from profile")
		tx, err := derand.UnregisterProfile(auth, flagUnregisterProfileId, flagUnregisterProfileVersion)
		if err != nil {
			name, args, err := utils.DecodeCustomError(gen.DeRandABI, err)
			if err != nil {
				return err
			}

			return fmt.Errorf("%s%v", name, args)
		}

		utils.PrintSubtitle("Transaction:", tx.Hash())
		return nil
	},
}

func init() {
	unregisterCmd.Flags().Uint64Var(&flagUnregisterProfileId, "profile-id", 0, "profile id")
	unregisterCmd.Flags().Uint32Var(&flagUnregisterProfileVersion, "profile-version", 0, "profile version")

	unregisterCmd.MarkFlagRequired("profile-id")
	unregisterCmd.MarkFlagRequired("profile-version")
}
