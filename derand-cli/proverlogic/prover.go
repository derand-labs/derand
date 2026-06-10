package proverlogic

import (
	"bytes"
	"context"
	"crypto/sha256"
	"derand-cli/backend"
	"derand-cli/config"
	"derand-cli/gen"
	"derand-cli/profile"
	"derand-cli/utils"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"os"
	"os/exec"
	"path"
	"slices"
	"strconv"
	"time"
	zkvdfcommon "zkvdf/cmd/zkvdf/common"
	zkvdfprove "zkvdf/cmd/zkvdf/prove"
	"zkvdf/nonzk"
	"zkvdf/vdf"

	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/constraint"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/shirou/gopsutil/cpu"
)

// Prover is designed for single thread use. Do not spawn it in multithreads.
type Prover struct {
	cfg     *config.Config
	backend *backend.BackendPool
	auth    *bind.TransactOpts
	derand  *gen.DeRand

	requestID     uint64
	submitPreview bool

	x *big.Int
	t uint64

	remoteProfile        gen.ProfileView
	standardLocalProfile profile.StandardClassgroupZKPlonkBn254LocalProfile

	// tmp value
	proofID  string
	r1cs     constraint.ConstraintSystem
	pk       plonk.ProvingKey
	nextr1cs constraint.ConstraintSystem
	nextpk   plonk.ProvingKey
}

func NewProver(requestID uint64, submitPreview bool, password string) (*Prover, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	currentChain, err := cfg.GetCurrentChain()
	if err != nil {
		return nil, err
	}

	backend, err := cfg.GetCurrentChainBackend()
	if err != nil {
		return nil, err
	}

	derand, err := gen.NewDeRand(currentChain.DeRand.Address, backend)
	if err != nil {
		return nil, err
	}

	auth, err := cfg.GetCurrentDefaultTxOpts(password)
	if err != nil {
		return nil, err
	}

	request, err := derand.RequestOf(nil, requestID)
	if err != nil {
		return nil, err
	}

	remoteProfile, _, err := derand.ProfileOf(nil, request.ProfileId, request.ProfileVersion)
	if err != nil {
		return nil, err
	}

	localProfileName, ok := currentChain.RemoteProfileMap[int(request.ProfileId)]
	if !ok {
		return nil, fmt.Errorf("not set local profile for remote profile %d", request.ProfileId)
	}

	localProfile, ok := cfg.LocalProfiles[localProfileName]
	if !ok {
		return nil, fmt.Errorf("invalid configuration: not found local profile name (%s)", localProfileName)
	}

	if localProfile.Data.Type != "standard_classgroup_zk_plonk_bn254" {
		return nil, fmt.Errorf("not support local profile type %s", localProfile.Data.Type)
	}

	return &Prover{
		requestID:            requestID,
		submitPreview:        submitPreview,
		cfg:                  cfg,
		backend:              backend,
		auth:                 auth,
		derand:               derand,
		x:                    request.Seed,
		t:                    uint64(request.Delay) * remoteProfile.DelayScale,
		standardLocalProfile: *localProfile.Data.StandardClassgroupZKPlonkBn254,
	}, nil
}

func NewFindParameterProver(localProfileName string) (*Prover, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	localProfile, ok := cfg.LocalProfiles[localProfileName]
	if !ok {
		return nil, fmt.Errorf("invalid configuration: not found local profile name (%s)", localProfileName)
	}

	if localProfile.Data.Type != "standard_classgroup_zk_plonk_bn254" {
		return nil, fmt.Errorf("not support local profile type %s", localProfile.Data.Type)
	}

	return &Prover{
		cfg:                  cfg,
		x:                    big.NewInt(0),
		standardLocalProfile: *localProfile.Data.StandardClassgroupZKPlonkBn254,
	}, nil
}

