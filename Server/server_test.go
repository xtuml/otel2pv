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

// MockPushable is a mock implementation of the Pushable interface
type MockPushable struct {
	isError  bool
	receiver Receiver
}

func (p *MockPushable) GetReceiver() (Receiver, error) {
	if p.isError {
		return nil, errors.New("test error")
	}
	return p.receiver, nil
}

func TestPushable(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		pushable := &MockPushable{
			isError:  false,
			receiver: &MockReceiver{},
		}

		receiver, err := pushable.GetReceiver()

		if receiver == nil {
			t.Errorf("Expected receiver to be non-nil")
		}
		if err != nil {
			t.Errorf("Expected no error from GetReceiver, got %v", err)
		}
		_, pushableOk := interface{}(pushable).(Pushable)
		if !pushableOk {
			t.Errorf("Expected pushable to implement Pushable interface")
		}
		_, ok := interface{}(pushable.receiver).(Receiver)
		if !ok {
			t.Errorf("Expected pushable's receiver to implement Receiver interface")
		}
	})
	t.Run("Error", func(t *testing.T) {
		pushable := &MockPushable{
			isError:  true,
			receiver: nil,
		}

		receiver, err := pushable.GetReceiver()

		if receiver != nil {
			t.Errorf("Expected receiver to be nil")
		}
		if err == nil {
			t.Errorf("Expected error from GetReceiver, got '%v'", err)
		}
		_, pushableOk := interface{}(pushable).(Pushable)
		if !pushableOk {
			t.Errorf("Expected pushable to implement Pushable interface")
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
type MockReceiverForPullable struct {
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

// MockSourceServer is a mock implementation of the SourceServer interface
type MockSourceServer struct {
	isError  bool
	Receiver Receiver
}

func (s *MockSourceServer) AddReceiver(receiver Receiver) error {
	if s.isError {
		return errors.New("test error")
	}
	s.Receiver = receiver
	return nil
}

func (s *MockSourceServer) Serve() error {
	if s.isError {
		return errors.New("test error")
	}
	return nil
}

func TestSourceServer(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		sourceServer := MockSourceServer{
			isError: false,
		}
		mockReceiver := MockReceiverForPullable{}
		err := sourceServer.AddReceiver(&mockReceiver)

		_, ok := interface{}(&sourceServer).(SourceServer)
		if !ok {
			t.Errorf("Expected sourceServer to implement SourceServer interface")
		}

		_, pullableOk := interface{}(&sourceServer).(Pullable)
		if !pullableOk {
			t.Errorf("Expected sourceServer to implement Pullable interface")
		}

		_, receiverOk := interface{}(&mockReceiver).(Receiver)
		if !receiverOk {
			t.Errorf("Expected mockReceiver to implement Receiver interface")
		}

		if err != nil {
			t.Errorf("Expected no error from AddReceiver, got %v", err)
		}
		if sourceServer.Receiver != &mockReceiver {
			t.Errorf("Expected receiver to be set, got %v", sourceServer.Receiver)
		}
		appData := AppData{
			data:    "test data",
			handler: &MockCompletionHandlerForPullable{},
		}
		_ = sourceServer.Receiver.SendTo(appData)
		if mockReceiver.AddedData != appData {
			t.Errorf("Expected data to be sent to receiver, got %v", mockReceiver.AddedData)
		}

		err = sourceServer.Serve()
		if err != nil {
			t.Errorf("Expected no error from Serve, got %v", err)
		}
	})
	t.Run("Error", func(t *testing.T) {
		sourceServer := MockSourceServer{
			isError: true,
		}

		err := sourceServer.AddReceiver(&MockReceiverForPullable{})

		if err == nil {
			t.Errorf("Expected error from AddReceiver, got '%v'", err)
		}

		err = sourceServer.Serve()

		if err == nil {
			t.Errorf("Expected error from Serve, got '%v'", err)
		}
	})
}
