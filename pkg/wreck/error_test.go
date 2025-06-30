package wreck_test

import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"testing"

	"github.com/mgnsk/calendar/pkg/wreck"
)

func TestErrors(t *testing.T) {
	t.Run("creating new errors", func(t *testing.T) {
		base := wreck.New("base")
		err := base.New("new error")
		assert(t, errors.Is(err, base), true)
		assert(t, err.Error(), "new error")
	})

	t.Run("wrapping existing error", func(t *testing.T) {
		base := wreck.New("base")
		one := fmt.Errorf("one")
		err := base.New("new error", one)
		assert(t, errors.Is(err, base), true)
		assert(t, errors.Is(err, one), true)
		assert(t, err.Error(), "new error: one")
	})

	t.Run("wrapping multiple existing errors", func(t *testing.T) {
		base := wreck.New("base")
		one := fmt.Errorf("one")
		two := fmt.Errorf("two")
		err := base.New("new error", one, two)
		assert(t, errors.Is(err, base), true)
		assert(t, errors.Is(err, one), true)
		assert(t, errors.Is(err, two), true)
		assert(t, err.Error(), "new error: one\ntwo")
	})

	t.Run("wrapping multiple times", func(t *testing.T) {
		inner := wreck.New("inner")
		outer := wreck.New("outer")

		err1 := inner.New("one")
		err2 := outer.New("two", err1)

		assert(t, errors.Is(err1, inner), true)
		assert(t, errors.Is(err2, outer), true)
		assert(t, errors.Is(err2, inner), true)
		assert(t, err2.Error(), "two: one")
	})

	t.Run("safe error message", func(t *testing.T) {
		base := wreck.New("base")
		err := base.New("Message", fmt.Errorf("internal message"))

		assert(t, err.Error(), "Message: internal message")
		assert(t, err.Message(), "Message")
	})

	t.Run("storing arguments in base error", func(t *testing.T) {
		t.Run("values are stored on new base error", func(t *testing.T) {
			origBase := wreck.New("base")
			newBase := origBase.With("key", "value")
			err := newBase.New("Message")

			args := wreck.Args(err)
			assert(t, args, []any{"key", "value"})
		})

		t.Run("original base error is not modified", func(t *testing.T) {
			origBase := wreck.New("base")
			_ = origBase.With("key", "value")
			err := origBase.New("Message")

			args := wreck.Args(err)
			assert(t, len(args), 0)
		})

		t.Run("error matches original base error", func(t *testing.T) {
			origBase := wreck.New("base")
			newBase := origBase.With("key", "value")
			err := newBase.New("Message")

			args := wreck.Args(err)
			assert(t, args, []any{"key", "value"})

			assert(t, errors.Is(err, newBase), true)
			assert(t, errors.Is(err, origBase), true)
		})

		t.Run("all arguments can be collected", func(t *testing.T) {
			origBase := wreck.New("base").With(
				"a", "value1",
				"b", "value2",
			)
			newBase := origBase.With("c", "value3")
			err := newBase.New("Message")

			args := wreck.Args(err)
			assert(t, args, []any{
				"c", "value3",
				"a", "value1",
				"b", "value2",
			})
		})

		t.Run("all arguments can be interpreted as key-value pair attributes", func(t *testing.T) {
			origBase := wreck.New("base").With(
				"a", "value1",
				"b", "value2",
			)
			newBase := origBase.With("c", "value3")
			err := newBase.New("Message")

			attrs := wreck.Attrs(err)
			assert(t, attrs, []slog.Attr{
				slog.String("c", "value3"),
				slog.String("a", "value1"),
				slog.String("b", "value2"),
			})
		})

		t.Run("single attribute can be collected", func(t *testing.T) {
			origBase := wreck.New("base").With(
				"a", "value1",
				"b", "value2",
			)
			newBase := origBase.With("c", "value3")
			err := newBase.New("Message")

			value, ok := wreck.Value(err, "a")
			assert(t, ok, true)
			assert(t, value.String(), "value1")
		})
	})
}

func assert[T any](t testing.TB, a, b T) {
	t.Helper()

	if !reflect.DeepEqual(a, b) {
		t.Fatalf("expected '%v' to equal '%v'", a, b)
	}
}
