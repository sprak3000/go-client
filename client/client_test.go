package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sprak3000/go-glitch/glitch"
)

func TestUnit_Do(t *testing.T) {
	var (
		testServer *httptest.Server
	)

	tests := map[string]struct {
		method           string
		slug             string
		query            url.Values
		headers          http.Header
		body             io.Reader
		response         interface{}
		requestHandler   http.HandlerFunc
		finder           func(serviceName string, useTLS bool) (url.URL, error)
		expectedResponse interface{}
		expectedErr      glitch.DataError
		validate         func(t *testing.T, expectedResponse, actualResponse interface{}, expectedErr, actualErr glitch.DataError)
	}{
		"base path- GET": {
			method:   "GET",
			slug:     "1",
			response: new(map[string]string),
			requestHandler: func(w http.ResponseWriter, r *http.Request) {
				_, _ = fmt.Fprintf(w, `{"foo":"bar"}`)
			},
			finder: func(serviceName string, useTLS bool) (url.URL, error) {
				u, err := url.Parse(testServer.URL)
				return *u, err
			},
			expectedResponse: &map[string]string{"foo": "bar"},
			validate: func(t *testing.T, expectedResponse, actualResponse interface{}, expectedErr, actualErr glitch.DataError) {
				require.NoError(t, actualErr)
				require.Equal(t, expectedResponse, actualResponse)
			},
		},
		"base path- POST": {
			method:   "POST",
			slug:     "2",
			body:     bytes.NewBuffer([]byte(`{"test":true}`)),
			response: new(map[string]string),
			requestHandler: func(w http.ResponseWriter, r *http.Request) {
				type tmp struct {
					Test bool `json:"test"`
				}
				ret := tmp{}
				dec := json.NewDecoder(r.Body)
				err := dec.Decode(&ret)
				if err != nil {
					log.Printf("couldn't decode body in test server: %v", err)
					_, _ = fmt.Fprintf(w, `error`)
				}
				_, _ = fmt.Fprintf(w, `{"foo":"bar"}`)
			},
			finder: func(serviceName string, useTLS bool) (url.URL, error) {
				u, err := url.Parse(testServer.URL)
				return *u, err
			},
			expectedResponse: &map[string]string{"foo": "bar"},
			validate: func(t *testing.T, expectedResponse, actualResponse interface{}, expectedErr, actualErr glitch.DataError) {
				require.NoError(t, actualErr)
				require.Equal(t, expectedResponse, actualResponse)
			},
		},
		"base path- POST 2": {
			method:   "POST",
			slug:     "2",
			query:    url.Values{"foo": []string{"bar"}},
			body:     bytes.NewBuffer([]byte(`{"test":false}`)),
			response: new(map[string]string),
			requestHandler: func(w http.ResponseWriter, r *http.Request) {
				type tmp struct {
					Test bool `json:"test"`
				}
				ret := tmp{}
				dec := json.NewDecoder(r.Body)
				err := dec.Decode(&ret)
				if err != nil {
					log.Printf("couldn't decode body in test server: %v", err)
					_, _ = fmt.Fprintf(w, `error`)
				}
				_, _ = fmt.Fprintf(w, `{"foo":"baz"}`)
			},
			finder: func(serviceName string, useTLS bool) (url.URL, error) {
				u, err := url.Parse(testServer.URL)
				return *u, err
			},
			expectedResponse: &map[string]string{"foo": "baz"},
			validate: func(t *testing.T, expectedResponse, actualResponse interface{}, expectedErr, actualErr glitch.DataError) {
				require.NoError(t, actualErr)
				require.Equal(t, expectedResponse, actualResponse)
			},
		},
		"exceptional path- server returns an error": {
			method: "GET",
			slug:   "3",
			requestHandler: func(w http.ResponseWriter, r *http.Request) {
				ret := glitch.HTTPProblem{Code: "FOOBAR", Status: 500, Detail: "test error"}
				by, _ := json.Marshal(ret)
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = fmt.Fprintf(w, string(by))
			},
			finder: func(serviceName string, useTLS bool) (url.URL, error) {
				u, err := url.Parse(testServer.URL)
				return *u, err
			},
			expectedErr: glitch.FromHTTPProblem(glitch.HTTPProblem{Code: "FOOBAR", Status: 500, Detail: "test error"}, "Error from GET to foo - 3"),
			validate: func(t *testing.T, expectedResponse, actualResponse interface{}, expectedErr, actualErr glitch.DataError) {
				require.Error(t, actualErr)
				require.Equal(t, expectedErr, actualErr)
			},
		},
		"exceptional path- server returns an error with an improperly formatted error payload": {
			method: "GET",
			slug:   "3",
			requestHandler: func(w http.ResponseWriter, r *http.Request) {
				by, _ := json.Marshal(make(chan int))
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = fmt.Fprintf(w, string(by))
			},
			finder: func(serviceName string, useTLS bool) (url.URL, error) {
				u, err := url.Parse(testServer.URL)
				return *u, err
			},
			expectedErr: glitch.NewDataError(nil, ErrorDecodingError, "Could not decode error response"),
			validate: func(t *testing.T, expectedResponse, actualResponse interface{}, expectedErr, actualErr glitch.DataError) {
				require.Error(t, actualErr)
				require.Equal(t, expectedErr.Code(), actualErr.Code())
			},
		},
		"exceptional path- server returns success with an improperly formatted payload": {
			method:   "GET",
			slug:     "3",
			response: new(map[string]string),
			requestHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
				_, _ = fmt.Fprintf(w, `{bad:- "JSON`)
			},
			finder: func(serviceName string, useTLS bool) (url.URL, error) {
				u, err := url.Parse(testServer.URL)
				return *u, err
			},
			expectedErr: glitch.NewDataError(nil, ErrorDecodingResponse, "Could not decode response"),
			validate: func(t *testing.T, expectedResponse, actualResponse interface{}, expectedErr, actualErr glitch.DataError) {
				require.Error(t, actualErr)
				require.Equal(t, expectedErr.Code(), actualErr.Code())
			},
		},
		"exceptional path- error making request": {
			method:   "GET",
			slug:     "3",
			response: new(map[string]string),
			requestHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
				_, _ = fmt.Fprintf(w, `{bad:- "JSON`)
			},
			finder: func(serviceName string, useTLS bool) (url.URL, error) {
				u, _ := url.Parse(testServer.URL)
				return *u, errors.New("error finding service")
			},
			expectedErr: glitch.NewDataError(errors.New("error finding service"), ErrorCantFind, "Error finding service"),
			validate: func(t *testing.T, expectedResponse, actualResponse interface{}, expectedErr, actualErr glitch.DataError) {
				require.Error(t, actualErr)
				require.Equal(t, expectedErr, actualErr)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			testServer = httptest.NewServer(tc.requestHandler)

			defer testServer.Close()

			client := NewBaseClient(tc.finder, "foo", false, 10*time.Second, nil)
			err := client.Do(context.Background(), tc.method, tc.slug, tc.query, tc.headers, tc.body, tc.response)
			tc.validate(t, tc.expectedResponse, tc.response, tc.expectedErr, err)
		})
	}
}

