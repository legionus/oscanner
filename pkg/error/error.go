package error

import (
	"fmt"
)

var _ error = &Error{}

type Error struct {
	Type    string
	Message string
}

func NewError(code, format string, args ...interface{}) error {
	return &Error{
		Type:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// Error implements the error interface.
func (v *Error) Error() string {
	return fmt.Sprintf("%s: %s", v.Type, v.Message)
}

func IsError(code string, err error) bool {
	v, ok := err.(*Error)
	if !ok {
		return false
	}
	return v.Type == code
}
