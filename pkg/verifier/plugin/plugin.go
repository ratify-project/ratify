package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/notaryproject/hora/pkg/common"
	pluginCommon "github.com/notaryproject/hora/pkg/common/plugin"
	e "github.com/notaryproject/hora/pkg/executor"
	"github.com/notaryproject/hora/pkg/ocispecs"
	"github.com/notaryproject/hora/pkg/referrerstore"
	"github.com/notaryproject/hora/pkg/verifier"
	"github.com/notaryproject/hora/pkg/verifier/config"
	"github.com/notaryproject/hora/pkg/verifier/types"
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
	referrerStore referrerstore.ReferrerStore,
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
			return verifier.VerifierResult{IsSuccess: false, Name: vp.name, Results: []string{string(encodedResults)}}, nil
		}
	}

	referenceManifest, err := referrerStore.GetReferenceManifest(ctx, subjectReference, referenceDescriptor)

	if err != nil {
		return verifier.VerifierResult{}, err
	}

	result := verifier.VerifierResult{Name: vp.name, IsSuccess: true}
	for _, blobDesc := range referenceManifest.Blobs {
		refBlob, err := referrerStore.GetBlobContent(ctx, subjectReference, blobDesc.Digest)
		if err != nil {
			return verifier.VerifierResult{}, err
		}

		vr, err := vp.verifyReferenceBlob(ctx, subjectReference, refBlob)
		if err != nil {
			return verifier.VerifierResult{}, err
		}

		result.Results = append(result.Results, vr.Results...)
		if !vr.IsSuccess {
			result.IsSuccess = false
			break
		}
	}

	fmt.Fprintf(vp.OutWriter, "Verification of [%s]%s completed with status: %v\n", referenceDescriptor.ArtifactType, referenceDescriptor.Digest, result.IsSuccess)

	return result, nil
}

func (vp *VerifierPlugin) verifyReferenceBlob(ctx context.Context, subjectReference common.Reference, refBlob []byte) (*verifier.VerifierResult, error) {
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
		Config: vp.rawConfig,
		Blob:   refBlob,
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
