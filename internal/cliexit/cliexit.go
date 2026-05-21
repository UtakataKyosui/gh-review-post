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
	ErrCodeAuthNoBinary         ErrCode = "AUTH_GH_NOT_FOUND"
	ErrCodeAuthNoToken          ErrCode = "AUTH_NOT_LOGGED_IN"
	ErrCodeAuthOldGH            ErrCode = "AUTH_GH_VERSION_TOO_OLD"
	ErrCodeUsageBadArgs         ErrCode = "USAGE_INVALID_ARGS"
	ErrCodeUsageNoRepo          ErrCode = "USAGE_REPO_NOT_DETECTED"
	ErrCodeValidation           ErrCode = "VALIDATION_FAILED"
	ErrCodeBodyHasSuggestion    ErrCode = "VALIDATION_BODY_HAS_SUGGESTION"
	ErrCodeSuggestionFormat     ErrCode = "VALIDATION_SUGGESTION_FORMAT"
	ErrCodeSelfApprove          ErrCode = "VALIDATION_SELF_APPROVE"
	ErrCodeAPI                  ErrCode = "API_REQUEST_FAILED"
	ErrCodeGeneral              ErrCode = "INTERNAL_ERROR"
)

type Error struct {
	ExitCode int
	Code     ErrCode
	Message  string
	Details  map[string]any
	Wrapped  error
}

func (e *Error) Error() string { return e.Message }
func (e *Error) Unwrap() error { return e.Wrapped }

func errMsg(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func NewAuth(code ErrCode, err error) *Error {
	return &Error{ExitCode: CodeAuth, Code: code, Message: errMsg(err), Wrapped: err}
}

func NewUsage(code ErrCode, err error) *Error {
	return &Error{ExitCode: CodeUsage, Code: code, Message: errMsg(err), Wrapped: err}
}

func NewValidation(code ErrCode, err error, details map[string]any) *Error {
	return &Error{ExitCode: CodeValidation, Code: code, Message: errMsg(err), Details: details, Wrapped: err}
}

func NewAPI(code ErrCode, err error) *Error {
	return &Error{ExitCode: CodeAPI, Code: code, Message: errMsg(err), Wrapped: err}
}

func NewGeneral(err error) *Error {
	return &Error{ExitCode: CodeGeneral, Code: ErrCodeGeneral, Message: errMsg(err), Wrapped: err}
}

// Render writes the error to stdout (JSON mode) or stderr (plain mode).
func Render(err error, asJSON bool, stdout, stderr io.Writer) {
	if err == nil {
		return
	}
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
		b, marshalErr := json.Marshal(out)
		if marshalErr != nil {
			// fallback to plain if the details contain non-JSON-serialisable values
			fmt.Fprintf(stderr, "error: %s\n", e.Message)
			return
		}
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
