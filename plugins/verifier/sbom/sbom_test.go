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
package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProcessSPDXJsonMediaType(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "bom.json"))
	if err != nil {
		t.Fatalf("error reading %s", filepath.Join("testdata", "bom.json"))
	}
	vr, err := processSpdxJsonMediaType("test", b)
	if err != nil {
		t.Fatalf("expected to process spdx json file: %s", filepath.Join("testdata", "bom.json"))
	}
	if !vr.IsSuccess {
		t.Fatalf("expected to successfully verify schema")
	}
}

func TestProcessInvalidSPDXJsonMediaType(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "invalid-bom.json"))
	if err != nil {
		t.Fatalf("error reading %s", filepath.Join("testdata", "invalid-bom.json"))
	}
	_, err = processSpdxJsonMediaType("test", b)
	if err == nil {
		t.Fatalf("expected to have an error processing spdx json file: %s", filepath.Join("testdata", "bom.json"))
	}
}
