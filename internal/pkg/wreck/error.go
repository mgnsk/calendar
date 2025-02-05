package wreck

import (
	"errors"
	"fmt"
	"net/http"
)

// KeyHTTPCode is a HTTP code error key.
const KeyHTTPCode = "http_code"

// Base errors.
var (
	PreconditionFailed = NewBaseError("precondition_failed").With(KeyHTTPCode, http.StatusPreconditionFailed)
	InvalidValue       = NewBaseError("invalid_param").With(KeyHTTPCode, http.StatusBadRequest)
	AlreadyExists      = NewBaseError("already_exists").With(KeyHTTPCode, http.StatusConflict)
	NotFound           = NewBaseError("not_found").With(KeyHTTPCode, http.StatusNotFound)

	Internal = NewBaseError("internal").With(KeyHTTPCode, http.StatusInternalServerError)
)

// Value extracts a value from err's base error.
func Value(err error, key string) any {
	var werr *wreckError
	if errors.As(err, &werr) {
		return werr.base.values[key]
	}
	return nil
}

// BaseError is a base error.
type BaseError interface {
	error
	With(key string, value any) BaseError
	New(string, ...error) Error
}

// Error is an error with a safe error message.
type Error interface {
	error
	Message() string
}

// NewBaseError creates a new base error.
func NewBaseError(code string) BaseError {
	return &baseError{
		code:   code,
		values: map[string]any{},
	}
}

type baseError struct {
	code   string
	values map[string]any
}

func (e *baseError) Error() string {
	return e.code
}

func (e *baseError) With(key string, value any) BaseError {
	e.values[key] = value
	return e
}

func (e *baseError) New(msg string, errs ...error) Error {
	return &wreckError{
		base: e,
		msg:  msg,
		err:  errors.Join(errs...),
	}
}

type wreckError struct {
	base *baseError
	msg  string
	err  error
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
	if base, ok := target.(*baseError); ok {
		return e.base == base
	}
	return false
}
