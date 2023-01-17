package anymind

import "fmt"

type ErrorType int

const (
	InternalErr = iota
	ParameterErr
)

type Error struct {
	Type    int
	Cause   error
	Message string
}

func (e Error) Error() string {
	if e.Message != "" {
		return e.Message
	}

	switch e.Type {
	case InternalErr:
		return fmt.Sprintf("internal error: %s", e.Cause)
	case ParameterErr:
		return fmt.Sprintf("parameter error: %s", e.Cause)
	}

	return "error"
}

func ParameterError(err error) *Error {
	return &Error{
		Type:  ParameterErr,
		Cause: err,
	}
}

func InternalError(err error) *Error {
	return &Error{
		Type:  InternalErr,
		Cause: err,
	}
}