func TestUnit_MakeRequest(t *testing.T) {
	var (
		testServer *httptest.Server
	)

	tests := map[string]struct {
		method             string
		slug               string
		query              url.Values
		headers            http.Header
		body               io.Reader
		requestHandler     http.HandlerFunc
		finder             func(serviceName string, useTLS bool) (url.URL, error)
		expectedStatusCode int
		expectedResponse   interface{}
		expectedErr        glitch.DataError
		validate           func(t *testing.T, expectedStatusCode, actualStatusCode int, expectedResponse interface{}, actualResponse []byte, expectedErr, actualErr glitch.DataError)
	}{

		"base path- GET": {
			method: "GET",
			slug:   "1",
			requestHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
				_, _ = fmt.Fprintf(w, `{"foo":"bar"}`)
			},
			finder: func(serviceName string, useTLS bool) (url.URL, error) {
				u, err := url.Parse(testServer.URL)
				return *u, err
			},
			expectedStatusCode: http.StatusCreated,
			expectedResponse:   &map[string]string{"foo": "bar"},
			validate: func(t *testing.T, expectedStatusCode, actualStatusCode int, expectedResponse interface{}, actualResponse []byte, expectedErr, actualErr glitch.DataError) {
				require.NoError(t, actualErr)
				require.Equal(t, expectedStatusCode, actualStatusCode)

				resp := new(map[string]string)
				err := json.Unmarshal(actualResponse, &resp)
				require.NoError(t, err)
				require.Equal(t, expectedResponse, resp)
			},
		},
		"exceptional path- cannot find service": {
			method:         "GET",
			slug:           "3",
			requestHandler: func(w http.ResponseWriter, r *http.Request) {},
			finder: func(serviceName string, useTLS bool) (url.URL, error) {
				u, _ := url.Parse(testServer.URL)
				return *u, errors.New("error finding service")
			},
			expectedErr: glitch.NewDataError(errors.New("error finding service"), ErrorCantFind, "Error finding service"),
			validate: func(t *testing.T, expectedStatusCode, actualStatusCode int, expectedResponse interface{}, actualResponse []byte, expectedErr, actualErr glitch.DataError) {
				require.Error(t, actualErr)
				require.Equal(t, expectedErr, actualErr)
			},
		},
		"exceptional path- cannot create request object": {
			method:         ":",
			slug:           "3",
			requestHandler: func(w http.ResponseWriter, r *http.Request) {},
			finder: func(serviceName string, useTLS bool) (url.URL, error) {
				u, err := url.Parse(testServer.URL)
				return *u, err
			},
			expectedErr: glitch.NewDataError(errors.New(`net/http: invalid method ":"`), ErrorRequestError, "Could not make the request"),
			validate: func(t *testing.T, expectedStatusCode, actualStatusCode int, expectedResponse interface{}, actualResponse []byte, expectedErr, actualErr glitch.DataError) {
				require.Error(t, actualErr)
				require.Equal(t, expectedErr, actualErr)
			},
		},
		"exceptional path- could not make the request": {
			method:         "GET",
			slug:           "3",
			requestHandler: func(w http.ResponseWriter, r *http.Request) {},
			finder: func(serviceName string, useTLS bool) (url.URL, error) {
				u, err := url.Parse("")
				return *u, err
			},
			expectedErr: glitch.NewDataError(nil, ErrorRequestError, "Could not make the request"),
			validate: func(t *testing.T, expectedStatusCode, actualStatusCode int, expectedResponse interface{}, actualResponse []byte, expectedErr, actualErr glitch.DataError) {
				require.Error(t, actualErr)
				require.Equal(t, expectedErr.Code(), actualErr.Code())
			},
		},
		"exceptional path- cannot read response body": {
			method: "GET",
			slug:   "3",
			requestHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Length", "1")
			},
			finder: func(serviceName string, useTLS bool) (url.URL, error) {
				u, err := url.Parse(testServer.URL)
				return *u, err
			},
			expectedErr: glitch.NewDataError(errors.New("unexpected EOF"), ErrorDecodingResponse, "Could not read response body"),
			validate: func(t *testing.T, expectedStatusCode, actualStatusCode int, expectedResponse interface{}, actualResponse []byte, expectedErr, actualErr glitch.DataError) {
				require.Error(t, actualErr)
				require.Equal(t, expectedErr, actualErr)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			testServer = httptest.NewServer(tc.requestHandler)

			defer testServer.Close()

			client := NewBaseClient(tc.finder, "foo", false, 10*time.Second, nil)
			statusCode, respBytes, err := client.MakeRequest(context.Background(), tc.method, tc.slug, tc.query, tc.headers, tc.body)
			tc.validate(t, tc.expectedStatusCode, statusCode, tc.expectedResponse, respBytes, tc.expectedErr, err)
		})
	}
}
