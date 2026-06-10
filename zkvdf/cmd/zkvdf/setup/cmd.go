package setup

import (
	"fmt"
	"math/bits"
	"zkvdf/cmd/zkvdf/common"

	"github.com/consensys/gnark-crypto/kzg"

	bn254kzg "github.com/consensys/gnark-crypto/ecc/bn254/kzg"
	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/constraint"

	"github.com/spf13/cobra"
)

var (
	flagSystemID    string
	flagCircuitName string
	flagSRSSource   string
	flagExportSol   bool
)

var Cmd = &cobra.Command{
	Use:   "setup",
	Short: "generate pk/vk",

	RunE: func(cmd *cobra.Command, args []string) error {
		common.PrintHeading("Circuit:", flagCircuitName)
		common.PrintHeading("System ID:", flagSystemID)

		circuitName := common.CircuitName(flagCircuitName)

		cs, err := common.NewStep1[constraint.ConstraintSystem]("Loading constraint system").
			Do1(func() (constraint.ConstraintSystem, error) {
				return common.LoadCS(circuitName, flagSystemID)
			})
		if err != nil {
			return fmt.Errorf("failed to load constraint system file: %w", err)
		}

		canonicalSRS, lagrangeSRS, err := loadSRS(flagSRSSource, cs)
		if err != nil {
			return err
		}

		pk, vk, err := common.NewStep2[plonk.ProvingKey, plonk.VerifyingKey]("Setting up keys").
			Do2(func() (plonk.ProvingKey, plonk.VerifyingKey, error) {
				return plonk.Setup(cs, canonicalSRS, lagrangeSRS)
			})
		if err != nil {
			return fmt.Errorf("failed to setup: %w", err)
		}

		err = common.NewStep0("Saving proving key").
			OkMessage(common.PKPath(flagSRSSource, circuitName, flagSystemID)).
			Do0(func() error { return common.SavePK(flagSRSSource, circuitName, flagSystemID, pk) })
		if err != nil {
			return fmt.Errorf("failed to save proving key: %w", err)
		}

		err = common.NewStep0("Saving verifying key").
			OkMessage(common.VKPath(flagSRSSource, circuitName, flagSystemID)).
			Do0(func() error { return common.SaveVK(flagSRSSource, circuitName, flagSystemID, vk) })
		if err != nil {
			return fmt.Errorf("failed to save verifying key: %w", err)
		}

		if flagExportSol {
			err = common.NewStep0("Saving Solidity verifier").
				OkMessage(common.SolPath(flagSRSSource, circuitName, flagSystemID)).
				Do0(func() error { return common.SaveSol(flagSRSSource, circuitName, flagSystemID, vk) })
			if err != nil {
				return fmt.Errorf("failed to save solidity verifier: %w", err)
			}
		}

		return nil
	},
}

func init() {
	Cmd.Flags().StringVar(&flagSystemID, "system", "", "system id")
	Cmd.Flags().StringVar(&flagCircuitName, "circuit", "", "circuit name")
	Cmd.Flags().StringVar(&flagSRSSource, "srs-source", "unsafe", "unsafe/snarkjs/perpetual (do not use unsafe in production)")
	Cmd.Flags().BoolVar(&flagExportSol, "sol", false, "export solidity verifier")
	Cmd.MarkFlagRequired("system")
	Cmd.MarkFlagRequired("circuit")
}

func loadSRS(source string, cs constraint.ConstraintSystem) (kzg.SRS, kzg.SRS, error) {
	var err error
	var canonicalSRS, lagrangeSRS kzg.SRS

	power := bits.Len(uint(cs.GetNbConstraints() + cs.GetNbPublicVariables() - 1))

	canonicalSRS, err = common.NewStep1[kzg.SRS]("Loading canonical srs").
		OkMessageFunc1(func(s kzg.SRS) string {
			if s == nil {
				return "Not found " + common.CanonicalSRSPath(source, power) + ", generate now!"
			} else {
				return "Use " + common.CanonicalSRSPath(source, power)
			}
		}).
		Do1(func() (kzg.SRS, error) { return common.LoadCanonicalSRS(source, power) })
	if err != nil {
		return nil, nil, err
	}

	if canonicalSRS == nil {
		canonicalSRS, lagrangeSRS, err = generateSRS(source, power)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate unsafe SRS: %w", err)
		}

		err = common.NewStep0("Saving canonical srs").
			OkMessage(common.CanonicalSRSPath(source, power)).
			Do0(func() error { return common.SaveCanonicalSRS(source, canonicalSRS.(*bn254kzg.SRS)) })
		if err != nil {
			return nil, nil, fmt.Errorf("failed to save canonical SRS: %w", err)
		}

		err = common.NewStep0("Saving lagrange srs").
			OkMessage(common.LagrangeSRSPath(source, power)).
			Do0(func() error { return common.SaveLagrangeSRS(source, lagrangeSRS.(*bn254kzg.SRS)) })
		if err != nil {
			return nil, nil, fmt.Errorf("failed to save lagrange SRS: %w", err)
		}
	}

	if lagrangeSRS == nil {
		lagrangeSRS, err = common.NewStep1[kzg.SRS]("Loading lagrange srs").
			OkMessageFunc1(func(s kzg.SRS) string {
				if s == nil {
					return "Not found " + common.LagrangeSRSPath(source, power) + ", generate now!"
				} else {
					return "Use " + common.LagrangeSRSPath(source, power)
				}
			}).
			Do1(func() (kzg.SRS, error) { return common.LoadLagrangeSRS(source, power) })
		if err != nil {
			return nil, nil, err
		}
	}

	// If still not found lagrange srs, we can generate it from the canonical one.
	if lagrangeSRS == nil {
		lagrangeSRS, err = common.NewStep1[*bn254kzg.SRS]("Generating lagrange srs from canonical srs").
			OkMessageFunc1(func(s *bn254kzg.SRS) string {
				return fmt.Sprintf("Size: %d (2^%d)", len(s.Pk.G1), power)
			}).
			Do1(func() (*bn254kzg.SRS, error) { return canonicalSRSToLagrange(canonicalSRS.(*bn254kzg.SRS)) })
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate Lagrange SRS: %w", err)
		}

		err = common.NewStep0("Saving lagrange srs").
			OkMessage(common.LagrangeSRSPath(source, power)).
			Do0(func() error { return common.SaveLagrangeSRS(source, lagrangeSRS.(*bn254kzg.SRS)) })
		if err != nil {
			return nil, nil, fmt.Errorf("failed to save lagrange SRS: %w", err)
		}
	}

	return canonicalSRS, lagrangeSRS, nil
}