func NewBenchmarkProver(profileID uint64) (*Prover, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	currentChain, err := cfg.GetCurrentChain()
	if err != nil {
		return nil, err
	}

	backend, err := cfg.GetCurrentChainBackend()
	if err != nil {
		return nil, err
	}

	derand, err := gen.NewDeRand(currentChain.DeRand.Address, backend)
	if err != nil {
		return nil, err
	}

	auth, err := cfg.GetCurrentDefaultTxOpts("")
	if err != nil {
		return nil, err
	}

	remoteProfile, _, err := derand.ProfileOf(nil, profileID, 0)
	if err != nil {
		return nil, err
	}

	localProfileName, ok := currentChain.RemoteProfileMap[int(profileID)]
	if !ok {
		return nil, fmt.Errorf("not set local profile for remote profile %d", profileID)
	}

	localProfile, ok := cfg.LocalProfiles[localProfileName]
	if !ok {
		return nil, fmt.Errorf("invalid configuration: not found local profile name (%s)", localProfileName)
	}

	if localProfile.Data.Type != "standard_classgroup_zk_plonk_bn254" {
		return nil, fmt.Errorf("not support local profile type %s", localProfile.Data.Type)
	}

	return &Prover{
		cfg:                  cfg,
		backend:              backend,
		auth:                 auth,
		derand:               derand,
		x:                    big.NewInt(0),
		t:                    remoteProfile.DelayScale,
		remoteProfile:        remoteProfile,
		standardLocalProfile: *localProfile.Data.StandardClassgroupZKPlonkBn254,
	}, nil
}

func (p *Prover) Prove() error {
	zkvdfcommon.SetupRootDir(config.GetVDFDir())

	// Phase 0:
	// 1. Proving by corevdf
	// 2. Loading CS, PK of hash_to_form
	if err := p.proveCircuitAndPrepareNext("", zkvdfcommon.HashToFormCircuitName); err != nil {
		return err
	}

	// Phase 1:
	// 1. Proving hash_to_form
	// 2. Prepare CS, PK of intermediate_pow
	if err := p.proveCircuitAndPrepareNext(zkvdfcommon.HashToFormCircuitName, zkvdfcommon.IntermediatePowCircuitName); err != nil {
		return err
	}

	// Phase 2:
	// 1. Proving intermediate_pow
	// 2. Prepare CS, PK of rc_verifier (or rc_verifier_phase_1)
	circuitPhase3 := zkvdfcommon.RCVerifierCircuitName
	if _, ok := p.standardLocalProfile.Signature.Circuits["rc_verifier"]; !ok {
		circuitPhase3 = zkvdfcommon.RCVerifierPhase1CircuitName
	}
	if err := p.proveCircuitAndPrepareNext(zkvdfcommon.IntermediatePowCircuitName, circuitPhase3); err != nil {
		return err
	}

	// Phase 3:
	// 1. Proving rc_verifier (or rc_verifier_phase_1)
	// 2. Prepare CS, PK of rc_verifier_phase_2 (if any)
	circuitLastPhase := zkvdfcommon.RCVerifierCircuitName
	if circuitPhase3 == zkvdfcommon.RCVerifierPhase1CircuitName {
		circuitLastPhase = zkvdfcommon.RCVerifierPhase2CircuitName
	}

	if circuitLastPhase == zkvdfcommon.RCVerifierCircuitName {
		// Phase 4a: Proving rc_verifier
		if err := p.proveCircuit(circuitPhase3); err != nil {
			return err
		}
	} else {
		// Phase 4a: Proving rc_verifier_phase_1 and prepare rc_verifier_phase_2
		if err := p.proveCircuitAndPrepareNext(circuitPhase3, circuitLastPhase); err != nil {
			return err
		}

		// Proving rc_verifier_phase_2
		if err := p.proveCircuit(circuitLastPhase); err != nil {
			return err
		}
	}

	// Verify to create the final zk proof.
	if err := p.verifyCircuit(circuitLastPhase); err != nil {
		return err
	}

	data, err := p.loadProofData(circuitLastPhase)
	if err != nil {
		return err
	}

	if err := p.submit(data); err != nil {
		return err
	}

	return nil
}

