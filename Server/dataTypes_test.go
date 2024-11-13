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
		// Type assertion to check if handler implements CompletionHandler
		_, ok := interface{}(handler).(CompletionHandler)
		if !ok {
			t.Errorf("Expected handler to implement CompletionHandler interface")
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
		// Type assertion to check if handler implements CompletionHandler
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

type MockReceiver struct {
	isGetOutChanError bool
	isSendToError     bool
	channel           chan *AppData
	DataReceived      any
}

func (t *MockReceiver) SendTo(receiver *AppData) error {
	if t.isSendToError {
		return errors.New("completion handler failed")
	}
	t.DataReceived = receiver.GetData()
	return nil
}

func (t *MockReceiver) GetOutChan() (<-chan *AppData, error) {
	if t.isGetOutChanError {
		return nil, errors.New("error getting channel")
	}
	return t.channel, nil
}

func TestReceiver(t *testing.T) {
	t.Run("SendTo", func(t *testing.T) {
		mockReceiver := &MockReceiver{}
		testData := &MockReceiver{}
		appData := &AppData{data: testData, handler: &MockCompletionHandler{}}

		err := mockReceiver.SendTo(appData)

		if err != nil {
			t.Errorf("Expected no error from Receive, got %v", err)
		}
		if mockReceiver.DataReceived != testData {
			t.Errorf("Expected DataReceived to be '%v', got '%v'", testData, mockReceiver.DataReceived)
		}
		// Type assertion to check if mockReceiver implements Receiver
		_, ok := interface{}(mockReceiver).(Receiver)
		if !ok {
			t.Errorf("Expected mockReceiver to implement Receiver interface")
		}
	})
	t.Run("SendToError", func(t *testing.T) {
		mockReceiver := &MockReceiver{
			isSendToError:     true,
			isGetOutChanError: false,
		}
		testData := "test data"
		appData := &AppData{data: testData, handler: &MockCompletionHandler{}}

		err := mockReceiver.SendTo(appData)

		if err == nil || err.Error() != "completion handler failed" {
			t.Errorf("Expected error 'completion handler failed', got %v", err)
		}
		// Type assertion to check if errorReceiver implements Receiver
		_, ok := interface{}(mockReceiver).(Receiver)
		if !ok {
			t.Errorf("Expected errorReceiver to implement Receiver interface")
		}
	})
	t.Run("GetOutChan", func(t *testing.T) {
		mockReceiver := &MockReceiver{channel: make(chan *AppData)}

		channel, err := mockReceiver.GetOutChan()

		if err != nil {
			t.Errorf("Expected no error from GetOutChan, got %v", err)
		}
		if channel == nil {
			t.Errorf("Expected channel to be non-nil")
		}

		_, ok := interface{}(mockReceiver).(Receiver)
		if !ok {
			t.Errorf("Expected mockReceiver to implement Receiver interface")
		}
	})
	t.Run("GetOutChanError", func(t *testing.T) {
		mockReceiver := &MockReceiver{isGetOutChanError: true}

		channel, err := mockReceiver.GetOutChan()

		if err == nil || err.Error() != "error getting channel" {
			t.Errorf("Expected error 'error getting channel', got %v", err)
		}
		if channel != nil {
			t.Errorf("Expected channel to be nil")
		}
	})
}
