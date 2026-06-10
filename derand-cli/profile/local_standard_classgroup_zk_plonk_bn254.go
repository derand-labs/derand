package profile

import (
	"bytes"
	"crypto/sha256"
	"derand-cli/sol"
	"derand-cli/utils"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

var StandardClassgroupZKPlonkBn254CircuitOrder = []string{
	"hash_to_form",
	"intermediate_pow",
	"rc_verifier",
	"rc_verifier_phase_1",
	"rc_verifier_phase_2",
}

type StandardClassgroupZKPlonkBn254LocalProfile struct {
	// System configuration
	Seed                 hexutil.Bytes `json:"seed"`
	DBits                uint16        `json:"d_bits"`
	LimbBits             uint16        `json:"limb_bits"`
	SplitExp             uint16        `json:"split_exp"`
	HashToFormGenerators uint16        `json:"hash_to_form_generators"`
	HashToFormSteps      uint16        `json:"hash_to_form_steps"`

	// zkSNARK setup configuration
	SRSSource string `json:"srs_source"`

	Signature StandardClassgroupZKPlonkBn254Signature `json:"signature"`
}

func (p *StandardClassgroupZKPlonkBn254LocalProfile) GetWarnings() []ProfileWarning {
	warnings := []ProfileWarning{}

	switch {
	case p.DBits < 1024:
		warnings = append(warnings, ProfileWarning{
			Level:   ProfileWarningLevelCritical,
			Message: "d-bits: dangerous (test purposes only)",
		})

	case 1024 <= p.DBits && p.DBits < 3072:
		warnings = append(warnings, ProfileWarning{
			Level:   ProfileWarningLevelMedium,
			Message: "d-bits: low (expected 6656)",
		})

	case 3072 <= p.DBits && p.DBits < 6656:
		warnings = append(warnings, ProfileWarning{
			Level:   ProfileWarningLevelLow,
			Message: "d-bits: below recommended security level (expected 6656)",
		})
	}

	expectedLimbBits := computeLimbBits(int(p.DBits))
	if p.LimbBits >= 128 {
		warnings = append(warnings, ProfileWarning{
			Level:   ProfileWarningLevelCritical,
			Message: "limb-bits: invalid value; proofs will fail if limb-bits is 128 or greater (must be less than 128)",
		})
	} else if p.LimbBits != uint16(expectedLimbBits) {
		warnings = append(warnings, ProfileWarning{
			Level:   ProfileWarningLevelLow,
			Message: fmt.Sprintf("limb-bits: optimal value is %d", expectedLimbBits),
		})
	}

	h2fSpace := int(Log2CombRepeat(int64(p.HashToFormGenerators), int64(p.HashToFormSteps)))
	switch {
	case h2fSpace < 128:
		warnings = append(warnings, ProfileWarning{
			Level:   ProfileWarningLevelCritical,
			Message: fmt.Sprintf("hash-to-form: the entropy space is too small (required at least 2^128, got 2^%d)", h2fSpace),
		})
	case h2fSpace < 256:
		warnings = append(warnings, ProfileWarning{
			Level:   ProfileWarningLevelMedium,
			Message: fmt.Sprintf("hash-to-form: below recommended entropy space (required 2^256, got 2^%d)", h2fSpace),
		})
	}

	if p.SRSSource == "unsafe" {
		warnings = append(warnings, ProfileWarning{
			Level:   ProfileWarningLevelCritical,
			Message: "srs source: use snarkjs or perpetual instead",
		})
	}

	return warnings
}

func (p *StandardClassgroupZKPlonkBn254LocalProfile) Validate(vdfDir string) error {
	if err := p.validateSystem(vdfDir, false); err != nil {
		return err
	}

	if err := p.validateCircuits(vdfDir, false); err != nil {
		return err
	}

	return nil
}

func (p *StandardClassgroupZKPlonkBn254LocalProfile) Install(buildDir, vdfDir string, reinstall bool) error {
	if err := p.installSystem(buildDir, vdfDir, reinstall); err != nil {
		return err
	}

	if err := p.installCircuits(buildDir, vdfDir, reinstall); err != nil {
		return err
	}

	return nil
}

func (p *StandardClassgroupZKPlonkBn254LocalProfile) DeployVerifier(
	vdfDir string,
	auth *bind.TransactOpts,
	backend bind.ContractBackend,
	hashToPrime128Address ethcommon.Address,
) (ethcommon.Address, ethcommon.Hash, error) {
	abi, bytecode, err := p.CompileVerifier(vdfDir)
	if err != nil {
		return ethcommon.Address{}, ethcommon.Hash{}, err
	}

	address, tx, _, err := bind.DeployContract(auth, abi, bytecode, backend, hashToPrime128Address)
	if err != nil {
		return ethcommon.Address{}, ethcommon.Hash{}, err
	}

	return address, tx.Hash(), nil
}

func (p *StandardClassgroupZKPlonkBn254LocalProfile) CompileVerifier(vdfDir string) (abi.ABI, []byte, error) {
	verifier, err := p.GetVerifierWrapper(vdfDir)
	if err != nil {
		return abi.ABI{}, nil, err
	}

	output, err := sol.Compile("Verifier.sol", verifier)
	if err != nil {
		return abi.ABI{}, nil, err
	}

	parsedABI, err := abi.JSON(strings.NewReader(string(output.Contracts["Verifier.sol"]["Verifier"].ABI)))
	if err != nil {
		return abi.ABI{}, nil, err
	}

	bytecode := common.FromHex(output.Contracts["Verifier.sol"]["Verifier"].EVM.Bytecode.Object)

	return parsedABI, bytecode, nil
}

func (p *StandardClassgroupZKPlonkBn254LocalProfile) GetVerifierWrapper(vdfDir string) (string, error) {
	name := "rc_verifier"
	if _, ok := p.Signature.Circuits["rc_verifier_phase_2"]; ok {
		name = "rc_verifier_phase_2"
	}

	path := path.Join(vdfDir, "zk", "systems", p.GetSystemID(), p.SRSSource, name+".sol")
	verifierContent, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	templateArgs := map[string]any{
		"plonk_verifier": string(verifierContent),
		"d_bits":         p.DBits,
		"limb_bits":      p.LimbBits,
	}

	var buf bytes.Buffer
	if err := standardClassgroupZKPlonkBn254Template.Execute(&buf, templateArgs); err != nil {
		return "", err
	}

	content := buf.String()
	content = strings.ReplaceAll(content, "pragma solidity ^0.8.0", "pragma solidity 0.8.34")

	return content, nil
}

func (p *StandardClassgroupZKPlonkBn254LocalProfile) installSystem(buildDir, vdfDir string, reinstall bool) error {
	allowSetupWhenSignatureEmpty := !reinstall

	utils.PrintTitle("Setup system")
	if p.validateSystem(vdfDir, allowSetupWhenSignatureEmpty) != nil {
		cmd := exec.Command(
			path.Join(buildDir, "corevdf"),
			"setup",
			"--setup-dir", vdfDir,
			"--seed", p.Seed.String()[2:],
			"--d-bits", strconv.Itoa(int(p.DBits)),
			"--l-bits", "128",
			"--limb-bits", strconv.Itoa(int(p.LimbBits)),
			"--split-exp", strconv.Itoa(int(p.SplitExp)),
			"--hash-to-form-nb-generators", strconv.Itoa(int(p.HashToFormGenerators)),
			"--hash-to-form-steps", strconv.Itoa(int(p.HashToFormSteps)),
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("setup vdf system error: %w", err)
		}
	}

	if err := p.validateSystem(vdfDir, true); err != nil {
		return fmt.Errorf("setup vdf system error: %w", err)
	}

	utils.PrintSubtitle(utils.Green("done"))
	return nil
}

func (p *StandardClassgroupZKPlonkBn254LocalProfile) GetSystemID() string {
	h := sha256.New()

	h.Write([]byte(p.Seed))

	var buf [2]byte

	binary.BigEndian.PutUint16(buf[:], p.DBits)
	h.Write(buf[:])

	binary.BigEndian.PutUint16(buf[:], 128) // standard l-bits
	h.Write(buf[:])

	binary.BigEndian.PutUint16(buf[:], p.LimbBits)
	h.Write(buf[:])

	binary.BigEndian.PutUint16(buf[:], p.SplitExp)
	h.Write(buf[:])

	binary.BigEndian.PutUint16(buf[:], p.HashToFormGenerators)
	h.Write(buf[:])

	binary.BigEndian.PutUint16(buf[:], p.HashToFormSteps)
	h.Write(buf[:])

	return hex.EncodeToString(h.Sum(nil))
}

func (p *StandardClassgroupZKPlonkBn254LocalProfile) validateSystem(vdfDir string, allowSetWhenEmpty bool) error {
	systemPath := path.Join(vdfDir, "systems", "system-"+p.GetSystemID()+".json")

	systemFile, err := os.Open(systemPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("system %w", ErrNotFound)
		}

		return err
	}

	standardSystem := StandardSystem{}
	if err := json.NewDecoder(systemFile).Decode(&standardSystem); err != nil {
		return err
	}

	gotSignature := standardSystem.GetHash()
	if bytes.Equal(gotSignature, p.Signature.System) {
		return nil
	}

	if len(p.Signature.System) == 0 && allowSetWhenEmpty {
		p.Signature.System = gotSignature
		return nil
	}

	return fmt.Errorf("mismatched system signature: expected %s != got %s", p.Signature.System, gotSignature)
}

