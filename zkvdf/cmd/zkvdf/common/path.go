package common

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"path"
	"time"
)

var (
	rootdir = ".vdf"
)

func SetupRootDir(d string) {
	rootdir = d
}

func SystemPath(id string) string {
	return pathOf(fmt.Sprintf("system-%s.json", id), "systems")
}

func TranscriptPath(systemID, id string) string {
	return pathOf(fmt.Sprintf("transcript-%s.json", id), "proofs", systemID)
}

func CSFileName(name CircuitName) string {
	return string(name) + ".r1cs"
}

func CSPath(name CircuitName, systemID string) string {
	return pathOf(CSFileName(name), "zk", "systems", systemID)
}

func PKFileName(name CircuitName) string {
	return string(name) + ".pk"
}

func PKPath(source string, name CircuitName, systemID string) string {
	return pathOf(PKFileName(name), "zk", "systems", systemID, source)
}

func VKFileName(name CircuitName) string {
	return string(name) + ".vk"
}

func VKPath(source string, name CircuitName, systemID string) string {
	return pathOf(VKFileName(name), "zk", "systems", systemID, source)
}

func SolFileName(name CircuitName) string {
	return string(name) + ".sol"
}

func SolPath(source string, name CircuitName, systemID string) string {
	return pathOf(SolFileName(name), "zk", "systems", systemID, source)
}

func CanonicalSRSFileName(power int) string {
	return fmt.Sprintf("canonical-%d.srs", power)
}

func CanonicalSRSPath(source string, power int) string {
	return pathOf(CanonicalSRSFileName(power), "zk", "srs", source)
}

func LagrangeSRSFileName(power int) string {
	return fmt.Sprintf("lagrange-%d.srs", power)
}

func LagrangeSRSPath(source string, power int) string {
	return pathOf(LagrangeSRSFileName(power), "zk", "srs", source)
}

func ProofPath(source string, circuitName CircuitName, systemID, proofID string) string {
	return pathOf(fmt.Sprintf("%s.proof", circuitName), "zk", "systems", systemID, source, "proofs", proofID)
}

func PublicWitnessPath(source string, circuitName CircuitName, systemID, name string) string {
	return pathOf(fmt.Sprintf("%s.witness", circuitName), "zk", "systems", systemID, source, "proofs", name)
}

func SolidityProofPath(source string, circuitName CircuitName, systemID, name string) string {
	return pathOf(fmt.Sprintf("solidity-proof-%s.json", circuitName), "zk", "systems", systemID, source, "proofs", name)
}

func pathOf(filename string, subdirs ...string) string {
	d := []string{rootdir}
	d = append(d, subdirs...)
	d = append(d, filename)
	return path.Join(d...)
}

func randomHex(nBytes int) string {
	b := make([]byte, nBytes)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Read(b)
	return hex.EncodeToString(b)
}
