package util

import (
	"testing"
)

var shiftPathTests = []struct {
	path, head, tail string
}{
	{"hello/world", "hello", "/world"},
	{"hello", "hello", "/"},
	{"/hello/world", "hello", "/world"},
	{"hello/", "hello", "/"},
	{"", "", "/"},
}

func TestShiftPath(t *testing.T) {
	for _, test := range shiftPathTests {
		head, tail := ShiftPath(test.path)
		if head != test.head || tail != test.tail {
			t.Errorf("test %+v: got head=%v, tail=%v", test, head, tail)
		}
	}
}
