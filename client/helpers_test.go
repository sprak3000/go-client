package client

import (
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnit_ObjectToJSONReader(t *testing.T) {
	type teststruct struct {
		Foo string `json:"foo"`
		Baz int    `json:"baz"`
	}
	type teststruct2 struct {
		A string     `json:"a"`
		B int        `json:"b"`
		C teststruct `json:"c"`
	}

	tests := map[string]struct {
		input       interface{}
		expected    []byte
		expectedErr error
	}{
		"bytes": {
			input:    []byte(`{"foo":"bar","baz":123}`),
			expected: []byte(`{"foo":"bar","baz":123}`),
		},
		"struct": {
			input:    teststruct2{A: "bar", B: 123, C: teststruct{Foo: "bar2", Baz: 456}},
			expected: []byte(`{"a":"bar","b":123,"c":{"foo":"bar2","baz":456}}`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
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

func TestUnit_PrefixRoute(t *testing.T) {
	tests := map[string]struct {
		serviceName              string
		pathPrefix               string
		appendServiceNameToRoute bool
		route                    string
		expectedRoute            string
	}{
		"do not append service name, empty prefix": {
			serviceName:              "testservice",
			appendServiceNameToRoute: false,
			route:                    "/foo/bar",
			expectedRoute:            "/foo/bar",
		},
		"append service name, empty prefix": {
			serviceName:              "testservice",
			appendServiceNameToRoute: true,
			route:                    "/foo/bar",
			expectedRoute:            "/testservice/foo/bar",
		},
		"do not append service name, has prefix": {
			serviceName:              "testservice",
			pathPrefix:               "v1",
			appendServiceNameToRoute: false,
			route:                    "/foo/bar",
			expectedRoute:            "/v1/foo/bar",
		},
		"append service name, has prefix": {
			serviceName:              "testservice",
			pathPrefix:               "v1",
			appendServiceNameToRoute: true,
			route:                    "/foo/bar",
			expectedRoute:            "/v1/testservice/foo/bar",
		},
		"append service name, has prefix, trailing slash in route": {
			serviceName:              "testservice",
			pathPrefix:               "v1",
			appendServiceNameToRoute: true,
			route:                    "/foo/bar/",
			expectedRoute:            "/v1/testservice/foo/bar",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			r := PrefixRoute(tc.serviceName, tc.pathPrefix, tc.appendServiceNameToRoute, tc.route)
			require.Equal(t, tc.expectedRoute, r)
		})
	}
}
