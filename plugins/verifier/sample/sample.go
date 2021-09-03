package main

import (
	"encoding/json"
	"fmt"

	"github.com/deislabs/hora/pkg/common"
	"github.com/deislabs/hora/pkg/verifier"
	"github.com/deislabs/hora/pkg/verifier/plugin/skel"
)

type PluginConfig struct {
	Name string `json:"name"`
}

type PluginInput struct {
	Config PluginConfig `json:"config"`
	Blob   []byte       `json:"blob"`
}

func main() {
	skel.PluginMain("sample", "1.0.0", VerifyReference, []string{"1.0.0"})
}

func parseInput(stdin []byte) (*PluginInput, error) {
	conf := PluginInput{}

	if err := json.Unmarshal(stdin, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse stdin for the input: %v", err)
	}

	return &conf, nil
}

func VerifyReference(args *skel.CmdArgs, subjectReference common.Reference) (*verifier.VerifierResult, error) {
	input, err := parseInput(args.StdinData)
	if err != nil {
		return nil, err
	}

	return &verifier.VerifierResult{
		Name:      input.Config.Name,
		IsSuccess: len(input.Blob) > 0,
		Results:   []string{"Sample verification success"},
	}, nil
}
