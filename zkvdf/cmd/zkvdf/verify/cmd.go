package verify

import (
	"fmt"
	"zkvdf/cmd/zkvdf/common"
	"zkvdf/vdf"
	"zkvdf/vdfcircuit"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/plonk"
	rcplonk "github.com/consensys/gnark/std/recursion/plonk"
	"github.com/spf13/cobra"
)

var (
	flagCircuitName string
	flagSystemID    string
	flagProofID     string
	flagSRSSource   string
)

var Cmd = &cobra.Command{
	Use:   "verify",
	Short: "verify proof",

	RunE: func(cmd *cobra.Command, args []string) error {
		common.PrintHeading("Circuit:", flagCircuitName)
		common.PrintHeading("System ID:", flagSystemID)
		common.PrintHeading("Proof ID:", flagProofID)

		circuitName := common.CircuitName(flagCircuitName)

		suffixes := []string{}
		if circuitName == common.IntermediatePowCircuitName {
			setup, err := vdf.LoadSetup(common.SystemPath(flagSystemID))
			if err != nil {
				return err
			}

			for i := range setup.SplitExp {
				suffixes = append(suffixes, fmt.Sprintf("pil-%d", i))
				suffixes = append(suffixes, fmt.Sprintf("xr-%d", i))
			}
		} else {
			suffixes = append(suffixes, "")
		}

		vk, err := common.LoadVK(flagSRSSource, circuitName, flagSystemID)
		if err != nil {
			return fmt.Errorf("failed to load vk file: %w", err)
		}

		options := []backend.VerifierOption{}
		if circuitName == common.HashToFormCircuitName || circuitName == common.IntermediatePowCircuitName || circuitName == common.RCVerifierPhase1CircuitName {
			options = append(options, rcplonk.GetNativeVerifierOptions(ecc.BN254.ScalarField(), ecc.BN254.ScalarField()))
		}

		for _, suffix := range suffixes {
			title := flagCircuitName
			currentCircuitName := circuitName
			if suffix != "" {
				currentCircuitName += common.CircuitName("-" + suffix)
				title = fmt.Sprintf("%s (%s)", flagCircuitName, suffix)
			}

			proof, err := common.LoadProof(flagSRSSource, currentCircuitName, flagSystemID, flagProofID)
			if err != nil {
				return fmt.Errorf("failed to load proof file: %w", err)
			}

			publicWitness, err := common.LoadPublicWitness(flagSRSSource, currentCircuitName, flagSystemID, flagProofID)
			if err != nil {
				return fmt.Errorf("failed to load public witness file: %w", err)
			}

			common.NewStep0("Verifying for " + title).
				FailMessage(common.ProofPath(flagSRSSource, currentCircuitName, flagSystemID, flagProofID)).
				FailMessage(common.PublicWitnessPath(flagSRSSource, currentCircuitName, flagSystemID, flagProofID)).
				FailMessageFunc(func(err error) string {
					return err.Error()
				}).
				Do0(func() error {
					return plonk.Verify(proof, vk, publicWitness, options...)
				})

			if circuitName == common.RCVerifierCircuitName || circuitName == common.RCVerifierPhase2CircuitName {
				err = common.NewStep0("Saving Solidity proof for " + title).
					OkMessage(common.SolidityProofPath(flagSRSSource, currentCircuitName, flagSystemID, flagProofID)).
					Do0(func() error {
						setup, err := vdf.LoadSetup(common.SystemPath(flagSystemID))
						if err != nil {
							return err
						}

						transcript, err := vdfcircuit.LoadTranscript(setup, common.TranscriptPath(flagSystemID, flagProofID))
						if err != nil {
							return err
						}

						solidityProof, err := common.NewSolidityProof(setup, proof, transcript)
						if err != nil {
							return fmt.Errorf("failed to create solidity proof: %w", err)
						}

						return common.SaveSolidityProof(
							flagSRSSource, currentCircuitName, flagSystemID, flagProofID, solidityProof)
					})
				if err != nil {
					return fmt.Errorf("failed to save solidity proof: %w", err)
				}
			}
		}

		return nil
	},
}

func init() {
	Cmd.Flags().StringVar(&flagSystemID, "system", "", "system id")
	Cmd.Flags().StringVar(&flagCircuitName, "circuit", "", "circuit name")
	Cmd.Flags().StringVar(&flagProofID, "proof", "", "proof id")
	Cmd.Flags().StringVar(&flagSRSSource, "srs-source", "unsafe", "unsafe/snarkjs/perpetual")
	Cmd.MarkFlagRequired("system")
	Cmd.MarkFlagRequired("circuit")
	Cmd.MarkFlagRequired("proof")
}
