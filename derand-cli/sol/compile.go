package sol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

func Compile(name, source string) (*SolcOutput, error) {
	jsonInput, err := json.Marshal(buildInput(name, source))
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("solc", "--standard-json")
	cmd.Stdin = bytes.NewReader(jsonInput)

	raw, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("solc error: %s", string(raw))
	}

	var out SolcOutput
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}

	if len(out.Errors) != 0 {
		return nil, fmt.Errorf("%s", solcErrorsString(out.Errors))
	}

	return &out, nil
}

func buildInput(name, source string) solcInput {
	return solcInput{
		Language: "Solidity",
		Sources: map[string]sourceUnit{
			name: {
				Content: source,
			},
		},
		Settings: fixedSetting,
	}
}
