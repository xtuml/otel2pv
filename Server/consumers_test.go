package Server

import (
	"errors"
	"testing"
)

// Mock ConsumerConfig is a mock implementation of the ConsumerConfig interface
type MockConsumerConfig struct {
	isError bool
}

func (c *MockConsumerConfig) IngestConfig(map[string]any) error {
	if c.isError {
		return errors.New("test error")
	}
	return nil
}

func TestConsumerConfig(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		config := MockConsumerConfig{
			isError: false,
		}

		_, ok := interface{}(&config).(ConsumerConfig)
		if !ok {
			t.Errorf("Expected config to implement ConsumerConfig interface")
		}

		err := config.IngestConfig(map[string]any{})
		if err != nil {
			t.Errorf("Expected no error from IngestConfig, got %v", err)
		}
	})
	t.Run("Error", func(t *testing.T) {
		config := MockConsumerConfig{
			isError: true,
		}

		err := config.IngestConfig(map[string]any{})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
	})
}

// MockConsumer is a mock implementation of the Consumer interface
type MockConsumer struct {
	isAddError   bool
	isSetupError bool
	isServeError bool
	Pushable     Pushable
}

func (c *MockConsumer) AddPushable(pushable Pushable) error {
	if c.isAddError {
		return errors.New("test error")
	}
	c.Pushable = pushable
	return nil
}

func (c *MockConsumer) Setup(config ConsumerConfig) error {
	if c.isSetupError {
		return errors.New("test error")
	}
	return nil
}

func (c *MockConsumer) Serve() error {
	if c.isServeError {
		return errors.New("test error")
	}
	return nil
}

func TestConsumer(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		consumer := MockConsumer{
			isAddError:   false,
			isSetupError: false,
			isServeError: false,
		}

		if _, ok := interface{}(&consumer).(Consumer); !ok {
			t.Fatalf("Expected consumer to implement Consumer interface")
		}

		config := &MockConsumerConfig{isError: false}
		err := consumer.Setup(config)
		if err != nil {
			t.Fatalf("Expected no error from Setup, got %v", err)
		}
		if _, configOk := interface{}(config).(ConsumerConfig); !configOk {
			t.Fatalf("Expected consumer to implement Consumer interface")
		}

		mockPushable := &MockPushable{}
		err = consumer.AddPushable(mockPushable)
		if err != nil {
			t.Fatalf("Expected no error from AddPushable, got %v", err)
		}

		if consumer.Pushable != mockPushable {
			t.Fatalf("Expected AddPushable to set the Pushable")
		}

		err = consumer.Serve()
		if err != nil {
			t.Fatalf("Expected no error from Serve, got %v", err)
		}
	})

	t.Run("Error", func(t *testing.T) {
		consumer := MockConsumer{
			isAddError:   true,
			isSetupError: true,
			isServeError: true,
		}

		err := consumer.Setup(&MockConsumerConfig{isError: true})
		if err == nil {
			t.Fatalf("Expected error from Setup, got nil")
		}

		mockPushable := &MockPushable{}
		err = consumer.AddPushable(mockPushable)
		if err == nil {
			t.Fatalf("Expected error from AddPushable, got nil")
		}

		err = consumer.Serve()
		if err == nil {
			t.Fatalf("Expected error from Serve, got nil")
		}
	})
}
