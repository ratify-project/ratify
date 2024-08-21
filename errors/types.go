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
	originalError error
	code          ErrorCode
	message       string
	detail        interface{}
	componentType ComponentType
	remediation   string
	pluginName    string
	stack         string
	description   string
	isRootError   bool // isRootError is true if the originalError is either nil or not an Error type
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

// Description returned the description of this error code.
func (ec ErrorCode) Description() string {
	return ec.Descriptor().Description
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

// WithRemediation returns a new Error object with remediation.
func (ec ErrorCode) WithRemediation(link string) Error {
	return newError(ec, ec.Message()).WithRemediation(link)
}

func (ec ErrorCode) WithDescription() Error {
	return newError(ec, ec.Message()).WithDescription()
}

// WithPluginName returns a new Error object with pluginName set.
func (ec ErrorCode) WithPluginName(pluginName string) Error {
	return newError(ec, ec.Message()).WithPluginName(pluginName)
}

// NewError returns a new Error object.
func (ec ErrorCode) NewError(componentType ComponentType, pluginName, remediation string, err error, detail interface{}, printStackTrace bool) Error {
	stack := ""
	if printStackTrace {
		stack = getStackTrace()
	}
	return Error{
		code:          ec,
		message:       ec.Message(),
		originalError: err,
		detail:        detail,
		componentType: componentType,
		pluginName:    pluginName,
		remediation:   remediation,
		stack:         stack,
		isRootError:   err == nil || !errors.As(err, &Error{}),
	}
}

func newError(code ErrorCode, message string) Error {
	return Error{
		code:        code,
		message:     message,
		isRootError: true,
	}
}

// Is returns true if the error is the same type of the target error.
func (e Error) Is(target error) bool {
	t := &Error{}
	if errors.As(target, t) {
		return e.code.ErrorCode() == t.code.ErrorCode()
	}
	return false
}

// ErrorCode returns the ID/Value of this Error
func (e Error) ErrorCode() ErrorCode {
	return e.code
}

// Unwrap returns the original error
func (e Error) Unwrap() error {
	return e.originalError
}

// Error returns a human readable representation of the error.
// An Error message includes the error code, detail from nested errors, root cause and remediation, all separated by ": ".
func (e Error) Error() string {
	err, details := e.getRootError()
	if err.detail != nil {
		details = append(details, fmt.Sprintf("%s", err.detail))
	}
	if err.originalError != nil {
		details = append(details, err.originalError.Error())
	}

	if err.remediation != "" {
		details = append(details, err.remediation)
	}
	return fmt.Sprintf("%s: %s", err.ErrorCode().Descriptor().Value, strings.Join(details, ": "))
}

// GetDetail returns details from all nested errors.
func (e Error) GetDetail() string {
	err, details := e.getRootError()
	if err.originalError != nil && err.detail != nil {
		details = append(details, fmt.Sprintf("%s", err.detail))
	}

	return strings.Join(details, ": ")
}

// GetErrorReason returns the root cause of the error.
func (e Error) GetErrorReason() string {
	err, _ := e.getRootError()
	if err.originalError != nil {
		return err.originalError.Error()
	}
	return fmt.Sprintf("%s", err.detail)
}

// GetRemiation returns the remediation of the root error.
func (e Error) GetRemediation() string {
	err, _ := e.getRootError()
	return err.remediation
}

// GetConciseError returns a formatted error message consisting of the error code and reason.
// If the generated error message exceeds the specified maxLength, it truncates the message and appends an ellipsis ("...").
// The function ensures that the returned error message is concise and within the length limit.
func (e Error) GetConciseError(maxLength int) string {
	err, _ := e.getRootError()
	errMsg := fmt.Sprintf("%s: %s", err.ErrorCode().Descriptor().Value, e.GetErrorReason())
	if len(errMsg) > maxLength {
		return fmt.Sprintf("%s...", errMsg[:maxLength-3])
	}
	return errMsg
}

func (e Error) getRootError() (err Error, details []string) {
	err = e
	for !err.isRootError {
		if err.detail != nil {
			details = append(details, fmt.Sprintf("%s", err.detail))
		}
		var ratifyError Error
		if errors.As(err.originalError, &ratifyError) {
			err = ratifyError
		} else {
			// break is unnecessary, but added for safety
			break
		}
	}
	return err, details
}

// WithDetail will return a new Error, based on the current one, but with
// some Detail info added
func (e Error) WithDetail(detail interface{}) Error {
	e.detail = detail
	return e
}

// WithPluginName returns a new Error object with pluginName set.
func (e Error) WithPluginName(pluginName string) Error {
	e.pluginName = pluginName
	return e
}

// WithComponentType returns a new Error object with ComponentType set.
func (e Error) WithComponentType(componentType ComponentType) Error {
	e.componentType = componentType
	return e
}

// WithError returns a new Error object with original error.
func (e Error) WithError(err error) Error {
	e.originalError = err
	e.isRootError = err == nil || !errors.As(err, &Error{})
	return e
}

// WithRemediation returns a new Error object attached with remediation.
func (e Error) WithRemediation(remediation string) Error {
	e.remediation = remediation
	return e
}

func (e Error) WithDescription() Error {
	e.description = e.code.Description()
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
