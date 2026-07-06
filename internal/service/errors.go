package service

import (
	"errors"
	"net/http"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrInvalidInput = errors.New("invalid input")
)

type AppError struct {
	Status  int
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewError(status int, code, message string, err error) *AppError {
	return &AppError{Status: status, Code: code, Message: message, Err: err}
}

func InvalidInput(message string) *AppError {
	return NewError(http.StatusBadRequest, "invalid_request", message, ErrInvalidInput)
}

func NotFound(message string) *AppError {
	return NewError(http.StatusNotFound, "not_found", message, ErrNotFound)
}

func Conflict(message string) *AppError {
	return NewError(http.StatusConflict, "conflict", message, ErrConflict)
}

func InvalidAPIKey() *AppError {
	return NewError(http.StatusUnauthorized, "invalid_api_key", "Your API key is invalid", nil)
}

func Forbidden(code, message string) *AppError {
	return NewError(http.StatusForbidden, code, message, nil)
}
