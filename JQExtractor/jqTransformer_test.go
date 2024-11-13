package jqextractor

import (
	"testing"

	"github.com/itchyny/gojq"

	"github.com/SmartDCSITlimited/CDS-OTel-To-PV/Server"
)

func TestGetDataFromJQIterator(t *testing.T) {
	t.Run("getDataFromJQIterator", func(t *testing.T) {
		// Test when there is no data
		iter := gojq.NewIter()
		data, err := getDataFromJQIterator(&iter)
		if data != nil {
			t.Errorf("Expected data to be nil, got %v", data)
		}
		if err == nil {
			t.Errorf("Expected error from getDataFromJQIterator, got nil")
		}
		// Test when there is data
		iter = gojq.NewIter(1)
		data, err = getDataFromJQIterator(&iter)
		if data != 1 {
			t.Errorf("Expected data to be 1, got %v", data)
		}
		if err != nil {
			t.Errorf("Expected no error from getDataFromJQIterator, got %v", err)
		}
		// Test when there is more than one data
		iter = gojq.NewIter(1, 2)
		data, err = getDataFromJQIterator(&iter)
		if data != nil {
			t.Errorf("Expected data to be nil, got %v", data)
		}
		if err == nil {
			t.Errorf("Expected error from getDataFromJQIterator, got nil")
		}
	})
}

type MockReceiver struct {
	// MockReceiver is a struct that provides a mock implementation of the Receiver interface
	mychan chan *Server.AppData
}

func (mr *MockReceiver) SendTo(data *Server.AppData) error {
	mr.mychan <- data
	return nil
}

func (mr *MockReceiver) GetOutChan() (<-chan *Server.AppData, error) {
	return mr.mychan, nil
}

type MockCompletionHandler struct {
	// MockCompletionHandler is a struct that provides a mock implementation of the CompletionHandler interface
}

func (mch *MockCompletionHandler) Complete(data interface{}, err error) error {
	return nil
}

func TestJQTransformer(t *testing.T) {
	jqQuery, err := gojq.Parse(".key")
	if err != nil {
		t.Errorf("Error parsing JQ program: %v", err)
	}
	jqProgram, err := gojq.Compile(jqQuery)
	if err != nil {
		t.Errorf("Error compiling JQ program: %v", err)
	}
	inChan := make(chan *Server.AppData, 1)
	outChan := make(chan *Server.AppData, 1)
	inReceiver := MockReceiver{
		mychan: inChan,
	}
	outReceiver := MockReceiver{
		mychan: outChan,
	}
	t.Run("Instantiation", func(t *testing.T) {
		jqTransformer := JQTransformer{
			jqProgram:   jqProgram,
			inReceiver:  &inReceiver,
			outReceiver: &outReceiver,
		}
		if jqTransformer.jqProgram != jqProgram {
			t.Errorf("Expected jqProgram to be %v, got %v", jqProgram, jqTransformer.jqProgram)
		}
		if jqTransformer.inReceiver != &inReceiver {
			t.Errorf("Expected inReceiver to be %v, got %v", &inReceiver, jqTransformer.inReceiver)
		}
		if jqTransformer.outReceiver != &outReceiver {
			t.Errorf("Expected outReceiver to be %v, got %v", &outReceiver, jqTransformer.outReceiver)
		}
	})
	t.Run("GetReceiver", func(t *testing.T) {
		jqTransformer := JQTransformer{jqProgram: jqProgram}
		// Test when inReceiver is nil
		_, err := jqTransformer.GetReceiver()
		if err == nil {
			t.Errorf("Expected error from GetReceiver, got nil")
		}
		// Test when inReceiver is set
		jqTransformer.inReceiver = &inReceiver
		receiver, err := jqTransformer.GetReceiver()
		if err != nil {
			t.Errorf("Expected no error from GetReceiver, got %v", err)
		}
		if receiver != &inReceiver {
			t.Errorf("Expected receiver to be %v, got %v", &inReceiver, receiver)
		}
	})
	t.Run("AddReceiver", func(t *testing.T) {
		jqTransformer := JQTransformer{jqProgram: jqProgram}
		// Test when outReceiver is nil
		err := jqTransformer.AddReceiver(&outReceiver)
		if err != nil {
			t.Errorf("Expected no error from AddReceiver, got %v", err)
		}
		if jqTransformer.outReceiver != &outReceiver {
			t.Errorf("Expected outReceiver to be %v, got %v", &outReceiver, jqTransformer.outReceiver)
		}
		// Test when outReceiver is set
		err = jqTransformer.AddReceiver(&outReceiver)
		if err == nil {
			t.Errorf("Expected error from AddReceiver, got nil")
		}
	})
	t.Run("HandleIncomingData", func(t *testing.T) {
		jqTransformer := JQTransformer{
			jqProgram: jqProgram,
		}
		// Test when jqProgram is nil
		err := jqTransformer.HandleIncomingData(&Server.AppData{})
		if err == nil {
			t.Errorf("Expected error from HandleIncomingData, got nil")
		}
		// Test when outReceiver is nil
		jqTransformer.jqProgram = jqProgram
		err = jqTransformer.HandleIncomingData(&Server.AppData{})
		if err == nil {
			t.Errorf("Expected error from HandleIncomingData, got nil")
		}
		// Test when jqProgram and outReceiver are set
		jqTransformer.outReceiver = &outReceiver
		completionHandler := &MockCompletionHandler{}
		appData := Server.NewAppData(map[string]interface{}{"key": 1}, completionHandler)
		err = jqTransformer.HandleIncomingData(appData)
		if err != nil {
			t.Errorf("Expected no error from HandleIncomingData, got %v", err)
		}
		data := <-outChan
		if data.GetData() != 1 {
			t.Errorf("Expected data to be 1, got %v", data.GetData())
		}
		dataHandler, err := data.GetHandler()
		if err != nil {
			t.Errorf("Expected no error from GetHandler, got %v", err)
		}

		if dataHandler != completionHandler {
			t.Errorf("Expected handler to be %v, got %v", completionHandler, dataHandler)
		}

		close(inChan)
		close(outChan)
	})
	t.Run("Serve", func(t *testing.T) {
		jqTransformer := JQTransformer{}
		// Test when inReceiver is nil
		err := jqTransformer.Serve()
		if err == nil {
			t.Errorf("Expected error from Serve, got nil")
		}
		// Test when outReceiver is nil
		jqTransformer.inReceiver = &inReceiver
		err = jqTransformer.Serve()
		if err == nil {
			t.Errorf("Expected error from Serve, got nil")
		}
		// Test when jqProgram is nil
		jqTransformer.outReceiver = &outReceiver
		err = jqTransformer.Serve()
		if err == nil {
			t.Errorf("Expected error from Serve, got %v", err)
		}
		// Test when all fields are set
		jqTransformer.jqProgram = jqProgram
		err = jqTransformer.Serve()
		if err != nil {
			t.Errorf("Expected no error from Serve, got %v", err)
		}
	})
	t.Run("ImplementsPipeServer", func(t *testing.T) {
		jqTransformer := JQTransformer{}
		// Type assertion to check if JQTransformer implements PipeServer
		_, ok := interface{}(&jqTransformer).(Server.PipeServer)
		if !ok {
			t.Errorf("Expected JQTransformer to implement PipeServer interface")
		}
	})
}