func (p *Prover) FindBestParameter(targetDelayTime time.Duration, initT uint64, repeats int) error {
	cpuInfo, err := cpu.Info()
	if err != nil {
		return err
	}

	zkvdfcommon.SetupRootDir(config.GetVDFDir())

	t, err := p.findDelayScaleForTargetDelay(targetDelayTime, initT, repeats)
	if err != nil {
		return err
	}

	logicalCores, _ := cpu.Counts(true)
	ghz := cpuInfo[0].Mhz / 1000.0

	utils.PrintTitle("CPU Info")
	utils.PrintSubtitle("CPU:", cpuInfo[0].ModelName)
	utils.PrintSubtitle("CPU Frequency:", ghz, "Ghz")
	utils.PrintSubtitle("Cores:", logicalCores)
	utils.PrintSubtitle("Found suitable delay scale on this machine:", t)

	p.t = uint64(float64(t*30) / ghz)
	if err := p.corevdfProve(p.t, false, false); err != nil {
		return err
	}

	zkSnarkProve := func() error {
		if err := p.corevdfGenerateTranscript(false); err != nil {
			return err
		}

		if err := p.prepareCircuit(zkvdfcommon.HashToFormCircuitName); err != nil {
			return err
		}

		p.r1cs = p.nextr1cs
		p.pk = p.nextpk

		if err := p.proveCircuitAndPrepareNext(zkvdfcommon.HashToFormCircuitName, zkvdfcommon.IntermediatePowCircuitName); err != nil {
			return err
		}

		circuitPhase3 := zkvdfcommon.RCVerifierCircuitName
		if _, ok := p.standardLocalProfile.Signature.Circuits["rc_verifier"]; !ok {
			circuitPhase3 = zkvdfcommon.RCVerifierPhase1CircuitName
		}
		if err := p.proveCircuitAndPrepareNext(zkvdfcommon.IntermediatePowCircuitName, circuitPhase3); err != nil {
			return err
		}

		circuitLastPhase := zkvdfcommon.RCVerifierCircuitName
		if circuitPhase3 == zkvdfcommon.RCVerifierPhase1CircuitName {
			circuitLastPhase = zkvdfcommon.RCVerifierPhase2CircuitName
		}

		if circuitLastPhase == zkvdfcommon.RCVerifierCircuitName {
			if err := p.proveCircuit(circuitPhase3); err != nil {
				return err
			}
		} else {
			if err := p.proveCircuitAndPrepareNext(circuitPhase3, circuitLastPhase); err != nil {
				return err
			}

			if err := p.proveCircuit(circuitLastPhase); err != nil {
				return err
			}
		}

		return nil
	}

	benchResult, err := benchmark("zkSNARK", repeats, zkSnarkProve)
	if err != nil {
		return err
	}

	utils.PrintTitle("Recommended for the profile")
	utils.PrintSubtitle("Delay scale:", p.t)
	utils.PrintSubtitle("Base time:", int(benchResult.P90/time.Second))

	return nil
}