func (p *StandardClassgroupZKPlonkBn254LocalProfile) installCircuits(buildDir, vdfDir string, allowSetWhenEmpty bool) error {
	for _, name := range StandardClassgroupZKPlonkBn254CircuitOrder {
		if _, ok := p.Signature.Circuits[name]; !ok {
			continue
		}

		exportSol := false
		if name == "rc_verifier" || name == "rc_verifier_phase_2" {
			exportSol = true
		}

		if err := p.installSingleCircuit(buildDir, vdfDir, name, exportSol, allowSetWhenEmpty); err != nil {
			return err
		}
	}

	return nil
}

func (p *StandardClassgroupZKPlonkBn254LocalProfile) validateCircuits(vdfDir string, allowSetWhenEmpty bool) error {
	for _, name := range StandardClassgroupZKPlonkBn254CircuitOrder {
		if _, ok := p.Signature.Circuits[name]; !ok {
			continue
		}

		if err := p.validateSingleCircuit(vdfDir, name, allowSetWhenEmpty); err != nil {
			return err
		}
	}

	return nil
}

func (p *StandardClassgroupZKPlonkBn254LocalProfile) installSingleCircuit(buildDir, vdfDir, name string, exportSol, reinstall bool) error {
	allowSetWhenIfFileExists := !reinstall

	utils.PrintTitle("Compile circuit " + name)
	if p.validateSingleCircuitFile(vdfDir, name, "r1cs", allowSetWhenIfFileExists) != nil {
		cmd := exec.Command(
			path.Join(buildDir, "zkvdf"),
			"compile",
			"--setup-dir", vdfDir,
			"--srs-source", p.SRSSource,
			"--system", p.GetSystemID(),
			"--circuit", name,
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("compile %s error: %w", name, err)
		}
	}

	if err := p.validateSingleCircuitFile(vdfDir, name, "r1cs", true); err != nil {
		return fmt.Errorf("compile %s error: %w", name, err)
	}

	utils.PrintSubtitle(utils.Green("done"))

	utils.PrintTitle("Setup circuit " + name)
	pkErr := p.validateSingleCircuitFile(vdfDir, name, "pk", allowSetWhenIfFileExists)
	vkErr := p.validateSingleCircuitFile(vdfDir, name, "vk", allowSetWhenIfFileExists)
	var solErr error
	if exportSol {
		solErr = p.validateSingleCircuitFile(vdfDir, name, "sol", allowSetWhenIfFileExists)
	}

	if pkErr != nil || vkErr != nil || solErr != nil {
		args := []string{
			"setup",
			"--setup-dir", vdfDir,
			"--srs-source", p.SRSSource,
			"--system", p.GetSystemID(),
			"--circuit", name,
		}
		if exportSol {
			args = append(args, "--sol")
		}

		cmd := exec.Command(path.Join(buildDir, "zkvdf"), args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("setup %s error: %w", name, err)
		}
	}

	if err := p.validateSingleCircuitFile(vdfDir, name, "pk", true); err != nil {
		return fmt.Errorf("setup %s error: %w", name, err)
	}

	if err := p.validateSingleCircuitFile(vdfDir, name, "vk", true); err != nil {
		return fmt.Errorf("setup %s error: %w", name, err)
	}

	if exportSol {
		if err := p.validateSingleCircuitFile(vdfDir, name, "sol", true); err != nil {
			return fmt.Errorf("setup %s error: %w", name, err)
		}
	}

	utils.PrintSubtitle(utils.Green("done"))

	return p.validateSingleCircuit(vdfDir, name, false)
}

