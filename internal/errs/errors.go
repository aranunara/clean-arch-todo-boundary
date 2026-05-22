// Package errs provides application error categories and classification helpers.
package errs

import (
	stderrors "errors"
	"fmt"
	"strings"
)

var (
	ErrNotFound        = New("not found")
	ErrInvalidInput    = New("invalid input")
	ErrConflict        = New("conflict")
	ErrExternalService = New("external service error")
	ErrInternal        = New("internal error")
)

type markedError struct {
	err      error
	category error
}

func (e *markedError) Error() string {
	return e.err.Error()
}

func (e *markedError) Unwrap() error {
	return e.err
}

func (e *markedError) Is(target error) bool {
	return target == e.category
}

func New(message string) error {
	return stderrors.New(message)
}

func Errorf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

func Mark(err, category error) error {
	if err == nil {
		return nil
	}
	return &markedError{err: err, category: category}
}

func Is(err, target error) bool {
	return stderrors.Is(err, target)
}

func As(err error, target any) bool {
	return stderrors.As(err, target)
}

// ValidationError is a single field validation error.
type ValidationError struct {
	Field   string
	Message string
	Value   any
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func (e *ValidationError) Is(target error) bool {
	return target == ErrInvalidInput
}

// ValidationErrors is a collection of field validation errors.
type ValidationErrors struct {
	Errors []ValidationError
}

func (e *ValidationErrors) Error() string {
	if len(e.Errors) == 0 {
		return ErrInvalidInput.Error()
	}

	msgs := make([]string, 0, len(e.Errors))
	for _, err := range e.Errors {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

func (e *ValidationErrors) Is(target error) bool {
	return target == ErrInvalidInput
}

func NewValidationError(field, message string, value any) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	}
}

func NewValidationErrors(errs []ValidationError) error {
	return &ValidationErrors{Errors: errs}
}

func AsValidationErrors(err error) (*ValidationErrors, bool) {
	var validationErrors *ValidationErrors
	if As(err, &validationErrors) {
		return validationErrors, true
	}
	return nil, false
}

func IsInvalidInput(err error) bool {
	return Is(err, ErrInvalidInput)
}

func IsNotFound(err error) bool {
	return Is(err, ErrNotFound)
}

func IsConflict(err error) bool {
	return Is(err, ErrConflict)
}

func IsExternalService(err error) bool {
	return Is(err, ErrExternalService)
}

// UIError is an error with a user-facing message.
type UIError struct {
	err       error
	uiMessage string
}

func (e *UIError) Error() string {
	return e.err.Error()
}

func (e *UIError) Unwrap() error {
	return e.err
}

func (e *UIError) UIMessage() string {
	return e.uiMessage
}

func WithUIMessage(err error, msg string) error {
	return &UIError{err: err, uiMessage: msg}
}

func GetUIMessage(err error) string {
	var uiError *UIError
	if As(err, &uiError) {
		return uiError.UIMessage()
	}
	return ""
}
