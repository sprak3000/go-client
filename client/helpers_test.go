package client

import "testing"
import "reflect"
import "io/ioutil"

func TestUnit_ObjectToJSONReader(t *testing.T) {
	type testcase struct {
		name        string
		input       interface{}
		expected    []byte
		expectedErr error
	}

	type teststruct struct {
		Foo string `json:"foo"`
		Baz int    `json:"baz"`
	}
	type teststruct2 struct {
		A string     `json:"a"`
		B int        `json:"b"`
		C teststruct `json:"c"`
	}

	testcases := []testcase{
		{
			name:     "bytes",
			input:    []byte(`{"foo":"bar","baz":123}`),
			expected: []byte(`{"foo":"bar","baz":123}`),
		},
		{
			name:     "struct",
			input:    teststruct2{A: "bar", B: 123, C: teststruct{Foo: "bar2", Baz: 456}},
			expected: []byte(`{"a":"bar","b":123,"c":{"foo":"bar2","baz":456}}`),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ret, err := ObjectToJSONReader(tc.input)
			if !reflect.DeepEqual(err, tc.expectedErr) {
				t.Fatalf("Error actual (%v) did not match expected (%v)", err, tc.expectedErr)
			}

			by, _ := ioutil.ReadAll(ret)
			if !reflect.DeepEqual(by, tc.expected) {
				t.Fatalf("Actual (%s) did not match expected (%s)", by, tc.expected)
			}
		})
	}

}
