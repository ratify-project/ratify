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

package schemavalidation

import (
	"os"
	"testing"
)

var schema_url = "https://json.schemastore.org/sarif-2.1.0-rtm.5.json"
var schema_file_bytes []byte
var schema_file_mismatch_bytes []byte
var schema_file_bad_bytes []byte
var trivy_scan_report []byte

func init() {
	trivy_scan_report, _ = os.ReadFile("./testdata/trivy_scan_report.json")
	schema_file_bytes, _ = os.ReadFile("./schemas/sarif-2.1.0-rtm.5.json")
	schema_file_mismatch_bytes, _ = os.ReadFile("./testdata/mismatch_schema.json")
	schema_file_bad_bytes, _ = os.ReadFile("./testdata/bad_schema.json")
}

func TestProperSchemaValidates(t *testing.T) {
	expected := true
	result := Validate(schema_url, trivy_scan_report) == nil

	if expected != result {
		t.Logf("expected: %v, got: %v", expected, result)
		t.FailNow()
	}
}

func TestInvalidSchemaFailsValidation(t *testing.T) {
	expected := false
	result := Validate("bad schema", trivy_scan_report) == nil

	if expected != result {
		t.Logf("expected: %v, got: %v", expected, result)
		t.FailNow()
	}
}

func TestProperSchemaValidatesFromFile(t *testing.T) {
	expected := true
	result := ValidateAgainstOfflineSchema(schema_file_bytes, trivy_scan_report) == nil

	if expected != result {
		t.Logf("expected: %v, got: %v", expected, result)
		t.FailNow()
	}
}

func TestSchemaMismatchFromFile(t *testing.T) {
	expected := false
	result := ValidateAgainstOfflineSchema(schema_file_mismatch_bytes, trivy_scan_report) == nil

	if expected != result {
		t.Logf("expected: %v, got: %v", expected, result)
		t.FailNow()
	}
}

func TestBadSchemaValidatesFromFile(t *testing.T) {
	expected := false
	result := ValidateAgainstOfflineSchema(schema_file_bad_bytes, trivy_scan_report) == nil

	if expected != result {
		t.Logf("expected: %v, got: %v", expected, result)
		t.FailNow()
	}
}
