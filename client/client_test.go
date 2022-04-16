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
			expectedErr: glitch.NewDataError(nil, "ERROR_MAKING_REQUEST", "Could not decode error response"),
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
			expectedErr: glitch.NewDataError(nil, "ERROR_DECODING_RESPONSE", "Could not decode response"),
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
			expectedErr: glitch.NewDataError(errors.New("error finding service"), "CANT_FIND_SERVICE", "Error finding service"),
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
