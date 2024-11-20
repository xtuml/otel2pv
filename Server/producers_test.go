package Server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

// MockProducer is a mock implementation of the Producer interface
type MockProducer struct {
	isSetupError  bool
	isSendToError bool
	isServeError  bool
	AppData       *AppData
}

func (p *MockProducer) Setup(Config) error {
	if p.isSetupError {
		return errors.New("test error")
	}
	return nil
}
func (p *MockProducer) SendTo(data *AppData) error {
	if p.isSendToError {
		return errors.New("test error")
	}
	p.AppData = data
	return nil
}
func (p *MockProducer) Serve() error {
	if p.isServeError {
		return errors.New("test error")
	}
	return nil
}

func TestProducer(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		producer := MockProducer{
			isSetupError:  false,
			isSendToError: false,
			isServeError:  false,
		}

		_, ok := interface{}(&producer).(Producer)
		if !ok {
			t.Errorf("Expected producer to implement Producer interface")
		}
		_, sinkServerOk := interface{}(&producer).(SinkServer)
		if !sinkServerOk {
			t.Errorf("Expected producer to implement SinkServer interface")
		}

		err := producer.Setup(&MockConfig{})
		if err != nil {
			t.Errorf("Expected no error from Setup, got %v", err)
		}

		appData := &AppData{
			data:    "test data",
			handler: &MockCompletionHandler{},
		}
		err = producer.SendTo(appData)
		if err != nil {
			t.Errorf("Expected no error from Send, got %v", err)
		}
		if producer.AppData != appData {
			t.Errorf("Expected producer.AppData to be equal to appData")
		}

		err = producer.Serve()
		if err != nil {
			t.Errorf("Expected no error from Serve, got %v", err)
		}
	})
	t.Run("Error", func(t *testing.T) {
		producer := MockProducer{
			isSetupError:  true,
			isSendToError: true,
			isServeError:  true,
		}

		err := producer.Setup(&MockConfig{})
		if err == nil {
			t.Errorf("Expected error from Setup, got '%v'", err)
		}

		err = producer.SendTo(&AppData{})
		if err == nil {
			t.Errorf("Expected error from Send, got '%v'", err)
		}

		err = producer.Serve()
		if err == nil {
			t.Errorf("Expected error from Serve, got '%v'", err)
		}
	})
}

// Tests for HTTPConfig
func TestHTTPConfig(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		config := &HTTPProducerConfig{}
		// With default values
		err := config.IngestConfig(map[string]any{
			"URL": "http://test.com",
		})
		if err != nil {
			t.Errorf("Expected no error from IngestConfig, got %v", err)
		}
		if config.URL != "http://test.com" {
			t.Errorf("Expected URL to be 'http://test.com', got '%v'", config.URL)
		}
		if config.numRetries != 3 {
			t.Errorf("Expected numRetries to be 3, got %v", config.numRetries)
		}
		if config.timeout != 10 {
			t.Errorf("Expected timeout to be 10, got %v", config.timeout)
		}
		// With custom values
		err = config.IngestConfig(map[string]any{
			"URL":        "http://test.com",
			"numRetries": 5,
			"timeout":    20,
		})
		if err != nil {
			t.Errorf("Expected no error from IngestConfig, got %v", err)
		}
		if config.URL != "http://test.com" {
			t.Errorf("Expected URL to be 'http://test.com', got '%v'", config.URL)
		}
		if config.numRetries != 5 {
			t.Errorf("Expected numRetries to be 5, got %v", config.numRetries)
		}
		if config.timeout != 20 {
			t.Errorf("Expected timeout to be 20, got %v", config.timeout)
		}
	})
	t.Run("Error", func(t *testing.T) {
		config := &HTTPProducerConfig{}

		err := config.IngestConfig(map[string]any{
			"URL":        "http://test.com",
			"numRetries": "3",
		})
		if err.Error() != "invalid numRetries - must be an integer" {
			t.Errorf("Expected specified error from IngestConfig, got '%v'", err)
		}

		err = config.IngestConfig(map[string]any{
			"URL":     "http://test.com",
			"timeout": "10",
		})
		if err.Error() != "invalid timeout - must be an integer" {
			t.Errorf("Expected specified error from IngestConfig, got '%v'", err)
		}
	})
}

