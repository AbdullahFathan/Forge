package apperr

import (
	"errors"
	"fmt"
	"net/http"
)

type Error struct {
	Code    string
	Status  int
	Message string
	Details any
}

func (e *Error) Error() string {
	return e.Message
}

func New(code string, status int, message string) *Error {
	return &Error{Code: code, Status: status, Message: message}
}

func (e *Error) WithDetails(details any) *Error {
	cp := *e
	cp.Details = details
	return &cp
}

func (e *Error) WithMessage(msg string) *Error {
	cp := *e
	cp.Message = msg
	return &cp
}

var (
	ErrInvalidCredentials = New("INVALID_CREDENTIALS", http.StatusUnauthorized, "invalid email or password")
	ErrUnauthorized       = New("UNAUTHORIZED", http.StatusUnauthorized, "unauthorized")
	ErrForbidden          = New("FORBIDDEN", http.StatusForbidden, "forbidden")
	ErrInactiveUser       = New("USER_INACTIVE", http.StatusForbidden, "user is inactive")
	ErrNotFound           = New("NOT_FOUND", http.StatusNotFound, "resource not found")
	ErrConflict           = New("CONFLICT", http.StatusConflict, "resource already exists")
	ErrValidation         = New("VALIDATION_ERROR", http.StatusBadRequest, "validation error")
	ErrLastSuperAdmin     = New("LAST_SUPER_ADMIN", http.StatusBadRequest, "cannot demote or remove the last Super Admin")
	ErrRateLimited        = New("RATE_LIMITED", http.StatusTooManyRequests, "too many requests")
	ErrInternal           = New("INTERNAL_ERROR", http.StatusInternalServerError, "internal server error")
	ErrInvalidToken       = New("INVALID_TOKEN", http.StatusUnauthorized, "invalid or expired token")
)

func As(err error) (*Error, bool) {
	var ae *Error
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}

func WrapInternal(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := As(err); ok {
		return err
	}
	return fmt.Errorf("%w: %v", ErrInternal, err)
}
