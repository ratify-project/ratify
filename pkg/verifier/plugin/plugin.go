package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/deislabs/hora/pkg/common"
	pluginCommon "github.com/deislabs/hora/pkg/common/plugin"
	e "github.com/deislabs/hora/pkg/executor"
	"github.com/deislabs/hora/pkg/ocispecs"
	rc "github.com/deislabs/hora/pkg/referrerstore/config"
	"github.com/deislabs/hora/pkg/verifier"
	"github.com/deislabs/hora/pkg/verifier/config"
	"github.com/deislabs/hora/pkg/verifier/types"
)

type VerifierPlugin struct {
	name          string
	artifactTypes []string
	// TODO Nested reference types
	nestedReferences []string
	version          string
	path             []string
	rawConfig        config.VerifierConfig
	exec             pluginCommon.Exec
	OutWriter        io.Writer
	ErrWriter        io.Writer
}

func NewVerifier(version string, verifierConfig config.VerifierConfig, pluginPaths []string) (verifier.ReferenceVerifier, error) {
	verifierName, ok := verifierConfig[types.Name]
	if !ok {
		return nil, fmt.Errorf("failed to find verifier name in the verifier config with key %s", "name")
	}

	// TODO throw error?
	var nestedReferences []string
	if vs, ok := verifierConfig[types.NestedReferences]; ok {
		nestedReferences = strings.Split(fmt.Sprintf("%s", vs), ",")
	}

	var artifactTypes []string
	if at, ok := verifierConfig[types.ArtifactTypes]; ok {
		// TODO can we get []string directly
		artifactTypes = strings.Split(fmt.Sprintf("%s", at), ",")
	}

	if len(artifactTypes) == 0 {
		artifactTypes = append(artifactTypes, "*")
	}

	return &VerifierPlugin{
		name:             fmt.Sprintf("%s", verifierName),
		version:          version,
		path:             pluginPaths,
		rawConfig:        verifierConfig,
		artifactTypes:    artifactTypes,
		nestedReferences: nestedReferences,
		exec:             &pluginCommon.DefaultExec{Stderr: os.Stderr},
		OutWriter:        os.Stdout,
		ErrWriter:        os.Stderr,
	}, nil
}

func (vp *VerifierPlugin) CanVerify(ctx context.Context, referenceDescriptor ocispecs.ReferenceDescriptor) bool {
	for _, at := range vp.artifactTypes {
		if at == "*" || at == referenceDescriptor.ArtifactType {
			return true
		}
	}
	return false
}

func (vp *VerifierPlugin) Name() string {
	return vp.name
}

func (vp *VerifierPlugin) Verify(ctx context.Context,
	subjectReference common.Reference,
	referenceDescriptor ocispecs.ReferenceDescriptor,
	referrerStoreConfig *rc.StoreConfig,
	executor e.Executor) (verifier.VerifierResult, error) {

	if len(vp.nestedReferences) > 0 {
		verifyParameters := e.VerifyParameters{
			Subject:        fmt.Sprintf("%s@%s", subjectReference.Path, referenceDescriptor.Digest),
			ReferenceTypes: vp.nestedReferences,
		}
		nestedVerifyResult, err := executor.VerifySubject(ctx, verifyParameters)

		if err != nil {
			return verifier.VerifierResult{}, err
		}

		encodedResults, err := json.Marshal(nestedVerifyResult.VerifierReports)
		if err != nil {
			return verifier.VerifierResult{}, err
		}
		if !nestedVerifyResult.IsSuccess {
			return verifier.VerifierResult{
				Subject:   subjectReference.Original,
				IsSuccess: false,
				Name:      vp.name,
				Results:   []string{string(encodedResults)},
			}, nil
		}
	}

	vr, err := vp.verifyReference(ctx, subjectReference, referenceDescriptor, referrerStoreConfig)
	if err != nil {
		return verifier.VerifierResult{}, err
	}

	return *vr, nil

	// fmt.Fprintf(vp.ErrWriter,
	// 	"Verification of [%s]%s completed with status: %v\n",
	// 	referenceDescriptor.ArtifactType,
	// 	referenceDescriptor.Digest,
	// 	result.IsSuccess)
}

func (vp *VerifierPlugin) verifyReference(
	ctx context.Context,
	subjectReference common.Reference,
	referenceDescriptor ocispecs.ReferenceDescriptor,
	referrerStoreConfig *rc.StoreConfig) (*verifier.VerifierResult, error) {
	pluginPath, err := vp.exec.FindInPath(vp.name, vp.path)
	if err != nil {
		return nil, err
	}

	pluginArgs := VerifierPluginArgs{
		Command:          VerifyCommand,
		Version:          vp.version,
		SubjectReference: subjectReference.String(),
		PluginArgs:       nil,
	}

	inputConfig := config.PluginInputConfig{
		Config:       vp.rawConfig,
		StoreConfig:  *referrerStoreConfig,
		ReferencDesc: referenceDescriptor,
	}

	verifierConfigBytes, err := json.Marshal(inputConfig)
	if err != nil {
		return nil, err
	}

	// TODO std writer
	stdoutBytes, err := vp.exec.ExecPlugin(ctx, pluginPath, nil, verifierConfigBytes, pluginArgs.AsEnv())
	if err != nil {
		return nil, err
	}

	result, err := types.GetVerifierResult(stdoutBytes)
	if err != nil {
		return nil, err
	}

	return result, nil
}
