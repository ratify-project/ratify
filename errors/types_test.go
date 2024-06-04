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
	testGroup         = "test-group"
	testValue         = "test-value"
	testMessage       = "test-message"
	testDescription   = "test-description"
	testDetail        = "test-detail"
	testComponentType = "test-component-type"
	testLink          = "test-link"
	testPluginName    = "test-plugin-name"
	testErrorString   = "Original Error: (Error: , Code: UNKNOWN), Error: test-message, Code: test-value, Plugin Name: test-plugin-name, Component Type: test-component-type, Documentation: test-link, Detail: test-detail"
	nonexistentEC     = 2000
)

var (
	testEC = Register(testGroup, ErrorDescriptor{
		Value:       testValue,
		Message:     testMessage,
		Description: testDescription,
	})

	testEC2 = Register(testGroup, ErrorDescriptor{})
)

func TestErrorCode(t *testing.T) {
	ec := ErrorCode(1)
	if ec.ErrorCode() != 1 {
		t.Fatalf("ErrorCode() should return 1")
	}
}

func TestError(t *testing.T) {
	if testEC.Error() != testValue {
		t.Fatalf("expected: %s, got: %s", testValue, testEC.Error())
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
			expectedValue: testValue,
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
	err := testEC.WithDetail(testDetail)
	if err.Detail != testDetail {
		t.Fatalf("expected detail: %s, got: %s", testDetail, err.Detail)
	}
}

func TestWithError(t *testing.T) {
	err := testEC.WithError(testEC2)
	if !errors.Is(errors.Unwrap(err), testEC2) {
		t.Fatalf("expected error: %s, got: %s", testEC2, errors.Unwrap(err))
	}
}

func TestWithComponentType(t *testing.T) {
	err := testEC.WithComponentType(testComponentType)
	if err.ComponentType != testComponentType {
		t.Fatalf("expected component type: %s, got: %s", testComponentType, err.ComponentType)
	}
}

func TestWithLinkToDoc(t *testing.T) {
	err := testEC.WithLinkToDoc(testLink)
	if err.LinkToDoc != testLink {
		t.Fatalf("expected link to doc: %s, got: %s", testLink, err.LinkToDoc)
	}
}

func TestWithPluginName(t *testing.T) {
	err := testEC.WithPluginName(testPluginName)
	if err.PluginName != testPluginName {
		t.Fatalf("expected plugin name: %s, got: %s", testPluginName, err.PluginName)
	}
}

func TestWithDescription(t *testing.T) {
	err := testEC.WithDescription()
	if err.Description != testDescription {
		t.Fatalf("expected description: %s, got: %s", testDescription, err.Description)
	}
}

func TestIs(t *testing.T) {
	err := testEC.WithDetail(testDetail)
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
		Code: 1,
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
	err := testEC.WithPluginName(testPluginName).WithComponentType(testComponentType).WithLinkToDoc(testLink).WithDetail(testDetail).WithError(Error{}).WithDescription()
	result := err.Error()
	if !strings.HasPrefix(result, testErrorString) {
		t.Fatalf("expected string starts with: %s, but got: %s", testErrorString, result)
	}
}

func TestNewError(t *testing.T) {
	err := testEC.NewError(testComponentType, testPluginName, testLink, Error{}, testDetail, false)

	if err.ComponentType != testComponentType || err.PluginName != testPluginName || err.LinkToDoc != testLink || err.Detail != testDetail {
		t.Fatalf("expected component type: %s, plugin name: %s, link to doc: %s, detail: %s, but got: %s, %s, %s, %s", testComponentType, testPluginName, testLink, testDetail, err.ComponentType, err.PluginName, err.LinkToDoc, err.Detail)
	}
}
