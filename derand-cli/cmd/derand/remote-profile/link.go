package remoteprofile

import (
	"bytes"
	"derand-cli/config"
	"derand-cli/gen"
	"derand-cli/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	flagLinkLocalProfileName string
	flagLinkRemoteProfileId  uint64
)

var linkCmd = &cobra.Command{
	Use:   "link",
	Short: "link a new remote profiles with a local profile and verify it",

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

		localProfile, ok := cfg.LocalProfiles[flagLinkLocalProfileName]
		if !ok {
			return fmt.Errorf("not found local profile")
		}

		if localProfile.Data.Type != "standard_classgroup_zk_plonk_bn254" {
			return fmt.Errorf("not support map the remote profile with a local profile type %s", localProfile.Data.Type)
		}

		if _, ok := localProfile.Data.VerifierByChainID[currentChain.ChainID]; !ok {
			return fmt.Errorf("the local profile has not deployed yet, please run `derand local-profile deploy`")
		}

		profiles, err := derand.ListProfiles(nil, flagLinkRemoteProfileId, 1)
		if err != nil {
			return err
		}

		address := localProfile.Data.VerifierByChainID[currentChain.ChainID].Address
		if !bytes.Equal(profiles[0].Verifier[:], address[:]) {
			return fmt.Errorf("remote and local profile verifier addresses do not match")
		}

		currentChain.RemoteProfileMap[int(flagLinkRemoteProfileId)] = flagLinkLocalProfileName
		if err := cfg.Save(); err != nil {
			return err
		}

		utils.PrintTitle("OK")

		return nil
	},
}

func init() {
	linkCmd.Flags().Uint64Var(&flagLinkRemoteProfileId, "profile-id", 0, "remote profile id")
	linkCmd.Flags().StringVar(&flagLinkLocalProfileName, "local-profile", "", "local profile name")

	linkCmd.MarkFlagRequired("profile-id")
	linkCmd.MarkFlagRequired("local-profile")
}
