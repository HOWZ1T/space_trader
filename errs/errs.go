package errs

import (
	"fmt"
)

type ApiError struct {
	Err  string `json:"error"`
	Msg  string `json:"message"`
	Code int    `json:"code"`
}

func New(err string, msg string) *ApiError {
	return &ApiError{
		Err:  err,
		Msg:  msg,
		Code: 0,
	}
}

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