func (p *StandardClassgroupZKPlonkBn254LocalProfile) validateSingleCircuit(vdfDir, name string, allowSetWhenEmpty bool) error {
	ts := []string{"r1cs", "pk", "vk"}
	if name == "rc_verifier" || name == "rc_verifier_phase_2" {
		ts = append(ts, "sol")
	}

	for _, t := range ts {
		if err := p.validateSingleCircuitFile(vdfDir, name, t, allowSetWhenEmpty); err != nil {
			return err
		}
	}

	return nil
}

func (p *StandardClassgroupZKPlonkBn254LocalProfile) validateSingleCircuitFile(vdfDir, name, t string, allowSetWhenEmpty bool) error {
	dir := path.Join(vdfDir, "zk", "systems", p.GetSystemID())
	if t == "pk" || t == "vk" || t == "sol" {
		dir = path.Join(dir, p.SRSSource)
	}

	signature, err := utils.SHA256File(path.Join(dir, name+"."+t))
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s %s %w", name, t, ErrNotFound)
		}

		return fmt.Errorf("%s %s %w", name, t, err)
	}

	if bytes.Equal(signature, p.Signature.Circuits[name][t]) {
		return nil
	}

	if len(p.Signature.Circuits[name][t]) == 0 && allowSetWhenEmpty {
		p.Signature.Circuits[name][t] = signature
	} else {
		return fmt.Errorf("mismatched %s %s signature: expected %s != got %s",
			name, t, p.Signature.Circuits[name][t], hexutil.Bytes(signature))
	}

	return nil
}