func (p *Prover) Benchmark(repeats int) error {
	zkvdfcommon.SetupRootDir(config.GetVDFDir())

	delayBenchResult, err := benchmark("VDF delay", repeats, func() error {
		if err := p.corevdfProve(p.t, false, false); err != nil {
			return err
		}

		return p.corevdfGenerateTranscript(false)
	})
	if err != nil {
		return err
	}

	zkSnarkProve := func() error {
		if err := p.prepareCircuit(zkvdfcommon.HashToFormCircuitName); err != nil {
			return err
		}

		p.r1cs = p.nextr1cs
		p.pk = p.nextpk

		if err := p.proveCircuitAndPrepareNext(zkvdfcommon.HashToFormCircuitName, zkvdfcommon.IntermediatePowCircuitName); err != nil {
			return err
		}

		circuitPhase3 := zkvdfcommon.RCVerifierCircuitName
		if _, ok := p.standardLocalProfile.Signature.Circuits["rc_verifier"]; !ok {
			circuitPhase3 = zkvdfcommon.RCVerifierPhase1CircuitName
		}
		if err := p.proveCircuitAndPrepareNext(zkvdfcommon.IntermediatePowCircuitName, circuitPhase3); err != nil {
			return err
		}

		circuitLastPhase := zkvdfcommon.RCVerifierCircuitName
		if circuitPhase3 == zkvdfcommon.RCVerifierPhase1CircuitName {
			circuitLastPhase = zkvdfcommon.RCVerifierPhase2CircuitName
		}

		if circuitLastPhase == zkvdfcommon.RCVerifierCircuitName {
			if err := p.proveCircuit(circuitPhase3); err != nil {
				return err
			}
		} else {
			if err := p.proveCircuitAndPrepareNext(circuitPhase3, circuitLastPhase); err != nil {
				return err
			}

			if err := p.proveCircuit(circuitLastPhase); err != nil {
				return err
			}
		}

		return nil
	}

	zkBenchResult, err := benchmark("zkSNARK", repeats, zkSnarkProve)
	if err != nil {
		return err
	}

	delayTime := time.Duration(p.remoteProfile.DelayTime) * time.Second

	utils.PrintTitle("Generate preview random number")
	utils.PrintSubtitle("Ideal time:", delayTime)
	utils.PrintSubtitle("Expected time:", delayTime*30)
	utils.PrintSubtitle("Actual time:", delayBenchResult.P90)
	switch {
	case delayBenchResult.P90 < 15*delayTime:
		utils.PrintSubtitle("Conclusion:", utils.Green("(safe)"))
	case delayBenchResult.P90 < 18*delayTime:
		utils.PrintSubtitle("Conclusion:", utils.Yellow("(tight - buffer is thin; small jitter may push it close to the limit)"))
	case delayBenchResult.P90 < 20*delayTime:
		utils.PrintSubtitle("Conclusion:", utils.Yellow("(dangerous - high chance of missing the allowed window)"))
	default:
		utils.PrintSubtitle("Conclusion:", utils.Red("(failed - likely to miss the requirement)"))
	}

	baseTime := time.Duration(p.remoteProfile.BaseTime) * time.Second
	utils.PrintTitle("Generate final proof")
	utils.PrintSubtitle("Ideal time:", baseTime)
	utils.PrintSubtitle("Expected time:", baseTime*2)
	utils.PrintSubtitle("Actual time:", zkBenchResult.P90)
	switch {
	case zkBenchResult.P90 < 15*baseTime:
		utils.PrintSubtitle("Conclusion:", utils.Green("(safe)"))
	case zkBenchResult.P90 < 18*baseTime:
		utils.PrintSubtitle("Conclusion:", utils.Yellow("(tight - buffer is thin; small jitter may push it close to the limit)"))
	case zkBenchResult.P90 < 20*baseTime:
		utils.PrintSubtitle("Conclusion:", utils.Yellow("(dangerous - high chance of missing the allowed window)"))
	default:
		utils.PrintSubtitle("Conclusion:", utils.Red("(failed - likely to miss the requirement)"))
	}

	return nil
}

func (p *Prover) proveCircuitAndPrepareNext(circuitThisPhase, circuitNextPhase zkvdfcommon.CircuitName) error {
	wg := newWaitGroup()
	wg.Run(func() error {
		if circuitThisPhase == "" {
			if err := p.corevdfProve(p.t, false, true); err != nil {
				return err
			}

			if p.submitPreview {
				proofPath := path.Join(
					config.GetVDFDir(), "proofs", p.standardLocalProfile.GetSystemID(),
					fmt.Sprintf("proof-%s.json", p.proofID),
				)

				proofFile, err := os.Open(proofPath)
				if err != nil {
					return err
				}

				var proof struct {
					Y  nonzk.Form `json:"y"`
					Pi nonzk.Form `json:"pi"`
				}
				if err := json.NewDecoder(proofFile).Decode(&proof); err != nil {
					return err
				}

				setup, err := vdf.LoadSetup(zkvdfcommon.SystemPath(p.standardLocalProfile.GetSystemID()))
				if err != nil {
					return err
				}

				args := abi.Arguments{{Type: yABI}}
				data, err := args.Pack(zkvdfcommon.NewSolidityForm(setup, proof.Y))
				if err != nil {
					return err
				}

				utils.PrintTitle("Submit preview:")
				p.auth.GasLimit = 1000000
				submitTx, err := p.derand.SubmitPreviewRandomNumber(p.auth, p.requestID, data)
				if err != nil {
					name, args, err := utils.DecodeCustomError(gen.DeRandABI, err)
					if err != nil {
						return fmt.Errorf("failed to submit random number: %w", err)
					}

					return fmt.Errorf("%s%v", name, args)
				}
				utils.PrintSubtitle("Transaction:", submitTx.Hash())
			}

			return p.corevdfGenerateTranscript(true)
		}
		return p.proveCircuit(circuitThisPhase)
	})
	if circuitNextPhase != "" {
		wg.Run(func() error {
			return p.prepareCircuit(circuitNextPhase)
		})
	}
	if err := wg.Wait(); err != nil {
		return err
	}

	p.r1cs = p.nextr1cs
	p.pk = p.nextpk

	return nil
}

