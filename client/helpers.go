package client

import (
	"bytes"
	"encoding/json"
	"io"
)

// ObjectToJSONReader will v to a io.Reader of the JSON representation of v
func ObjectToJSONReader(v interface{}) (io.Reader, error) {
	if by, ok := v.([]byte); ok {
		return bytes.NewBuffer(by), nil
	}
	by, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(by), nil
}
