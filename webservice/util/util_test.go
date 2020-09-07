package util

import (
	"encoding/base64"
	"net/http"
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

// below copied from "net/http"
var parseBasicAuthTests = []struct {
	header, username, password string
	ok                         bool
}{
	{"Basic " + base64.StdEncoding.EncodeToString([]byte("Aladdin:open sesame")), "Aladdin", "open sesame", true},

	// Case doesn't matter:
	{"BASIC " + base64.StdEncoding.EncodeToString([]byte("Aladdin:open sesame")), "Aladdin", "open sesame", true},
	{"basic " + base64.StdEncoding.EncodeToString([]byte("Aladdin:open sesame")), "Aladdin", "open sesame", true},

	{"Basic " + base64.StdEncoding.EncodeToString([]byte("Aladdin:open:sesame")), "Aladdin", "open:sesame", true},
	{"Basic " + base64.StdEncoding.EncodeToString([]byte(":")), "", "", true},
	{"Basic" + base64.StdEncoding.EncodeToString([]byte("Aladdin:open sesame")), "", "", false},
	{base64.StdEncoding.EncodeToString([]byte("Aladdin:open sesame")), "", "", false},
	{"Basic ", "", "", false},
	{"Basic Aladdin:open sesame", "", "", false},
	{`Digest username="Aladdin"`, "", "", false},
}

func TestParseBasicAuth(t *testing.T) {
	for _, tt := range parseBasicAuthTests {
		r, _ := http.NewRequest("GET", "http://example.com/", nil)
		r.Header.Set("Authorization", tt.header)
		username, password, ok := r.BasicAuth()
		if ok != tt.ok || username != tt.username || password != tt.password {
			t.FailNow()
		}
	}
}
