package errors

import (
	"fmt"
	"io"
)

type ErrorC struct {
	message string
	code    int
}

// check formatter implementation
var _ fmt.Formatter = (*ErrorC)(nil)

func New(code int, message string) *ErrorC {
	return &ErrorC{
		code:    code,
		message: message,
	}
}

// Error implements the error interface.
// A result is the same as with %s formatter and does not contain a stack trace.
func (e *ErrorC) Error() string {
	return e.fullMessage()
}

func (e *ErrorC) Code() int {
	return e.code
}

func (e *ErrorC) Format(s fmt.State, verb rune) {
	message := e.fullMessage()
	switch verb {
	case 'v':
		io.WriteString(s, message)
		// if s.Flag('+') {
		// 	e.stackTrace.Format(s, verb)
		// }
	case 's':
		io.WriteString(s, message)
	case 'd':
		io.WriteString(s, fmt.Sprintf("%d", e.code))
	}
}

func (e *ErrorC) fullMessage() string {
	return fmt.Sprintf("[%d]%s", e.code, e.message)
}
