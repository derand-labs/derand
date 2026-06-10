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
	flagCreateLocalProfileName string
	flagCreateBaseFee          string
	flagCreateDelayFee         string
	flagCreateBaseTime         uint64
	flagCreateDelayTime        uint64
	flagCreateDelayScale       uint64
	flagCreateMaxDelay         uint16
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new remote profile from a local profile",

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

		localProfile, ok := cfg.LocalProfiles[flagCreateLocalProfileName]
		if !ok {
			return fmt.Errorf("not found local profile")
		}

		if localProfile.Data.Type != "standard_classgroup_zk_plonk_bn254" {
			return fmt.Errorf("not support create remote profile from local profile type %s", localProfile.Data.Type)
		}

		if _, ok := localProfile.Data.VerifierByChainID[currentChain.ChainID]; !ok {
			return fmt.Errorf("the local profile has not deployed yet, please run `derand local-profile deploy`")
		}

		baseFeeInGwei, err := utils.ETHStringToGwei(flagCreateBaseFee)
		if err != nil {
			return err
		}
		delayFeeInGwei, err := utils.ETHStringToGwei(flagCreateDelayFee)
		if err != nil {
			return err
		}

		utils.PrintTitle("Added profile")
		profileTx, err := derand.AddProfile(auth, gen.Profile{
			Verifier:       localProfile.Data.VerifierByChainID[currentChain.ChainID].Address,
			DelayScale:     flagCreateDelayScale,
			MaximumDelay:   flagCreateMaxDelay,
			BaseTime:       flagCreateBaseTime,
			DelayTime:      flagCreateDelayTime,
			BaseFeeInGwei:  baseFeeInGwei.Uint64(),
			DelayFeeInGwei: delayFeeInGwei.Uint64(),
		})
		if err != nil {
			return err
		}

		utils.PrintSubtitle("Transaction:", profileTx.Hash())

		receipt, err := bind.WaitMined(cmd.Context(), backend, profileTx)
		if err != nil {
			return err
		}

		remoteProfileID := -1
		for _, vLog := range receipt.Logs {
			ev, err := derand.ParseProfileCreated(*vLog)
			if err == nil {
				remoteProfileID = int(ev.ProfileId)
				break
			}
		}

		if remoteProfileID == -1 {
			return fmt.Errorf("not found the event to catch profile id")
		}

		utils.PrintSubtitle("Profile ID:", remoteProfileID)

		if currentChain.RemoteProfileMap == nil {
			currentChain.RemoteProfileMap = make(map[int]string)
		}

		currentChain.RemoteProfileMap[remoteProfileID] = flagCreateLocalProfileName
		if err := cfg.Save(); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	createCmd.Flags().StringVar(&flagCreateLocalProfileName, "local-profile", "", "local profile name")
	createCmd.Flags().StringVar(&flagCreateBaseFee, "base-fee", "0", "base fee in ETH")
	createCmd.Flags().StringVar(&flagCreateDelayFee, "delay-fee", "0", "delay fee in ETH")
	createCmd.Flags().Uint64Var(&flagCreateBaseTime, "base-time", 0, "base time in second")
	createCmd.Flags().Uint64Var(&flagCreateDelayTime, "delay-time", 12, "delay time in second")
	createCmd.Flags().Uint16Var(&flagCreateMaxDelay, "delay-max", 300, "maximum delay value")
	createCmd.Flags().Uint64Var(&flagCreateDelayScale, "delay-scale", 0, "maximum delay value")

	createCmd.MarkFlagRequired("local-profile")
	createCmd.MarkFlagRequired("base-fee")
	createCmd.MarkFlagRequired("delay-fee")
	createCmd.MarkFlagRequired("base-time")
	createCmd.MarkFlagRequired("delay-scale")
}