func (p *Prover) corevdfGenerateTranscript(showOutput bool) error {
	cmd := exec.Command(
		path.Join(config.GetBuildDir(), "corevdf"),
		"verify",
		"--setup-dir", config.GetVDFDir(),
		"--system", p.standardLocalProfile.GetSystemID(),
		"--x-seed", ethcommon.Bytes2Hex(p.x.Bytes()),
		"-t", strconv.FormatUint(p.t, 10),
	)

	if showOutput {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run corevdf verify error: %w", err)
	}

	return nil
}

func (p *Prover) corevdfProve(t uint64, evalOnly bool, showOutput bool) error {
	args := []string{
		"prove",
		"--setup-dir", config.GetVDFDir(),
		"--system", p.standardLocalProfile.GetSystemID(),
		"--x-seed", ethcommon.Bytes2Hex(p.x.Bytes()),
		"-t", strconv.FormatUint(t, 10),
	}
	if evalOnly {
		args = append(args, "--eval-only")
	}

	cmd := exec.Command(path.Join(config.GetBuildDir(), "corevdf"), args...)
	if showOutput {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run corevdf prove error: %w", err)
	}

	// Compute proof_id
	hash := sha256.New()
	hash.Write(p.x.Bytes())

	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], p.t)
	hash.Write(buf[:])

	p.proofID = ethcommon.Bytes2Hex(hash.Sum(nil))

	return nil
}

func (p *Prover) prepareCircuit(name zkvdfcommon.CircuitName) error {
	cs, err := zkvdfcommon.LoadCS(name, p.standardLocalProfile.GetSystemID())
	if err != nil {
		return fmt.Errorf("load cs of %s: %w", name, err)
	}

	pk, err := zkvdfcommon.LoadPK(p.standardLocalProfile.SRSSource, name, p.standardLocalProfile.GetSystemID())
	if err != nil {
		return fmt.Errorf("load pk of %s: %w", name, err)
	}

	p.nextr1cs = cs
	p.nextpk = pk

	return nil
}

func (p *Prover) proveCircuit(name zkvdfcommon.CircuitName) error {
	assignments, err := zkvdfprove.LoadAssignments(
		p.standardLocalProfile.SRSSource, name, p.standardLocalProfile.GetSystemID(), p.proofID)
	if err != nil {
		return err
	}

	for i := range assignments {
		if err := zkvdfprove.RunProve(
			p.r1cs, p.pk,
			p.standardLocalProfile.SRSSource,
			p.standardLocalProfile.GetSystemID(),
			p.proofID,
			assignments[i],
		); err != nil {
			return err
		}
	}

	return nil
}

func (p *Prover) verifyCircuit(name zkvdfcommon.CircuitName) error {
	cmd := exec.Command(
		path.Join(config.GetBuildDir(), "zkvdf"),
		"verify",
		"--setup-dir", config.GetVDFDir(),
		"--system", p.standardLocalProfile.GetSystemID(),
		"--circuit", string(name),
		"--srs-source", p.standardLocalProfile.SRSSource,
		"--proof", p.proofID,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run corevdf verify error: %w", err)
	}

	return nil
}

