package wreck

import (
	"errors"
	"fmt"
	"net/http"
)

// Keys for error data in errors.
const (
	KeyHTTPCode = "http_code"
	Stack       = "stack"
)

// Base errors.
var (
	PreconditionFailed = New("precondition_failed").With(KeyHTTPCode, http.StatusPreconditionFailed)
	InvalidValue       = New("invalid_param").With(KeyHTTPCode, http.StatusBadRequest)
	AlreadyExists      = New("already_exists").With(KeyHTTPCode, http.StatusConflict)
	NotFound           = New("not_found").With(KeyHTTPCode, http.StatusNotFound)
	Timeout            = New("timeout").With(KeyHTTPCode, http.StatusRequestTimeout)
	Forbidden          = New("forbidden").With(KeyHTTPCode, http.StatusForbidden)

	Internal = New("internal").With(KeyHTTPCode, http.StatusInternalServerError)
)

// Error is an error with a safe error message.
type Error interface {
	// Error returns the internal error message.
	Error() string

	// Message returns the public error message.
	Message() string

	// With returns a clone of the error with the key-value pair arguments added.
	With(args ...any) Error

	// New creates a new error from the base error.
	New(msg string, errs ...error) Error
}

// New creates a new error.
func New(msg string) Error {
	return &wreckError{
		msg: msg,
	}
}

type wreckError struct {
	base *wreckError
	msg  string
	err  error
	args []any
}

func (e *wreckError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("%s: %s", e.msg, e.err.Error())
	}
	return e.msg
}

func (e *wreckError) Message() string {
	return e.msg
}

func (e *wreckError) Unwrap() error {
	return e.err
}

func (e *wreckError) Is(target error) bool {
	if base, ok := target.(*wreckError); ok {
		b := e.base
		for b != nil {
			if b == base {
				return true
			}
			b = b.base
		}
	}
	return false
}

func (e *wreckError) With(args ...any) Error {
	return &wreckError{
		base: e,
		msg:  e.msg,
		err:  e.err,
		args: args,
	}
}

func (e *wreckError) New(msg string, errs ...error) Error {
	return &wreckError{
		base: e,
		msg:  msg,
		err:  errors.Join(errs...),
	}
}
