package Server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	rabbitmq "github.com/rabbitmq/amqp091-go"
)

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
		_, ok := interface{}(producer).(SinkServer)
		if !ok {
			t.Errorf("Expected producer to implement SinkServer interface")
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

// Tests for SelectProducerConfig
func TestSelectProducerConfig(t *testing.T) {
	t.Run("ImplementsConfig", func(t *testing.T) {
		config := &SelectProducerConfig{}
		_, ok := interface{}(config).(Config)
		if !ok {
			t.Errorf("Expected config to implement Config interface")
		}
	})
	t.Run("IngestConfigHTTP", func(t *testing.T) {
		config := &SelectProducerConfig{}
		// test ingest with valid config and default for
		// Map
		err := config.IngestConfig(map[string]any{
			"Type": "HTTP",
			"ProducerConfig": map[string]any{
				"URL": "http://test.com",
			},
		})
		if err != nil {
			t.Errorf("Expected no error from IngestConfig, got %v", err)
		}
		if config.Type != "HTTP" {
			t.Errorf("Expected Type to be 'HTTP', got '%v'", config.Type)
		}
		if config.Map != "" {
			t.Errorf("Expected Map to be '', got '%v'", config.Map)
		}
		if config.ProducerConfig == nil {
			t.Errorf("Expected ProducerConfig to be populated, got nil")
		}
		_, ok := config.ProducerConfig.(*HTTPProducerConfig)
		if !ok {
			t.Errorf("Expected ProducerConfig to be of type HTTPProducerConfig")
		}
		// test ingest with valid config and value set for map
		err = config.IngestConfig(map[string]any{
			"Type": "HTTP",
			"Map":  "test",
			"ProducerConfig": map[string]any{
				"URL": "http://test.com",
			},
		})
		if err != nil {
			t.Errorf("Expected no error from IngestConfig, got %v", err)
		}
		if config.Type != "HTTP" {
			t.Errorf("Expected Type to be 'HTTP', got '%v'", config.Type)
		}
		if config.Map != "test" {
			t.Errorf("Expected Map to be 'test', got '%v'", config.Map)
		}
		if config.ProducerConfig == nil {
			t.Errorf("Expected ProducerConfig to be populated, got nil")
		}
		_, ok = config.ProducerConfig.(*HTTPProducerConfig)
		if !ok {
			t.Errorf("Expected ProducerConfig to be of type HTTPProducerConfig")
		}
		// tests case where ProducerConfig.IngestConfig returns an error
		err = config.IngestConfig(map[string]any{
			"Type": "HTTP",
			"ProducerConfig": map[string]any{
				"URL": 1,
			},
		})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "invalid URL - must be a string and must be set" {
			t.Errorf("Expected specified error from IngestConfig, got '%v'", err)
		}
	})
	t.Run("IngestConfigInvalid", func(t *testing.T) {
		config := &SelectProducerConfig{}
		// tests case where Type is not set
		err := config.IngestConfig(map[string]any{
			"ProducerConfig": map[string]any{
				"URL": "http://test.com",
			},
		})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "invalid producer type" {
			t.Errorf("Expected specified error from IngestConfig, got '%v'", err)
		}
		// tests case where Type is not a string
		err = config.IngestConfig(map[string]any{
			"Type": 1,
			"ProducerConfig": map[string]any{
				"URL": "http://test.com",
			},
		})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "invalid producer type" {
			t.Errorf("Expected specified error from IngestConfig, got '%v'", err)
		}
		// tests case where type is not in PRODUCERCONFIGMAP
		err = config.IngestConfig(map[string]any{
			"Type": "test",
			"ProducerConfig": map[string]any{
				"URL": "http://test.com",
			},
		})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "invalid producer type: test" {
			t.Errorf("Expected specified error from IngestConfig, got '%v'", err)
		}
		// tests case where ProducerConfig is not set
		err = config.IngestConfig(map[string]any{
			"Type": "HTTP",
		})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "Producer config not set correctly" {
			t.Errorf("Expected specified error from IngestConfig, got '%v'", err)
		}
		// tests case where ProducerConfig is not a map[string]any
		err = config.IngestConfig(map[string]any{
			"Type":           "HTTP",
			"ProducerConfig": "test",
		})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "Producer config not set correctly" {
			t.Errorf("Expected specified error from IngestConfig, got '%v'", err)
		}
	})
}

// Tests for SetupProducersConfig
func TestSetupProducersConfig(t *testing.T) {
	t.Run("ImplementsConfig", func(t *testing.T) {
		config := &SetupProducersConfig{}
		_, ok := interface{}(config).(Config)
		if !ok {
			t.Errorf("Expected config to implement Config interface")
		}
	})
	t.Run("IngestConfig", func(t *testing.T) {
		config := &SetupProducersConfig{}
		// test ingest with valid config but default for IsMapping
		err := config.IngestConfig(map[string]any{
			"ProducerConfigs": []map[string]any{
				{
					"Type": "HTTP",
					"ProducerConfig": map[string]any{
						"URL": "http://test.com",
					},
				},
				{
					"Type": "HTTP",
					"ProducerConfig": map[string]any{
						"URL": "http://test2.com",
					},
				},
			},
		})
		if err != nil {
			t.Errorf("Expected no error from IngestConfig, got %v", err)
		}
		if len(config.SelectProducerConfigs) != 2 {
			t.Errorf("Expected SelectProducerConfigs to have 2 elements, got %v", len(config.SelectProducerConfigs))
		}
		if config.IsMapping {
			t.Errorf("Expected IsMapping to be false, got true")
		}
		// test ingest with valid config and value set for IsMapping
		err = config.IngestConfig(map[string]any{
			"IsMapping": true,
			"ProducerConfigs": []map[string]any{
				{
					"Type": "HTTP",
					"ProducerConfig": map[string]any{
						"URL": "http://test.com",
					},
				},
			},
		})
		if err != nil {
			t.Errorf("Expected no error from IngestConfig, got %v", err)
		}
		if len(config.SelectProducerConfigs) != 1 {
			t.Errorf("Expected SelectProducerConfigs to have 1 element, got %v", len(config.SelectProducerConfigs))
		}
		if !config.IsMapping {
			t.Errorf("Expected IsMapping to be true, got false")
		}
		// test ingest when ProducerConfigs is not set
		err = config.IngestConfig(map[string]any{})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "ProducerConfigs not set correctly" {
			t.Errorf("Expected specified error from IngestConfig, got '%v'", err)
		}
		// test ingest when ProducerConfigs is not a slice of map[string]any
		err = config.IngestConfig(map[string]any{
			"ProducerConfigs": "test",
		})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "ProducerConfigs not set correctly" {
			t.Errorf("Expected specified error from IngestConfig, got '%v'", err)
		}
		// test ingest when ProducerConfigs is an empty slice
		err = config.IngestConfig(map[string]any{
			"ProducerConfigs": []map[string]any{},
		})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "ProducerConfigs is empty" {
			t.Errorf("Expected specified error from IngestConfig, got '%v'", err)
		}
		// test when there is an error in one of the SelectProducerConfig.IngestConfig
		err = config.IngestConfig(map[string]any{
			"ProducerConfigs": []map[string]any{
				{
					"Type": "HTTP",
					"ProducerConfig": map[string]any{
						"URL": "http://test.com",
					},
				},
				{
					"Type": "HTTP",
					"ProducerConfig": map[string]any{
						"URL": 1,
					},
				},
			},
		})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
	})
}

// Tests for RabbitMQProducerConfig
func TestRabbitMQProducerConfig(t *testing.T) {
	t.Run("ImplementConfig", func(t *testing.T) {
		config := &RabbitMQProducerConfig{}
		_, ok := interface{}(config).(Config)
		if !ok {
			t.Errorf("Expected config to implement Config interface")
		}
	})
	t.Run("IngestConfig", func(t *testing.T) {
		config := &RabbitMQProducerConfig{}
		// Tests when nothing is set
		err := config.IngestConfig(map[string]any{})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "invalid Connection - must be a string and must be set" {
			t.Errorf("Expected specified error from IngestConfig, got '%v'", err)
		}
		// Test when Connection is not a string
		err = config.IngestConfig(map[string]any{
			"Connection": 1,
		})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "invalid Connection - must be a string and must be set" {
			t.Errorf("Expected specified error from IngestConfig, got '%v'", err)
		}
		// Test when Connection is set correctly and Exchange is set but not a string
		err = config.IngestConfig(map[string]any{
			"Connection": "test",
			"Exchange":   1,
		})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "invalid Exchange - must be a string" {
			t.Errorf("Expected specified error from IngestConfig, got '%v'", err)
		}
		// Test when Connection is set and Exchange is set correctly but RoutingKey is not set
		err = config.IngestConfig(map[string]any{
			"Connection": "test",
			"Exchange":   "test",
		})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "invalid Queue - must be a string and must be set" {
			t.Errorf("Expected specified error from IngestConfig, got '%v'", err)
		}
		// Test when Connection and Exchange are set correctly and RoutingKey is set but not a string
		err = config.IngestConfig(map[string]any{
			"Connection": "test",
			"Exchange":   "test",
			"RoutingKey": 1,
		})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "invalid Queue - must be a string and must be set" {
			t.Errorf("Expected specified error from IngestConfig, got '%v'", err)
		}
		// Test when everything is set correctly with default value for Exchange
		config = &RabbitMQProducerConfig{}
		err = config.IngestConfig(map[string]any{
			"Connection": "test",
			"Queue": "test",
		})
		if err != nil {
			t.Errorf("Expected no error from IngestConfig, got %v", err)
		}
		if config.Connection != "test" {
			t.Errorf("Expected Connection to be 'test', got '%v'", config.Connection)
		}
		if config.Exchange != "" {
			t.Errorf("Expected Exchange to be '', got '%v'", config.Exchange)
		}
		if config.RoutingKey != "test" {
			t.Errorf("Expected RoutingKey to be 'test', got '%v'", config.RoutingKey)
		}
		// Test when everything is set correctly
		config = &RabbitMQProducerConfig{}
		err = config.IngestConfig(map[string]any{
			"Connection": "test",
			"Exchange":   "test",
			"Queue": "test",
		})
		if err != nil {
			t.Errorf("Expected no error from IngestConfig, got %v", err)
		}
		if config.Connection != "test" {
			t.Errorf("Expected Connection to be 'test', got '%v'", config.Connection)
		}
		if config.Exchange != "test" {
			t.Errorf("Expected Exchange to be 'test', got '%v'", config.Exchange)
		}
		if config.RoutingKey != "test" {
			t.Errorf("Expected RoutingKey to be 'test', got '%v'", config.RoutingKey)
		}
	})
}

// MockRabbitMQProducerChannel is a mock implementation of the RabbitMQProducerChannel interface
type MockRabbitMQProducerChannel struct {
	isQueueDeclareError bool
	isPublishError      bool
	isCloseError        bool
	incomingData        rabbitmq.Publishing
}

func (c *MockRabbitMQProducerChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args rabbitmq.Table) (rabbitmq.Queue, error) {
	if c.isQueueDeclareError {
		return rabbitmq.Queue{}, errors.New("error getting queue")
	}
	return rabbitmq.Queue{Name: "test"}, nil
}

func (c *MockRabbitMQProducerChannel) Publish(exchange, key string, mandatory, immediate bool, msg rabbitmq.Publishing) error {
	if c.isPublishError {
		return errors.New("error publishing message")
	}
	c.incomingData = msg
	return nil
}

func (c *MockRabbitMQProducerChannel) Close() error {
	if c.isCloseError {
		return errors.New("test error")
	}
	return nil
}

// MockRabbitMQProducerConnection is a mock implementation of the RabbitMQProducerConnection interface
type MockRabbitMQProducerConnection struct {
	isChannelError bool
	isCloseError   bool
	channel        *MockRabbitMQProducerChannel
}

func (c *MockRabbitMQProducerConnection) Channel() (RabbitMQProducerChannel, error) {
	if c.isChannelError {
		return nil, errors.New("error getting channel")
	}
	if c.channel == nil {
		return nil, errors.New("channel not set")
	}
	return c.channel, nil
}

func (c *MockRabbitMQProducerConnection) Close() error {
	if c.isCloseError {
		return errors.New("test error")
	}
	return nil
}

// MockRabbitMQProducerDialWrapper
func MockRabbitMQProducerDialWrapper(url string, mockRabbitMQProducerConnection *MockRabbitMQProducerConnection, isError bool) func(string) (RabbitMQProducerConnection, error) {
	return func(url string) (RabbitMQProducerConnection, error) {
		if isError {
			return nil, errors.New("error dialing")
		}
		return mockRabbitMQProducerConnection, nil
	}
}

// Tests for RabbitMQProducer
func TestRabbitMQProducer(t *testing.T) {
	t.Run("ImplementsSinkServer", func(t *testing.T) {
		producer := &RabbitMQProducer{}
		_, ok := interface{}(producer).(SinkServer)
		if !ok {
			t.Errorf("Expected producer to implement SinkServer interface")
		}
	})
	t.Run("Setup", func(t *testing.T) {
		producer := &RabbitMQProducer{}
		config := &RabbitMQProducerConfig{
			Connection: "test",
			RoutingKey: "test",
		}

		err := producer.Setup(config)
		if err != nil {
			t.Errorf("Expected no error from Setup, got %v", err)
		}
		if producer.config != config {
			t.Errorf("Expected producer.config to be equal to config")
		}
		if producer.dial == nil {
			t.Errorf("Expected dial to be set, got nil")
		}
		if producer.ctx == nil {
			t.Errorf("Expected context to be set, got nil")
		}
		if producer.cancel == nil {
			t.Errorf("Expected cancel to be set, got nil")
		}
		// Try with invalid config
		err = producer.Setup(&MockConfig{})
		if err.Error() != "config is not a RabbitMQProducerConfig" {
			t.Errorf("Expected specified error from Setup, got '%v'", err)
		}
	})
	t.Run("Serve", func(t *testing.T) {
		producer := &RabbitMQProducer{}
		err := producer.Serve()
		// check error case when config is not set
		if err == nil {
			t.Errorf("Expected error from Serve, got nil")
		}
		if err.Error() != "config not set" {
			t.Errorf("Expected specified error from Serve, got '%v'", err)
		}
		// check error case when dial is not set
		producer.config = &RabbitMQProducerConfig{}
		err = producer.Serve()
		if err == nil {
			t.Errorf("Expected error from Serve, got nil")
		}
		if err.Error() != "dial not set" {
			t.Errorf("Expected specified error from Serve, got '%v'", err)
		}
		// check error case when ctx is not set
		producer.dial = MockRabbitMQProducerDialWrapper("", &MockRabbitMQProducerConnection{}, false)
		err = producer.Serve()
		if err == nil {
			t.Errorf("Expected error from Serve, got nil")
		}
		if err.Error() != "context not set" {
			t.Errorf("Expected specified error from Serve, got '%v'", err)
		}
		// check error case when cancel is not set
		ctx, cancel := context.WithCancelCause(context.Background())
		producer.ctx = ctx
		err = producer.Serve()
		if err == nil {
			t.Errorf("Expected error from Serve, got nil")
		}
		if err.Error() != "context cancel not set" {
			t.Errorf("Expected specified error from Serve, got '%v'", err)
		}
		// check case when there is an error in dial
		producer.cancel = cancel
		producer.dial = MockRabbitMQProducerDialWrapper("", &MockRabbitMQProducerConnection{}, true)
		err = producer.Serve()
		if err == nil {
			t.Errorf("Expected error from Serve, got nil")
		}
		if err.Error() != "error dialing" {
			t.Errorf("Expected specified error from Serve, got '%v'", err)
		}
		// check case when there is an error in getting channel
		producer.dial = MockRabbitMQProducerDialWrapper("", &MockRabbitMQProducerConnection{isChannelError: true}, false)
		err = producer.Serve()
		if err == nil {
			t.Errorf("Expected error from Serve, got nil")
		}
		if err.Error() != "error getting channel" {
			t.Errorf("Expected specified error from Serve, got '%v'", err)
		}
		// check case when the context is cancelled but with an error
		producer.dial = MockRabbitMQProducerDialWrapper("", &MockRabbitMQProducerConnection{
			channel: &MockRabbitMQProducerChannel{},
		}, false)
		producer.cancel(errors.New("check context with error"))
		err = producer.Serve()
		if err == nil {
			t.Errorf("Expected error from Serve, got nil")
		}
		if err.Error() != "check context with error" {
			t.Errorf("Expected specified error from Serve, got '%v'", err)
		}
		// check case when everything is set correctly
		channel := &MockRabbitMQProducerChannel{}
		producer = &RabbitMQProducer{
			config: &RabbitMQProducerConfig{
				Connection: "test",
				RoutingKey: "test",
			},
			dial: MockRabbitMQProducerDialWrapper("", &MockRabbitMQProducerConnection{
				channel: channel,
			}, false),
		}

		ctx, cancel = context.WithCancelCause(context.Background())
		producer.ctx = ctx
		producer.cancel = cancel
		producer.cancel(nil)
		err = producer.Serve()
		if err != nil {
			t.Errorf("Expected error from Serve, got nil")
		}
		if producer.channel != channel {
			t.Errorf("Expected producer.channel to be equal to channel")
		}
	})
	t.Run("SendTo", func(t *testing.T) {
		producer := &RabbitMQProducer{}
		appData := &AppData{
			data: map[string]any{"test": "data"},
		}
		// check error case when config is not set
		err := producer.SendTo(appData)
		if err == nil {
			t.Fatalf("Expected error from SendTo, got nil")
		}
		if err.Error() != "config not set" {
			t.Fatalf("Expected specified error from SendTo, got '%v'", err)
		}
		// check error case when context is not set 
		producer.config = &RabbitMQProducerConfig{}
		err = producer.SendTo(appData)
		if err == nil {
			t.Fatalf("Expected error from SendTo, got nil")
		}
		if err.Error() != "context not set" {
			t.Fatalf("Expected specified error from SendTo, got '%v'", err)
		}
		// check error case when cancel not set
		ctx, cancel := context.WithCancelCause(context.Background())
		producer.ctx = ctx
		err = producer.SendTo(appData)
		if err == nil {
			t.Fatalf("Expected error from SendTo, got nil")
		}
		if err.Error() != "context cancel not set" {
			t.Fatalf("Expected specified error from SendTo, got '%v'", err)
		}
		// check error case when channel is not set
		producer.cancel = cancel
		err = producer.SendTo(appData)
		if err == nil {
			t.Fatalf("Expected error from SendTo, got nil")
		}
		if err.Error() != "channel not set" {
			t.Fatalf("Expected specified error from SendTo, got '%v'", err)
		}
		// check error case when appData.GetHandler() returns an error
		producer.channel = &MockRabbitMQProducerChannel{}
		err = producer.SendTo(appData)
		if err.Error() != "handler not set" {
			t.Fatalf("Expected specified error from SendTo, got '%v'", err)
		}
		// check error when the data cannot be marshalled to json
		appData = &AppData{
			data: make(chan int),
			handler: &MockCompletionHandler{},
		}
		err = producer.SendTo(appData)
		if err == nil {
			t.Fatalf("Expected error from SendTo, got nil")
		}
		if err.Error() != "json: unsupported type: chan int" {
			t.Fatalf("Expected specified error from SendTo, got '%v'", err)
		}
		// check error case when Publish() returns an error
		appData = &AppData{
			data: map[string]any{"test": "data"},
			handler: &MockCompletionHandler{},
		}
		producer.channel = &MockRabbitMQProducerChannel{isPublishError: true}
		err = producer.SendTo(appData)
		if err.Error() != "error publishing message" {
			t.Fatalf("Expected specified error from SendTo, got '%v'", err)
		}
		<-producer.ctx.Done()
		if context.Cause(producer.ctx).Error() != "error publishing message" {
			t.Fatalf("Expected specified error from SendTo, got '%v'", context.Cause(producer.ctx).Error())
		}
		// check case when everything is set correctly
		ctx, cancel = context.WithCancelCause(context.Background())
		producer = &RabbitMQProducer{
			config: &RabbitMQProducerConfig{
				Connection: "test",
				RoutingKey: "test",
			},
			ctx:    ctx,
			cancel: cancel,
			channel: &MockRabbitMQProducerChannel{},
		}
		appData = &AppData{
			data: map[string]any{"test": "data"},
			handler: &MockCompletionHandler{},
		}
		err = producer.SendTo(appData)
		if err != nil {
			t.Fatalf("Expected no error from SendTo, got %v", err)
		}
		marshalledData, err := json.Marshal(appData.data)
		if err != nil {
			t.Fatalf("Expected no error from json.Marshal, got %v", err)
		}
		publishedData := rabbitmq.Publishing{
			Body:    marshalledData,
			ContentType: "application/json",
		}
		if !reflect.DeepEqual(producer.channel.(*MockRabbitMQProducerChannel).incomingData, publishedData) {
			t.Fatalf("Expected incomingData to be '%v', got '%v'", publishedData, producer.channel.(*MockRabbitMQProducerChannel).incomingData)
		}
		if producer.ctx.Err() != nil {
			t.Fatalf("Expected context to still be active, got %v", producer.ctx.Err())
		}
		// check case when there is an error in Complete method of the handler
		appData = &AppData{
			data: map[string]any{"test": "data"},
			handler: &MockCompletionHandler{
				isError: true,
			},
		}
		err = producer.SendTo(appData)
		if err == nil {
			t.Fatalf("Expected error from SendTo, got nil")
		}
		if err.Error() != "error" {
			t.Fatalf("Expected specified error from SendTo, got '%v'", err)
		}
		<-producer.ctx.Done()
		if context.Cause(producer.ctx).Error() != "error" {
			t.Fatalf("Expected specified error from SendTo, got '%v'", context.Cause(producer.ctx).Error())
		}
	})



}
