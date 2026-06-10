package remoteprofile

import (
	"derand-cli/config"
	"derand-cli/gen"
	"derand-cli/utils"
	"fmt"
	"math/big"

	"github.com/spf13/cobra"
)

var (
	flagPage      int
	flagProfileID int
)

var Cmd = &cobra.Command{
	Use:   "remote-profile",
	Short: "list remote profiles by page",

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

		if flagProfileID == -1 {
			profiles, err := derand.ListProfiles(nil, uint64(flagPage)*10, 10)
			if err != nil {
				return err
			}

			if len(profiles) == 0 {
				utils.PrintTitle("(empty)")
			} else {
				for i, profile := range profiles {
					printRemoteProfile(currentChain, i, profile)
				}
			}
		} else {
			profiles, err := derand.ListProfiles(nil, uint64(flagProfileID), 1)
			if err != nil {
				return err
			}

			printRemoteProfile(currentChain, flagProfileID, profiles[0])

			versions, err := derand.ListProfileVersions(
				nil, uint64(flagProfileID), uint32(flagPage)*10, 10)
			if err != nil {
				return err
			}

			if len(versions) == 0 {
				utils.PrintTitle("(empty)")
			} else {
				for i, version := range versions {
					utils.PrintTitle(utils.Bold(fmt.Sprintf("Profile Version [%d]", i)))
					utils.PrintSubtitle("base fee:", utils.WeiToETHString(version.BaseFee), currentChain.Symbol)
					utils.PrintSubtitle("delay fee:", utils.WeiToETHString(version.DelayFee), currentChain.Symbol)
					utils.PrintSubtitle("current pool size:", version.PoolSize)
					utils.PrintSubtitle(
						"required prover collateral:",
						utils.WeiToETHString(
							calRequiredCollateral(
								version.BaseFee,
								version.DelayFee,
								uint64(profiles[0].MaximumDelay)),
						),
						currentChain.Symbol,
					)
				}
			}
		}

		return nil
	},
}

func init() {
	Cmd.Flags().IntVar(&flagPage, "page", 0, "page number (max 10 profiles per page)")
	Cmd.Flags().IntVar(&flagProfileID, "profile-id", -1, "list versions in a profile (max 10 versions per page)")

	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(linkCmd)
	Cmd.AddCommand(createVersionCmd)
	Cmd.AddCommand(benchmarkCmd)
}

func printRemoteProfile(currentChain *config.ChainInfo, index int, profile gen.ProfileView) {
	if localName, ok := currentChain.RemoteProfileMap[index]; ok {
		utils.PrintTitle(
			utils.Bold(fmt.Sprintf("Profile [%d]", index)),
			utils.Green(fmt.Sprintf("-> Local [%s]", localName)),
		)
	} else {
		utils.PrintTitle(utils.Bold(fmt.Sprintf("Profile [%d]", index)))
	}
	utils.PrintSubtitle("verifier:", profile.Verifier)
	utils.PrintSubtitle("base time:", profile.BaseTime)
	utils.PrintSubtitle("delay time:", profile.DelayTime)
	utils.PrintSubtitle("delay scale:", profile.DelayScale)
	utils.PrintSubtitle("delay max:", profile.MaximumDelay)
	utils.PrintSubtitle("num versions:", profile.VersionCount)
}

func calRequiredCollateral(baseFee, delayFee *big.Int, maxDelay uint64) *big.Int {
	maxDelayFee := new(big.Int).Mul(delayFee, big.NewInt(int64(maxDelay)))
	maxRequestFee := new(big.Int).Add(baseFee, maxDelayFee)
	return new(big.Int).Mul(maxRequestFee, big.NewInt(106))
}