// Tests for HTTPProducer
func TestHTTPProducer(t *testing.T) {
	t.Run("ImplementsProducer", func(t *testing.T) {
		producer := &HTTPProducer{}
		_, ok := interface{}(producer).(Producer)
		if !ok {
			t.Errorf("Expected producer to implement Producer interface")
		}
	})
	t.Run("Setup", func(t *testing.T) {
		producer := &HTTPProducer{}
		config := &HTTPProducerConfig{
			URL:     "http://test.com",
			timeout: 10,
		}

		err := producer.Setup(config)
		if err != nil {
			t.Errorf("Expected no error from Setup, got %v", err)
		}
		if producer.config != config {
			t.Errorf("Expected producer.config to be equal to config")
		}
		if producer.client.Timeout != time.Duration(10)*time.Second {
			t.Errorf("Expected client.Timeout to be 10, got %v", producer.client.Timeout)
		}
		// Try with invalid config
		err = producer.Setup(&MockConfig{})
		if err.Error() != "invalid config" {
			t.Errorf("Expected specified error from Setup, got '%v'", err)
		}
	})
	t.Run("Serve", func(t *testing.T) {
		producer := &HTTPProducer{}
		err := producer.Serve()
		if err != nil {
			t.Errorf("Expected no error from Serve, got %v", err)
		}
	})
	t.Run("SendTo", func(t *testing.T) {
		sentData := `{"test":"data"}`
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/test" {
				http.Error(w, "invalid path", http.StatusNotFound)
				return
			}
			if r.Header.Get("Content-Type") != "application/json" {
				http.Error(w, "invalid content type", http.StatusUnsupportedMediaType)
				return
			}
			if r.Body == nil {
				http.Error(w, "no body", http.StatusBadRequest)
				return
			}
			defer r.Body.Close()
			buf := make([]byte, 1024)
			n, err := r.Body.Read(buf)
			if err.Error() != "EOF" {
				http.Error(w, "error reading body", http.StatusInternalServerError)
				return
			}
			if string(buf[:n]) != sentData {
				t.Errorf("Expected body to be '%v', got '%v'", sentData, string(buf[:n]))
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()
		producer := &HTTPProducer{
			config: &HTTPProducerConfig{
				URL:        server.URL + "/test",
				numRetries: 1,
				timeout:    1,
			},
			client: &http.Client{},
		}
		appData := &AppData{
			data:    map[string]any{"test": "data"},
			handler: &MockCompletionHandler{},
		}

		err := producer.SendTo(appData)
		if err != nil {
			t.Errorf("Expected no error from SendTo, got %v", err)
		}
		dataComplete, ok := appData.handler.(*MockCompletionHandler).DataReceived.(map[string]any)
		if !ok {
			t.Errorf("Expected data to be a map, got %v", appData.handler.(*MockCompletionHandler).DataReceived)
		}
		expectedDataComplete := map[string]any{"test": "data"}
		if !reflect.DeepEqual(dataComplete, expectedDataComplete) {
			t.Errorf("Expected data to be '%v', got '%v'", expectedDataComplete, dataComplete)
		}
		// Error case in which data is not a map[string]any
		appData = &AppData{
			data:    "test",
			handler: &MockCompletionHandler{},
		}
		err = producer.SendTo(appData)
		if err.Error() != "invalid data" {
			t.Errorf("Expected specified error from SendTo, got '%v'", err)
		}
		// Error case in which appData.GetHandler() returns an error
		appData = &AppData{
			data: map[string]any{"test": "data"},
		}
		err = producer.SendTo(appData)
		if err.Error() != "handler not set" {
			t.Errorf("Expected specified error from SendTo, got '%v'", err)
		}
		// Error case in which url is not found (2 retries to test)
		producer.config.URL = "invalid"
		producer.config.numRetries = 2
		appData.handler = &MockCompletionHandler{}
		err = producer.SendTo(appData)
		if err.Error() != "failed to send data" {
			t.Errorf("Expected specified error from SendTo, got '%v'", err)
		}
		// Error case in which http.NewRequest() returns an error
		producer.config.URL = server.URL + "/%%"
		err = producer.SendTo(appData)
		if err.Error() != "parse "+`"`+server.URL+`/%%": invalid URL escape "%%"` {
			t.Errorf("Expected specified error from SendTo, got full '%v'", err)
		}
	})
	t.Run("SendToPanic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic from SendTo, got nil")
			}
		}()
		producer := &HTTPProducer{
			config: &HTTPProducerConfig{
				URL:        "/test",
				numRetries: 0,
				timeout:    0,
			},
			client: &http.Client{},
		}
		appData := &AppData{
			data: map[string]any{"test": "data"},
			handler: &MockCompletionHandler{
				isError: true,
			},
		}
		_ = producer.SendTo(appData)
	})
}
