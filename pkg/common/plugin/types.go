package plugin

import (
	"encoding/json"
	"fmt"
	"os"
)

type Error struct {
	Code    uint   `json:"code"`
	Msg     string `json:"msg"`
	Details string `json:"details,omitempty"`
}

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
