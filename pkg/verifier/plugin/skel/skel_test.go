/*
Copyright The Ratify Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package skel

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore"

	"github.com/deislabs/ratify/pkg/verifier/plugin"
	"github.com/deislabs/ratify/pkg/verifier/types"

	"github.com/deislabs/ratify/pkg/verifier"
)

func TestPluginMain_VerifyReference_ReturnsExpected(t *testing.T) {
	verifyReference := func(args *CmdArgs, subjectReference common.Reference, referenceDescriptor ocispecs.ReferenceDescriptor, referrerStore referrerstore.ReferrerStore) (*verifier.VerifierResult, error) {
		if referenceDescriptor.ArtifactType != "test-type" {
			t.Fatalf("expected artifact type %s actual %s", "test-type", referenceDescriptor.ArtifactType)
		}

		if referrerStore.Name() != "test-store" {
			t.Fatalf("expected store name %s actual %s", "test-store", referrerStore.Name())
		}
		return &verifier.VerifierResult{IsSuccess: true}, nil
	}

	environment := map[string]string{
		plugin.CommandEnvKey: plugin.VerifyCommand,
		plugin.VersionEnvKey: "1.0.0",
		plugin.SubjectEnvKey: "localhost:5000/net-monitor:v1@sha256:a0fc570a245b09ed752c42d600ee3bb5b4f77bbd70d8898780b7ab43454530eb",
	}

	stdinData := `{ "storeConfig" : {"store": {"name":"test-store", "some": "config"}}, "config": {"name": "skel-test-case", "some":"config"}, "referenceDesc": {"artifactType": "test-type"}}`
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	pluginContext := &pcontext{
		GetEnviron: func(key string) string { return environment[key] },
		Stdin:      strings.NewReader(stdinData),
		Stdout:     stdout,
		Stderr:     stderr,
	}

	err := pluginContext.pluginMainCore("skel-test-case", "1.0.0", verifyReference, []string{"1.0.0"})
	if err != nil {
		t.Fatalf("plugin execution failed %v", err)
	}

	out := fmt.Sprintf("%s", stdout)
	if !strings.Contains(out, `"isSuccess":true`) {
		t.Fatalf("plugin execution failed. expected %v actual %v", "isSuccess: true", out)
	}
}

func TestPluginMain_ErrorCases(t *testing.T) {
	verifyReference := func(args *CmdArgs, subjectReference common.Reference, referenceDescriptor ocispecs.ReferenceDescriptor, referrerStore referrerstore.ReferrerStore) (*verifier.VerifierResult, error) {
		return nil, fmt.Errorf("simulated error")
	}
	environment := map[string]string{
		plugin.CommandEnvKey: plugin.VerifyCommand,
		plugin.ArgsEnvKey:    "digest=sha256:a0fc570a245b09ed752c42d600ee3bb5b4f77bbd70d8898780b7ab43454530eb",
		plugin.SubjectEnvKey: "localhost:5000/net-monitor:v1@sha256:a0fc570a245b09ed752c42d600ee3bb5b4f77bbd70d8898780b7ab43454530eb",
	}

	stdinData := `{ "storeConfig" : {"store": {"name":"test-store", "some": "config"}}, "config": {"name": "skel-test-case", "some":"config"}, "referenceDesc": {"artifactType": "test-type"}}`
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	pluginContext := &pcontext{
		GetEnviron: func(key string) string { return environment[key] },
		Stdin:      strings.NewReader(stdinData),
		Stdout:     stdout,
		Stderr:     stderr,
	}

	err := pluginContext.pluginMainCore("skel-test-case", "1.0.0", verifyReference, []string{"1.0.0"})
	if err == nil || err.Code != types.ErrMissingEnvironmentVariables {
		t.Fatalf("plugin execution expected to fail with error code %d", types.ErrMissingEnvironmentVariables)
	}

	environment[plugin.VersionEnvKey] = "1.0.0"
	environment[plugin.SubjectEnvKey] = "localhost&300"

	err = pluginContext.pluginMainCore("skel-test-case", "1.0.0", verifyReference, []string{"1.0.0"})
	if err == nil || err.Code != types.ErrArgsParsingFailure {
		t.Fatalf("plugin execution expected to fail with error code %d for invalid subject", types.ErrArgsParsingFailure)
	}

	environment[plugin.SubjectEnvKey] = "localhost:5000/net-monitor:v1@sha256:a0fc570a245b09ed752c42d600ee3bb5b4f77bbd70d8898780b7ab43454530eb"
	environment[plugin.VersionEnvKey] = "2.0.0"
	err = pluginContext.pluginMainCore("skel-test-case", "1.0.0", verifyReference, []string{"1.0.0"})
	if err == nil || err.Code != types.ErrVersionNotSupported {
		t.Fatalf("plugin execution expected to fail with error code %d for unsupported version", types.ErrVersionNotSupported)
	}

	environment[plugin.VersionEnvKey] = "1.0.0"

	stdinData = `"storeConfig" : {"store": {"name":"test-store", "some": "config"}}, "config": {"name": "skel-test-case", "some":"config"}, "referenceDesc": {"artifactType": "test-type"}}`
	pluginContext.Stdin = strings.NewReader(stdinData)
	err = pluginContext.pluginMainCore("skel-test-case", "1.0.0", verifyReference, []string{"1.0.0"})
	if err == nil || err.Code != types.ErrConfigParsingFailure {
		t.Fatalf("plugin execution expected to fail with error code %d for invalid config", types.ErrConfigParsingFailure)
	}

	stdinData = `{"storeConfig" : {"store": {"name":"test-store", "some": "config"}}, "config": {"some":"config"}, "referenceDesc": {"artifactType": "test-type"}}`
	pluginContext.Stdin = strings.NewReader(stdinData)
	err = pluginContext.pluginMainCore("skel-test-case", "1.0.0", verifyReference, []string{"1.0.0"})
	if err == nil || err.Code != types.ErrInvalidVerifierConfig {
		t.Fatalf("plugin execution expected to fail with error code %d for missing verifier name", types.ErrInvalidVerifierConfig)
	}

	environment[plugin.CommandEnvKey] = "unknown"
	stdinData = `{"storeConfig" : {"store": {"name":"test-store", "some": "config"}}, "config": {"name": "skel-test-case", "some":"config"}, "referenceDesc": {"artifactType": "test-type"}}`
	pluginContext.Stdin = strings.NewReader(stdinData)
	err = pluginContext.pluginMainCore("skel-test-case", "1.0.0", verifyReference, []string{"1.0.0"})
	if err == nil || err.Code != types.ErrUnknownCommand {
		t.Fatalf("plugin execution expected to fail with error code %d for invalid command", types.ErrUnknownCommand)
	}

	environment[plugin.CommandEnvKey] = plugin.VerifyCommand
	stdinData = `{"storeConfig" : {"store": {"name":"test-store", "some": "config"}}, "config": {"name": "skel-test-case", "some":"config"}, "referenceDesc": {"artifactType": "test-type"}}`
	pluginContext.Stdin = strings.NewReader(stdinData)
	err = pluginContext.pluginMainCore("skel-test-case", "1.0.0", verifyReference, []string{"1.0.0"})
	if err == nil || err.Code != types.ErrPluginCmdFailure {
		t.Fatalf("plugin execution expected to fail with error code %d for cmd failure", types.ErrPluginCmdFailure)
	}
}
