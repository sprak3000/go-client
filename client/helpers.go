package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/healthimation/go-glitch/glitch"
)

// ObjectToJSONReader will v to a io.Reader of the JSON representation of v
func ObjectToJSONReader(v interface{}) (io.Reader, glitch.DataError) {
	if by, ok := v.([]byte); ok {
		return bytes.NewBuffer(by), nil
	}
	by, err := json.Marshal(v)
	if err != nil {
		return nil, glitch.NewDataError(err, ErrorMarshallingObject, "Error marshalling object to json")
	}
	return bytes.NewBuffer(by), nil
}

func PrefixRoute(serviceName string, pathPrefix string, appendServiceNameToRoute bool, route string) string {
	if !appendServiceNameToRoute && pathPrefix == "" {
		return normalizePathPart(route)
	} else if appendServiceNameToRoute && pathPrefix == "" {
		return fmt.Sprintf("%s%s", normalizePathPart(serviceName), normalizePathPart(route))
	} else if !appendServiceNameToRoute && pathPrefix != "" {
		return fmt.Sprintf("%s%s", normalizePathPart(pathPrefix), normalizePathPart(route))
	}

	return fmt.Sprintf("%s%s%s", normalizePathPart(pathPrefix), normalizePathPart(serviceName), normalizePathPart(route))
}

func normalizePathPart(route string) string {
	// if there is a trailing / delete it
	if string(route[len(route)-1]) == "/" {
		route = route[:len(route)-1]
	}

	// if there already is a prepended / just return otherwise add one
	if string(route[0]) == "/" {
		return route
	}

	return fmt.Sprintf("/%s", route)
}
