package localprofile

import (
	"derand-cli/config"
	"derand-cli/profile"
	"derand-cli/utils"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/spf13/cobra"
)

var (
	flagCreateName              string
	flagCreateSeed              string
	flagCreateDBits             uint16
	flagCreateLimbBits          uint16
	flagCreateSplitExp          uint16
	flagCreateH2fNbGenerators   uint16
	flagCreateH2fSteps          uint16
	flagCreateSRSSource         string
	flagCreateZKVerifierVersion int
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new local profile",

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if _, ok := cfg.LocalProfiles[flagCreateName]; ok {
			return fmt.Errorf("duplicate name")
		}

		data := profile.LocalProfile{
			Type: "standard_classgroup_zk_plonk_bn254",
			StandardClassgroupZKPlonkBn254: &profile.StandardClassgroupZKPlonkBn254LocalProfile{
				Seed:                 hexutil.MustDecode(flagCreateSeed),
				DBits:                flagCreateDBits,
				LimbBits:             flagCreateLimbBits,
				SplitExp:             flagCreateSplitExp,
				HashToFormGenerators: flagCreateH2fNbGenerators,
				HashToFormSteps:      flagCreateH2fSteps,
				SRSSource:            flagCreateSRSSource,
			},
		}

		data.StandardClassgroupZKPlonkBn254.Signature.Circuits = make(map[string]map[string]hexutil.Bytes)
		data.StandardClassgroupZKPlonkBn254.Signature.Circuits["hash_to_form"] = make(map[string]hexutil.Bytes)
		data.StandardClassgroupZKPlonkBn254.Signature.Circuits["intermediate_pow"] = make(map[string]hexutil.Bytes)

		switch flagCreateZKVerifierVersion {
		case 1:
			data.StandardClassgroupZKPlonkBn254.Signature.Circuits["rc_verifier"] = make(map[string]hexutil.Bytes)
		case 2:
			data.StandardClassgroupZKPlonkBn254.Signature.Circuits["rc_verifier_phase_1"] = make(map[string]hexutil.Bytes)
			data.StandardClassgroupZKPlonkBn254.Signature.Circuits["rc_verifier_phase_2"] = make(map[string]hexutil.Bytes)
		default:
			return fmt.Errorf("invalid zk verifier version")
		}

		if cfg.LocalProfiles == nil {
			cfg.LocalProfiles = make(map[string]config.LocalProfileInfo)
		}

		cfg.LocalProfiles[flagCreateName] = config.LocalProfileInfo{
			Path: "[created]",
			Data: data,
		}

		if err := cfg.Save(); err != nil {
			return err
		}

		utils.PrintTitle("OK")
		return nil
	},
}

func init() {
	createCmd.Flags().StringVar(&flagCreateName, "name", "", "profile name")
	createCmd.Flags().StringVar(&flagCreateSeed, "seed", "0x000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f", "seed")
	createCmd.Flags().Uint16Var(&flagCreateDBits, "d-bits", 16, "classgroup discriminant size")
	createCmd.Flags().Uint16Var(&flagCreateLimbBits, "limb-bits", 16, "zk circuit limb size")
	createCmd.Flags().Uint16Var(
		&flagCreateSplitExp,
		"split-exp",
		1,
		"higher value reduces intermediate pow circuit size but increases rc verifier final circuit size",
	)
	createCmd.Flags().Uint16Var(&flagCreateH2fNbGenerators, "h2f-generators", 8, "the number of generators for hash-to-form")
	createCmd.Flags().Uint16Var(&flagCreateH2fSteps, "h2f-steps", 2, "the number of steps for hash-to-form")
	createCmd.Flags().IntVar(&flagCreateZKVerifierVersion, "zk-vv", 1, "single phase or two phase rc_verifier (1/2)")
	createCmd.Flags().StringVar(&flagCreateSRSSource, "srs-source", "unsafe", "unsafe/snarkjs/perpetual")

	createCmd.MarkFlagRequired("name")
}
