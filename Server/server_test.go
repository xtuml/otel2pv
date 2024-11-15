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
	isSendToError bool
	incomingData          *AppData
}

func (p *MockPushable) SendTo(data *AppData) error {
	if p.isSendToError {
		return errors.New("test error")
	}
	p.incomingData = data
	return nil
}

func TestPushable(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		pushable := &MockPushable{
			isSendToError:  false,
		}


		_, pushableOk := interface{}(pushable).(Pushable)
		if !pushableOk {
			t.Errorf("Expected pushable to implement Pushable interface")
		}

		appData := &AppData{
			data:    "test data",
			handler: &MockCompletionHandler{},
		}
		err := pushable.SendTo(appData)
		if err != nil {
			t.Errorf("Expected no error from SendTo, got %v", err)
		}
		if pushable.incomingData != appData {
			t.Errorf("Expected incoming data to be set, got %v", pushable.incomingData)
		}
	})
	t.Run("Error", func(t *testing.T) {
		pushable := &MockPushable{
			isSendToError:               true,
		}
		appData := AppData{
			data:    "test data",
			handler: &MockCompletionHandler{},
		}

		err := pushable.SendTo(&appData)
		if err == nil {
			t.Errorf("Expected error from SendTo, got '%v'", err)
		}
	})
}

// MockPullable is a mock implementation of the Pullable interface
type MockPullable struct {
	isError  bool
	Pushable Pushable
}

func (p *MockPullable) AddPushable(pushable Pushable) error {
	if p.isError {
		return errors.New("test error")
	}
	p.Pushable = pushable
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
		mockPushable := MockPushable{
			isSendToError: false,
		}
		err := pullable.AddPushable(&mockPushable)
		_, ok := interface{}(&pullable).(Pullable)
		if !ok {
			t.Errorf("Expected pullable to implement Pullable interface")
		}
		if err != nil {
			t.Errorf("Expected no error from AddPushable, got %v", err)
		}
		if pullable.Pushable != &mockPushable {
			t.Errorf("Expected Pushable to be set, got %v", pullable.Pushable)
		}
		appData := &AppData{
			data:    "test data",
			handler: &MockCompletionHandlerForPullable{},
		}
		_ = pullable.Pushable.SendTo(appData)
		if mockPushable.incomingData != appData {
			t.Errorf("Expected data to be sent to Pushable, got %v", mockPushable.incomingData)
		}

	})
	t.Run("Error", func(t *testing.T) {
		pullable := &MockPullable{
			isError: true,
		}

		err := pullable.AddPushable(&MockPushable{})

		if err == nil {
			t.Errorf("Expected error from AddPushable, got '%v'", err)
		}
	})
}

// MockSourceServer is a mock implementation of the SourceServer interface
type MockSourceServer struct {
	isAddError  bool
	isServeError bool
	Pushable Pushable
}

func (s *MockSourceServer) AddPushable(pushable Pushable) error {
	if s.isAddError {
		return errors.New("test error")
	}
	s.Pushable = pushable
	return nil
}

func (s *MockSourceServer) Serve() error {
	if s.isServeError {
		return errors.New("test error")
	}
	return nil
}

