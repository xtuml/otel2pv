package Server

import "errors"

// FullError is a struct that holds an error message.
type FullError struct {
	message string
}

// Error returns the error message stored in the FullError struct.
//
// It returns:
//
// 1. string. The error message.
func (f *FullError) Error() string {
	return f.message
}

// NewFullError is a function that creates a new FullError struct.
//
// It takes as args:
//
// 1. message: string. The error message to be stored in the FullError struct.
//
// It returns:
//
// 1. *FullError. A pointer to the FullError struct.
func NewFullError(message string) *FullError {
	return &FullError{
		message: message,
	}
}

// NewFullErrorFromError is a function that creates a new FullError struct from an error.
//
// It takes as args:
//
// 1. err: error. The error to be stored in the FullError struct.
//
// It returns:
//
// 1. *FullError. A pointer to the FullError struct.
func NewFullErrorFromError(err error) *FullError {
	return &FullError{
		message: err.Error(),
	}
}

// InvalidError is a struct that holds an error message.
type InvalidError struct {
	message string
}

// Error returns the error message stored in the InvalidError struct.
//
// It returns:
//
// 1. string. The error message.
func (i *InvalidError) Error() string {
	return i.message
}

// NewInvalidError is a function that creates a new InvalidError struct.
//
// It takes as args:
//
// 1. message: string. The error message to be stored in the InvalidError struct.
//
// It returns:
//
// 1. *InvalidError. A pointer to the InvalidError struct.
func NewInvalidError(message string) *InvalidError {
	return &InvalidError{
		message: message,
	}
}

// NewInvalidErrorFromError is a function that creates a new InvalidError struct from an error.
//
// It takes as args:
//
// 1. err: error. The error to be stored in the InvalidError struct.
//
// It returns:
//
// 1. *InvalidError. A pointer to the InvalidError struct.
func NewInvalidErrorFromError(err error) *InvalidError {
	return &InvalidError{
		message: err.Error(),
	}
}

// SendError is a struct that holds an error message.
type SendError struct {
	message string
}

// Error returns the error message stored in the SendError struct.
//
// It returns:
//
// 1. string. The error message.
func (s *SendError) Error() string {
	return s.message
}

// NewSendError is a function that creates a new SendError struct.
//
// It takes as args:
//
// 1. message: string. The error message to be stored in the SendError struct.
//
// It returns:
//
// 1. *SendError. A pointer to the SendError struct.
func NewSendError(message string) *SendError {
	return &SendError{
		message: message,
	}
}

// NewSendErrorFromError is a function that creates a new SendError struct from an error.
//
// It takes as args:
//
// 1. err: error. The error to be stored in the SendError struct.
//
// It returns:
//
// 1. *SendError. A pointer to the SendError struct.
func NewSendErrorFromError(err error) *SendError {
	return &SendError{
		message: err.Error(),
	}
}

// ProcessError is a struct that holds an error message.
type ProcessError struct {
	message string
}

// Error returns the error message stored in the ProcessError struct.
//
// It returns:
//
// 1. string. The error message.
func (p *ProcessError) Error() string {
	return p.message
}

// NewProcessError is a function that creates a new ProcessError struct.
//
// It takes as args:
//
// 1. message: string. The error message to be stored in the ProcessError struct.
//
// It returns:
//
// 1. *ProcessError. A pointer to the ProcessError struct.
func NewProcessError(message string) *ProcessError {
	return &ProcessError{
		message: message,
	}
}

// NewProcessErrorFromError is a function that creates a new ProcessError struct from an error.
//
// It takes as args:
//
// 1. err: error. The error to be stored in the ProcessError struct.
//
// It returns:
//
// 1. *ProcessError. A pointer to the ProcessError struct.
func NewProcessErrorFromError(err error) *ProcessError {
	return &ProcessError{
		message: err.Error(),
	}
}

type ErrorType string

const (
	FullErrorType    ErrorType = "FullError"
	InvalidErrorType ErrorType = "InvalidError"
	SendErrorType    ErrorType = "SendError"
	ProcessErrorType ErrorType = "ProcessError"
	NoneErrorType    ErrorType = ""
)

// GetErrorType is a function that returns the type of error.
//
// It takes as args:
//
// 1. err: error. The error to get the type of.
//
// It returns:
//
// 1. ErrorType. The type of error.
func GetErrorType(err error) (ErrorType, error) {
	if err == nil {
		return NoneErrorType, errors.New("error is nil")
	}
	switch err.(type) {
	case *FullError:
		return FullErrorType, nil
	case *InvalidError:
		return InvalidErrorType, nil
	case *SendError:
		return SendErrorType, nil
	case *ProcessError:
		return ProcessErrorType, nil
	}
	return NoneErrorType, nil
}
