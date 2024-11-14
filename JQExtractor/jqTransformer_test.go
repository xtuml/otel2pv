package jqextractor

import (
	"testing"
	"errors"

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

// MockPushable is a mock implementation of the Pushable interface
type MockPushable struct {
	isSendToError bool
	incomingData          *Server.AppData
}

func (p *MockPushable) SendTo(data *Server.AppData) error {
	if p.isSendToError {
		return errors.New("test error")
	}
	p.incomingData = data
	return nil
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
	Pushable := MockPushable{
		isSendToError: false,
	}
	t.Run("Instantiation", func(t *testing.T) {
		jqTransformer := JQTransformer{
			jqProgram:   jqProgram,
			pushable:   &Pushable,
		}
		if jqTransformer.jqProgram != jqProgram {
			t.Errorf("Expected jqProgram to be %v, got %v", jqProgram, jqTransformer.jqProgram)
		}
		if jqTransformer.pushable != &Pushable {
			t.Errorf("Expected pushable to be %v, got %v", &Pushable, jqTransformer.pushable)
		}
	})
	t.Run("AddPushable", func(t *testing.T) {
		jqTransformer := JQTransformer{jqProgram: jqProgram}
		// Test when pushable is nil
		err := jqTransformer.AddPushable(&Pushable)
		if err != nil {
			t.Errorf("Expected no error from AddPushable, got %v", err)
		}
		if jqTransformer.pushable != &Pushable {
			t.Errorf("Expected pushable to be %v, got %v", &Pushable, jqTransformer.pushable)
		}
		// Test when pushable is set
		err = jqTransformer.AddPushable(&Pushable)
		if err == nil {
			t.Errorf("Expected error from AddPushable, got nil")
		}
	})
	t.Run("SendTo", func(t *testing.T) {
		jqTransformer := JQTransformer{
		}
		// Test when jqProgram is nil
		err := jqTransformer.SendTo(&Server.AppData{})
		if err == nil {
			t.Errorf("Expected error from SendTo, got nil")
		}
		// Test when pushable is nil
		jqTransformer.jqProgram = jqProgram
		err = jqTransformer.SendTo(&Server.AppData{})
		if err == nil {
			t.Errorf("Expected error from SendTo, got nil")
		}
		// Test when jqProgram and pushable are set
		jqTransformer.pushable = &Pushable
		completionHandler := &MockCompletionHandler{}
		appData := Server.NewAppData(map[string]interface{}{"key": 1}, completionHandler)
		err = jqTransformer.SendTo(appData)
		if err != nil {
			t.Errorf("Expected no error from SendTo, got %v", err)
		}
		data := Pushable.incomingData
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
	})
	t.Run("Serve", func(t *testing.T) {
		jqTransformer := JQTransformer{}
		// Test when pushable is nil
		err := jqTransformer.Serve()
		if err == nil {
			t.Errorf("Expected error from Serve, got nil")
		}
		// Test when jqProgram is nil
		jqTransformer.pushable = &Pushable
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
