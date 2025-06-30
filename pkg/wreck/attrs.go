package wreck

import (
	"errors"
	"log/slog"
)

// Value extracts a single value from key-value pair attributes.
func Value(err error, key string) (slog.Value, bool) {
	rawArgs := collectRawArgs(err)
	attrs := slog.Group("", rawArgs...).Value.Group()

	for _, a := range attrs {
		if a.Key == key {
			return a.Value, true
		}
	}

	return slog.Value{}, false
}

// Attrs extracts arguments as key-value pair attributes.
func Attrs(err error) []slog.Attr {
	rawArgs := collectRawArgs(err)
	return slog.Group("", rawArgs...).Value.Group()
}

// Args extracts arguments as key-value pair arguments.
func Args(err error) (args []any) {
	rawArgs := collectRawArgs(err)
	attrs := slog.Group("", rawArgs...).Value.Group()

	for _, a := range attrs {
		args = append(args, a.Key, a.Value.Any())
	}

	return args
}

func collectRawArgs(err error) (args []any) {
	var werr *wreckError
	if errors.As(err, &werr) {
		for werr != nil {
			args = append(args, werr.args...)
			werr = werr.base
		}
	}
	return args
}
