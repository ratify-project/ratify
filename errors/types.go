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
	"fmt"
	"runtime"
	"strings"
	"sync"
)

const (
	EmptyLink       = ""
	PrintStackTrace = true
	HideStackTrace  = false
)

var (
	nextCode     = 1000
	registerLock sync.Mutex
)

var (
	errorCodeToDescriptors = map[ErrorCode]ErrorDescriptor{}
	idToDescriptors        = map[string]ErrorDescriptor{}
	groupToDescriptors     = map[string][]ErrorDescriptor{}
)

type ComponentType string

const (
	Verifier              ComponentType = "verifier"
	ReferrerStore         ComponentType = "referrerStore"
	Policy                ComponentType = "policy"
	Executor              ComponentType = "executor"
	Cache                 ComponentType = "cache"
	AuthProvider          ComponentType = "authProvider"
	PolicyProvider        ComponentType = "policyProvider"
	CertProvider          ComponentType = "certProvider"
	KeyManagementProvider ComponentType = "keyManagementProvider"
)

// ErrorCode represents the error type. The errors are serialized via strings
// and the integer format may change and should *never* be exported.
type ErrorCode int

// Error provides a wrapper around ErrorCode with extra Details provided.
type Error struct {
	OriginalError error         `json:"originalError,omitempty"`
	Code          ErrorCode     `json:"code"`
	Message       string        `json:"message"`
	Detail        interface{}   `json:"detail,omitempty"`
	ComponentType ComponentType `json:"componentType,omitempty"`
	PluginName    string        `json:"pluginName,omitempty"`
	LinkToDoc     string        `json:"linkToDoc,omitempty"`
	Stack         string        `json:"stack,omitempty"`
}

// ErrorDescriptor provides relevant information about a given error code.
type ErrorDescriptor struct {
	// Code is the error code that this descriptor describes.
	Code ErrorCode

	// Value provides a unique, string key, often captilized with
	// underscores, to identify the error code. This value is used as the
	// keyed value when serializing api errors.
	Value string

	// Message is a short, human readable description of the error condition
	// included in API responses.
	Message string

	// Description provides a complete account of the errors purpose, suitable
	// for use in documentation.
	Description string

	// ComponentType specifies the type of the component that the error is from.
	ComponentType ComponentType

	// PluginName specifies the name of the plugin that the error is from.
	PluginName string
}

var emptyError = Error{}

// ErrorCode returns itself
func (ec ErrorCode) ErrorCode() ErrorCode {
	return ec
}

// Error returns the ID/Value
func (ec ErrorCode) Error() string {
	// NOTE(stevvooe): Cannot use message here since it may have unpopulated args.
	return strings.ToLower(strings.Replace(ec.String(), "_", " ", -1))
}

// Descriptor returns the descriptor for the error code.
func (ec ErrorCode) Descriptor() ErrorDescriptor {
	d, ok := errorCodeToDescriptors[ec]

	if !ok {
		return ErrorCodeUnknown.Descriptor()
	}

	return d
}

// Message returned the human-readable error message for this error code.
func (ec ErrorCode) Message() string {
	return ec.Descriptor().Message
}

// String returns the canonical identifier for this error code.
func (ec ErrorCode) String() string {
	return ec.Descriptor().Value
}

// WithDetail returns a new Error object with details about the error.
func (ec ErrorCode) WithDetail(detail interface{}) Error {
	return newError(ec, ec.Message()).WithDetail(detail)
}

// WithError returns a new Error object with original error.
func (ec ErrorCode) WithError(err error) Error {
	return newError(ec, ec.Message()).WithError(err)
}

// WithComponentType returns a new Error object with ComponentType set.
func (ec ErrorCode) WithComponentType(componentType ComponentType) Error {
	return newError(ec, ec.Message()).WithComponentType(componentType)
}

// WithLinkToDoc returns a new Error object with attached link to the documentation.
func (ec ErrorCode) WithLinkToDoc(link string) Error {
	return newError(ec, ec.Message()).WithLinkToDoc(link)
}

