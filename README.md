# go-client
This package works in concert with [go-glitch](https://github.com/healthimation/go-glitch) to encourage code based error handling during inter-service
communication.  If a service returns a [problem](https://github.com/healthimation/go-glitch/blob/master/glitch/http-problem.go) detail or http problem with a `code` field
this client will facilitate calling that service and parsing the response into a `glitch.DataError` or a successful response.

**Note** that this package looks up the service using the provided finder every time a request is made.  This allows it to work in more ephemeral environments where 
services might move frequently.  If you have performance concerns about looking up service urls we suggest implementing a short cache in the `ServiceFinder` function.

## Usage

```go 
finder := func(serviceName string, useTLS bool) (url.URL, error) {
    u, err := url.Parse("http://example.com/")
    return *u, err
}
bc := NewBaseClient(finder, "example-service", false, 10*time.Second)

type user struct {
    ID int `json:"id"`
    Name string `json:"name"`
}
u := user{}
err := tc.client.Do(r.Context(), "GET", "v1/user/1", nil, nil, &u)
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
```

