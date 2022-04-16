package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/sprak3000/go-glitch/glitch"
)

// Error codes
const (
	ErrorCantFind          = "CANT_FIND_SERVICE"
	ErrorRequestCreation   = "CANT_CREATE_REQUEST"
	ErrorRequestError      = "ERROR_MAKING_REQUEST"
	ErrorDecodingError     = "ERROR_DECODING_ERROR"
	ErrorDecodingResponse  = "ERROR_DECODING_RESPONSE"
	ErrorMarshallingObject = "ERROR_MARSHALLING_OBJECT"
)

// ServiceFinder can find a service's base URL
type ServiceFinder func(serviceName string, useTLS bool) (url.URL, error)

//go:generate go run -mod=mod github.com/golang/mock/mockgen -package clientmock -destination=./clientmock/client-mock.go -source=../client/client.go -build_flags=-mod=mod

// BaseClient can do requests
type BaseClient interface {
	Do(ctx context.Context, method string, slug string, query url.Values, headers http.Header, body io.Reader, response interface{}) glitch.DataError
	MakeRequest(ctx context.Context, method string, slug string, query url.Values, headers http.Header, body io.Reader) (int, []byte, glitch.DataError)
}

type client struct {
	finder      ServiceFinder
	useTLS      bool
	serviceName string
	client      *http.Client
}

// NewBaseClient creates a new BaseClient
func NewBaseClient(finder ServiceFinder, serviceName string, useTLS bool, timeout time.Duration, rt http.RoundTripper) BaseClient {
	if rt == nil {
		rt = http.DefaultTransport
	}
	c := &http.Client{
		Timeout:   timeout,
		Transport: rt,
	}

	return &client{finder: finder, serviceName: serviceName, useTLS: useTLS, client: c}
}

// Do parses the request body into the response provider if in the 2xx range; otherwise, parses it into a glitch.DataError
func (c *client) Do(ctx context.Context, method string, slug string, query url.Values, headers http.Header, body io.Reader, response interface{}) glitch.DataError {
	status, ret, err := c.MakeRequest(ctx, method, slug, query, headers, body)
	if err != nil {
		return err
	}

	if status >= 400 || status < 200 {
		prob := glitch.HTTPProblem{}
		err := json.Unmarshal(ret, &prob)
		if err != nil {
			return glitch.NewDataError(err, ErrorDecodingError, "Could not decode error response")
		}
		return glitch.FromHTTPProblem(prob, fmt.Sprintf("Error from %s to %s - %s", method, c.serviceName, slug))
	}

	if response != nil {
		err := json.Unmarshal(ret, response)
		if err != nil {
			return glitch.NewDataError(err, ErrorDecodingResponse, "Could not decode response")
		}
	}

	return nil
}

// MakeRequest does the request and returns the status, body, and any error.
// This should be used only if the API doesn't return errors in the glitch.DataError format.
func (c *client) MakeRequest(ctx context.Context, method string, slug string, query url.Values, headers http.Header, body io.Reader) (int, []byte, glitch.DataError) {
	u, err := c.finder(c.serviceName, c.useTLS)
	if err != nil {
		return 0, nil, glitch.NewDataError(err, ErrorCantFind, "Error finding service")
	}
	u.Path = slug
	u.RawQuery = query.Encode()

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return 0, nil, glitch.NewDataError(err, ErrorRequestCreation, "Error creating request object")
	}

	req.Header = headers

	if ctx != nil {
		req = req.WithContext(ctx)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, nil, glitch.NewDataError(err, ErrorRequestError, "Could not make the request")
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	ret, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, glitch.NewDataError(err, ErrorDecodingResponse, "Could not read response body")
	}

	return resp.StatusCode, ret, nil
}
