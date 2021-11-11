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
	"encoding/json"
	"fmt"
	"os"
)

// Error describes an error during plugin execution
type Error struct {
	Code    uint   `json:"code"`
	Msg     string `json:"msg"`
	Details string `json:"details,omitempty"`
}

// NewError creates new Error
func NewError(code uint, msg, details string) *Error {
	return &Error{
		Code:    code,
		Msg:     msg,
		Details: details,
	}
}

func (e *Error) Error() string {
	details := ""
	if e.Details != "" {
		details = fmt.Sprintf("; %v", e.Details)
	}
	return fmt.Sprintf("%v%v", e.Msg, details)
}

func (e *Error) Print() error {
	return prettyPrint(e)
}

func prettyPrint(obj interface{}) error {
	data, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(data)
	return err
}
