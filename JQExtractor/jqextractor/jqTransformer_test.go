package jqextractor

import (
	"errors"
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/itchyny/gojq"

	"github.com/SmartDCSITlimited/CDS-OTel-To-PV/Server"
)

// TestJQTransformerConfig tests the IngestConfig method of the JQTransformerConfig struct
func TestJQTransformerConfig(t *testing.T) {
	t.Run("getJQStringFromInput", func(t *testing.T) {
		// Test when the input is not a map
		_, err := getJQStringFromInput(1)
		if err == nil {
			t.Errorf("Expected error from getJQStringFromInput, got nil")
		}
		if err.Error() != "invalid JQQueryStrings map in config. Must map a string identifier to a string" {
			t.Errorf("Expected error message to be \"invalid JQQueryStrings map in config. Must map a string identifier to a string\", got %v", err)
		}
		// Test when the input is a map but the jq field is not present
		_, err = getJQStringFromInput(map[string]any{"key": "value"})
		if err == nil {
			t.Errorf("Expected error from getJQStringFromInput, got nil")
		}
		if err.Error() != "invalid JQQueryStrings map in config. Must include field: \"jq\" and this field must be a string" {
			t.Errorf("Expected error message to be \"invalid JQQueryStrings map in config. Must include field: \"jq\" and this field must be a string\", got %v", err)
		}
		// Test when the input is a map and jq field is not a string
		_, err = getJQStringFromInput(map[string]any{"jq": 1})
		if err == nil {
			t.Errorf("Expected error from getJQStringFromInput, got nil")
		}
		if err.Error() != "invalid JQQueryStrings map in config. Must include field: \"jq\" and this field must be a string" {
			t.Errorf("Expected error message to be \"invalid JQQueryStrings map in config. Must include field: \"jq\" and this field must be a string\", got %v", err)
		}
		// Test when the input is a map and jq field is a string but type field is not correct type
		_, err = getJQStringFromInput(map[string]any{"jq": "value", "type": 1})
		if err == nil {
			t.Errorf("Expected error from getJQStringFromInput, got nil")
		}
		if err.Error() != "invalid JQQueryStrings map in config. If field: \"type\" is present it must be a string" {
			t.Errorf("Expected error message to be \"invalid JQQueryStrings map in config. If field: \"type\" is present it must be a string\", got %v", err)
		}
		// Test when the input is a map and jq field is a string but type field is not a valid type
		_, err = getJQStringFromInput(map[string]any{"jq": "value", "type": "invalid"})
		if err == nil {
			t.Errorf("Expected error from getJQStringFromInput, got nil")
		}
		if err.Error() != "invalid JQQueryStrings map in config. Invalid field \"type\": invalid" {
			t.Errorf("Expected error message to be \"invalid JQQueryStrings map in config. Invalid field \"type\": invalid\", got %v", err)
		}
		// Test when the input is a map and jq field is a string and type field is "file" but there is an error reading the file
		_, err = getJQStringFromInput(map[string]any{"jq": "value", "type": "file"})
		if err == nil {
			t.Errorf("Expected error from getJQStringFromInput, got nil")
		}
		if err.Error() != "invalid JQQueryStrings map in config. open value: no such file or directory" {
			t.Errorf("Expected error message to be \"invalid JQQueryStrings map in config. open value: no such file or directory\", got %v", err)
		}
		// Test when the input is a map and jq field is a string and type field is "file" and the file is read correctly
		tmpFile, err := os.CreateTemp("", "testfile")
		if err != nil {
			t.Errorf("Error creating temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())
		data := []byte("value")
		err = os.WriteFile(tmpFile.Name(), data, 0644)
		if err != nil {
			t.Errorf("Error writing to temp file: %v", err)
		}
		jqString, err := getJQStringFromInput(map[string]any{"jq": tmpFile.Name(), "type": "file"})
		if err != nil {
			t.Errorf("Expected no error from getJQStringFromInput, got %v", err)
		}
		if jqString != "value" {
			t.Errorf("Expected jqString to be \"value\", got %v", jqString)
		}
		// Test when the input is a map and jq field is a string and type field is not present
		jqString, err = getJQStringFromInput(map[string]any{"jq": "value"})
		if err != nil {
			t.Errorf("Expected no error from getJQStringFromInput, got %v", err)
		}
		if jqString != "value" {
			t.Errorf("Expected jqString to be \"value\", got %v", jqString)
		}
	})
	t.Run("IngestConfig", func(t *testing.T) {
		jqtConfig := JQTransformerConfig{}
		// Test when the config is invalid
		err := jqtConfig.IngestConfig(map[string]interface{}{})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		// Test when there is an error from getJQStringFromInput
		err = jqtConfig.IngestConfig(map[string]interface{}{"JQQueryStrings": map[string]any{"key": map[string]any{"jq": ".key", "type": 1}}})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "invalid JQQueryStrings map in config. If field: \"type\" is present it must be a string. Key: key" {
			t.Errorf("Expected error message to be \"invalid JQQueryStrings map in config. If field: \"type\" is present it must be a string. Key: key\", got %v", err)
		}
		// Test when the config is valid
		err = jqtConfig.IngestConfig(map[string]interface{}{"JQQueryStrings": map[string]any{"key": map[string]any{"jq": ".key"}}})
		if err != nil {
			t.Errorf("Expected no error from IngestConfig, got %v", err)
		}
		if len(jqtConfig.JQQueryStrings) == 0 {
			t.Errorf("Expected JQQueryStrings to be populated, got empty")
		}
		// Tests when the JQQueryStrings map is empty
		err = jqtConfig.IngestConfig(map[string]interface{}{"JQQueryStrings": map[string]any{}})
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
		// Test when there is data but it is not a map of string to array of any
		iter = gojq.NewIter(1)
		_, err = getDataFromJQIterator(&iter)
		if err == nil {
			t.Errorf("Expected error from getDataFromJQIterator, got %v", err)
		}
		if err.Error() != "data is not a map of strings to arrays" {
			t.Errorf("Expected error message to be \"data is not a map of strings to arrays\", got %v", err)
		}
		// Test when there is more than one data
		iter = gojq.NewIter(1, 2)
		_, err = getDataFromJQIterator(&iter)
		if err == nil {
			t.Errorf("Expected error from getDataFromJQIterator, got nil")
		}
		if err.Error() != "more than one data" {
			t.Errorf("Expected error message to be \"more than one data\", got %v", err)
		}
		// Test when there is one data and it is a map of string to array of any
		iter = gojq.NewIter(map[string][]any{"key": {1}})
		data, err = getDataFromJQIterator(&iter)
		if err != nil {
			t.Errorf("Expected no error from getDataFromJQIterator, got %v", err)
		}
		if data["key"][0] != 1 {
			t.Errorf("Expected data to be 1, got %v", data["key"][0])
		}
	})
}

// MockPushable is a mock implementation of the Pushable interface
type MockPushable struct {
	isSendToError bool
	incomingData  chan(*Server.AppData)
}

func (p *MockPushable) SendTo(data *Server.AppData) error {
	if p.isSendToError {
		return errors.New("test error")
	}
	p.incomingData <- data
	return nil
}

type MockCompletionHandler struct {
	// MockCompletionHandler is a struct that provides a mock implementation of the CompletionHandler interface
}

func (mch *MockCompletionHandler) Complete(data interface{}, err error) error {
	return nil
}

// NotJQTransformerConfig is a struct that is not JQTransformerConfig struct
type NotJQTransformerConfig struct{}

func (s *NotJQTransformerConfig) IngestConfig(config map[string]any) error {
	return nil
}

// TestJQTransformer tests the JQTransformer struct and its methods
func TestJQTransformer(t *testing.T) {
	jqQuery, err := gojq.Parse(". |\n\n .key")
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
		// Tests when jqProgram and pushable are set but GetData returns an error
		jqTransformer.pushable = &Pushable
		err = jqTransformer.SendTo(&Server.AppData{})
		if err == nil {
			t.Fatalf("Expected error from SendTo, got nil")
		}
		if err.Error() != "data is not set" {
			t.Errorf("Expected error message to be \"data is not set\", got %v", err)
		}
		// Test when jqProgram and pushable are set but the jqProgram does not return a map of string to an array
		jqTransformer.pushable = &Pushable
		appData := Server.NewAppData(map[string]interface{}{"key": 1}, "")
		err = jqTransformer.SendTo(appData)
		if err == nil {
			t.Errorf("Expected error from SendTo, got nil")
		}
		// Test when the pushable SendTo method returns an error
		Pushable.isSendToError = true
		// this provides a test that the gojq parser can handle comments and new lines 
		// i.e. such as in a file
		newjqQuery, err := gojq.Parse("{\"X\": [(.|###\n .key)]}")
		if err != nil {
			t.Errorf("Error parsing JQ program: %v", err)
		}
		newjqProgram, err := gojq.Compile(newjqQuery)
		if err != nil {
			t.Errorf("Error compiling JQ program: %v", err)
		}
		jqTransformer.jqProgram = newjqProgram
		err = jqTransformer.SendTo(appData)
		if err == nil {
			t.Errorf("Expected error from SendTo, got nil")
		}
		if err.Error() != "test error" {
			t.Errorf("Expected error message to be \"test error\", got %v", err)
		}
		// Test when jqProgram and pushable are set and the jqProgram returns a map of string to an array
		Pushable.isSendToError = false
		incomingChannel := make(chan(*Server.AppData), 1)
		Pushable.incomingData = incomingChannel
		err = jqTransformer.SendTo(appData)
		if err != nil {
			t.Errorf("Expected no error from SendTo, got %v", err)
		}
		data := <- Pushable.incomingData
		close(incomingChannel)
		gotData, err := data.GetData()
		if err != nil {
			t.Errorf("Expected no error from GetData, got %v", err)
		}
		if gotData != 1 {
			t.Errorf("Expected data to be 1, got %v", gotData)
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

// MockConfig is a mock implementation of the Config interface
type MockConfig struct {}

func (mc *MockConfig) IngestConfig(config map[string]any) error {
	return nil
}

// MockSourceServer is a mock implementation of the SourceServer interface
type MockSourceServer struct {
	dataToSend []any
	pushable  Server.Pushable
}

// AddPushable is a method that adds a pushable to the SourceServer
func (mss *MockSourceServer) AddPushable(pushable Server.Pushable) error {
	mss.pushable = pushable
	return nil
}

// Serve is a method that serves the SourceServer
func (mss *MockSourceServer) Serve() error {
	for _, data := range mss.dataToSend {
		err := mss.pushable.SendTo(Server.NewAppData(data, ""))
		if err != nil {
			return err
		}
	}
	return nil
}

// Setup is a method that sets up the SourceServer
func (mss *MockSourceServer) Setup(config Server.Config) error {
	return nil
}

// MockSinkServer is a mock implementation of the SinkServer interface
type MockSinkServer struct {
	dataReceived chan(any)
}

// SendTo is a method that sends data to the SinkServer
func (mss *MockSinkServer) SendTo(data *Server.AppData) error {
	if data == nil {
		return errors.New("data is nil")
	}
	gotData, err := data.GetData()
	if err != nil {
		return err
	}
	mss.dataReceived <- gotData
	return nil
}

// Serve is a method that serves the SinkServer
func (mss *MockSinkServer) Serve() error {
	return nil
}

// Setup is a method that sets up the SinkServer
func (mss *MockSinkServer) Setup(config Server.Config) error {
	return nil
}



// TestJQTransformer integrating with RunApp
func TestJQTransformerRunApp(t *testing.T) {
	// Setup
	dataToSend := []any{}
	for i := 0; i < 10; i++ {
		dataToSend = append(dataToSend, map[string]interface{}{"key": i})
	}
	mockSourceServer := &MockSourceServer{
		dataToSend: dataToSend,
	}
	chanForData := make(chan(any), 10)
	mockSinkServer := &MockSinkServer{
		dataReceived: chanForData,
	}
	jqTransformer := &JQTransformer{}
	jqTransformerConfig := &JQTransformerConfig{}
	producerConfigMap := map[string]func() Server.Config{
		"MockSink": func() Server.Config {
			return &MockConfig{}
		},
	}
	consumerConfigMap := map[string]func() Server.Config{
		"MockSource": func() Server.Config {
			return &MockConfig{}
		},
	}
	producerMap := map[string]func() Server.SinkServer{
		"MockSink": func() Server.SinkServer {
			return mockSinkServer
		},
	}
	consumerMap := map[string]func() Server.SourceServer{
		"MockSource": func() Server.SourceServer {
			return mockSourceServer
		},
	}
	// set config file
	tmpFile, err := os.CreateTemp("", "config.json")
	if err != nil {
		t.Errorf("Error creating temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	data := []byte(
		`{"AppConfig":{"JQQueryStrings":{"outData":{"jq":".key"}}},"ProducersSetup":{"ProducerConfigs":[{"Type":"MockSink","ProducerConfig":{}}]},"ConsumersSetup":{"ConsumerConfigs":[{"Type":"MockSource","ConsumerConfig":{}}]}}`,
	)
	err = os.WriteFile(tmpFile.Name(), data, 0644)
	if err != nil {
		t.Errorf("Error writing to temp file: %v", err)
	}
	err = flag.Set("config", tmpFile.Name())
	if err != nil {
		t.Errorf("Error setting flag: %v", err)
	}
	// Run the app
	err = Server.RunApp(
		jqTransformer, jqTransformerConfig,
		producerConfigMap, consumerConfigMap,
		producerMap, consumerMap,
	)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	close(chanForData)
	// Check if the data was transformed correctly
	counter := 0
	for data := range mockSinkServer.dataReceived {
		if data != counter {
			t.Errorf("Expected data to be %d, got %v", counter, data)
		}
		counter++
	}
	if counter != 10 {
		t.Errorf("Expected 10 data points, got %d", counter)
	}

}