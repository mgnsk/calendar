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
	if e.Err != nil {
		return fmt.Sprintf("precondition failed: %s", e.Err.Error())
	}
	return "precondition failed"
}

// InvalidInput is an invalid input error.
type InvalidInput struct {
	Err error
}

func (e *InvalidInput) Unwrap() error {
	return e.Err
}

func (e *InvalidInput) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("invalid input: %s", e.Err.Error())
	}
	return "invalid input"
}

// AlreadyExists is an invalid input error.
type AlreadyExists struct {
	Err error
}

func (e *AlreadyExists) Unwrap() error {
	return e.Err
}

func (e *AlreadyExists) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("already exists: %s", e.Err.Error())
	}
	return "already exists"
}

// NotFound is a not found error.
type NotFound struct {
	Err error
}

func (e *NotFound) Unwrap() error {
	return e.Err
}

func (e *NotFound) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("not found: %s", e.Err.Error())
	}
	return "not found"
}
