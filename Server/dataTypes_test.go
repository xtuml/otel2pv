package Server

import (
	"errors"
	"testing"
)

type MockCompletionHandler struct {
	DataReceived  any
	ErrorReceived error
}

func (t *MockCompletionHandler) Complete(data any, err error) error {
	t.DataReceived = data
	t.ErrorReceived = err
	if err != nil {
		return err
	}
	return nil
}

func TestCompletionHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		handler := &MockCompletionHandler{}

		testData := "test data"
		err := handler.Complete(testData, nil)

		if err != nil {
			t.Errorf("Expected no error from Complete, got %v", err)
		}
		if handler.DataReceived != testData {
			t.Errorf("Expected DataReceived to be '%v', got '%v'", testData, handler.DataReceived)
		}
		if handler.ErrorReceived != nil {
			t.Errorf("Expected ErrorReceived to be nil, got %v", handler.ErrorReceived)
		}
	})
	t.Run("Error", func(t *testing.T) {
		handler := &MockCompletionHandler{}

		testError := errors.New("test error")
		err := handler.Complete(nil, testError)

		if err != testError {
			t.Errorf("Expected error '%v' from Complete, got '%v'", testError, err)
		}
		if handler.DataReceived != nil {
			t.Errorf("Expected no data, got %v", handler.DataReceived)
		}
		if handler.ErrorReceived != testError {
			t.Errorf("Expected error '%v', got '%v'", testError, handler.ErrorReceived)
		}
	})
}

func TestAppData(t *testing.T) {
	t.Run("GetData", func(t *testing.T) {
		testData := "test data"
		appData := &AppData{data: testData, handler: &MockCompletionHandler{}}

		data := appData.GetData()

		if data != testData {
			t.Errorf("Expected data to be '%v', got '%v'", testData, data)
		}
	})
	t.Run("Complete", func(t *testing.T) {
		mockHandler := &MockCompletionHandler{}
		appData := &AppData{data: "test data", handler: mockHandler}

		err := appData.handler.Complete(appData.data, nil)

		if err != nil {
			t.Errorf("Expected no error from Complete, got %v", err)
		}
		if mockHandler.DataReceived != appData.data {
			t.Errorf("Expected DataReceived to be '%v', got '%v'", appData.data, mockHandler.DataReceived)
		}
		if mockHandler.ErrorReceived != nil {
			t.Errorf("Expected ErrorReceived to be nil, got %v", mockHandler.ErrorReceived)
		}
	})
}
