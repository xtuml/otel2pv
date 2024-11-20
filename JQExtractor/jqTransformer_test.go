package jqextractor

import (
	"errors"
	"strings"
	"testing"

	"github.com/itchyny/gojq"

	"github.com/SmartDCSITlimited/CDS-OTel-To-PV/Server"
)

// TestJQTransformerConfig tests the IngestConfig method of the JQTransformerConfig struct
func TestJQTransformerConfig(t *testing.T) {
	t.Run("IngestConfig", func(t *testing.T) {
		jqtConfig := JQTransformerConfig{}
		// Test when the config is invalid
		err := jqtConfig.IngestConfig(map[string]interface{}{})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		// Test when the config is valid
		err = jqtConfig.IngestConfig(map[string]interface{}{"JQQueryStrings": map[string]string{"key": ".key"}})
		if err != nil {
			t.Errorf("Expected no error from IngestConfig, got %v", err)
		}
		if len(jqtConfig.JQQueryStrings) == 0 {
			t.Errorf("Expected JQQueryStrings to be populated, got empty")
		}
		// Tests when the JQQueryStrings map is empty
		err = jqtConfig.IngestConfig(map[string]interface{}{"JQQueryStrings": map[string]string{}})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
	})
}

// TestGetDataFromJQIterator tests the getDataFromJQIterator function
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
	incomingData  *Server.AppData
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

// NotJQTransformerConfig is a struct that does is not JQTransformerConfig struct
type NotJQTransformerConfig struct{}

func (s *NotJQTransformerConfig) IngestConfig(config map[string]any) error {
	return nil
}

// TestJQTransformer tests the JQTransformer struct and its methods
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
			jqProgram: jqProgram,
			pushable:  &Pushable,
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
		jqTransformer := JQTransformer{}
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
	t.Run("Setup", func(t *testing.T) {
		jqTransformer := JQTransformer{}
		// Test when config is not JQTransformerConfig
		err := jqTransformer.Setup(&NotJQTransformerConfig{})
		if err == nil {
			t.Errorf("Expected error from Setup, got nil")
		}
		// Test when config is JQTransformerConfig and JQQueryStrings is not set
		err = jqTransformer.Setup(&JQTransformerConfig{})
		if err == nil {
			t.Errorf("Expected error from Setup, got nil")
		}
		// Test when config is JQTransformerConfig and JQQueryStrings is empty map
		err = jqTransformer.Setup(&JQTransformerConfig{JQQueryStrings: map[string]string{}})
		if err == nil {
			t.Errorf("Expected error from Setup, got nil")
		}
		// Tests when the JQQueryStrings map is populated but there is a parse error
		err = jqTransformer.Setup(&JQTransformerConfig{JQQueryStrings: map[string]string{"key": `.key | %`}})
		if err == nil {
			t.Errorf("Expected error from Setup, got nil")
		}
		if !strings.Contains(err.Error(), "The following JQ string for key \"key\" failed to be parsed correctly") {
			t.Errorf("Expected error message to contain \"The following JQ string for key \"key\" failed to be parsed correctly\", got %v", err)
		}
		// Tests when the JQQueryStrings map is populated but there is a compile error
		err = jqTransformer.Setup(&JQTransformerConfig{JQQueryStrings: map[string]string{"key": `.key | $var`}})
		if err == nil {
			t.Errorf("Expected error from Setup, got nil")
		}
		if !strings.Contains(err.Error(), "The following JQ string for key \"key\" failed to be compiled correctly") {
			t.Errorf("Expected error message to contain \"The following JQ string for key \"key\" failed to be compiled correctly\", got %v", err)
		}
		// Test when config is JQTransformerConfig and JQQueryStrings is set
		err = jqTransformer.Setup(&JQTransformerConfig{JQQueryStrings: map[string]string{"key": ".key", "key2": ".key2"}})
		if err != nil {
			t.Errorf("Expected no error from Setup, got %v", err)
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
