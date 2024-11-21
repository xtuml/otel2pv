package Server

import (
	"errors"
	"strings"
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
	t.Run("IngestConfig", func(t *testing.T) {
		scc := &SelectConsumerConfig{}
		// Tests valid case
		err := scc.IngestConfig(map[string]any{"Type": "RabbitMQ", "ConsumerConfig": map[string]any{
			"Connection": "test",
			"Queue":      "test",
		}})
		if err != nil {
			t.Errorf("Expected no error from IngestConfig, got %v", err)
		}
		if scc.Type != "RabbitMQ" {
			t.Errorf("Expected Type to be 'RabbitMQ', got %v", scc.Type)
		}
		if scc.ConsumerConfig == nil {
			t.Errorf("Expected ConsumerConfig to be set, got nil")
		}
		if _, ok := scc.ConsumerConfig.(*RabbitMQConsumerConfig); !ok {
			t.Errorf("Expected ConsumerConfig to be of type RabbitMQConsumerConfig, got %T", scc.ConsumerConfig)
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
		// Test when Type is not a string
		err = scc.IngestConfig(map[string]any{"Type": 1})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "invalid Type - must be a string and must be set" {
			t.Errorf("Expected error message to be 'invalid Type - must be a string and must be set', got %v", err.Error())
		}
		// Tests when Type is not in the CONSUMERCONFIGMAP
		err = scc.IngestConfig(map[string]any{"Type": "test"})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "invalid consumer type: test" {
			t.Errorf("Expected error message to be 'invalid consumer type: test', got %v", err.Error())
		}
		// Test when ConsumerConfig is not set
		err = scc.IngestConfig(map[string]any{"Type": "RabbitMQ"})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "Consumer config not set correctly" {
			t.Errorf("Expected error message to be 'Consumer config not set correctly', got %v", err.Error())
		}
		// Test when ConsumerConfig is not a map
		err = scc.IngestConfig(map[string]any{"Type": "RabbitMQ", "ConsumerConfig": "test"})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "Consumer config not set correctly" {
			t.Errorf("Expected error message to be 'Consumer config not set correctly', got %v", err.Error())
		}
		// Test when ConsumerConfig returns an error
		err = scc.IngestConfig(map[string]any{"Type": "RabbitMQ", "ConsumerConfig": map[string]any{}})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if !strings.Contains(err.Error(), "Consumer config not set correctly:") {
			t.Errorf("Expected error message to contain 'Consumer config not set correctly:', got %v", err.Error())
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
		err = scc.IngestConfig(map[string]any{"ConsumerConfigs": "test"})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "ConsumerConfigs not set correctly" {
			t.Errorf("Expected error message to be 'ConsumerConfigs not set correctly', got %v", err.Error())
		}
		// Test when SelectConsumerConfigs is an empty slice
		err = scc.IngestConfig(map[string]any{"ConsumerConfigs": []map[string]any{}})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "ConsumerConfigs is empty" {
			t.Errorf("Expected error message to be 'ConsumerConfigs is empty', got %v", err.Error())
		}
		// Test when SelectConsumerConfigs is a slice of SelectConsumerConfig but one of them is invalid
		err = scc.IngestConfig(map[string]any{"ConsumerConfigs": []map[string]any{
			{"Type": "RabbitMQ", "ConsumerConfig": map[string]any{
				"Connection": "test",
				"Queue":      "test",
			}},
			{},
		}})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got ")
		}
		// Tests the valid case
		err = scc.IngestConfig(map[string]any{"ConsumerConfigs": []map[string]any{
			{"Type": "RabbitMQ", "ConsumerConfig": map[string]any{
				"Connection": "test",
				"Queue":      "test",
			}},
			{"Type": "RabbitMQ", "ConsumerConfig": map[string]any{
				"Connection": "test",
				"Queue":      "test",
			}},
		}})
		if err != nil {
			t.Errorf("Expected no error from IngestConfig, got %v", err)
		}
		if len(scc.SelectConsumerConfigs) != 2 {
			t.Errorf("Expected SelectConsumerConfigs to have 2 elements, got %v", len(scc.SelectConsumerConfigs))
		}
	})
}

// Tests for RabbitMQConsumerConfig
func TestRabbitMQConsumerConfig(t *testing.T) {
	t.Run("ImplementsConfig", func(t *testing.T) {
		rcc := &RabbitMQConsumerConfig{}
		// Type assertion to check if RabbitMQConsumerConfig implements Config
		_, ok := interface{}(rcc).(Config)
		if !ok {
			t.Errorf("Expected RabbitMQConsumerConfig to implement Config interface")
		}
	})
	t.Run("IngestConfig", func(t *testing.T) {
		rcc := &RabbitMQConsumerConfig{}
		// Test when the RabbitMQConsumerConfig has no fields
		err := rcc.IngestConfig(map[string]any{})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "invalid Connection - must be a string and must be set" {
			t.Errorf("Expected error message to be 'invalid Connection - must be a string and must be set', got %v", err.Error())
		}
		// Test when the RabbitMQConsumerConfig has the Connection field set but it is not a string
		err = rcc.IngestConfig(map[string]any{"Connection": 1})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "invalid Connection - must be a string and must be set" {
			t.Errorf("Expected error message to be 'invalid Connection - must be a string and must be set', got %v", err.Error())
		}
		// Test when the RabbitMQConsumerConfig has a valid Connection field but the Queue field is not set
		err = rcc.IngestConfig(map[string]any{"Connection": "test"})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "invalid Queue - must be a string and must be set" {
			t.Errorf("Expected error message to be 'invalid Queue - must be a string and must be set', got %v", err.Error())
		}
		// Tests when the RabbitMQConsumerConfig has a valid Connection and Queue field is set but not a string
		err = rcc.IngestConfig(map[string]any{"Connection": "test", "Queue": 1})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "invalid Queue - must be a string and must be set" {
			t.Errorf("Expected error message to be 'invalid Queue - must be a string and must be set', got %v", err.Error())
		}
		// Test when the RabbitMQConsumerConfig has a valid Connection, Queue, and ConsumerTag field set but ConsumerTag is not a string
		err = rcc.IngestConfig(map[string]any{"Connection": "test", "Queue": "test", "ConsumerTag": 1})
		if err == nil {
			t.Errorf("Expected errnilor from IngestConfig, got nil")
		}
		if err.Error() != "invalid ConsumerTag - must be a string" {
			t.Errorf("Expected error message to be 'invalid ConsumerTag - must be a string', got %v", err.Error())
		}
		// Test when the RabbitMQConsumerConfig has a valid Connection and Queue field set but ConsumerTag is not set
		rcc = &RabbitMQConsumerConfig{}
		err = rcc.IngestConfig(map[string]any{"Connection": "test", "Queue": "test"})
		if err != nil {
			t.Errorf("Expected no error from IngestConfig, got %v", err)
		}
		if rcc.ConsumerTag != "" {
			t.Errorf("Expected ConsumerTag to be empty, got %v", rcc.ConsumerTag)
		}
		if rcc.Connection != "test" {
			t.Errorf("Expected Connection to be 'test', got %v", rcc.Connection)
		}
		if rcc.Queue != "test" {
			t.Errorf("Expected Queue to be 'test', got %v", rcc.Queue)
		}
		// Test when the RabbitMQConsumerConfig has a valid Connection, Queue, and ConsumerTag field set
		rcc = &RabbitMQConsumerConfig{}
		err = rcc.IngestConfig(map[string]any{"Connection": "test", "Queue": "test", "ConsumerTag": "test"})
		if err != nil {
			t.Errorf("Expected no error from IngestConfig, got %v", err)
		}
		if rcc.ConsumerTag != "test" {
			t.Errorf("Expected ConsumerTag to be 'test', got %v", rcc.ConsumerTag)
		}
		if rcc.Connection != "test" {
			t.Errorf("Expected Connection to be 'test', got %v", rcc.Connection)
		}
		if rcc.Queue != "test" {
			t.Errorf("Expected Queue to be 'test', got %v", rcc.Queue)
		}
	})
}
