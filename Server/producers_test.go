package Server

import (
	"errors"
	"testing"
)

// MockProducerConfig is a mock implementation of the ProducerConfig interface
type MockProducerConfig struct {
	isError bool
}

func (c *MockProducerConfig) IngestConfig(map[string]any) error {
	if c.isError {
		return errors.New("test error")
	}
	return nil
}

func TestProducerConfig(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		config := MockProducerConfig{
			isError: false,
		}

		_, ok := interface{}(&config).(ProducerConfig)
		if !ok {
			t.Errorf("Expected config to implement ProducerConfig interface")
		}

		err := config.IngestConfig(map[string]any{})
		if err != nil {
			t.Errorf("Expected no error from IngestConfig, got %v", err)
		}
	})
	t.Run("Error", func(t *testing.T) {
		config := MockProducerConfig{
			isError: true,
		}

		err := config.IngestConfig(map[string]any{})

		if err == nil {
			t.Errorf("Expected error from IngestConfig, got '%v'", err)
		}
	})
}

// MockProducer is a mock implementation of the Producer interface
type MockProducer struct {
	isSetupError  bool
	isSendToError bool
	isServeError  bool
	AppData       *AppData
}

func (p *MockProducer) Setup(ProducerConfig) error {
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

		err := producer.Setup(&MockProducerConfig{})
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

		err := producer.Setup(&MockProducerConfig{})
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

// Tests for HTTPProducerConfig
func TestHTTPProducerConfig(t *testing.T) {
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