type StandardClassgroupZKPlonkBn254Signature struct {
	System hexutil.Bytes `json:"system"`

	// circuit -> type -> hash
	// circuit: hash_to_form, intermediate_pow, rc_verifier, rc_verifier_phase_1, rc_verifier_phase_2
	// type: r1cs, pk, vk, solidity
	Circuits map[string]map[string]hexutil.Bytes `json:"circuits"`
}

func Log2CombRepeat(n, k int64) float64 {
	if k < 0 || n < 0 {
		panic("negative input")
	}

	if k == 0 {
		return 0
	}

	if n == 0 {
		panic("no generators")
	}

	return log2Comb(n+k-1, k)
}

func log2Comb(n, k int64) float64 {
	if n < 0 || k < 0 || k > n {
		panic("invalid n,k")
	}

	if k > n-k {
		k = n - k
	}

	sum := 0.0

	for i := int64(1); i <= k; i++ {
		sum += math.Log2(float64(n - k + i))
		sum -= math.Log2(float64(i))
	}

	return sum
}

func log2ceil(x int) int {
	if x <= 0 {
		panic("x must be positive")
	}
	return int(math.Ceil(math.Log2(float64(x))))
}

func ceilDiv(a, b int) int {
	return (a + b - 1) / b
}

const fieldBits = 253

func valid(limbBits, nLimbs int) bool {
	// require: 2*limbBits + log2(nLimbs) <= 253
	return 2*limbBits+log2ceil(nLimbs) <= fieldBits
}

func limbs(dbits, limbBits int) int {
	return ceilDiv(dbits, limbBits)
}

func findNumLimbOfMaxLimbBits(dbits int) int {
	bestLimbs := 0

	for limbBits := 1; limbBits <= fieldBits; limbBits++ {
		n := limbs(dbits, limbBits)
		if valid(limbBits, n) {
			bestLimbs = n
		} else {
			break
		}
	}

	return bestLimbs
}

func findMinLimbBits(dbits, fixedLimbs int) int {
	for limbBits := 1; limbBits <= fieldBits; limbBits++ {
		n := limbs(dbits, limbBits)
		if n <= fixedLimbs && valid(limbBits, fixedLimbs) {
			return limbBits
		}
	}
	return 0
}

// full pipeline
func computeLimbBits(dbits int) int {
	limbs := findNumLimbOfMaxLimbBits(dbits)
	return findMinLimbBits(dbits, limbs)
}
