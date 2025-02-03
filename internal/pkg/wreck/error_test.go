package wreck_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/mgnsk/calendar/internal/pkg/wreck"
)

func TestErrors(t *testing.T) {
	t.Run("creating new errors", func(t *testing.T) {
		base := wreck.NewBaseError("base")
		err := base.New("new error")
		assertTrue(t, errors.Is(err, base))
		assertTrue(t, err.Error() == "new error")
	})

	t.Run("wrapping existing error", func(t *testing.T) {
		base := wreck.NewBaseError("base")
		one := fmt.Errorf("one")
		err := base.New("new error", one)
		assertTrue(t, errors.Is(err, base))
		assertTrue(t, errors.Is(err, one))
		assertTrue(t, err.Error() == "new error: one")
	})

	t.Run("wrapping multiple existing errors", func(t *testing.T) {
		base := wreck.NewBaseError("base")
		one := fmt.Errorf("one")
		two := fmt.Errorf("two")
		err := base.New("new error", one, two)
		assertTrue(t, errors.Is(err, base))
		assertTrue(t, errors.Is(err, one))
		assertTrue(t, errors.Is(err, two))
		assertTrue(t, err.Error() == "new error: one\ntwo")
	})

	t.Run("wrapping multiple times", func(t *testing.T) {
		inner := wreck.NewBaseError("inner")
		outer := wreck.NewBaseError("outer")

		err1 := inner.New("one")
		err2 := outer.New("two", err1)

		assertTrue(t, errors.Is(err1, inner))
		assertTrue(t, errors.Is(err2, outer))
		assertTrue(t, errors.Is(err2, inner))
		assertTrue(t, err2.Error() == "two: one")
	})

	t.Run("safe error message", func(t *testing.T) {
		base := wreck.NewBaseError("base")
		err := base.New("Message", fmt.Errorf("internal message"))

		assertTrue(t, err.Error() == "Message: internal message")
		assertTrue(t, err.Message() == "Message")
	})

	t.Run("storing values in base error", func(t *testing.T) {
		base := wreck.NewBaseError("base").With("key", "value")
		err := base.New("Message")

		value := wreck.Value(err, "key")
		assertTrue(t, value == "value")
	})
}

func assertTrue(t testing.TB, result bool, msg ...any) {
	if result {
		return
	}

	if len(msg) > 0 {
		t.Fatal(msg...)
	} else {
		t.Fatal("expected true")
	}
}

func assertFalse(t testing.TB, result bool, msg ...any) {
	if !result {
		return
	}

	if len(msg) > 0 {
		t.Fatal(msg...)
	} else {
		t.Fatal("expected true")
	}
}
