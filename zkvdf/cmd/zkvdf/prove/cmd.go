package prove

import (
	"fmt"
	"zkvdf/cmd/zkvdf/common"
	"zkvdf/vdf"
	"zkvdf/vdfcircuit"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	rcplonk "github.com/consensys/gnark/std/recursion/plonk"
	"github.com/spf13/cobra"
)

var (
	flagSystemID    string
	flagCircuitName string
	flagProofID     string
	flagSRSSource   string
)

var Cmd = &cobra.Command{
	Use:   "prove",
	Short: "generate proof",
	RunE: func(cmd *cobra.Command, args []string) error {
		common.PrintHeading("Circuit:", flagCircuitName)
		common.PrintHeading("System ID:", flagSystemID)
		common.PrintHeading("Proof ID:", flagProofID)

		circuitName := common.CircuitName(flagCircuitName)

		assignments, err := LoadAssignments(flagSRSSource, circuitName, flagSystemID, flagProofID)
		if err != nil {
			return err
		}

		cs, err := common.NewStep1[constraint.ConstraintSystem]("Loading constraint system").
			Do1(func() (constraint.ConstraintSystem, error) {
				return common.LoadCS(circuitName, flagSystemID)
			})
		if err != nil {
			return fmt.Errorf("failed to load constraint system: %w", err)
		}

		pk, err := common.NewStep1[plonk.ProvingKey]("Loading proving key").
			Do1(func() (plonk.ProvingKey, error) {
				return common.LoadPK(flagSRSSource, circuitName, flagSystemID)
			})
		if err != nil {
			return fmt.Errorf("failed to load constraint system: %w", err)
		}

		for i := range assignments {
			if err := RunProve(cs, pk, flagSRSSource, flagSystemID, flagProofID, assignments[i]); err != nil {
				return fmt.Errorf("fail to run prove: %w", err)
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

type CircuitAssignment struct {
	title                   string
	circuitName             common.CircuitName
	assignment              frontend.Circuit
	requireNativeProverFlag bool
}

func LoadAssignments(
	srsSource string,
	circuitName common.CircuitName,
	systemID, proofID string,
) ([]CircuitAssignment, error) {
	setup, err := vdf.LoadSetup(common.SystemPath(systemID))
	if err != nil {
		return nil, fmt.Errorf("failed to load system file: %w", err)
	}

	transcript, err := vdfcircuit.LoadTranscript(setup, common.TranscriptPath(systemID, proofID))
	if err != nil {
		return nil, fmt.Errorf("failed to load transcript: %w", err)
	}

	switch circuitName {
	case common.HashToFormCircuitName:
		assignment := vdfcircuit.NewVDFHashToFormCircuit(setup)
		assignment.Assign(transcript)

		return []CircuitAssignment{{
			circuitName:             circuitName,
			assignment:              assignment,
			requireNativeProverFlag: true,
		}}, nil

	case common.IntermediatePowCircuitName:
		assignments := []CircuitAssignment{}

		for i := range setup.SplitExp {
			assignment := vdfcircuit.NewVDFIntermediatePow(setup)
			assignment.AssignPiL(transcript, i)
			assignments = append(assignments, CircuitAssignment{
				title:                   fmt.Sprintf("%s round %d (pi^l)", circuitName, i),
				circuitName:             circuitName + common.CircuitName(fmt.Sprintf("-pil-%d", i)),
				assignment:              assignment,
				requireNativeProverFlag: true,
			})

			assignment = vdfcircuit.NewVDFIntermediatePow(setup)
			assignment.AssignXR(transcript, i)
			assignments = append(assignments, CircuitAssignment{
				title:                   fmt.Sprintf("%s round %d (x^r)", circuitName, i),
				circuitName:             circuitName + common.CircuitName(fmt.Sprintf("-xr-%d", i)),
				assignment:              assignment,
				requireNativeProverFlag: true,
			})
		}

		return assignments, nil

	case common.RCVerifierCircuitName:
		hashToFormSignature, err := common.LoadCircuitSignature(srsSource, common.HashToFormCircuitName, systemID)
		if err != nil {
			return nil, fmt.Errorf("failed to load hash to form signature: %w", err)
		}

		intermediatePowSignature, err := common.LoadCircuitSignature(srsSource, common.IntermediatePowCircuitName, systemID)
		if err != nil {
			return nil, fmt.Errorf("failed to load hash to form signature: %w", err)
		}

		hashToFormProof, err := common.LoadCircuitProof(
			srsSource,
			common.HashToFormCircuitName,
			systemID,
			proofID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load hash to form proof: %w", err)
		}

		intermediatePowPiLProofs := []vdfcircuit.CircuitProof{}
		intermediatePowXRProofs := []vdfcircuit.CircuitProof{}
		for i := range setup.SplitExp {
			p, err := common.LoadCircuitProof(
				srsSource,
				common.IntermediatePowCircuitName+common.CircuitName(fmt.Sprintf("-pil-%d", i)),
				systemID,
				proofID,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to load PiL intermediate pow proof: %w", err)
			}
			intermediatePowPiLProofs = append(intermediatePowPiLProofs, p)

			p, err = common.LoadCircuitProof(
				srsSource,
				common.IntermediatePowCircuitName+common.CircuitName(fmt.Sprintf("-xr-%d", i)),
				systemID,
				proofID,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to load XR intermediate pow proof: %w", err)
			}
			intermediatePowXRProofs = append(intermediatePowXRProofs, p)
		}

		assignment := vdfcircuit.NewRCVerifier(setup, hashToFormSignature, intermediatePowSignature)
		err = assignment.Assign(transcript, hashToFormProof, intermediatePowPiLProofs, intermediatePowXRProofs)
		if err != nil {
			return nil, fmt.Errorf("failed to assign vdf verifier circuit: %w", err)
		}

		return []CircuitAssignment{{
			circuitName: circuitName,
			assignment:  assignment,
		}}, nil

	case common.RCVerifierPhase1CircuitName:
		hashToFormSignature, err := common.LoadCircuitSignature(srsSource, common.HashToFormCircuitName, systemID)
		if err != nil {
			return nil, fmt.Errorf("failed to load hash to form signature: %w", err)
		}

		intermediatePowSignature, err := common.LoadCircuitSignature(srsSource, common.IntermediatePowCircuitName, systemID)
		if err != nil {
			return nil, fmt.Errorf("failed to load intermediate pow signature: %w", err)
		}

		hashToFormProof, err := common.LoadCircuitProof(
			srsSource,
			common.HashToFormCircuitName,
			systemID,
			proofID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load hash to form proof: %w", err)
		}

		intermediatePowPiLProofs := []vdfcircuit.CircuitProof{}
		intermediatePowXRProofs := []vdfcircuit.CircuitProof{}
		for i := range setup.SplitExp / 2 {
			p, err := common.LoadCircuitProof(
				srsSource,
				common.IntermediatePowCircuitName+common.CircuitName(fmt.Sprintf("-pil-%d", i)),
				systemID,
				proofID,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to load PiL intermediate pow proof: %w", err)
			}
			intermediatePowPiLProofs = append(intermediatePowPiLProofs, p)

			p, err = common.LoadCircuitProof(
				srsSource,
				common.IntermediatePowCircuitName+common.CircuitName(fmt.Sprintf("-xr-%d", i)),
				systemID,
				proofID,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to load XR intermediate pow proof: %w", err)
			}
			intermediatePowXRProofs = append(intermediatePowXRProofs, p)
		}

		assignment := vdfcircuit.NewRCVerifierPhase1(setup, hashToFormSignature, intermediatePowSignature)
		err = assignment.Assign(transcript, hashToFormProof, intermediatePowPiLProofs, intermediatePowXRProofs)
		if err != nil {
			return nil, fmt.Errorf("failed to assign rc verifier phase 1 circuit: %w", err)
		}

		return []CircuitAssignment{{
			circuitName:             circuitName,
			assignment:              assignment,
			requireNativeProverFlag: true,
		}}, nil

	case common.RCVerifierPhase2CircuitName:
		phase1Signature, err := common.LoadCircuitSignature(srsSource, common.RCVerifierPhase1CircuitName, systemID)
		if err != nil {
			return nil, fmt.Errorf("failed to load phase1 signature: %w", err)
		}

		intermediatePowSignature, err := common.LoadCircuitSignature(srsSource, common.IntermediatePowCircuitName, systemID)
		if err != nil {
			return nil, fmt.Errorf("failed to load intermediate pow signature: %w", err)
		}

		phase1Proof, err := common.LoadCircuitProof(
			srsSource,
			common.RCVerifierPhase1CircuitName,
			systemID,
			proofID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load hash to form proof: %w", err)
		}

		intermediatePowPiLProofs := []vdfcircuit.CircuitProof{}
		intermediatePowXRProofs := []vdfcircuit.CircuitProof{}
		for i := setup.SplitExp / 2; i < setup.SplitExp; i++ {
			p, err := common.LoadCircuitProof(
				srsSource,
				common.IntermediatePowCircuitName+common.CircuitName(fmt.Sprintf("-pil-%d", i)),
				systemID,
				proofID,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to load PiL intermediate pow proof: %w", err)
			}
			intermediatePowPiLProofs = append(intermediatePowPiLProofs, p)

			p, err = common.LoadCircuitProof(
				srsSource,
				common.IntermediatePowCircuitName+common.CircuitName(fmt.Sprintf("-xr-%d", i)),
				systemID,
				proofID,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to load XR intermediate pow proof: %w", err)
			}
			intermediatePowXRProofs = append(intermediatePowXRProofs, p)
		}

		assignment := vdfcircuit.NewRCVerifierPhase2(setup, phase1Signature, intermediatePowSignature)
		err = assignment.Assign(transcript, phase1Proof, intermediatePowPiLProofs, intermediatePowXRProofs)
		if err != nil {
			return nil, fmt.Errorf("failed to assign rc verifier phase 2 circuit: %w", err)
		}

		return []CircuitAssignment{{
			circuitName: circuitName,
			assignment:  assignment,
		}}, nil

	case common.VerifierCircuitName:
		assignment := vdfcircuit.NewVerifier(setup)
		assignment.Assign(transcript)

		return []CircuitAssignment{{
			circuitName: circuitName,
			assignment:  assignment,
		}}, nil

	default:
		return nil, fmt.Errorf("invalid circuit: %s", circuitName)
	}
}

func RunProve(
	cs constraint.ConstraintSystem,
	pk plonk.ProvingKey,
	srsSource string,
	systemID string,
	proofID string,
	assignment CircuitAssignment,
) error {
	title := assignment.title
	if title == "" {
		title = string(assignment.circuitName)
	}

	witness, err := common.NewStep1[witness.Witness]("Generating witness for " + title).
		Do1(func() (witness.Witness, error) {
			return frontend.NewWitness(assignment.assignment, ecc.BN254.ScalarField())
		})
	if err != nil {
		return fmt.Errorf("failed to generate witness: %w", err)
	}

	proof, err := common.NewStep1[plonk.Proof]("Proving for " + title).
		Do1(func() (plonk.Proof, error) {
			options := []backend.ProverOption{}
			if assignment.requireNativeProverFlag {
				options = append(options, rcplonk.GetNativeProverOptions(ecc.BN254.ScalarField(), ecc.BN254.ScalarField()))
			}

			return plonk.Prove(cs, pk, witness, options...)
		})
	if err != nil {
		return fmt.Errorf("failed to prove: %w", err)
	}

	err = common.NewStep0("Saving proof").
		OkMessage(common.ProofPath(srsSource, assignment.circuitName, systemID, proofID)).
		Do0(func() error {
			return common.SaveProof(srsSource, assignment.circuitName, systemID, proofID, proof)
		})
	if err != nil {
		return fmt.Errorf("failed to save proof: %w", err)
	}

	err = common.NewStep0("Saving public witness").
		OkMessage(common.PublicWitnessPath(srsSource, assignment.circuitName, systemID, proofID)).
		Do0(func() error {
			publicWitness, err := witness.Public()
			if err != nil {
				return fmt.Errorf("failed to get public witness: %w", err)
			}
			return common.SavePublicWitness(srsSource, assignment.circuitName, systemID, proofID, publicWitness)
		})
	if err != nil {
		return fmt.Errorf("failed to save witness: %w", err)
	}

	return nil
}
