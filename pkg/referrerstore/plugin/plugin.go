package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/notaryproject/hora/pkg/common"
	pluginCommon "github.com/notaryproject/hora/pkg/common/plugin"
	"github.com/notaryproject/hora/pkg/ocispecs"
	"github.com/notaryproject/hora/pkg/referrerstore"
	"github.com/notaryproject/hora/pkg/referrerstore/config"
	"github.com/notaryproject/hora/pkg/referrerstore/types"
	"github.com/opencontainers/go-digest"
)

type StorePlugin struct {
	name      string
	version   string
	path      []string
	rawConfig config.StoreConfig
	exec      pluginCommon.Exec
}

func NewStore(version string, storeConfig config.StoreConfig, pluginPaths []string) (referrerstore.ReferrerStore, error) {
	storeName, ok := storeConfig[types.Name]
	if !ok {
		return nil, fmt.Errorf("failed to find store name in the stores config with key %s", "name")
	}

	return &StorePlugin{
		name:      fmt.Sprintf("%s", storeName),
		version:   version,
		path:      pluginPaths,
		rawConfig: storeConfig,
		exec:      &pluginCommon.DefaultExec{Stderr: os.Stderr},
	}, nil
}

func (sp *StorePlugin) ListReferrers(ctx context.Context, subjectReference common.Reference, artifactTypes []string, nextToken string) (referrerstore.ListReferrersResult, error) {
	pluginPath, err := sp.exec.FindInPath(sp.name, sp.path)
	if err != nil {
		return referrerstore.ListReferrersResult{}, err
	}

	args := [][2]string{
		{"nextToken", nextToken},
		{"artifactTypes", strings.Join(artifactTypes, ",")},
	}

	pluginArgs := ReferrerStorePluginArgs{
		Command:          ListReferrersCommand,
		Version:          sp.version,
		SubjectReference: subjectReference.String(),
		PluginArgs:       args,
	}

	storeConfigBytes, err := json.Marshal(sp.rawConfig)
	if err != nil {
		return referrerstore.ListReferrersResult{}, err
	}

	// TODO std writer
	stdoutBytes, err := sp.exec.ExecPlugin(ctx, pluginPath, nil, storeConfigBytes, pluginArgs.AsEnv())
	if err != nil {
		return referrerstore.ListReferrersResult{}, err
	}

	listResult, err := types.GetListReferrersResult(stdoutBytes)
	if err != nil {
		return referrerstore.ListReferrersResult{}, err
	}

	return listResult, nil
}

func (sp *StorePlugin) Name() string {
	return sp.name
}

func (sp *StorePlugin) GetBlobContent(ctx context.Context, subjectReference common.Reference, digest digest.Digest) ([]byte, error) {
	pluginPath, err := sp.exec.FindInPath(sp.name, sp.path)
	if err != nil {
		return nil, err
	}

	args := [][2]string{
		{"digest", digest.String()},
	}

	pluginArgs := ReferrerStorePluginArgs{
		Command:          GetBlobContentCommand,
		Version:          sp.version,
		SubjectReference: subjectReference.String(),
		PluginArgs:       args,
	}

	storeConfigBytes, err := json.Marshal(sp.rawConfig)
	if err != nil {
		return nil, err
	}

	stdoutBytes, err := sp.exec.ExecPlugin(ctx, pluginPath, nil, storeConfigBytes, pluginArgs.AsEnv())
	if err != nil {
		return nil, err
	}

	return stdoutBytes, nil
}

func (sp *StorePlugin) GetReferenceManifest(ctx context.Context, subjectReference common.Reference, referenceDesc ocispecs.ReferenceDescriptor) (ocispecs.ReferenceManifest, error) {
	pluginPath, err := sp.exec.FindInPath(sp.name, sp.path)
	if err != nil {
		return ocispecs.ReferenceManifest{}, err
	}

	args := [][2]string{
		{"digest", referenceDesc.Digest.String()},
	}

	pluginArgs := ReferrerStorePluginArgs{
		Command:          GetRefManifestCommand,
		Version:          sp.version,
		SubjectReference: subjectReference.String(),
		PluginArgs:       args,
	}

	storeConfigBytes, err := json.Marshal(sp.rawConfig)
	if err != nil {
		return ocispecs.ReferenceManifest{}, err
	}

	// TODO std writer
	stdoutBytes, err := sp.exec.ExecPlugin(ctx, pluginPath, nil, storeConfigBytes, pluginArgs.AsEnv())
	if err != nil {
		return ocispecs.ReferenceManifest{}, err
	}

	manifest, err := types.GetReferenceManifestResult(stdoutBytes)
	if err != nil {
		return ocispecs.ReferenceManifest{}, err
	}

	return manifest, nil
}
