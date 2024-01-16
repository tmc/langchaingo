package qwenclient

import (
	"errors"
	"strconv"
)

var ErrNetwork = errors.New("network error")

type DashscopeError struct {
	Message string
	Cause   error
	Code    int
}

func (e *DashscopeError) Error() string {
	if e.Cause == nil {
		return e.Message + ": " + strconv.Itoa(e.Code)
	}
	return e.Message + ": " + strconv.Itoa(e.Code) + " " + e.Cause.Error()
}

type WrapMessageError struct {
	Message string
	Cause   error
}

func (e *WrapMessageError) Error() string {
	if e.Cause == nil {
		return e.Message
	}
	return e.Message + ": " + e.Cause.Error()
}

type EmptyRequestBodyError struct {
	Method string
}

func (e *EmptyRequestBodyError) Error() string {
	return "POST or PUT request body cannot be empty for method " + e.Method
}
