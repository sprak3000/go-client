package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/sprak3000/go-glitch/glitch"
)

func TestUnit_Do(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/1":
			fmt.Fprintf(w, `{"foo":"bar"}`)

		case "/2":
			type tmp struct {
				Test bool `json:"test"`
			}
			ret := tmp{}
			dec := json.NewDecoder(r.Body)
			err := dec.Decode(&ret)
			if err != nil {
				log.Printf("couldn't decode body in test server: %v", err)
				fmt.Fprintf(w, `error`)
			}
			if ret.Test {
				fmt.Fprintf(w, `{"foo":"bar"}`)
			} else {
				fmt.Fprintf(w, `{"foo":"baz"}`)
			}
		case "/3":
			ret := glitch.HTTPProblem{Code: "FOOBAR", Status: 500, Detail: "test error"}
			by, _ := json.Marshal(ret)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, string(by))
		}
	}))
	defer testServer.Close()

	finder := func(serviceName string, useTLS bool) (url.URL, error) {
		u, err := url.Parse(testServer.URL)
		return *u, err
	}

	tests := map[string]struct {
		client           BaseClient
		ctx              context.Context
		method           string
		slug             string
		query            url.Values
		headers          http.Header
		body             io.Reader
		response         interface{}
		expectedResponse interface{}
		expectedErr      error
	}{
		"GET": {
			client:           NewBaseClient(finder, "foo", false, 10*time.Second, nil),
			method:           "GET",
			slug:             "1",
			response:         new(map[string]string),
			expectedResponse: &map[string]string{"foo": "bar"},
		},
		"POST": {
			client:           NewBaseClient(finder, "foo", false, 10*time.Second, nil),
			method:           "POST",
			slug:             "2",
			body:             bytes.NewBuffer([]byte(`{"test":true}`)),
			response:         new(map[string]string),
			expectedResponse: &map[string]string{"foo": "bar"},
		},
		"POST2": {
			client:           NewBaseClient(finder, "foo", false, 10*time.Second, nil),
			method:           "POST",
			slug:             "2",
			query:            url.Values{"foo": []string{"bar"}},
			body:             bytes.NewBuffer([]byte(`{"test":false}`)),
			response:         new(map[string]string),
			expectedResponse: &map[string]string{"foo": "baz"},
		},
		"error": {
			client:      NewBaseClient(finder, "foo", false, 10*time.Second, nil),
			method:      "GET",
			slug:        "3",
			expectedErr: glitch.FromHTTPProblem(glitch.HTTPProblem{Code: "FOOBAR", Status: 500, Detail: "test error"}, "Error from GET to foo - 3"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := tc.client.Do(tc.ctx, tc.method, tc.slug, tc.query, tc.headers, tc.body, tc.response)
			if !reflect.DeepEqual(err, tc.expectedErr) {
				t.Fatalf("Error actual (%v) did not match expected (%v)", err, tc.expectedErr)
			}
			if !reflect.DeepEqual(tc.response, tc.expectedResponse) {
				t.Fatalf("Actual (%#v) did not match expected (%#v)", tc.response, tc.expectedResponse)
			}
		})
	}
}
