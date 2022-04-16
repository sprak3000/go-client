package client

import (
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sprak3000/go-glitch/glitch"
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
		expectedErr glitch.DataError
		validate    func(t *testing.T, expected []byte, actual io.Reader, expectedErr, actualErr glitch.DataError)
	}{
		"base path- bytes": {
			input:    []byte(`{"foo":"bar","baz":123}`),
			expected: []byte(`{"foo":"bar","baz":123}`),
			validate: func(t *testing.T, expected []byte, actual io.Reader, expectedErr, actualErr glitch.DataError) {
				require.NoError(t, actualErr)
				by, iErr := ioutil.ReadAll(actual)
				require.NoError(t, iErr)
				require.Equal(t, expected, by)
			},
		},
		"base path- struct": {
			input:    teststruct2{A: "bar", B: 123, C: teststruct{Foo: "bar2", Baz: 456}},
			expected: []byte(`{"a":"bar","b":123,"c":{"foo":"bar2","baz":456}}`),
			validate: func(t *testing.T, expected []byte, actual io.Reader, expectedErr, actualErr glitch.DataError) {
				require.NoError(t, actualErr)
				by, iErr := ioutil.ReadAll(actual)
				require.NoError(t, iErr)
				require.Equal(t, expected, by)
			},
		},
		"exceptional path- cannot marshal JSON": {
			input:       make(chan int),
			expectedErr: glitch.NewDataError(nil, "ERROR_MARSHALLING_OBJECT", "Error marshalling object to json"),
			validate: func(t *testing.T, expected []byte, actual io.Reader, expectedErr, actualErr glitch.DataError) {
				require.Error(t, actualErr)
				require.Equal(t, expectedErr.Code(), actualErr.Code())
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ret, err := ObjectToJSONReader(tc.input)
			tc.validate(t, tc.expected, ret, tc.expectedErr, err)
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
