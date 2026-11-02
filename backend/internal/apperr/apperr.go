package apperr

import "net/http"

type Error struct {
	Status  int
	Code    string
	Message string
}

func (e *Error) Error() string { return e.Code + ": " + e.Message }

func New(status int, code, message string) *Error {
	return &Error{Status: status, Code: code, Message: message}
}

func Invalid(code, message string) *Error {
	return New(http.StatusBadRequest, code, message)
}
func NotFound(code, message string) *Error {
	return New(http.StatusNotFound, code, message)
}
func Conflict(code, message string) *Error {
	return New(http.StatusConflict, code, message)
}
func Upstream(code, message string) *Error {
	return New(http.StatusBadGateway, code, message)
}
