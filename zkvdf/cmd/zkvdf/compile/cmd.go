package compile

import (
	"fmt"
	"strconv"
	"zkvdf/cmd/zkvdf/common"
	"zkvdf/vdf"
	"zkvdf/vdfcircuit"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"

	"github.com/spf13/cobra"
)

var (
	flagSystemID    string
	flagCircuitName string
	flagSRSSource   string
)

var Cmd = &cobra.Command{
	Use:   "compile",
	Short: "compile circuit",

	RunE: func(cmd *cobra.Command, args []string) error {
		common.PrintHeading("Circuit:", flagCircuitName)
		common.PrintHeading("System ID:", flagSystemID)

		circuitName := common.CircuitName(flagCircuitName)

		circuit, err := loadCircuit(circuitName, flagSystemID)
		if err != nil {
			return err
		}

		if _, err := compileCircuit(circuitName, flagSystemID, circuit); err != nil {
			return err
		}

		return nil
	},
}

func loadCircuit(circuitName common.CircuitName, systemID string) (frontend.Circuit, error) {
	setup, err := vdf.LoadSetup(common.SystemPath(systemID))
	if err != nil {
		return nil, fmt.Errorf("failed to load system file: %w", err)
	}

	switch circuitName {
	case common.HashToFormCircuitName:
		return vdfcircuit.NewVDFHashToFormCircuit(setup), nil

	case common.IntermediatePowCircuitName:
		return vdfcircuit.NewVDFIntermediatePow(setup), nil

	case common.RCVerifierCircuitName:
		hashToFormSignature, err := common.LoadCircuitSignature(flagSRSSource, common.HashToFormCircuitName, systemID)
		if err != nil {
			return nil, fmt.Errorf("failed to load hash to form signature: %w", err)
		}

		intermediatePowSignature, err := common.LoadCircuitSignature(flagSRSSource, common.IntermediatePowCircuitName, systemID)
		if err != nil {
			return nil, fmt.Errorf("failed to load hash to form signature: %w", err)
		}

		return vdfcircuit.NewRCVerifier(setup, hashToFormSignature, intermediatePowSignature), nil

	case common.RCVerifierPhase1CircuitName:
		hashToFormSignature, err := common.LoadCircuitSignature(flagSRSSource, common.HashToFormCircuitName, systemID)
		if err != nil {
			return nil, fmt.Errorf("failed to load hash to form signature: %w", err)
		}

		intermediatePowSignature, err := common.LoadCircuitSignature(flagSRSSource, common.IntermediatePowCircuitName, systemID)
		if err != nil {
			return nil, fmt.Errorf("failed to load intermediate pow signature: %w", err)
		}

		return vdfcircuit.NewRCVerifierPhase1(setup, hashToFormSignature, intermediatePowSignature), nil

	case common.RCVerifierPhase2CircuitName:
		phase1Signature, err := common.LoadCircuitSignature(flagSRSSource, common.RCVerifierPhase1CircuitName, systemID)
		if err != nil {
			return nil, fmt.Errorf("failed to load phase 1 signature: %w", err)
		}

		intermediatePowSignature, err := common.LoadCircuitSignature(flagSRSSource, common.IntermediatePowCircuitName, systemID)
		if err != nil {
			return nil, fmt.Errorf("failed to load intermediate pow signature: %w", err)
		}

		return vdfcircuit.NewRCVerifierPhase2(setup, phase1Signature, intermediatePowSignature), nil

	case common.VerifierCircuitName:
		return vdfcircuit.NewVerifier(setup), nil

	default:
		return nil, fmt.Errorf("invalid circuit name: %s", circuitName)
	}
}

func compileCircuit(circuitName common.CircuitName, systemID string, circuit frontend.Circuit) (constraint.ConstraintSystem, error) {
	cs, err := common.NewStep1[constraint.ConstraintSystem]("Compiling circuit " + circuitName).
		OkMessageFunc1(func(cs constraint.ConstraintSystem) string {
			return "Number of public variables: " + strconv.Itoa(cs.GetNbPublicVariables())
		}).
		OkMessageFunc1(func(cs constraint.ConstraintSystem) string {
			return "Number of constraints: " + strconv.Itoa(cs.GetNbConstraints())
		}).
		Do1(func() (constraint.ConstraintSystem, error) {
			return frontend.Compile(ecc.BN254.ScalarField(), scs.NewBuilder, circuit)
		})
	if err != nil {
		return nil, fmt.Errorf("failed to compile circuit: %w", err)
	}

	err = common.NewStep0("Saving constraint system").
		OkMessage(common.CSPath(circuitName, systemID)).
		Do0(func() error { return common.SaveCS(circuitName, systemID, cs) })
	if err != nil {
		return nil, fmt.Errorf("failed to save constraint system: %w", err)
	}

	return cs, nil
}

func init() {
	Cmd.Flags().StringVar(&flagSystemID, "system", "", "system id")
	Cmd.Flags().StringVar(&flagCircuitName, "circuit", "", "circuit name (hash_to_form/intermediate_pow/rc_verifier)")
	Cmd.Flags().StringVar(&flagSRSSource, "srs-source", "unsafe", "unsafe/snarkjs/perpetual")
	Cmd.MarkFlagRequired("system")
	Cmd.MarkFlagRequired("circuit")
}
