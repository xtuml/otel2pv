package Server

import (
	"errors"
	"testing"
)

// MockProducerConfig is a mock implementation of the ProducerConfig interface
type MockProducerConfig struct {
	isError bool
}

func (c *MockProducerConfig) IngestConfig(any) error {
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

		err := config.IngestConfig(1)

		if err != nil {
			t.Errorf("Expected no error from IngestConfig, got %v", err)
		}
	})
	t.Run("Error", func(t *testing.T) {
		config := MockProducerConfig{
			isError: true,
		}

		err := config.IngestConfig(struct{}{})

		if err == nil {
			t.Errorf("Expected error from IngestConfig, got '%v'", err)
		}
	})
}


// MockProducer is a mock implementation of the Producer interface
type MockProducer struct {
	isSetupError bool
	isStartError bool
	isSendError  bool
}

func (p *MockProducer) Setup(ProducerConfig) error {
	if p.isSetupError {
		return errors.New("test error")
	}
	return nil
}
func (p *MockProducer) Start() error {
	if p.isStartError {
		return errors.New("test error")
	}
	return nil
}
func (p *MockProducer) Send(any) error {
	if p.isSendError {
		return errors.New("test error")
	}
	return nil
}

func TestProducer(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		producer := MockProducer{
			isSetupError: false,
			isStartError: false,
			isSendError:  false,
		}
		
		_, ok := interface{}(&producer).(Producer)
		if !ok {
			t.Errorf("Expected producer to implement Producer interface")
		}

		err := producer.Setup(&MockProducerConfig{})

		if err != nil {
			t.Errorf("Expected no error from Setup, got %v", err)
		}

		err = producer.Start()
		if err != nil {
			t.Errorf("Expected no error from Start, got %v", err)
		}

		err = producer.Send(nil)
		if err != nil {
			t.Errorf("Expected no error from Send, got %v", err)
		}
	})
	t.Run("Error", func(t *testing.T) {
		producer := MockProducer{
			isSetupError: true,
			isStartError: true,
			isSendError:  true,
		}

		err := producer.Setup(&MockProducerConfig{})

		if err == nil {
			t.Errorf("Expected error from Setup, got '%v'", err)
		}

		err = producer.Start()
		if err == nil {
			t.Errorf("Expected error from Start, got '%v'", err)
		}

		err = producer.Send(nil)
		if err == nil {
			t.Errorf("Expected error from Send, got '%v'", err)
		}
	})
}
