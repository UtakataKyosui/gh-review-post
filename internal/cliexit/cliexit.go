package cliexit

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

const (
	CodeSuccess    = 0
	CodeGeneral    = 1
	CodeUsage      = 2
	CodeAuth       = 3
	CodeValidation = 4
	CodeAPI        = 5
)

type ErrCode string

const (
	ErrCodeAuthNoBinary  ErrCode = "AUTH_GH_NOT_FOUND"
	ErrCodeAuthNoToken   ErrCode = "AUTH_NOT_LOGGED_IN"
	ErrCodeAuthOldGH     ErrCode = "AUTH_GH_VERSION_TOO_OLD"
	ErrCodeUsageBadArgs  ErrCode = "USAGE_INVALID_ARGS"
	ErrCodeUsageNoRepo   ErrCode = "USAGE_REPO_NOT_DETECTED"
	ErrCodeValidation    ErrCode = "VALIDATION_FAILED"
	ErrCodeAPI           ErrCode = "API_REQUEST_FAILED"
	ErrCodeGeneral       ErrCode = "INTERNAL_ERROR"
)

type Error struct {
	ExitCode int
	Code     ErrCode
	Message  string
	Details  map[string]any
	Wrapped  error
}

func (e *Error) Error() string  { return e.Message }
func (e *Error) Unwrap() error  { return e.Wrapped }

func NewAuth(code ErrCode, err error) *Error {
	msg := err.Error()
	if err != nil {
		msg = err.Error()
	}
	return &Error{ExitCode: CodeAuth, Code: code, Message: msg, Wrapped: err}
}

func NewUsage(code ErrCode, err error) *Error {
	return &Error{ExitCode: CodeUsage, Code: code, Message: err.Error(), Wrapped: err}
}

func NewValidation(code ErrCode, err error, details map[string]any) *Error {
	return &Error{ExitCode: CodeValidation, Code: code, Message: err.Error(), Details: details, Wrapped: err}
}

func NewAPI(code ErrCode, err error) *Error {
	return &Error{ExitCode: CodeAPI, Code: code, Message: err.Error(), Wrapped: err}
}

func NewGeneral(err error) *Error {
	return &Error{ExitCode: CodeGeneral, Code: ErrCodeGeneral, Message: err.Error(), Wrapped: err}
}

// Render writes the error to stdout (JSON mode) or stderr (plain mode).
func Render(err error, asJSON bool, stdout, stderr io.Writer) {
	var e *Error
	if !errors.As(err, &e) {
		e = NewGeneral(err)
	}
	if asJSON {
		out := struct {
			Error struct {
				Code    ErrCode        `json:"code"`
				Message string         `json:"message"`
				Details map[string]any `json:"details,omitempty"`
			} `json:"error"`
		}{}
		out.Error.Code = e.Code
		out.Error.Message = e.Message
		out.Error.Details = e.Details
		b, _ := json.Marshal(out)
		fmt.Fprintf(stdout, "%s\n", b)
		return
	}
	fmt.Fprintf(stderr, "error: %s\n", e.Message)
}

// ExitCodeOf returns the exit code for the given error.
func ExitCodeOf(err error) int {
	if err == nil {
		return CodeSuccess
	}
	var e *Error
	if errors.As(err, &e) {
		return e.ExitCode
	}
	return CodeGeneral
}
