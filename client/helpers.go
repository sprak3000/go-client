package client

import (
	"bytes"
	"encoding/json"
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
