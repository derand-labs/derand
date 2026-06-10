package utils

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

func DecodeCustomError(contractABI string, err error) (string, []any, error) {
	parsed, err2 := abi.JSON(strings.NewReader(contractABI))
	if err2 != nil {
		return "", nil, err2
	}

	// Catch two cases:
	// 1) execution reverted: custom error 0x3001c4a9
	// 2) execution reverted: custom error 0x3001c4a9: <args hex>
	re := regexp.MustCompile(`custom error (0x[0-9a-fA-F]{8})(?::\s*(?:0x)?([0-9a-fA-F]*))?`)
	m := re.FindStringSubmatch(err.Error())
	if m == nil {
		return "", nil, err
	}

	selectorBytes, err2 := hex.DecodeString(strings.TrimPrefix(m[1], "0x"))
	if err2 != nil {
		return "", nil, err2
	}

	data := selectorBytes
	if len(m) >= 3 && m[2] != "" {
		bodyBytes, err2 := hex.DecodeString(strings.TrimPrefix(m[2], "0x"))
		if err2 != nil {
			return "", nil, err2
		}
		data = append(data, bodyBytes...)
	}

	if len(data) < 4 {
		return "", nil, fmt.Errorf("revert data too short")
	}

	var id [4]byte
	copy(id[:], data[:4])

	er, err2 := parsed.ErrorByID(id)
	if err2 != nil {
		return "", nil, err2
	}

	// Custom error has no parameter
	if len(data) == 4 {
		return er.Name, nil, nil
	}

	args, err2 := er.Inputs.Unpack(data[4:])
	if err2 != nil {
		return er.Name, nil, err2
	}

	return er.Name, args, nil
}
