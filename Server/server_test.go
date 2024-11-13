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
	isError               bool
	isHandleIncomingError bool
	incomingData          *AppData
	receiver              Receiver
}

func (p *MockPushable) GetReceiver() (Receiver, error) {
	if p.isError {
		return nil, errors.New("test error")
	}
	return p.receiver, nil
}

func (p *MockPushable) HandleIncomingData(data *AppData) error {
	if p.isHandleIncomingError {
		return errors.New("test error")
	}
	p.incomingData = data
	return nil
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

		appData := &AppData{
			data:    "test data",
			handler: &MockCompletionHandler{},
		}
		err = pushable.HandleIncomingData(appData)
		if err != nil {
			t.Errorf("Expected no error from HandleIncomingData, got %v", err)
		}
		if pushable.incomingData != appData {
			t.Errorf("Expected incoming data to be set, got %v", pushable.incomingData)
		}
	})
	t.Run("Error", func(t *testing.T) {
		pushable := &MockPushable{
			isError:               true,
			isHandleIncomingError: true,
			receiver:              nil,
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

		appData := AppData{
			data:    "test data",
			handler: &MockCompletionHandler{},
		}

		err = pushable.HandleIncomingData(&appData)
		if err == nil {
			t.Errorf("Expected error from HandleIncomingData, got '%v'", err)
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
	AddedData *AppData
}

func (r *MockReceiverForPullable) SendTo(data *AppData) error {
	r.AddedData = data
	return nil
}

func (r *MockReceiverForPullable) GetOutChan() (<-chan *AppData, error) {
	return nil, nil
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
		appData := &AppData{
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
		appData := &AppData{
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

// MockSinkServer is a mock implementation of the SinkServer interface
type MockSinkServer struct {
	isError  bool
	Receiver Receiver
}

func (s *MockSinkServer) GetReceiver() (Receiver, error) {
	if s.isError {
		return nil, errors.New("test error")
	}
	return s.Receiver, nil
}

func (s *MockSinkServer) Serve() error {
	if s.isError {
		return errors.New("test error")
	}
	return nil
}

func (s *MockSinkServer) HandleIncomingData(data *AppData) error {
	return nil
}

func TestSinkServer(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		sinkServer := MockSinkServer{
			isError: false,
		}
		mockReceiver := MockReceiverForPullable{}
		sinkServer.Receiver = &mockReceiver

		_, ok := interface{}(&sinkServer).(SinkServer)
		if !ok {
			t.Errorf("Expected sinkServer to implement SinkServer interface")
		}

		_, pushableOk := interface{}(&sinkServer).(Pushable)
		if !pushableOk {
			t.Errorf("Expected sinkServer to implement Pushable interface")
		}

		_, receiverOk := interface{}(&mockReceiver).(Receiver)
		if !receiverOk {
			t.Errorf("Expected mockReceiver to implement Receiver interface")
		}

		receiver, err := sinkServer.GetReceiver()

		if receiver == nil {
			t.Errorf("Expected receiver to be non-nil")
		}
		if err != nil {
			t.Errorf("Expected no error from GetReceiver, got %v", err)
		}
		if sinkServer.Receiver != &mockReceiver {
			t.Errorf("Expected receiver to be set, got %v", sinkServer.Receiver)
		}

		err = sinkServer.Serve()
		if err != nil {
			t.Errorf("Expected no error from Serve, got %v", err)
		}
	})
	t.Run("Error", func(t *testing.T) {
		sinkServer := MockSinkServer{
			isError: true,
		}

		_, err := sinkServer.GetReceiver()

		if err == nil {
			t.Errorf("Expected error from GetReceiver, got '%v'", err)
		}

		err = sinkServer.Serve()

		if err == nil {
			t.Errorf("Expected error from Serve, got '%v)", err)
		}
	})
}

// MockPipeServer is a mock implementation of the PipeServer interface
type MockPipeServer struct {
	isAddError    bool
	isServeError  bool
	MyReceiver    Receiver
	TheirReceiver Receiver
}

func (s *MockPipeServer) AddReceiver(receiver Receiver) error {
	if s.isAddError {
		return errors.New("test error")
	}
	s.TheirReceiver = receiver
	return nil
}

func (s *MockPipeServer) GetReceiver() (Receiver, error) {
	if s.MyReceiver == nil {
		return nil, errors.New("test error")
	}
	return s.MyReceiver, nil
}

func (s *MockPipeServer) Serve() error {
	if s.isServeError {
		return errors.New("test error")
	}
	return nil
}

func (s *MockPipeServer) HandleIncomingData(data *AppData) error {
	return nil
}

func TestPipeServer(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		pipeServer := MockPipeServer{
			isAddError:   false,
			isServeError: false,
			MyReceiver:   &MockReceiverForPullable{},
		}
		mockReceiver := MockReceiverForPullable{}
		err := pipeServer.AddReceiver(&mockReceiver)

		_, ok := interface{}(&pipeServer).(PipeServer)
		if !ok {
			t.Errorf("Expected pipeServer to implement PipeServer interface")
		}

		_, pullableOk := interface{}(&pipeServer).(Pullable)
		if !pullableOk {
			t.Errorf("Expected pipeServer to implement Pullable interface")
		}

		_, pushableOk := interface{}(&pipeServer).(Pushable)
		if !pushableOk {
			t.Errorf("Expected pipeServer to implement Pushable interface")
		}

		if err != nil {
			t.Errorf("Expected no error from AddReceiver, got %v", err)
		}
		if pipeServer.TheirReceiver != &mockReceiver {
			t.Errorf("Expected receiver to be set, got %v", pipeServer.TheirReceiver)
		}

		err = pipeServer.Serve()
		if err != nil {
			t.Errorf("Expected no error from Serve, got %v", err)
		}

		receiver, err := pipeServer.GetReceiver()
		if receiver == nil {
			t.Errorf("Expected receiver to be non-nil")
		}
		if err != nil {
			t.Errorf("Expected no error from GetReceiver, got %v", err)
		}
		if pipeServer.MyReceiver != receiver {
			t.Errorf("Expected receiver to be set, got %v", pipeServer.MyReceiver)
		}
	})
	t.Run("Error", func(t *testing.T) {
		pipeServer := MockPipeServer{
			isAddError:   true,
			isServeError: true,
		}

		err := pipeServer.AddReceiver(&MockReceiverForPullable{})
		if err == nil {
			t.Errorf("Expected error from AddReceiver, got '%v'", err)
		}

		err = pipeServer.Serve()
		if err == nil {
			t.Errorf("Expected error from Serve, got '%v'", err)
		}

		_, err = pipeServer.GetReceiver()
		if err == nil {
			t.Errorf("Expected error from GetReceiver, got '%v'", err)
		}
	})
}

// Mock structures
// MockReceiverForPushableDataReceipt is a struct that provides a mock implementation of the Receiver interface
type MockReceiverForPushableDataReceipt struct {
	outChan chan *AppData
	isError bool
}

func (mr *MockReceiverForPushableDataReceipt) SendTo(data *AppData) error {
	return nil
}

func (mr *MockReceiverForPushableDataReceipt) GetOutChan() (<-chan *AppData, error) {
	if mr.isError {
		return nil, errors.New("error getting channel")
	}
	return mr.outChan, nil
}

// MockPushableForPushableDataReceipt is a struct that provides a mock implementation of the Pushable interface
type MockPushableForPushableDataReceipt struct {
	isReceiverError       bool
	isHandleIncomingError bool
	receiver              Receiver
	incomingData          []*AppData
}

func (p *MockPushableForPushableDataReceipt) GetReceiver() (Receiver, error) {
	if p.isReceiverError {
		return nil, errors.New("error getting receiver")
	}
	return p.receiver, nil
}

func (p *MockPushableForPushableDataReceipt) HandleIncomingData(data *AppData) error {
	if p.isHandleIncomingError {
		return errors.New("error handling incoming data")
	}
	p.incomingData = append(p.incomingData, data)
	return nil
}

func TestHandlePushableDataReceipt(t *testing.T) {
	tests := []struct {
		name                 string
		receiverError        bool
		handleIncomingError  bool
		expectedError        error
		numberChannels       int
		expectedIncomingData []*AppData
	}{
		{
			name:                "Successful data handling with single item",
			receiverError:       false,
			handleIncomingError: false,
			expectedError:       nil,
			expectedIncomingData: []*AppData{
				{data: "test data 1"},
			},
		},
		{
			name:                "Successful data handling with multiple items",
			receiverError:       false,
			handleIncomingError: false,
			expectedError:       nil,
			expectedIncomingData: []*AppData{
				{data: "test data 1"},
				{data: "test data 2"},
			},
		},
		{
			name:                 "Error retrieving receiver",
			receiverError:        true,
			handleIncomingError:  false,
			expectedError:        errors.New("error getting receiver"),
			expectedIncomingData: []*AppData{},
		},
		{
			name:                "Error handling incoming data",
			receiverError:       false,
			handleIncomingError: true,
			expectedError:       errors.New("error handling incoming data"),
			expectedIncomingData: []*AppData{
				{data: "test data 1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReceiver := &MockReceiverForPushableDataReceipt{
				outChan: make(chan *AppData, len(tt.expectedIncomingData)),
				isError: tt.receiverError,
			}
			if !tt.receiverError {
				for _, data := range tt.expectedIncomingData {
					mockReceiver.outChan <- data
				}
				close(mockReceiver.outChan)
			}

			mockPushable := &MockPushableForPushableDataReceipt{
				isReceiverError:       tt.receiverError,
				isHandleIncomingError: tt.handleIncomingError,
				receiver:              mockReceiver,
			}

			err := HandlePushableDataReceipt(mockPushable)

			// Verify the returned error matches the expected error
			if err != nil && tt.expectedError == nil || err == nil && tt.expectedError != nil {
				t.Fatalf("expected error: %v, got: %v", tt.expectedError, err)
			}
			if err != nil && tt.expectedError != nil && err.Error() != tt.expectedError.Error() {
				t.Fatalf("expected error message: %v, got: %v", tt.expectedError.Error(), err.Error())
			}

			// Early error return will prevent the incoming data from being processed,
			// causing mismatch between the expected and actual incoming data
			if tt.handleIncomingError {
				if len(mockPushable.incomingData) != 0 {
					t.Fatalf("expected no incoming data to be handled, but got %d", len(mockPushable.incomingData))
				}
				return
			}
			// Check that the processed data matches the expected data
			if len(mockPushable.incomingData) != len(tt.expectedIncomingData) {
				t.Fatalf("expected %d incoming data items to be handled, but got %d", len(tt.expectedIncomingData), len(mockPushable.incomingData))
			}

			for i, expectedData := range tt.expectedIncomingData {
				if mockPushable.incomingData[i] != expectedData {
					t.Fatalf("expected incoming data at index %d to be %v, but got %v", i, expectedData, mockPushable.incomingData[i])
				}
			}
		})
	}
}
