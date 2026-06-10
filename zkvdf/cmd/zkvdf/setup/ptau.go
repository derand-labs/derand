package setup

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"math/bits"
	"os"
	"os/exec"
	"path"
	"time"
	"zkvdf/cmd/zkvdf/common"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fp"
	bn254fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	bn254fft "github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
	bn254kzg "github.com/consensys/gnark-crypto/ecc/bn254/kzg"
	"github.com/consensys/gnark-crypto/kzg"
)

func generateSRS(source string, power int) (kzg.SRS, kzg.SRS, error) {
	if power > 26 {
		common.PrintWarn("This circuit is too large and may cause the build to fail")
		common.PrintWarn("Please consider stopping this process and optimizing your circuit")
		time.Sleep(5 * time.Second)
	}

	if source == "unsafe" {
		return generateUnsafeSRS(power)
	}

	path := ""
	url := ""
	switch source {
	case "snarkjs":
		switch {
		case power >= 8 && power < 28:
			url = fmt.Sprintf("https://storage.googleapis.com/zkevm/ptau/powersOfTau28_hez_final_%02d.ptau", power)
		case power == 28:
			url = "https://storage.googleapis.com/zkevm/ptau/powersOfTau28_hez_final.ptau"
		default:
			return nil, nil, fmt.Errorf("unsupported snarkjs ptau power")
		}

	case "perpetual":
		switch {
		case power >= 1 && power < 28:
			url = fmt.Sprintf("https://pse-trusted-setup-ppot.s3.eu-central-1.amazonaws.com/pot28_0080/ppot_0080_%02d.ptau", power)
		case power == 28:
			url = "https://pse-trusted-setup-ppot.s3.eu-central-1.amazonaws.com/pot28_0080/ppot_0080_final.ptau"
		default:
			return nil, nil, fmt.Errorf("unsupported perpetual ptau power")
		}

	default:
		return nil, nil, fmt.Errorf("invalid source: %s", source)
	}

	path, err := downloadFile(url)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to download ptau from snarkjs: %w", err)
	}

	return generateSRSFromPtau(path)
}

func generateSRSFromPtau(path string) (kzg.SRS, kzg.SRS, error) {
	canonicalSRS, err := common.NewStep1[*bn254kzg.SRS]("Converting ptau to canonical srs").
		OkMessageFunc1(func(s *bn254kzg.SRS) string {
			return fmt.Sprintf("Size: %d (2^%d+3)", len(s.Pk.G1), bits.Len(uint(len(s.Pk.G1)-4)))
		}).
		Do1(func() (*bn254kzg.SRS, error) { return canonicalSRSFromPtau(path) })
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert ptau to canonical SRS: %w", err)
	}

	lagrangeSRS, err := common.NewStep1[*bn254kzg.SRS]("Generating lagrange srs from canonical srs").
		OkMessageFunc1(func(s *bn254kzg.SRS) string {
			return fmt.Sprintf("Size: %d (2^%d)", len(s.Pk.G1), bits.Len(uint(len(s.Pk.G1)-1)))
		}).
		Do1(func() (*bn254kzg.SRS, error) { return canonicalSRSToLagrange(canonicalSRS) })
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate Lagrange SRS: %w", err)
	}

	return canonicalSRS, lagrangeSRS, nil
}