func TestSourceServer(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		sourceServer := MockSourceServer{
			isAddError: false,
			isServeError: false,
		}
		mockPushable := MockPushable{
			isSendToError: false,
		}
		err := sourceServer.AddPushable(&mockPushable)

		_, ok := interface{}(&sourceServer).(SourceServer)
		if !ok {
			t.Errorf("Expected sourceServer to implement SourceServer interface")
		}

		_, pullableOk := interface{}(&sourceServer).(Pullable)
		if !pullableOk {
			t.Errorf("Expected sourceServer to implement Pullable interface")
		}

		if err != nil {
			t.Errorf("Expected no error from AddPushable, got %v", err)
		}
		if sourceServer.Pushable != &mockPushable {
			t.Errorf("Expected Pushable to be set, got %v", sourceServer.Pushable)
		}
		appData := &AppData{
			data:    "test data",
			handler: &MockCompletionHandlerForPullable{},
		}
		err = sourceServer.Pushable.SendTo(appData)
		if err != nil {
			t.Errorf("Expected no error from SendTo, got %v", err)
		}
		if mockPushable.incomingData != appData {
			t.Errorf("Expected data to be sent to Pushable, got %v", mockPushable.incomingData)
		}

		err = sourceServer.Serve()
		if err != nil {
			t.Errorf("Expected no error from Serve, got %v", err)
		}
	})
	t.Run("Error", func(t *testing.T) {
		sourceServer := MockSourceServer{
			isAddError: true,
			isServeError: true,
		}

		err := sourceServer.AddPushable(&MockPushable{})

		if err == nil {
			t.Errorf("Expected error from AddPushable, got '%v'", err)
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
	incomingData		  *AppData
}

func (s *MockSinkServer) Serve() error {
	if s.isError {
		return errors.New("test error")
	}
	return nil
}

func (s *MockSinkServer) SendTo(data *AppData) error {
	if s.isError {
		return errors.New("test error")
	}
	s.incomingData = data
	return nil
}

func TestSinkServer(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		sinkServer := MockSinkServer{
			isError: false,
		}

		_, ok := interface{}(&sinkServer).(SinkServer)
		if !ok {
			t.Errorf("Expected sinkServer to implement SinkServer interface")
		}

		_, pushableOk := interface{}(&sinkServer).(Pushable)
		if !pushableOk {
			t.Errorf("Expected sinkServer to implement Pushable interface")
		}

		err := sinkServer.Serve()
		if err != nil {
			t.Errorf("Expected no error from Serve, got %v", err)
		}

		appData := &AppData{
			data:    "test data",
			handler: &MockCompletionHandlerForPullable{},
		}
		err = sinkServer.SendTo(appData)
		if err != nil {
			t.Errorf("Expected no error from SendTo, got %v", err)
		}
		if sinkServer.incomingData != appData {
			t.Errorf("Expected data to be sent to SinkServer, got %v", sinkServer.incomingData)
		}
	})
	t.Run("Error", func(t *testing.T) {
		sinkServer := MockSinkServer{
			isError: true,
		}

		err := sinkServer.Serve()

		if err == nil {
			t.Errorf("Expected error from Serve, got '%v)", err)
		}

		appData := AppData{
			data:    "test data",
			handler: &MockCompletionHandlerForPullable{},
		}
		err = sinkServer.SendTo(&appData)
		if err == nil {
			t.Errorf("Expected error from SendTo, got '%v'", err)
		}
	})
}

// MockPipeServer is a mock implementation of the PipeServer interface
type MockPipeServer struct {
	isAddError    bool
	isServeError  bool
	Pushable	  Pushable
}

func (s *MockPipeServer) AddPushable(pushable Pushable) error {
	if s.isAddError {
		return errors.New("test error")
	}
	s.Pushable = pushable
	return nil
}

func (s *MockPipeServer) Serve() error {
	if s.isServeError {
		return errors.New("test error")
	}
	return nil
}

func (s *MockPipeServer) SendTo(data *AppData) error {
	return nil
}

func TestPipeServer(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		pipeServer := MockPipeServer{
			isAddError:   false,
			isServeError: false,
		}
		mockPushable := MockPushable{}
		err := pipeServer.AddPushable(&mockPushable)

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
			t.Errorf("Expected no error from AddPushable got %v", err)
		}
		if pipeServer.Pushable != &mockPushable {
			t.Errorf("Expected Pushable to be set, got %v", pipeServer.Pushable)
		}

		err = pipeServer.Serve()
		if err != nil {
			t.Errorf("Expected no error from Serve, got %v", err)
		}
	})
	t.Run("Error", func(t *testing.T) {
		pipeServer := MockPipeServer{
			isAddError:   true,
			isServeError: true,
		}

		err := pipeServer.AddPushable(&MockPushable{})
		if err == nil {
			t.Errorf("Expected error from AddPushable, got '%v'", err)
		}

		err = pipeServer.Serve()
		if err == nil {
			t.Errorf("Expected error from Serve, got '%v'", err)
		}
	})
}

