package wreck_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/mgnsk/calendar/pkg/wreck"
)

func TestErrors(t *testing.T) {
	t.Run("creating new errors", func(t *testing.T) {
		base := wreck.NewBaseError("base")
		err := base.New("new error")
		assert(t, errors.Is(err, base), true)
		assert(t, err.Error(), "new error")
	})

	t.Run("wrapping existing error", func(t *testing.T) {
		base := wreck.NewBaseError("base")
		one := fmt.Errorf("one")
		err := base.New("new error", one)
		assert(t, errors.Is(err, base), true)
		assert(t, errors.Is(err, one), true)
		assert(t, err.Error(), "new error: one")
	})

	t.Run("wrapping multiple existing errors", func(t *testing.T) {
		base := wreck.NewBaseError("base")
		one := fmt.Errorf("one")
		two := fmt.Errorf("two")
		err := base.New("new error", one, two)
		assert(t, errors.Is(err, base), true)
		assert(t, errors.Is(err, one), true)
		assert(t, errors.Is(err, two), true)
		assert(t, err.Error(), "new error: one\ntwo")
	})

	t.Run("wrapping multiple times", func(t *testing.T) {
		inner := wreck.NewBaseError("inner")
		outer := wreck.NewBaseError("outer")

		err1 := inner.New("one")
		err2 := outer.New("two", err1)

		assert(t, errors.Is(err1, inner), true)
		assert(t, errors.Is(err2, outer), true)
		assert(t, errors.Is(err2, inner), true)
		assert(t, err2.Error(), "two: one")
	})

	t.Run("safe error message", func(t *testing.T) {
		base := wreck.NewBaseError("base")
		err := base.New("Message", fmt.Errorf("internal message"))

		assert(t, err.Error(), "Message: internal message")
		assert(t, err.Message(), "Message")
	})

	t.Run("storing values in base error", func(t *testing.T) {
		base := wreck.NewBaseError("base").With("key", "value")
		err := base.New("Message")

		value := wreck.Value(err, "key")
		assert(t, value, "value")
	})
}

func assert[T any](t testing.TB, a, b T) {
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("expected '%v' to equal '%v'", a, b)
	}
}
