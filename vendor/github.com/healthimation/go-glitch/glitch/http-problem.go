package glitch

import "fmt"

// HTTPProblem should be used as the response in case of an error during an HTTP request.
// It implements the https://datatracker.ietf.org/doc/rfc7807 spec with an addition code
// field which is meant to be machine readable and give clients enough information to handle the error appropriately.
type HTTPProblem struct {
	Type     string `json:"type,omitempty"`
	Title    string `json:"title,omitempty"`
	Status   int    `json:"status,omitempty"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
	Code     string `json:"code,omitempty"`
}

func (h HTTPProblem) Error() string {
	return fmt.Sprintf("HTTPProblem: [%d - %s] - %s - %s", h.Status, h.Code, h.Title, h.Detail)
}