func generateUnsafeSRS(power int) (kzg.SRS, kzg.SRS, error) {
	tau, err := rand.Int(rand.Reader, ecc.BN254.ScalarField())
	if err != nil {
		return nil, nil, err
	}

	canonicalSRS, err := bn254kzg.NewSRS(1<<power+3, tau)
	if err != nil {
		return nil, nil, err
	}

	ttau := new(bn254fr.Element).SetBigInt(tau)
	lagrangeSRS := &bn254kzg.SRS{Vk: canonicalSRS.Vk}
	size := uint64(1 << power)

	// instead of using ToLagrangeG1 we can directly do a fft on the powers of alpha
	// since we know the randomness in test.
	pAlpha := make([]bn254fr.Element, size)
	pAlpha[0].SetUint64(1)
	for i := 1; i < len(pAlpha); i++ {
		pAlpha[i].Mul(&pAlpha[i-1], ttau)
	}
	// do a fft on this.
	d := bn254fft.NewDomain(size)
	d.FFTInverse(pAlpha, bn254fft.DIF)
	bn254fft.BitReverse(pAlpha)

	// bath scalar mul
	_, _, g1gen, _ := bn254.Generators()
	lagrangeSRS.Pk.G1 = bn254.BatchScalarMultiplicationG1(&g1gen, pAlpha)

	return canonicalSRS, lagrangeSRS, nil
}

func downloadFile(url string) (string, error) {
	out, err := os.CreateTemp("", "*.download")
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	out.Close()

	fdir, fname := path.Split(out.Name())
	cmd := exec.Command(
		"aria2c",
		"-x", "16",
		"-s", "16",
		"-c", url,
		"-d", fdir,
		"-o", fname,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}

	return out.Name(), nil
}

// canonicalSRSFromPtau converts a SnarkJS PTAU file to a Gnark bn254 KZG SRS.
//
// See https://github.com/iden3/snarkjs/blob/e44656d9e7b451250038211e44c1a7d80dd76b89/src/powersoftau_new.js#L20-L66.
func canonicalSRSFromPtau(filename string) (*bn254kzg.SRS, error) {
	reader, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	var ptauStr = make([]byte, 4)
	if _, err := io.ReadFull(reader, ptauStr); err != nil {
		return nil, err
	}

	// version
	if _, err := readULE32(reader); err != nil {
		return nil, err
	}

	// number of sections
	sections, err := readULE32(reader)
	if err != nil {
		return nil, err
	}
	if sections < 3 {
		return nil, fmt.Errorf("unexpected section count %d, expected at least 3", sections)
	}

	var srs bn254kzg.SRS
	power := uint32(0)
	expectedSectionLengths := []func() uint64{
		func() uint64 { return bn254fr.Bytes + 12 },                           // 32-byte prime + 3x 32-bit ints
		func() uint64 { return (uint64(1<<power)*2 - 1) * bn254fr.Bytes * 2 }, // 2^power*2-1 G1 elements
		func() uint64 { return uint64(1<<power) * bn254fr.Bytes * 4 },         // 2^power G2 elements
	}
	for i := range 3 {
		section, err := readULE32(reader)
		if err != nil {
			return nil, err
		}
		if section != uint32(i+1) {
			return nil, fmt.Errorf("unexpected section %d, expected %d", section, i+1)
		}
		length, err := readULE64(reader)
		if err != nil {
			return nil, err
		}
		expectedLength := expectedSectionLengths[i]()
		if length != expectedLength {
			return nil, fmt.Errorf("unexpected length %d for section %d, expected %d", length, section, expectedLength)
		}
		switch i {
		case 0:
			power, err = readHeader(reader)
		case 1:
			err = readG1Array(reader, power, &srs)
		case 2:
			err = readG2Array(reader, &srs)
		}
		if err != nil {
			return nil, err
		}
	}

	srs.Vk.Lines[0] = bn254.PrecomputeLines(srs.Vk.G2[0])
	srs.Vk.Lines[1] = bn254.PrecomputeLines(srs.Vk.G2[1])

	return &srs, nil
}

func canonicalSRSToLagrange(canonicalSRS *bn254kzg.SRS) (*bn254kzg.SRS, error) {
	lagrangeSRSG1, err := bn254kzg.ToLagrangeG1(canonicalSRS.Pk.G1[:len(canonicalSRS.Pk.G1)-3])
	if err != nil {
		return nil, err
	}

	return &bn254kzg.SRS{Vk: canonicalSRS.Vk, Pk: bn254kzg.ProvingKey{G1: lagrangeSRSG1}}, nil
}