func (p *Prover) loadProofData(circuitLastPhase zkvdfcommon.CircuitName) ([]byte, error) {
	proofPath := path.Join(
		config.GetVDFDir(), "zk", "systems", p.standardLocalProfile.GetSystemID(),
		p.standardLocalProfile.SRSSource, "proofs", p.proofID,
	)
	proofPath = path.Join(proofPath, fmt.Sprintf("solidity-proof-%s.json", circuitLastPhase))

	proofFile, err := os.Open(proofPath)
	if err != nil {
		return nil, err
	}

	proof := zkvdfcommon.SolidityProof{}
	if err := json.NewDecoder(proofFile).Decode(&proof); err != nil {
		return nil, err
	}

	args := abi.Arguments{{Type: proofABI}}
	data, err := args.Pack(proof)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (p *Prover) submit(data []byte) error {
	utils.PrintTitle("Submit proof")
	p.auth.GasLimit = 1000000
	submitTx, err := p.derand.SubmitRandomNumber(p.auth, p.requestID, data)
	if err != nil {
		name, args, err := utils.DecodeCustomError(gen.DeRandABI, err)
		if err != nil {
			return fmt.Errorf("failed to submit random number: %w", err)
		}

		return fmt.Errorf("%s%v", name, args)
	}

	utils.PrintSubtitle("Transaction:", submitTx.Hash())

	receipt, err := bind.WaitMined(context.Background(), p.backend, submitTx)
	if err != nil {
		return err
	}

	var randomNumber *big.Int
	for _, vLog := range receipt.Logs {
		ev, err := p.derand.ParseRandomNumber(*vLog)
		if err == nil {
			randomNumber = ev.RandomNumber
		}
	}

	if randomNumber == nil {
		return fmt.Errorf("error when submit, check the transaction!")
	}

	utils.PrintSubtitle("Random number:", "0x"+hex.EncodeToString(randomNumber.Bytes()))
	return nil
}

func (p *Prover) findDelayScaleForTargetDelay(targetDelayTime time.Duration, initT uint64, repeats int) (uint64, error) {
	if initT == 0 {
		initT = 30000000
	}

	bench, err := benchmark(fmt.Sprintf("VDF delay with init T (%d)", initT), repeats, func() error {
		return p.corevdfProve(initT, true, false)
	})
	if err != nil {
		return 0, err
	}

	return uint64(targetDelayTime * time.Duration(initT) / bench.Min), nil
}

type benchmarkResult struct {
	Min    time.Duration
	Max    time.Duration
	Mean   time.Duration
	Median time.Duration
	P90    time.Duration
}

func benchmark(name string, n int, f func() error) (*benchmarkResult, error) {
	var sum time.Duration
	results := make([]time.Duration, n)

	utils.PrintTitle("Benchmark", name, "in", n, "times")
	for i := range n {
		start := time.Now()

		oldStdout := os.Stdout
		r, w, err := os.Pipe()
		if err != nil {
			return nil, err
		}

		os.Stdout = w

		if err := f(); err != nil {
			os.Stdout = oldStdout
			var buf bytes.Buffer
			io.Copy(&buf, r)
			return nil, fmt.Errorf("%w: %s", err, buf.String())
		}

		os.Stdout = oldStdout
		d := time.Since(start)
		utils.PrintSubtitle(fmt.Sprintf("[%d]: %s", i, d))

		results[i] = d
		sum += d
	}

	slices.Sort(results)

	min := results[0]
	max := results[len(results)-1]

	var median time.Duration
	if n%2 == 1 {
		median = results[n/2]
	} else {
		median = (results[n/2-1] + results[n/2]) / 2
	}

	p90Index := int(math.Ceil(0.90*float64(n))) - 1
	if p90Index < 0 {
		p90Index = 0
	}
	if p90Index >= n {
		p90Index = n - 1
	}

	result := &benchmarkResult{
		Min:    min,
		Max:    max,
		Mean:   sum / time.Duration(n),
		Median: median,
		P90:    results[p90Index],
	}

	utils.PrintSubtitle("Min:", result.Min)
	utils.PrintSubtitle("Max:", result.Max)
	utils.PrintSubtitle("Mean:", result.Mean)
	utils.PrintSubtitle("Median:", result.Median)
	utils.PrintSubtitle("P90:", result.P90)

	return result, nil
}
