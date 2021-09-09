package main

import (
	"encoding/json"
	"fmt"

	"github.com/deislabs/hora/pkg/common"
	"github.com/deislabs/hora/pkg/ocispecs"
	"github.com/deislabs/hora/pkg/referrerstore"
	"github.com/deislabs/hora/pkg/verifier"
	"github.com/deislabs/hora/pkg/verifier/plugin/skel"
)

type PluginConfig struct {
	Name string `json:"name"`
	// config specific to the plugin
}

type PluginInputConfig struct {
	Config PluginConfig `json:"config"`
}

func main() {
	skel.PluginMain("sample", "1.0.0", VerifyReference, []string{"1.0.0"})
}

func parseInput(stdin []byte) (*PluginConfig, error) {
	conf := PluginInputConfig{}

	if err := json.Unmarshal(stdin, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse stdin for the input: %v", err)
	}

	return &conf.Config, nil
}

func VerifyReference(args *skel.CmdArgs, subjectReference common.Reference, referenceDescriptor ocispecs.ReferenceDescriptor, referrerStore referrerstore.ReferrerStore) (*verifier.VerifierResult, error) {
	input, err := parseInput(args.StdinData)
	if err != nil {
		return nil, err
	}

	return &verifier.VerifierResult{
		Name:      input.Name,
		IsSuccess: referenceDescriptor.Size > 0,
		Results:   []string{"Sample verification success"},
	}, nil
}
