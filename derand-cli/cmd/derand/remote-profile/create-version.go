package remoteprofile

import (
	"derand-cli/config"
	"derand-cli/gen"
	"derand-cli/utils"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/spf13/cobra"
)

var (
	flagCreateVersionRemoteProfileID uint64
	flagCreateVersionBaseFee         string
	flagCreateVersionDelayFee        string
)

var createVersionCmd = &cobra.Command{
	Use:   "create-version",
	Short: "create a new version of a remote profile",

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		currentChain, err := cfg.GetCurrentChain()
		if err != nil {
			return err
		}

		if currentChain.DeRand == nil {
			return fmt.Errorf("DeRand has not been deployed yet, please run `derand deploy` or `derand chain setup`")
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

		baseFeeInGwei, err := utils.ETHStringToGwei(flagCreateVersionBaseFee)
		if err != nil {
			return err
		}
		delayFeeInGwei, _ := utils.ETHStringToGwei(flagCreateVersionDelayFee)
		if err != nil {
			return err
		}

		utils.PrintTitle("Added profile version")
		profileVersionTx, err := derand.AddProfileVersion(auth,
			flagCreateVersionRemoteProfileID,
			baseFeeInGwei.Uint64(),
			delayFeeInGwei.Uint64(),
		)
		if err != nil {
			return err
		}
		utils.PrintSubtitle("Transaction:", profileVersionTx.Hash())

		receipt, err := bind.WaitMined(cmd.Context(), backend, profileVersionTx)
		if err != nil {
			return err
		}

		remoteProfileVersion := -1
		for _, vLog := range receipt.Logs {
			ev, err := derand.ParseProfileVersionCreated(*vLog)
			if err == nil {
				remoteProfileVersion = int(ev.Version)
				break
			}
		}

		if remoteProfileVersion == -1 {
			return fmt.Errorf("not found the event to catch profile version")
		}

		utils.PrintSubtitle("Version:", remoteProfileVersion)

		return nil
	},
}

func init() {
	createVersionCmd.Flags().Uint64Var(&flagCreateVersionRemoteProfileID, "profile-id", 0, "remote profile id")
	createVersionCmd.Flags().StringVar(&flagCreateVersionBaseFee, "base-fee", "0", "base fee in ETH")
	createVersionCmd.Flags().StringVar(&flagCreateVersionDelayFee, "delay-fee", "0", "delay fee in ETH")

	createVersionCmd.MarkFlagRequired("profile-id")
	createVersionCmd.MarkFlagRequired("base-fee")
	createVersionCmd.MarkFlagRequired("delay-fee")
}
