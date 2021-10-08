package cmd

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
)

func TestAreSlicesEqualNoOrder(t *testing.T) {
	testCases := []struct {
		a     []string
		b     []string
		equal bool
	}{
		{[]string{}, []string{}, true},
		{[]string{"a"}, []string{"a"}, true},
		{[]string{"a", "b"}, []string{"a", "b"}, true},
		{[]string{"a", "b"}, []string{"a", "b", "c"}, false},
		{[]string{"b", "a", "c"}, []string{"a", "b", "c"}, true},
		{[]string{"a", "a", "b"}, []string{"a", "b", "a"}, true},
		{[]string{"a", "b"}, []string{"c", "d"}, false},
		{[]string{"a", "b"}, []string{"a", "c"}, false},
	}
	for _, tc := range testCases {
		msg := fmt.Sprintf("a: %v, b: %v", tc.a, tc.b)
		assert.Equal(t, tc.equal, areSlicesEqualNoOrder(tc.a, tc.b), msg)
	}
}
