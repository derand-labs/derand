package localprofile

import (
	"derand-cli/config"
	"derand-cli/profile"
	"derand-cli/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	flagDeployName string
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "deploy a local profile to chain",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		currentChain, err := cfg.GetCurrentChain()
		if err != nil {
			return err
		}

		localProfile, ok := cfg.LocalProfiles[flagDeployName]
		if !ok {
			return fmt.Errorf("not found local profile")
		}

		if localProfile.Data.Type != "standard_classgroup_zk_plonk_bn254" {
			return fmt.Errorf("unsupported profile type: %s", localProfile.Data.Type)
		}

		if _, ok := localProfile.Data.VerifierByChainID[currentChain.ChainID]; ok {
			if !utils.Confirm("This profile has already deployed on this chain. Continue? [y/N]: ") {
				return fmt.Errorf("cancelled")
			}
		}

		if currentChain.DeRand == nil {
			return fmt.Errorf("DeRand has not been deployed yet, please run `derand deploy` or `derand chain setup`")
		}

		if len(currentChain.HashToPrime128) == 0 {
			return fmt.Errorf("HashToPrime128 has not been deployed yet, please run `derand deploy` or `derand chain setup`")
		}

		backend, err := cfg.GetCurrentChainBackend()
		if err != nil {
			return err
		}

		auth, err := cfg.GetCurrentDefaultTxOpts(utils.AskPassword("Enter password: "))
		if err != nil {
			return err
		}

		address, txHash, err := localProfile.Data.StandardClassgroupZKPlonkBn254.DeployVerifier(
			config.GetVDFDir(),
			auth,
			backend,
			currentChain.HashToPrime128[0].Address,
		)
		if err != nil {
			return err
		}

		if localProfile.Data.VerifierByChainID == nil {
			localProfile.Data.VerifierByChainID = make(map[int]profile.ContractInfo)
		}

		localProfile.Data.VerifierByChainID[currentChain.ChainID] = profile.ContractInfo{
			Address:      address,
			DeployTxHash: txHash,
		}

		cfg.LocalProfiles[flagDeployName] = localProfile
		utils.PrintTitle("Deployed verifier")
		utils.PrintSubtitle("Address:", address)
		utils.PrintSubtitle("Transaction:", txHash)

		if err := cfg.Save(); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	deployCmd.Flags().StringVar(&flagDeployName, "name", "", "local profile name")
	deployCmd.MarkFlagRequired("name")
}