// WithPluginName returns a new Error object with pluginName set.
func (ec ErrorCode) WithPluginName(pluginName string) Error {
	return newError(ec, ec.Message()).WithPluginName(pluginName)
}

// NewError returns a new Error object.
func (ec ErrorCode) NewError(componentType ComponentType, pluginName, link string, err error, detail interface{}, printStackTrace bool) Error {
	stack := ""
	if printStackTrace {
		stack = getStackTrace()
	}
	return Error{
		Code:          ec,
		Message:       ec.Message(),
		OriginalError: err,
		Detail:        detail,
		ComponentType: componentType,
		PluginName:    pluginName,
		LinkToDoc:     link,
		Stack:         stack,
	}
}

func newError(code ErrorCode, message string) Error {
	return Error{
		Code:    code,
		Message: message,
	}
}

// Is returns true if the error is the same type of the target error.
func (e Error) Is(target error) bool {
	t := &Error{}
	if errors.As(target, t) {
		return e.Code.ErrorCode() == t.Code.ErrorCode()
	}
	return false
}

// ErrorCode returns the ID/Value of this Error
func (e Error) ErrorCode() ErrorCode {
	return e.Code
}

// Unwrap returns the original error
func (e Error) Unwrap() error {
	return e.OriginalError
}

// Error returns a human readable representation of the error.
func (e Error) Error() string {
	var errStr string
	if e.OriginalError != nil {
		errStr += fmt.Sprintf("Original Error: (%s), ", e.OriginalError.Error())
	}

	errStr += fmt.Sprintf("Error: %s, Code: %s", e.Message, e.Code.String())

	if e.PluginName != "" {
		errStr += fmt.Sprintf(", Plugin Name: %s", e.PluginName)
	}

	if e.ComponentType != "" {
		errStr += fmt.Sprintf(", Component Type: %s", e.ComponentType)
	}

	if e.LinkToDoc != "" {
		errStr += fmt.Sprintf(", Documentation: %s", e.LinkToDoc)
	}

	if e.Detail != nil {
		errStr += fmt.Sprintf(", Detail: %v", e.Detail)
	}

	if e.Stack != "" {
		errStr += fmt.Sprintf(", Stack trace: %s", e.Stack)
	}

	return errStr
}

// WithDetail will return a new Error, based on the current one, but with
// some Detail info added
func (e Error) WithDetail(detail interface{}) Error {
	e.Detail = detail
	return e
}

// WithPluginName returns a new Error object with pluginName set.
func (e Error) WithPluginName(pluginName string) Error {
	e.PluginName = pluginName
	return e
}

// WithComponentType returns a new Error object with ComponentType set.
func (e Error) WithComponentType(componentType ComponentType) Error {
	e.ComponentType = componentType
	return e
}

// WithError returns a new Error object with original error.
func (e Error) WithError(err error) Error {
	e.OriginalError = err
	return e
}

// WithLinkToDoc returns a new Error object attached with link to documentation.
func (e Error) WithLinkToDoc(link string) Error {
	e.LinkToDoc = link
	return e
}

// IsEmpty returns true if the error is empty.
func (e Error) IsEmpty() bool {
	return e == emptyError
}

func getStackTrace() string {
	stack := make([]byte, 4096)
	n := runtime.Stack(stack, false)
	return string(stack[:n])
}

// Register will make the passed-in error known to the environment and
// return a new ErrorCode
func Register(group string, descriptor ErrorDescriptor) ErrorCode {
	registerLock.Lock()
	defer registerLock.Unlock()

	descriptor.Code = ErrorCode(nextCode)

	if _, ok := idToDescriptors[descriptor.Value]; ok {
		panic(fmt.Sprintf("ErrorValue %q is already registered", descriptor.Value))
	}
	if _, ok := errorCodeToDescriptors[descriptor.Code]; ok {
		panic(fmt.Sprintf("ErrorCode %v is already registered", descriptor.Code))
	}

	groupToDescriptors[group] = append(groupToDescriptors[group], descriptor)
	errorCodeToDescriptors[descriptor.Code] = descriptor
	idToDescriptors[descriptor.Value] = descriptor

	nextCode++
	return descriptor.Code
}
