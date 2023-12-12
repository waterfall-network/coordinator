package apimiddleware

import (
	"net/http"

	"github.com/pkg/errors"
)

// ---------------
// Error handling.
// ---------------

// ErrorJSON describes common functionality of all JSON error representations.
type ErrorJSON interface {
	StatusCode() int
	SetCode(code int)
	Msg() string
	SetMsg(msg string)
}

// DefaultErrorJSON is a JSON representation of a simple error value, containing only a message and an error code.
type DefaultErrorJSON struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// InternalServerErrorWithMessage returns a DefaultErrorJSON with 500 code and a custom message.
func InternalServerErrorWithMessage(err error, message string) *DefaultErrorJSON {
	e := errors.Wrapf(err, message)
	return &DefaultErrorJSON{
		Message: e.Error(),
		Code:    http.StatusInternalServerError,
	}
}

// InternalServerError returns a DefaultErrorJSON with 500 code.
func InternalServerError(err error) *DefaultErrorJSON {
	return &DefaultErrorJSON{
		Message: err.Error(),
		Code:    http.StatusInternalServerError,
	}
}

func TimeoutError() *DefaultErrorJSON {
	return &DefaultErrorJSON{
		Message: "Request timeout",
		Code:    http.StatusRequestTimeout,
	}
}

// StatusCode returns the error's underlying error code.
func (e *DefaultErrorJSON) StatusCode() int {
	return e.Code
}

// Msg returns the error's underlying message.
func (e *DefaultErrorJSON) Msg() string {
	return e.Message
}

// SetCode sets the error's underlying error code.
func (e *DefaultErrorJSON) SetCode(code int) {
	e.Code = code
}

// SetMsg sets the error's underlying message.
func (e *DefaultErrorJSON) SetMsg(msg string) {
	e.Message = msg
}
