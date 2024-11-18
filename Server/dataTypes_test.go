package Server

import (
	"errors"
	"sync"
	"testing"
)

type MockCompletionHandler struct {
	DataReceived  any
	ErrorReceived error
	isError	   bool
}

func (t *MockCompletionHandler) Complete(data any, err error) error {
	if t.isError {
		return errors.New("error")
	}
	t.DataReceived = data
	t.ErrorReceived = err
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
		// Type assertion to check if handler implements CompletionHandler
		_, ok := interface{}(handler).(CompletionHandler)
		if !ok {
			t.Errorf("Expected handler to implement CompletionHandler interface")
		}
	})
	t.Run("Error", func(t *testing.T) {
		handler := &MockCompletionHandler{
			isError: true,
		}

		testError := errors.New("test error")
		err := handler.Complete(nil, testError)

		if err == nil {
			t.Errorf("Expected error '%v' from Complete, got '%v'", testError, err)
		}
		if handler.DataReceived != nil {
			t.Errorf("Expected no data, got %v", handler.DataReceived)
		}
		if handler.ErrorReceived != nil {
			t.Errorf("Expected no error '%v', got '%v'", testError, handler.ErrorReceived)
		}
		// Type assertion to check if handler implements CompletionHandler
		_, ok := interface{}(handler).(CompletionHandler)
		if !ok {
			t.Errorf("Expected handler to implement CompletionHandler interface")
		}
	})
}

// Tests for WaitGroupCompletionHandler
func TestWaitGroupCompletionHandler(t *testing.T) {
	t.Run("Complete", func(t *testing.T) {
		wg := &sync.WaitGroup{}
		handler := WaitGroupCompletionHandler{}
		// Test when wg is nil
		err := handler.Complete(nil, nil)
		if err == nil {
			t.Errorf("Expected error from Complete, got nil")
		}
		// Test when wg is set
		wg.Add(1)
		handler.wg = wg
		// Test when task is already complete
		handler.isComplete = true
		err = handler.Complete(nil, nil)
		if err == nil {
			t.Errorf("Expected error from Complete, got nil")
		}
		// Test when task is not complete
		handler.isComplete = false

		testData := "test data"
		err = handler.Complete(testData, nil)

		if err != nil {
			t.Errorf("Expected no error from Complete, got %v", err)
		}
		if handler.data != testData {
			t.Errorf("Expected data to be '%v', got '%v'", testData, handler.data)
		}
		if handler.err != nil {
			t.Errorf("Expected err to be nil, got %v", handler.err)
		}
		if !handler.isComplete {
			t.Errorf("Expected isComplete to be true, got false")
		}
		wg.Wait()
	})
	t.Run("GetData", func(t *testing.T) {
		handler := WaitGroupCompletionHandler{}
		// Test when task is not complete
		_, err := handler.GetData()
		if err == nil {
			t.Errorf("Expected error from GetData, got nil")
		}
		// Test when task is complete
		handler.isComplete = true
		testData := "test data"
		handler.data = testData

		data, err := handler.GetData()

		if err != nil {
			t.Errorf("Expected no error from GetData, got %v", err)
		}
		if data != testData {
			t.Errorf("Expected data to be '%v', got '%v'", testData, data)
		}
	})
	t.Run("GetError", func(t *testing.T) {
		handler := WaitGroupCompletionHandler{}
		// Test when task is not complete
		_, err := handler.GetError()
		if err == nil {
			t.Errorf("Expected error from GetError, got nil")
		}
		// Test when task is complete
		handler.isComplete = true
		testError := errors.New("test error")
		handler.err = testError

		err, err2 := handler.GetError()

		if err != testError {
			t.Errorf("Expected error to be '%v', got '%v'", testError, err)
		}
		if err2 != nil {
			t.Errorf("Expected no error from GetError, got %v", err2)
		}
	})
	t.Run("NewWaitGroupCompletionHandler", func(t *testing.T) {
		wg := &sync.WaitGroup{}

		handler := NewWaitGroupCompletionHandler(wg)

		if handler.wg != wg {
			t.Errorf("Expected wg to be '%v', got '%v'", wg, handler.wg)
		}
		if handler.isComplete {
			t.Errorf("Expected isComplete to be false, got true")
		}
		// Test for panic in WaitGroup.Done to make sure something has been added
		// to the WaitGroup
		wg.Done()
	})
	t.Run("ImplementCompletionHandler", func(t *testing.T) {
		handler := &WaitGroupCompletionHandler{}

		_, ok := interface{}(handler).(CompletionHandler)
		if !ok {
			t.Errorf("Expected handler to implement CompletionHandler interface")
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
		// Type assertion to check if handler implements CompletionHandler
		_, ok := interface{}(appData.handler).(CompletionHandler)
		if !ok {
			t.Errorf("Expected handler to implement CompletionHandler interface")
		}
	})
	t.Run("GetHandler", func(t *testing.T) {
		testHandler := &MockCompletionHandler{}
		appData := &AppData{}

		_, err := appData.GetHandler()

		if err == nil{
			t.Errorf("Expected error from GetHandler, got %v", err)
		}

		appData.handler = testHandler

		handler, err := appData.GetHandler()

		if err != nil {
			t.Errorf("Expected no error from GetHandler, got %v", err)
		}

		if handler != testHandler {
			t.Errorf("Expected handler to be '%v', got '%v'", testHandler, handler)
		}
	})
	t.Run("NewAppData", func(t *testing.T) {
		testData := "test data"
		testHandler := &MockCompletionHandler{}

		appData := NewAppData(testData, testHandler)

		if appData.data != testData {
			t.Errorf("Expected data to be '%v', got '%v'", testData, appData.data)
		}
		if appData.handler != testHandler {
			t.Errorf("Expected handler to be '%v', got '%v'", testHandler, appData.handler)
		}
	})
}
