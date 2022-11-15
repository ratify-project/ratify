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

package plugin

import (
	"context"
	"testing"
)

type TestExecutor struct {
	find    func(plugin string, paths []string) (string, error)
	execute func(ctx context.Context, pluginPath string, cmdArgs []string, stdinData []byte, environ []string) ([]byte, error)
}

func (e *TestExecutor) ExecutePlugin(ctx context.Context, pluginPath string, cmdArgs []string, stdinData []byte, environ []string) ([]byte, error) {
	return e.execute(ctx, pluginPath, cmdArgs, stdinData, environ)
}

func (e *TestExecutor) FindInPaths(plugin string, paths []string) (string, error) {
	return e.find(plugin, paths)
}

func TestNewVerifier_Expected(t *testing.T) {
	verifierConfig := map[string]interface{}{
		"name":             "test-verifier",
		"artifactTypes":    "test1,test2",
		"nestedReferences": "ref1,ref2",
	}

	verifier, err := NewVerifier("1.0.0", verifierConfig, []string{})
	if err != nil {
		t.Fatalf("failed to create plugin store %v", err)
	}

	if vc, ok := verifier.(*VerifierPlugin); !ok {
		t.Fatal("type assertion failed. expected plugin verifier")
	} else if len(vc.artifactTypes) != 2 {
		t.Fatalf("expected number of artifact Types 2, actual %d", len(vc.artifactTypes))
	} else if len(vc.nestedReferences) != 2 {
		t.Fatalf("expected number of nested references is 2, actual %d", len(vc.nestedReferences))
	}
}
