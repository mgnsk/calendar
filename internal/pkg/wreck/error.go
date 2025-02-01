package wreck

import "fmt"

// PreconditionFailed is a precondition failed error.
type PreconditionFailed struct {
	Err error
}

func (e *PreconditionFailed) Unwrap() error {
	return e.Err
}

func (e *PreconditionFailed) Error() string {
	return fmt.Sprintf("precondition failed: %s", e.Err.Error())
}

// InvalidInput is an invalid input error.
type InvalidInput struct {
	Err error
}

func (e *InvalidInput) Unwrap() error {
	return e.Err
}

func (e *InvalidInput) Error() string {
	return fmt.Sprintf("invalid input: %s", e.Err.Error())
}

// AlreadyExists is an invalid input error.
type AlreadyExists struct {
	Err error
}

func (e *AlreadyExists) Unwrap() error {
	return e.Err
}

func (e *AlreadyExists) Error() string {
	return fmt.Sprintf("already exists: %s", e.Err.Error())
}

// NotFound is a not found error.
type NotFound struct {
	Err error
}

func (e *NotFound) Unwrap() error {
	return e.Err
}

func (e *NotFound) Error() string {
	return fmt.Sprintf("not found: %s", e.Err.Error())
}
