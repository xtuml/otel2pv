package Server

import (
	"errors"
	"testing"
)

// MockServer is a mock implementation of the Server interface
type MockServer struct {
	isError bool
}

func (s *MockServer) Serve() error {
	if s.isError {
		return errors.New("test error")
	}
	return nil
}
func TestServer(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := MockServer{
			isError: false,
		}

		err := server.Serve()

		_, ok := interface{}(&server).(Server)
		if !ok {
			t.Errorf("Expected server to implement Server interface")
		}

		if err != nil {
			t.Errorf("Expected no error from Serve, got %v", err)
		}
	})
	t.Run("Error", func(t *testing.T) {
		server := MockServer{
			isError: true,
		}

		err := server.Serve()

		if err == nil {
			t.Errorf("Expected error from Serve, got '%v'", err)
		}
	})
}

// MockPullable is a mock implementation of the Pullable interface
type MockPullable struct {
	isError  bool
	Receiver Receiver
}

func (p *MockPullable) AddReceiver(receiver Receiver) error {
	if p.isError {
		return errors.New("test error")
	}
	p.Receiver = receiver
	return nil
}

// MockReceiver is a mock implementation of the Receiver struct
type MockReceiverForPullable struct{
	AddedData AppData
}

func (r *MockReceiverForPullable) SendTo(data AppData) error {
	r.AddedData = data
	return nil
}

// MockCompletionHandlerForPullable is a mock implementation of the CompletionHandler struct
type MockCompletionHandlerForPullable struct{}

func (c *MockCompletionHandlerForPullable) Complete(data any, err error) error {
	return nil
}

func TestPullable(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		pullable := MockPullable{
			isError: false,
		}
		mockReceiver := MockReceiverForPullable{}
		err := pullable.AddReceiver(&mockReceiver)
		_, ok := interface{}(&pullable).(Pullable)
		if !ok {
			t.Errorf("Expected pullable to implement Pullable interface")
		}
		if err != nil {
			t.Errorf("Expected no error from AddReceiver, got %v", err)
		}
		if pullable.Receiver != &mockReceiver {
			t.Errorf("Expected receiver to be set, got %v", pullable.Receiver)
		}
		appData := AppData{
			data:    "test data",
			handler: &MockCompletionHandlerForPullable{},
		}
		_ = pullable.Receiver.SendTo(appData)
		if mockReceiver.AddedData != appData {
			t.Errorf("Expected data to be sent to receiver, got %v", mockReceiver.AddedData)
		}

	})
	t.Run("Error", func(t *testing.T) {
		pullable := &MockPullable{
			isError: true,
		}

		err := pullable.AddReceiver(&MockReceiverForPullable{})

		if err == nil {
			t.Errorf("Expected error from AddReceiver, got '%v'", err)
		}
	})
}