// Test for ServersRun
func TestServersRun(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockSourceServer := MockSourceServer{
			isAddError: false,
		}
		mockPipeServer := MockPipeServer{
			isAddError:   false,
			isServeError: false,
		}
		mockSinkServer := MockSinkServer{
			isError: false,
		}

		err := ServersRun(&mockSourceServer, &mockPipeServer, &mockSinkServer)

		if err != nil {
			t.Errorf("Expected no error from ServersRun, got %v", err)
		}
	})
	t.Run("ErrorSourceAddPushable", func(t *testing.T) {
		mockSourceServer := MockSourceServer{
			isAddError: true,
		}
		mockPipeServer := MockPipeServer{
			isAddError:   false,
			isServeError: false,
		}
		mockSinkServer := MockSinkServer{
			isError: false,
		}

		err := ServersRun(&mockSourceServer, &mockPipeServer, &mockSinkServer)

		if err == nil {
			t.Errorf("Expected error from ServersRun, got nil")
		}
	})
	t.Run("ErrorPipeAddPushable", func(t *testing.T) {
		mockSourceServer := MockSourceServer{
			isAddError: false,
		}
		mockPipeServer := MockPipeServer{
			isAddError:   true,
			isServeError: false,
		}
		mockSinkServer := MockSinkServer{
			isError: false,
		}

		err := ServersRun(&mockSourceServer, &mockPipeServer, &mockSinkServer)

		if err == nil {
			t.Errorf("Expected error from ServersRun, got nil")
		}
	})
	t.Run("ErrorPipeServe", func(t *testing.T) {
		mockSourceServer := MockSourceServer{
			isAddError: false,
			isServeError: false,
		}
		mockPipeServer := MockPipeServer{
			isAddError:   false,
			isServeError: true,
		}
		mockSinkServer := MockSinkServer{
			isError: false,
		}

		err := ServersRun(&mockSourceServer, &mockPipeServer, &mockSinkServer)

		if err == nil {
			t.Errorf("Expected error from ServersRun, got nil")
		}
	})
	t.Run("ErrorSinkServe", func(t *testing.T) {
		mockSourceServer := MockSourceServer{
			isAddError: false,
			isServeError: false,
		}
		mockPipeServer := MockPipeServer{
			isAddError:   false,
			isServeError: false,
		}
		mockSinkServer := MockSinkServer{
			isError: true,
		}

		err := ServersRun(&mockSourceServer, &mockPipeServer, &mockSinkServer)

		if err == nil {
			t.Errorf("Expected error from ServersRun, got nil")
		}
	})
	t.Run("ErrorSourceServe", func(t *testing.T) {
		mockSourceServer := MockSourceServer{
			isAddError: false,
			isServeError: true,
		}
		mockPipeServer := MockPipeServer{
			isAddError:   false,
			isServeError: false,
		}
		mockSinkServer := MockSinkServer{
			isError: false,
		}

		err := ServersRun(&mockSourceServer, &mockPipeServer, &mockSinkServer)

		if err == nil {
			t.Errorf("Expected error from ServersRun, got nil")
		}
	})
	t.Run("ErrorServeAll", func(t *testing.T) {
		mockSourceServer := MockSourceServer{
			isAddError: false,
			isServeError: true,
		}
		mockPipeServer := MockPipeServer{
			isAddError:   false,
			isServeError: true,
		}
		mockSinkServer := MockSinkServer{
			isError: true,
		}

		err := ServersRun(&mockSourceServer, &mockPipeServer, &mockSinkServer)

		if err == nil {
			t.Errorf("Expected error from ServersRun, got nil")
		}
	})
}

// Test for MapSinkServer
func TestMapSinkServer(t *testing.T) {
	t.Run("Instantiation", func(t *testing.T) {
		sinkServerMap := make(map[string]SinkServer)
		sinkServerMap["test"] = &MockSinkServer{}
		mapSinkServer := MapSinkServer{
			sinkServerMap: sinkServerMap,
		}
		if mapSinkServer.sinkServerMap["test"] != sinkServerMap["test"] {
			t.Errorf("Expected sinkServerMap[\"test\"] to be %v, got %v", sinkServerMap["test"], mapSinkServer.sinkServerMap["test"])
		}
	})
}