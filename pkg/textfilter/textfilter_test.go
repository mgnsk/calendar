package textfilter_test

import (
	"reflect"
	"testing"

	"github.com/mgnsk/calendar/pkg/textfilter"
)

func TestPrepareFTSSearchStrings(t *testing.T) {
	type testcase struct {
		source   string
		expected []string
	}

	for _, tc := range []testcase{
		{
			source:   `a`,
			expected: []string{`"a"`},
		},
		{
			source:   `a b`,
			expected: []string{`"a"`, `"b"`},
		},
		{
			source:   `"one" "two"`,
			expected: []string{`"one"`, `"two"`},
		},
		{
			source:   `"one one" "two two"`,
			expected: []string{`"one one"`, `"two two"`},
		},
		{
			source:   `a AND b`,
			expected: []string{`"a"`, `AND`, `"b"`},
		},
		{
			source:   `AND AND a`,
			expected: []string{`AND`, `AND`, `"a"`},
		},
		{
			source:   `a AND AND`,
			expected: []string{`"a"`, `AND`, `AND`},
		},
	} {
		t.Run(tc.source, func(t *testing.T) {
			result := textfilter.PrepareFTSSearchStrings(tc.source)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Fatalf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestGetTags(t *testing.T) {
	type testcase struct {
		source   string
		expected []string
	}

	for _, tc := range []testcase{
		{
			source:   "a", // too short
			expected: nil,
		},
		{
			source:   "abc",
			expected: []string{"abc"},
		},
		{
			source:   "abc def",
			expected: []string{"abc", "def"},
		},
		{
			source:   `"abc"`,
			expected: []string{`abc`},
		},
		{
			source:   `"abc" "def"`,
			expected: []string{`abc`, `def`},
		},
		{
			source:   `""abc" "def""`,
			expected: []string{`abc`, `def`},
		},
		{
			source:   `"'abc' "def""`,
			expected: []string{`abc`, `def`},
		},
		{
			source:   `abc,`,
			expected: []string{`abc`},
		},
		{
			source:   `aab'c,`,
			expected: []string{`aab`},
		},
		{
			source:   `aab'ccc,`,
			expected: []string{`aab`, `ccc`},
		},
		{
			source:   `(abc)`,
			expected: []string{`abc`},
		},
		{
			source:   `aaa(bbbc)`,
			expected: []string{`aaa`, `bbbc`},
		},
	} {
		t.Run(tc.source, func(t *testing.T) {
			result := textfilter.GetTags(tc.source)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Fatalf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}
