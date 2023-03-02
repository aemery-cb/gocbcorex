package cbsearchx

import (
	"errors"
	"fmt"
)

var ErrInvalidArgument = errors.New("invalid argument")

type InvalidArgumentError struct {
	Message string
}

func (e InvalidArgumentError) Error() string {
	return fmt.Sprintf("invalid argument: %s", e.Message)
}

func (e InvalidArgumentError) Unwrap() error {
	return ErrInvalidArgument
}

type SearchError struct {
	InnerError     error       `json:"-"`
	Query          interface{} `json:"query,omitempty"`
	Endpoint       string      `json:"endpoint,omitempty"`
	ErrorText      string      `json:"error_text"`
	IndexName      string      `json:"index_name,omitempty"`
	HTTPStatusCode int         `json:"http_status_code,omitempty"`
}

// Unwrap returns the underlying cause for this error.
func (e SearchError) Unwrap() error {
	return e.InnerError
}
