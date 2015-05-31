package mgots

import (
	"fmt"
)

type Error struct {
	InnerError error
	Message    string
}

func NewError(err error, format string, a ...interface{}) *Error {
	return &Error{
		InnerError: err,
		Message:    fmt.Sprintf(format, a...),
	}
}

func (c *Error) Error() string {
	if c.InnerError == nil {
		return c.Message
	}

	return fmt.Sprintf("%s: %s", c.Message, c.InnerError.Error())
}