func readHeader(reader io.Reader) (uint32, error) {
	numberOfBytes, err := readULE32(reader)
	if err != nil {
		return 0, err
	}
	if numberOfBytes != bn254fr.Bytes {
		return 0, fmt.Errorf("unexpected n8 %d, expected %d", numberOfBytes, bn254fr.Bytes)
	}
	// prime
	if _, err := readElement(reader); err != nil {
		return 0, err
	}
	// power
	power, err := readULE32(reader)
	if err != nil {
		return 0, err
	}
	// ceremonyPower
	if _, err := readULE32(reader); err != nil {
		return 0, err
	}
	return power, nil
}

func readG1Array(reader io.Reader, power uint32, srs *bn254kzg.SRS) error {
	numPoints := uint64(1<<power)*2 - 1
	srs.Pk.G1 = make([]bn254.G1Affine, numPoints)

	var err error
	for i := range numPoints {
		srs.Pk.G1[i], err = readG1(reader)
		if err != nil {
			return err
		}
		if !srs.Pk.G1[i].IsOnCurve() {
			return fmt.Errorf(
				"G1 not on curve: \n index: %d g1Affine.X: %v \n g1Affine.Y: %v \n", i, srs.Pk.G1[i].X, srs.Pk.G1[i].Y)
		}
	}
	srs.Pk.G1 = srs.Pk.G1[:1<<power+3]
	srs.Vk.G1 = srs.Pk.G1[0]
	return nil
}

func readG2Array(reader io.Reader, srs *bn254kzg.SRS) error {
	var err error
	for i := range 2 {
		srs.Vk.G2[i], err = readG2(reader)
		if err != nil {
			return err
		}
		if !srs.Vk.G2[i].IsOnCurve() {
			return fmt.Errorf("tauG2: \n index: %d, g2Affine.X.A0: %v \n g2Affine.X.A1: %v \n g2Affine.Y.A0: %v \n g2Affine.Y.A1 %v \n", i, srs.Vk.G2[i].X.A0, srs.Vk.G2[i].X.A1, srs.Vk.G2[i].Y.A0, srs.Vk.G2[i].Y.A1)
		}
	}
	return nil
}

func readG1(reader io.Reader) (bn254.G1Affine, error) {
	var g1 bn254.G1Affine
	var err error
	g1.X, err = readElement(reader)
	if err != nil {
		return bn254.G1Affine{}, err
	}
	g1.Y, err = readElement(reader)
	if err != nil {
		return bn254.G1Affine{}, err
	}
	return g1, nil
}

func readG2(reader io.Reader) (bn254.G2Affine, error) {
	var g2 bn254.G2Affine
	var err error
	g2.X.A0, err = readElement(reader)
	if err != nil {
		return bn254.G2Affine{}, err
	}
	g2.X.A1, err = readElement(reader)
	if err != nil {
		return bn254.G2Affine{}, err
	}
	g2.Y.A0, err = readElement(reader)
	if err != nil {
		return bn254.G2Affine{}, err
	}
	g2.Y.A1, err = readElement(reader)
	if err != nil {
		return bn254.G2Affine{}, err
	}
	return g2, nil
}

func readULE32(reader io.Reader) (uint32, error) {
	var buffer = make([]byte, 4)
	_, err := io.ReadFull(reader, buffer)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buffer), nil
}

func readULE64(reader io.Reader) (uint64, error) {
	var buffer = make([]byte, 8)
	if _, err := io.ReadFull(reader, buffer); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(buffer), nil
}

func readElement(reader io.Reader) (fp.Element, error) {
	var b = make([]byte, bn254fr.Bytes)
	if _, err := io.ReadFull(reader, b); err != nil {
		return fp.Element{}, err
	}
	var z fp.Element
	z[0] = binary.LittleEndian.Uint64(b[0:8])
	z[1] = binary.LittleEndian.Uint64(b[8:16])
	z[2] = binary.LittleEndian.Uint64(b[16:24])
	z[3] = binary.LittleEndian.Uint64(b[24:32])
	return z, nil
}
