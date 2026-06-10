package prover

import (
	"derand-cli/config"
	"derand-cli/gen"
	"derand-cli/utils"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/spf13/cobra"
)

var (
	flagRegisterProfileId      uint64
	flagRegisterProfileVersion uint32
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "register to a profile",

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

		utils.PrintTitle("Register to profile")
		tx, err := derand.RegisterProfile(auth, flagRegisterProfileId, flagRegisterProfileVersion)
		if err != nil {
			name, args, err := utils.DecodeCustomError(gen.DeRandABI, err)
			if err != nil {
				return err
			}

			return fmt.Errorf("%s%v", name, args)
		}
		utils.PrintSubtitle("Transaction:", tx.Hash())

		receipt, err := bind.WaitMined(cmd.Context(), backend, tx)
		if err != nil {
			return err
		}

		if currentChain.LastProverWatchedBlock == 0 {
			currentChain.LastProverWatchedBlock = receipt.BlockNumber.Uint64()
		}

		return nil
	},
}

func init() {
	registerCmd.Flags().Uint64Var(&flagRegisterProfileId, "profile-id", 0, "profile id")
	registerCmd.Flags().Uint32Var(&flagRegisterProfileVersion, "profile-version", 0, "profile version")

	registerCmd.MarkFlagRequired("profile-id")
	registerCmd.MarkFlagRequired("profile-version")
}
