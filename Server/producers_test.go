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
	isSetupError              bool
	isReceiveError            bool
	isHandleIncomingDataError bool
	isServeError              bool
	Receiver                  Receiver
	AppData                   AppData
}

func (p *MockProducer) Setup(ProducerConfig) error {
	if p.isSetupError {
		return errors.New("test error")
	}
	return nil
}
func (p *MockProducer) GetReceiver() (Receiver, error) {
	if p.isReceiveError {
		return nil, errors.New("test error")
	}
	return p.Receiver, nil
}
func (p *MockProducer) HandleIncomingData(data AppData) error {
	if p.isHandleIncomingDataError {
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
			isSetupError:              false,
			isReceiveError:            false,
			isHandleIncomingDataError: false,
			isServeError:              false,
			Receiver:                  &MockReceiver{},
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

		receiver, err := producer.GetReceiver()
		if err != nil {
			t.Errorf("Expected no error from Start, got %v", err)
		}
		if receiver != producer.Receiver {
			t.Errorf("Expected receiver to be equal to producer.Receiver")
		}

		appData := AppData{
			data:    "test data",
			handler: &MockCompletionHandler{},
		}
		err = producer.HandleIncomingData(appData)
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
			isSetupError:              true,
			isReceiveError:            true,
			isHandleIncomingDataError: true,
			isServeError:              true,
		}

		err := producer.Setup(&MockProducerConfig{})
		if err == nil {
			t.Errorf("Expected error from Setup, got '%v'", err)
		}

		receiver, err := producer.GetReceiver()
		if err == nil {
			t.Errorf("Expected error from Start, got '%v'", err)
		}
		if receiver != nil {
			t.Errorf("Expected receiver to be nil")
		}

		err = producer.HandleIncomingData(AppData{})
		if err == nil {
			t.Errorf("Expected error from Send, got '%v'", err)
		}

		err = producer.Serve()
		if err == nil {
			t.Errorf("Expected error from Serve, got '%v'", err)
		}
	})
}
