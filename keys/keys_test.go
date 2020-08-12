package keys

import (
	"testing"

	"github.com/google/uuid"
)

func TestParseKey(t *testing.T) {
	type testCase struct {
		str string
		obj Key
	}

	var testCases = []testCase{
		// positive create
		{str: "create:userone:9c75a1da-2fad-4522-87aa-e5a0b1b35c7e",
			obj: CreateTimer{domain: "userone", timerUUID: uuid.UUID{156, 117, 161, 218, 47, 173, 69, 34, 135, 170, 229, 160, 177, 179, 92, 126}}},

		// positive execute
		{str: "execute:userone:9c75a1da-2fad-4522-87aa-e5a0b1b35c7e",
			obj: ExecuteTimer{domain: "userone", timerUUID: uuid.UUID{156, 117, 161, 218, 47, 173, 69, 34, 135, 170, 229, 160, 177, 179, 92, 126}}},

		// positive dummy
		{str: "dummy", obj: Dummy{}},

		// wrong number of fields
		{str: "create:userone:9c75a1da-2fad-4522-87aa-e5a0b1b35c7e:extra", obj: nil},

		// bad uuid (removed 'e' at end)
		{str: "create:userone:9c75a1da-2fad-4522-87aa-e5a0b1b35c7", obj: nil},
	}

	for idx, test := range testCases {
		key, err := ParseKey([]byte(test.str))

		if err != nil && test.obj != nil {
			t.Errorf("expected no error at idx %v, got err: %v\n", idx, err)
			continue
		}

		if err == nil && key != test.obj {
			t.Errorf("expected %v at idx %v, got: %v\n", test.obj, idx, key)
		}
	}
}
