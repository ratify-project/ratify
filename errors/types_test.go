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

package errors

import (
	"errors"
	"strings"
	"testing"
)

const (
	testGroup          = "test-group"
	testErrCode1       = "TEST_ERROR_CODE_1"
	testErrCode2       = "TEST_ERROR_CODE_2"
	testMessage        = "test-message"
	testDescription    = "test-description"
	testDetail1        = "test-detail-1"
	testDetail2        = "test-detail-2"
	testComponentType1 = "test-component-type-1"
	testComponentType2 = "test-component-type-2"
	testLink1          = "test-link-1"
	testLink2          = "test-link-2"
	testPluginName     = "test-plugin-name"
	nonexistentEC      = 2000
)

var (
	testEC = Register(testGroup, ErrorDescriptor{
		Value:       testErrCode1,
		Message:     testMessage,
		Description: testDescription,
	})

	testEC2 = Register(testGroup, ErrorDescriptor{
		Value: testErrCode2,
	})
)

func TestErrorCode(t *testing.T) {
	ec := ErrorCode(1)
	if ec.ErrorCode() != 1 {
		t.Fatalf("ErrorCode() should return 1")
	}
}

func TestError(t *testing.T) {
	expectedStr := "test error code 1"
	if testEC.Error() != expectedStr {
		t.Fatalf("expected: %s, got: %s", expectedStr, testEC.Error())
	}
}

func TestDescriptor(t *testing.T) {
	testCases := []struct {
		name          string
		ec            ErrorCode
		expectedValue string
	}{
		{
			name:          "existing error code",
			ec:            testEC,
			expectedValue: testErrCode1,
		},
		{
			name:          "nonexistent error code",
			ec:            ErrorCode(nonexistentEC),
			expectedValue: "UNKNOWN",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			desc := tc.ec.Descriptor()
			if desc.Value != tc.expectedValue {
				t.Fatalf("expected value: %s, got: %s", tc.expectedValue, desc.Value)
			}
		})
	}
}

func TestMessage(t *testing.T) {
	if testEC.Message() != testMessage {
		t.Fatalf("expected message: %s, got: %s", testMessage, testEC.Message())
	}
}

func TestWithDetail(t *testing.T) {
	err := testEC.WithDetail(testDetail1)
	if err.detail != testDetail1 {
		t.Fatalf("expected detail: %s, got: %s", testDetail1, err.detail)
	}
}

func TestWithError(t *testing.T) {
	err := testEC.WithError(testEC2)
	if !errors.Is(errors.Unwrap(err), testEC2) {
		t.Fatalf("expected error: %s, got: %s", testEC2, errors.Unwrap(err))
	}
}

func TestWithComponentType(t *testing.T) {
	err := testEC.WithComponentType(testComponentType1)
	if err.componentType != testComponentType1 {
		t.Fatalf("expected component type: %s, got: %s", testComponentType1, err.componentType)
	}
}

func TestWithRemediation(t *testing.T) {
	err := testEC.WithRemediation(testLink1)
	if err.remediation != testLink1 {
		t.Fatalf("expected remediation: %s, got: %s", testLink1, err.remediation)
	}
}

func TestWithPluginName(t *testing.T) {
	err := testEC.WithPluginName(testPluginName)
	if err.pluginName != testPluginName {
		t.Fatalf("expected plugin name: %s, got: %s", testPluginName, err.pluginName)
	}
}

func TestWithDescription(t *testing.T) {
	err := testEC.WithDescription()
	if err.description != testDescription {
		t.Fatalf("expected description: %s, got: %s", testDescription, err.description)
	}
}

func TestGetConciseError(t *testing.T) {
	err := testEC.WithDetail("long message, long message, long message")
	if err.GetConciseError(30) != "TEST_ERROR_CODE_1: long mes..." {
		t.Fatalf("expected: TEST_ERROR_CODE_1: long mes..., got: %s", err.GetConciseError(30))
	}

	err = testEC.WithDetail("short message")
	if err.GetConciseError(100) != "TEST_ERROR_CODE_1: short message" {
		t.Fatalf("expected: TEST_ERROR_CODE_1: short message, got: %s", err.GetConciseError(100))
	}
}

