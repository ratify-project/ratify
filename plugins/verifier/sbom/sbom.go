package main

import (
	"encoding/json"
	"fmt"

	"github.com/notaryproject/hora/pkg/common"
	"github.com/notaryproject/hora/pkg/verifier"
	"github.com/notaryproject/hora/pkg/verifier/plugin/skel"
)

type PluginConfig struct {
	Name string `json:"name"`
}

type PluginInput struct {
	Config PluginConfig `json:"config"`
	Blob   []byte       `json:"blob"`
}

type SbomContents struct {
	Contents string `json:"contents"`
}

func main() {
	skel.PluginMain("sbom", "1.0.0", VerifyReference, []string{"1.0.0"})
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

	var content SbomContents
	if err := json.Unmarshal(input.Blob, &content); err != nil {
		return nil, fmt.Errorf("failed to parse sbom: %v", err)
	}

	return &verifier.VerifierResult{
		Name:      input.Config.Name,
		IsSuccess: content.Contents == "good",
		Results:   []string{fmt.Sprintf("SBOM verification completed. contents %s", content.Contents)},
	}, nil
}
