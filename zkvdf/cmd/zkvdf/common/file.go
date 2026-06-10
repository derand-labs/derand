package common

import (
	"encoding/json"
	"fmt"
	"math/bits"
	"os"
	"path"
	"zkvdf/vdfcircuit"

	"github.com/consensys/gnark-crypto/ecc"
	bn254kzg "github.com/consensys/gnark-crypto/ecc/bn254/kzg"
	"github.com/consensys/gnark-crypto/kzg"
	"github.com/consensys/gnark/backend/plonk"
	be_bn254 "github.com/consensys/gnark/backend/plonk/bn254"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/constraint"
	cs_bn254 "github.com/consensys/gnark/constraint/bn254"
	"github.com/consensys/gnark/frontend"
)

func LoadCircuitSignature(source string, circuitName CircuitName, systemID string) (vdfcircuit.CircuitSignature, error) {
	sig := vdfcircuit.CircuitSignature{}

	var err error
	sig.CCS, err = LoadCS(circuitName, systemID)
	if err != nil {
		return sig, nil
	}

	sig.VK, err = LoadVK(source, circuitName, systemID)
	if err != nil {
		return sig, err
	}

	return sig, nil
}

func LoadCircuitProof(source string, circuitName CircuitName, systemID, proofID string) (vdfcircuit.CircuitProof, error) {
	proof := vdfcircuit.CircuitProof{}

	var err error
	proof.Proof, err = LoadProof(source, circuitName, systemID, proofID)
	if err != nil {
		return proof, nil
	}

	proof.Witness, err = LoadPublicWitness(source, circuitName, systemID, proofID)
	if err != nil {
		return proof, err
	}

	proof.Witness, err = proof.Witness.Public()
	if err != nil {
		return proof, err
	}

	return proof, nil
}

func SaveCS(
	circuitName CircuitName,
	systemID string,
	ccs constraint.ConstraintSystem,
) error {
	f, err := createFile(CSPath(circuitName, systemID))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = ccs.WriteTo(f)
	return err
}

func LoadCS(circuitName CircuitName, systemID string) (constraint.ConstraintSystem, error) {
	f, err := openFile(CSPath(circuitName, systemID))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ccs := new(cs_bn254.R1CS)

	_, err = ccs.ReadFrom(f)
	return ccs, err
}

func LoadVK(source string, circuitName CircuitName, systemID string) (plonk.VerifyingKey, error) {
	f, err := openFile(VKPath(source, circuitName, systemID))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	vk := plonk.NewVerifyingKey(ecc.BN254)

	_, err = vk.ReadFrom(f)
	return vk, err
}

func SaveCanonicalSRS(source string, srs *bn254kzg.SRS) error {
	power := bits.Len(uint(len(srs.Pk.G1) - 4))

	out, err := createFile(CanonicalSRSPath(source, power))
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := srs.WriteTo(out); err != nil {
		out.Close()
		return err
	}

	return nil
}

func SaveLagrangeSRS(source string, srs *bn254kzg.SRS) error {
	power := bits.Len(uint(len(srs.Pk.G1)) - 1)

	out, err := createFile(LagrangeSRSPath(source, power))
	if err != nil {
		return err
	}

	defer out.Close()

	if _, err := srs.WriteTo(out); err != nil {
		return err
	}

	return nil
}

func LoadCanonicalSRS(source string, power int) (kzg.SRS, error) {
	f, err := openFile(CanonicalSRSPath(source, power))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}
	defer f.Close()

	var srs bn254kzg.SRS
	if _, err := srs.ReadFrom(f); err != nil {
		return nil, fmt.Errorf("read canonical srs: %w", err)
	}

	return &srs, nil
}

func LoadLagrangeSRS(source string, power int) (kzg.SRS, error) {
	f, err := openFile(LagrangeSRSPath(source, power))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}
	defer f.Close()

	var srs bn254kzg.SRS
	if _, err := srs.ReadFrom(f); err != nil {
		return nil, fmt.Errorf("read lagrange srs: %w", err)
	}

	return &srs, nil
}

func SavePK(source string, circuitName CircuitName, systemID string, pk plonk.ProvingKey) error {
	f, err := createFile(PKPath(source, circuitName, systemID))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = pk.WriteTo(f)
	return err
}

func SaveVK(source string, circuitName CircuitName, systemID string, vk plonk.VerifyingKey) error {
	f, err := createFile(VKPath(source, circuitName, systemID))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = vk.WriteTo(f)
	return err
}

func SaveSol(source string, circuitName CircuitName, systemID string, vk plonk.VerifyingKey) error {
	f, err := createFile(SolPath(source, circuitName, systemID))
	if err != nil {
		return err
	}
	defer f.Close()

	return vk.ExportSolidity(f)
}

func LoadPK(source string, circuitName CircuitName, systemID string) (plonk.ProvingKey, error) {
	f, err := openFile(PKPath(source, circuitName, systemID))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	pk := plonk.NewProvingKey(ecc.BN254)

	_, err = pk.ReadFrom(f)
	return pk, err
}

func SaveProof(source string, circuitName CircuitName, systemID, proofID string, proof plonk.Proof) error {
	f, err := createFile(ProofPath(source, circuitName, systemID, proofID))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = proof.WriteTo(f)
	return err
}

func SavePublicWitness(source string, circuitName CircuitName, systemID, proofID string, witness witness.Witness) error {
	f, err := createFile(PublicWitnessPath(source, circuitName, systemID, proofID))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = witness.WriteTo(f)
	return err
}

func LoadProof(source string, circuitName CircuitName, systemID, proofID string) (*be_bn254.Proof, error) {
	f, err := openFile(ProofPath(source, circuitName, systemID, proofID))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	proof := &be_bn254.Proof{}
	_, err = proof.ReadFrom(f)
	return proof, err
}

func LoadPublicWitness(source string, circuitName CircuitName, systemID, proofID string) (witness.Witness, error) {
	f, err := openFile(PublicWitnessPath(source, circuitName, systemID, proofID))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	w, err := frontend.NewWitness(nil, ecc.BN254.ScalarField())
	if err != nil {
		return nil, err
	}

	_, err = w.ReadFrom(f)
	return w, err
}

func SaveSolidityProof(source string, circuitName CircuitName, systemID, proofID string, proof *SolidityProof) error {
	f, err := createFile(SolidityProofPath(source, circuitName, systemID, proofID))
	if err != nil {
		return err
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(proof); err != nil {
		return err
	}

	return nil
}

func openFile(f string) (*os.File, error) {
	return os.Open(f)
}

func createFile(f string) (*os.File, error) {
	mkdirp(f)
	return os.Create(f)
}

func mkdirp(f string) {
	d, _ := path.Split(f)
	if err := os.MkdirAll(d, 0o755); err != nil {
		panic("failed to create directory: " + rootdir)
	}
}