func TestIs(t *testing.T) {
	err := testEC.WithDetail(testDetail1)
	result := err.Is(err)
	if !result {
		t.Fatalf("expected true, got: %v", result)
	}

	err2 := errors.New(testMessage)
	result = err.Is(err2)
	if result {
		t.Fatalf("expected false, got: %v", result)
	}
}

func TestError_ErrorCode(t *testing.T) {
	err := Error{
		code: 1,
	}
	if err.ErrorCode() != 1 {
		t.Fatalf("expected 1, got: %d", err.ErrorCode())
	}
}

func TestIsEmpty(t *testing.T) {
	err := Error{}
	if !err.IsEmpty() {
		t.Fatalf("expected true, but got false")
	}

	if testEC.WithDetail("").IsEmpty() {
		t.Fatalf("expected false, but got true")
	}
}

func TestError_Error(t *testing.T) {
	// Nested errors.
	rootErr := testEC.NewError(testComponentType1, "", testLink1, errors.New(testMessage), testDetail1, false)
	err := testEC2.WithPluginName(testPluginName).WithComponentType(testComponentType2).WithRemediation(testLink2).WithDetail(testDetail2).WithError(rootErr)

	expectedErrStr := strings.Join([]string{testErrCode1, testDetail2, testDetail1, testMessage, testLink1}, ": ")
	if err.Error() != expectedErrStr {
		t.Fatalf("expected string: %s, but got: %s", expectedErrStr, err.Error())
	}

	// Single error.
	err = testEC.WithDetail(testDetail1)
	expectedErrStr = "TEST_ERROR_CODE_1: test-detail-1"
	if err.Error() != expectedErrStr {
		t.Fatalf("expected string: %s, but got: %s", expectedErrStr, err.Error())
	}
}

func TestError_GetRootCause(t *testing.T) {
	// rootErr contains original error.
	rootErr := testEC.NewError(testComponentType1, "", testLink1, errors.New(testMessage), testDetail1, false)
	err := testEC.WithPluginName(testPluginName).WithComponentType(testComponentType2).WithRemediation(testLink2).WithDetail(testDetail2).WithError(rootErr)

	if err.GetErrorReason() != testMessage {
		t.Fatalf("expected root cause: %v, but got: %v", err.GetErrorReason(), testMessage)
	}

	// rootErr does not contain original error.
	rootErr = testEC.NewError(testComponentType1, "", testLink1, nil, testDetail1, false)
	err = testEC.WithPluginName(testPluginName).WithComponentType(testComponentType2).WithRemediation(testLink2).WithDetail(testDetail2).WithError(rootErr)

	if err.GetErrorReason() != testDetail1 {
		t.Fatalf("expected root cause: %v, but got: %v", err.GetErrorReason(), testDetail1)
	}
}

func TestError_GetFullDetails(t *testing.T) {
	rootErr := testEC.NewError(testComponentType1, "", testLink1, errors.New(testMessage), testDetail1, false)
	err := testEC.WithPluginName(testPluginName).WithComponentType(testComponentType2).WithRemediation(testLink2).WithDetail(testDetail2).WithError(rootErr)

	expectedDetails := strings.Join([]string{testDetail2, testDetail1}, ": ")
	if err.GetDetail() != expectedDetails {
		t.Fatalf("expected full details: %v, but got: %v", expectedDetails, err.GetDetail())
	}
}

func TestError_GetRootRemediation(t *testing.T) {
	rootErr := testEC.NewError(testComponentType1, "", testLink1, errors.New(testMessage), testDetail1, false)
	err := testEC.WithPluginName(testPluginName).WithComponentType(testComponentType2).WithRemediation(testLink2).WithDetail(testDetail2).WithError(rootErr)

	if err.GetRemediation() != testLink1 {
		t.Fatalf("expected root remediation: %v, but got: %v", err.GetRemediation(), testLink1)
	}
}

func TestNewError(t *testing.T) {
	err := testEC.NewError(testComponentType1, testPluginName, testLink1, Error{}, testDetail1, false)

	if err.componentType != testComponentType1 || err.pluginName != testPluginName || err.remediation != testLink1 || err.detail != testDetail1 {
		t.Fatalf("expected component type: %s, plugin name: %s, link to doc: %s, detail: %s, but got: %s, %s, %s, %s", testComponentType1, testPluginName, testLink1, testDetail1, err.componentType, err.pluginName, err.remediation, err.detail)
	}
}
