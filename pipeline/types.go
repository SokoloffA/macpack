package pipeline

import (
	"fmt"
	"strings"
)

type Error struct {
	message string
	wrapped error
}

func NewError(format string, a ...interface{}) *Error {
	return &Error{
		message: fmt.Sprintf(format, a...),
		wrapped: nil,
	}
}

func FieldNotFoundError(field string) *Error {
	return NewError(`The required field "%s" is missed.`, field)
}

func (e Error) Error() string {
	if e.wrapped == nil {
		return "• " + e.message
	}

	msg := strings.Replace(e.message, "\n", "\n  ", -1)
	wrp := strings.Replace(e.wrapped.Error(), "\n", "\n  ", -1)

	if !strings.HasPrefix(wrp, "•") {
		wrp = "• " + wrp
	}
	return "• " + msg + "\n\n  " + wrp
}

func (e *Error) Wrap(err error) Error {
	e.wrapped = err
	return *e
}

func (e Error) Unwrap() error {
	return e.wrapped
}
