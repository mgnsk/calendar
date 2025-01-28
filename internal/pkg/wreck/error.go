package wreck

import "fmt"

// PreconditionFailed is a precondition failed error.
type PreconditionFailed struct {
	Cause error
}

func (e *PreconditionFailed) Unwrap() error {
	return e.Cause
}

func (e *PreconditionFailed) Error() string {
	return fmt.Sprintf("precondition failed: %s", e.Cause.Error())
}
