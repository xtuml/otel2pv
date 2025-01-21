package Server

import (
	"errors"
	"testing"
)

// TestErrors is a function that tests the error structs and functions.
func TestErrors(t *testing.T) {
	t.Run("TestFullError", func(t *testing.T) {
		// Test implements error interface for NewFullError.
		testError := NewFullError("test error")
		if _, ok := interface{}(testError).(error); !ok {
			t.Errorf("NewFullError does not implement error interface")
		}
		if testError.Error() != "test error" {
			t.Errorf("Error() does not return the correct message")
		}
		// Test NewFullErrorFromError.
		testErrorFromError := NewFullErrorFromError(errors.New("test error from error"))
		if testErrorFromError.Error() != "test error from error" {
			t.Errorf("Error() does not return the correct message")
		}
	})
	t.Run("TestInvalidError", func(t *testing.T) {
		// Test implements error interface for InvalidError.
		testError := NewInvalidError("test error")
		if _, ok := interface{}(testError).(error); !ok {
			t.Errorf("InvalidError does not implement error interface")
		}
		if testError.Error() != "test error" {
			t.Errorf("Error() does not return the correct message")
		}
		// Test NewInvalidErrorFromError.
		testErrorFromError := NewInvalidErrorFromError(errors.New("test error from error"))
		if testErrorFromError.Error() != "test error from error" {
			t.Errorf("Error() does not return the correct message")
		}
	})
	t.Run("TestSendError", func(t *testing.T) {
		// Test implements error interface for SendError.
		testError := NewSendError("test error")
		if _, ok := interface{}(testError).(error); !ok {
			t.Errorf("SendError does not implement error interface")
		}
		if testError.Error() != "test error" {
			t.Errorf("Error() does not return the correct message")
		}
		// Test NewSendErrorFromError.
		testErrorFromError := NewSendErrorFromError(errors.New("test error from error"))
		if testErrorFromError.Error() != "test error from error" {
			t.Errorf("Error() does not return the correct message")
		}
	})
	t.Run("TestProcessError", func(t *testing.T) {
		// Test implements error interface for ProcessError.
		testError := NewProcessError("test error")
		if _, ok := interface{}(testError).(error); !ok {
			t.Errorf("ProcessError does not implement error interface")
		}
		if testError.Error() != "test error" {
			t.Errorf("Error() does not return the correct message")
		}
		// Test NewProcessErrorFromError.
		testErrorFromError := NewProcessErrorFromError(errors.New("test error from error"))
		if testErrorFromError.Error() != "test error from error" {
			t.Errorf("Error() does not return the correct message")
		}
	})
}

// Test GetErrorType is a function that tests the GetErrorType function.
func TestGetErrorType(t *testing.T) {
	// Test case: nil error.
	_, err := GetErrorType(nil)
	if err == nil {
		t.Errorf("GetErrorType did not return an error for nil error")
	}
	// Test case: FullError.
	fullError := NewFullError("test error")
	errorType, err := GetErrorType(fullError)
	if err != nil {
		t.Errorf("GetErrorType returned an error for FullError")
	}
	if errorType != FullErrorType {
		t.Errorf("GetErrorType did not return the correct error type")
	}
	// Test case: InvalidError.
	invalidError := NewInvalidError("test error")
	errorType, err = GetErrorType(invalidError)
	if err != nil {
		t.Errorf("GetErrorType returned an error for InvalidError")
	}
	if errorType != InvalidErrorType {
		t.Errorf("GetErrorType did not return the correct error type")
	}
	// Test case: SendError.
	sendError := NewSendError("test error")
	errorType, err = GetErrorType(sendError)
	if err != nil {
		t.Errorf("GetErrorType returned an error for SendError")
	}
	if errorType != SendErrorType {
		t.Errorf("GetErrorType did not return the correct error type")
	}
	// Test case: ProcessError.
	processError := NewProcessError("test error")
	errorType, err = GetErrorType(processError)
	if err != nil {
		t.Errorf("GetErrorType returned an error for ProcessError")
	}
	if errorType != ProcessErrorType {
		t.Errorf("GetErrorType did not return the correct error type")
	}
	// Test case: unknown error.
	unknownError := errors.New("test error")
	errorType, err = GetErrorType(unknownError)
	if err != nil {
		t.Errorf("GetErrorType returned an error for unknown error")
	}
	if errorType != NoneErrorType {
		t.Errorf("GetErrorType did not return the correct error type")
	}
}
