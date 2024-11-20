package Server

import (
	"errors"
	"testing"
)

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

func (c *MockConsumer) Setup(config Config) error {
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

		config := &MockConfig{isError: false}
		err := consumer.Setup(config)
		if err != nil {
			t.Fatalf("Expected no error from Setup, got %v", err)
		}
		if _, configOk := interface{}(config).(Config); !configOk {
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

		err := consumer.Setup(&MockConfig{isError: true})
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


// Tests for SelectConsumerConfig
func TestSelectConsumerConfig(t *testing.T) {
	t.Run("ImplementsConfig", func(t *testing.T) {
		scc := &SelectConsumerConfig{}
		// Type assertion to check if SelectConsumerConfig implements Config
		_, ok := interface{}(scc).(Config)
		if !ok {
			t.Errorf("Expected SelectConsumerConfig to implement Config interface")
		}
	})
	t.Run("IngestConfigInvalid", func(t *testing.T) {
		scc := &SelectConsumerConfig{}
		// Test when Type is not set
		err := scc.IngestConfig(map[string]any{})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "invalid Type - must be a string and must be set" {
			t.Errorf("Expected error message to be 'invalid Type - must be a string and must be set', got %v", err.Error())
		}
		// Test when ConsumerConfig is not set
		err = scc.IngestConfig(map[string]any{"Type": "test"})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "Consumer config not set correctly" {
			t.Errorf("Expected error message to be 'Consumer config not set correctly', got %v", err.Error())
		}
		// Test when ConsumerConfig is not a map
		err = scc.IngestConfig(map[string]any{"Type": "test", "ConsumerConfig": "test"})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "Consumer config not set correctly" {
			t.Errorf("Expected error message to be 'Consumer config not set correctly', got %v", err.Error())
		}
		// Test when ConsumerConfig is a map but the consumer type is invalid
		err = scc.IngestConfig(map[string]any{"Type": "test", "ConsumerConfig": map[string]any{}})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "invalid consumer type: test" {
			t.Errorf("Expected error message to be 'invalid consumer type: test', got %v", err.Error())
		}
	})
}

// Tests for SetupConsumersConfig
func TestSetupConsumersConfig(t *testing.T) {
	t.Run("ImplementsConfig", func(t *testing.T) {
		scc := &SetupConsumersConfig{}
		// Type assertion to check if SetupConsumersConfig implements Config
		_, ok := interface{}(scc).(Config)
		if !ok {
			t.Errorf("Expected SetupConsumersConfig to implement Config interface")
		}
	})
	t.Run("IngestConfig", func(t *testing.T) {
		scc := &SetupConsumersConfig{}
		// Test when SelectConsumerConfigs is not set
		err := scc.IngestConfig(map[string]any{})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "ConsumerConfigs not set correctly" {
			t.Errorf("Expected error message to be 'ConsumerConfigs not set correctly', got %v", err.Error())
		}
		// Test when SelectConsumerConfigs is not a slice
		err = scc.IngestConfig(map[string]any{"SelectConsumerConfigs": "test"})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "ConsumerConfigs not set correctly" {
			t.Errorf("Expected error message to be 'ConsumerConfigs not set correctly', got %v", err.Error())
		}
		// Test when SelectConsumerConfigs is a slice but the elements are not SelectConsumerConfig
		err = scc.IngestConfig(map[string]any{"SelectConsumerConfigs": []any{"test"}})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "ConsumerConfigs not set correctly" {
			t.Errorf("Expected error message to be 'ConsumerConfigs not set correctly', got %v", err.Error())
		}
		// Test when SelectConsumerConfigs is a slice of SelectConsumerConfig but one of them is invalid
		err = scc.IngestConfig(map[string]any{"SelectConsumerConfigs": []any{map[string]any{}}})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
	})
}