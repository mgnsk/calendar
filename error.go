package calendar

import (
	"github.com/mgnsk/webkit"
)

// Keys for error data in errors.
const (
	KeyHTTPCode = webkit.KeyHTTPCode
)

// Base errors, shared with webkit so that its middleware maps them to
// HTTP status codes.
var (
	PreconditionFailed = webkit.PreconditionFailed
	InvalidValue       = webkit.InvalidValue
	AlreadyExists      = webkit.AlreadyExists
	NotFound           = webkit.NotFound
	Timeout            = webkit.Timeout
	Forbidden          = webkit.Forbidden

	Internal = webkit.Internal
)
