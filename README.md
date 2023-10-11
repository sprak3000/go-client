# go-client

[![Code Quality & Tests](https://github.com/sprak3000/go-client/actions/workflows/quality-and-tests.yml/badge.svg)](https://github.com/sprak3000/go-client/actions/workflows/quality-and-tests.yml)
[![Maintainability](https://api.codeclimate.com/v1/badges/61b8abbabfa223658774/maintainability)](https://codeclimate.com/github/sprak3000/go-client/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/61b8abbabfa223658774/test_coverage)](https://codeclimate.com/github/sprak3000/go-client/test_coverage)

This package works in concert with [go-glitch](https://github.com/sprak3000/go-glitch) to encourage code based error
handling during inter-service communication.  If a service returns a
[problem](https://github.com/HqOapp/go-glitch/blob/master/glitch/http-problem.go) detail or HTTP problem with a `code`
field, this client will facilitate calling that service and parsing the response into a `glitch.DataError` or a
successful response.

**NOTE:** that this package looks up the service using the provided finder every time a request is made. This allows it
to work in more ephemeral environments where services might move frequently. If you have performance concerns about
looking up service urls we suggest implementing a short cache in the `ServiceFinder` function.

Interested in making this library better? Read through our [development guide](docs/development.md).

## Usage

### Working with services returning the glitch.HTTPProblem (RFC 7807) format

Use this pattern if your service returns error conditions in the `glitch.HTTPProblem` format
([RFC 7807](https://datatracker.ietf.org/doc/rfc7807)):

```json
{
  "type": "error type",
  "title": "title",
  "status": 400,
  "detail": "More information about the error...",
  "instance": "More information about the error...",
  "code": "ERROR_CODE"
}
```

First, create a service finder closure. This will be used by the client to create the base service URL used
for requests.

```go
finder := func(serviceName string, useTLS bool) (url.URL, error) {
	var u *url.URL
	var err error

	switch serviceName {
	case "example-service":
	    u, err = url.Parse("https://example.com/")
	case "foo-service":
        u, err := url.Parse("https://foo.com/")
	default:
		return nil, glitch.NewDataError(nil, ErrorCantFind, "unknown service")
	}

    return *u, err
}
```

Create a base client to use to make calls by providing the finder closure, the name of the service, a boolean indicating
if TLS is to be used when connecting to the service, the amount of time before a call times out, and an optional HTTP
transport layer (`http.DefaultTransport` is used by default).

```go
bc := NewBaseClient(finder, "example-service", false, 10*time.Second)
```

You can now use the `Do()` method on the client to make a call to a service and process the result. As an example,
a service returns user data on a particular API endpoint. We define a type to contain the data from the response and
create a variable of that type.

```go
type user struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}
u := user{}
```

We pass our variable via pointer to `Do()` along with the HTTP method to use, the path to the endpoint, query
parameters, HTTP headers, and body data for `POST` requests.

```go
headers := http.Header{}
queryParams := url.Values{}
bodyReader, objErr := ObjectToJSONReader(bodyDataStruct)
if objErr != nil {
    // handle error
}

err := tc.client.Do(r.Context(), "GET", "v1/user/1", queryParams, headers, bodyReader, &u)
```

If the API call succeeds, `Do()` will populate the response variable `u`. Otherwise, `err` will be populated with the
details from the response, allowing you to process it as needed.

```go
// Handle any errors
if err != nil {
    switch err.Code() {
    case "USER_NOT_FOUND":
        w.WriteHeader(http.StatusNotFound)
        // ...
    case "PERMISSION_DENIED":
        w.WriteHeader(http.StatusForbidden)
        // ...
    case "USER_SETTING_PRIVATE"
        w.WriteHeader(http.StatusForbidden)
        // ...
    }
}

// Do things with your response
name := u.Name
```

### Working with services returning a non-glitch.HTTPProblem (RFC 7807) format

If the service returns a different error format, use `MakeRequest()` to make the service call. It is called nearly
identical to `Do()`, except it omits passing in a variable to hold the response from your service call.

```go
headers := http.Header{}
queryParams := url.Values{}
bodyReader, objErr := ObjectToJSONReader(bodyDataStruct)
if objErr != nil {
    // handle error
}

statusCode, respBytes, err := tc.client.MakeRequest(r.Context(), "GET", "v1/user/1", queryParams, headers, bodyReader)
```

`MakeRequest()` returns the HTTP status code from the response, the response payload as a byte slice, and a
`glitch.DataError` for any errors encountered trying to make the request or parsing the response. You will need to
handle converting the response byte slice (`respBytes`) yourself.

```go
type user struct {
    ID   int `json:"id"`
    Name string `json:"name"`
}
u := user{}

uErr := json.Unmarshal(respBytes, &u)
if uErr != nil {
    // handle error
}

name := u.Name
```
