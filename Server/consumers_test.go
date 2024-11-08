package Server

import (
	"errors"
	"testing"
)

// Mock ConsumerConfig is a mock implementation of the ConsumerConfig interface
type MockConsumerConfig struct {
	isError bool
}

func (c *MockConsumerConfig) IngestConfig(any) error {
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

		err := config.IngestConfig(nil)
		if err != nil {
			t.Errorf("Expected no error from IngestConfig, got %v", err)
		}
	})
	t.Run("Error", func(t *testing.T) {
		config := MockConsumerConfig{
			isError: true,
		}

		err := config.IngestConfig(nil)
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
	Receiver     Receiver
}

func (c *MockConsumer) AddReceiver(receiver Receiver) error {
	if c.isAddError {
		return errors.New("test error")
	}
	c.Receiver = receiver
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

		mockReceiver := &MockReceiver{}
		err = consumer.AddReceiver(mockReceiver)
		if err != nil {
			t.Fatalf("Expected no error from AddReceiver, got %v", err)
		}
		if _, receiverOk := interface{}(mockReceiver).(Receiver); !receiverOk {
			t.Fatalf("Expected receiver to implement Receiver interface")
		}

		if consumer.Receiver != mockReceiver {
			t.Fatalf("Expected AddReceiver to set the receiver")
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

		mockReceiver := &MockReceiver{}
		err = consumer.AddReceiver(mockReceiver)
		if err == nil {
			t.Fatalf("Expected error from AddReceiver, got nil")
		}

		if consumer.Receiver != nil {
			t.Fatalf("Expected AddReceiver to not set the receiver")
		}

		err = consumer.Serve()
		if err == nil {
			t.Fatalf("Expected error from Serve, got nil")
		}
	})
}
