// Provides implementation of errors for this project.
package errs

import (
	"fmt"
)

// ApiError is a general error struct for consuming and returning api errors.
//
// ApiError implements error interface.
type ApiError struct {
	Err  string `json:"error"`
	Msg  string `json:"message"`
	Code int    `json:"code"`
}

// Creates a new ApiError
func New(err string, msg string) *ApiError {
	return &ApiError{
		Err:  err,
		Msg:  msg,
		Code: 0,
	}
}

// implements the error interface for ApiError
func (e *ApiError) Error() string {
	s := ""
	if e.Code != 0 {
		s += fmt.Sprintf("[%3d] ", e.Code)
	}

	s += e.Err
	if len(e.Msg) > 0 {
		s += " - " + e.Msg
	}

	return s
}
