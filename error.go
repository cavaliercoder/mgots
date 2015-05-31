package mgots

import (
	"fmt"
)

type mgotsError struct {
	InnerError error
	Message    string
}

func newError(err error, format string, a ...interface{}) error {
	return &mgotsError{
		InnerError: err,
		Message:    fmt.Sprintf(format, a...),
	}
}

func (c *mgotsError) Error() string {
	if c.InnerError == nil {
		return c.Message
	}

	return fmt.Sprintf("%s: %s", c.Message, c.InnerError.Error())
}
