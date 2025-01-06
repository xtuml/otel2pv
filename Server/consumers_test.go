package Server

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/Azure/go-amqp"
	rabbitmq "github.com/rabbitmq/amqp091-go"
)

// Tests for SelectConsumerConfig
func TestSelectConsumerConfig(t *testing.T) {
	t.Run("IngestConfig", func(t *testing.T) {
		scc := &SelectConsumerConfig{}
		// Tests valid case
		err := scc.IngestConfig(map[string]any{"Type": "RabbitMQ", "ConsumerConfig": map[string]any{
			"Connection": "test",
			"Queue":      "test",
		}}, CONSUMERCONFIGMAP)
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
		err := scc.IngestConfig(map[string]any{}, CONSUMERCONFIGMAP)
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "invalid Type - must be a string and must be set" {
			t.Errorf("Expected error message to be 'invalid Type - must be a string and must be set', got %v", err.Error())
		}
		// Test when Type is not a string
		err = scc.IngestConfig(map[string]any{"Type": 1}, CONSUMERCONFIGMAP)
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "invalid Type - must be a string and must be set" {
			t.Errorf("Expected error message to be 'invalid Type - must be a string and must be set', got %v", err.Error())
		}
		// Tests when Type is not in the CONSUMERCONFIGMAP
		err = scc.IngestConfig(map[string]any{"Type": "test"}, CONSUMERCONFIGMAP)
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "invalid consumer type: test" {
			t.Errorf("Expected error message to be 'invalid consumer type: test', got %v", err.Error())
		}
		// Test when ConsumerConfig is not set
		err = scc.IngestConfig(map[string]any{"Type": "RabbitMQ"}, CONSUMERCONFIGMAP)
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "Consumer config not set correctly" {
			t.Errorf("Expected error message to be 'Consumer config not set correctly', got %v", err.Error())
		}
		// Test when ConsumerConfig is not a map
		err = scc.IngestConfig(map[string]any{"Type": "RabbitMQ", "ConsumerConfig": "test"}, CONSUMERCONFIGMAP)
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "Consumer config not set correctly" {
			t.Errorf("Expected error message to be 'Consumer config not set correctly', got %v", err.Error())
		}
		// Test when ConsumerConfig returns an error
		err = scc.IngestConfig(map[string]any{"Type": "RabbitMQ", "ConsumerConfig": map[string]any{}}, CONSUMERCONFIGMAP)
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
	t.Run("IngestConfig", func(t *testing.T) {
		scc := &SetupConsumersConfig{}
		// Test when SelectConsumerConfigs is not set
		err := scc.IngestConfig(map[string]any{}, CONSUMERCONFIGMAP)
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "ConsumerConfigs not set correctly" {
			t.Errorf("Expected error message to be 'ConsumerConfigs not set correctly', got %v", err.Error())
		}
		// Test when SelectConsumerConfigs is not a slice
		err = scc.IngestConfig(map[string]any{"ConsumerConfigs": "test"}, CONSUMERCONFIGMAP)
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "ConsumerConfigs not set correctly" {
			t.Errorf("Expected error message to be 'ConsumerConfigs not set correctly', got %v", err.Error())
		}
		// Test when SelectConsumerConfigs is an empty slice
		err = scc.IngestConfig(map[string]any{"ConsumerConfigs": []any{}}, CONSUMERCONFIGMAP)
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
		}}, CONSUMERCONFIGMAP)
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got ")
		}
		// Tests the valid case
		err = scc.IngestConfig(map[string]any{"ConsumerConfigs": []any{
			map[string]any{"Type": "RabbitMQ", "ConsumerConfig": map[string]any{
				"Connection": "test",
				"Queue":      "test",
			}},
			map[string]any{"Type": "RabbitMQ", "ConsumerConfig": map[string]any{
				"Connection": "test",
				"Queue":      "test",
			}},
		}}, CONSUMERCONFIGMAP)
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

// MockPushableForRabbitMQConsumer is a mock implementation of the Pushable interface
type MockPushableForRabbitMQConsumer struct {
	incomingData  []*AppData
	isSendToError bool
}

func (p *MockPushableForRabbitMQConsumer) SendTo(data *AppData) error {
	if p.isSendToError {
		return errors.New("test error")
	}
	p.incomingData = append(p.incomingData, data)
	return nil
}

// MockAcknowledger
type MockAcknowledger struct {
	isError          bool
	isNackError      bool
	hasAcknowledged  bool
	hasNacknowledged bool
}

func (a *MockAcknowledger) Ack(tag uint64, multiple bool) error {
	if a.isError {
		return errors.New("acknowledge error")
	}
	a.hasAcknowledged = true
	return nil
}

func (a *MockAcknowledger) Nack(tag uint64, multiple bool, requeue bool) error {
	if a.isNackError {
		return errors.New("nack error")
	}
	a.hasNacknowledged = true
	return nil
}

func (a *MockAcknowledger) Reject(tag uint64, requeue bool) error {
	return nil
}

// MockRabbitMQChannel is a mock implementation of the RabbitMQChannel interface
type MockRabbitMQChannel struct {
	isQueueDeclareError bool
	isConsumeError      bool
	isCloseError        bool
	channel             chan rabbitmq.Delivery
}

func (c *MockRabbitMQChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args rabbitmq.Table) (rabbitmq.Queue, error) {
	if c.isQueueDeclareError {
		return rabbitmq.Queue{}, errors.New("error getting queue")
	}
	return rabbitmq.Queue{Name: "test"}, nil
}

func (c *MockRabbitMQChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args rabbitmq.Table) (<-chan rabbitmq.Delivery, error) {
	if c.isConsumeError {
		return nil, errors.New("error getting consumer")
	}
	if c.channel == nil {
		c.channel = make(chan rabbitmq.Delivery)
	}
	return c.channel, nil
}

func (c *MockRabbitMQChannel) Close() error {
	if c.isCloseError {
		return errors.New("test error")
	}
	return nil
}

// MockRabbitMQConnection is a mock implementation of the RabbitMQConnection interface
type MockRabbitMQConnection struct {
	isChannelError bool
	isCloseError   bool
	channel        *MockRabbitMQChannel
}

func (c *MockRabbitMQConnection) Channel() (RabbitMQChannel, error) {
	if c.isChannelError {
		return nil, errors.New("error getting channel")
	}
	if c.channel == nil {
		return nil, errors.New("channel not set")
	}
	return c.channel, nil
}

func (c *MockRabbitMQConnection) Close() error {
	if c.isCloseError {
		return errors.New("test error")
	}
	return nil
}

// MockRabbitMQDialWrapper
func MockRabbitMQDialWrapper(url string, mockRabbitMQConnection *MockRabbitMQConnection, isError bool) func(string) (RabbitMQConnection, error) {
	return func(url string) (RabbitMQConnection, error) {
		if isError {
			return nil, errors.New("dial error")
		}
		return mockRabbitMQConnection, nil
	}
}

// Tests for RabbitMQConsumer
func TestRabbitMQConsumer(t *testing.T) {
	t.Run("ImplementsSourceServer", func(t *testing.T) {
		rc := &RabbitMQConsumer{}
		// Type assertion to check if RabbitMQConsumer implements SourceServer
		_, ok := interface{}(rc).(SourceServer)
		if !ok {
			t.Errorf("Expected RabbitMQConsumer to implement SourceServer interface")
		}
	})
	t.Run("AddPushable", func(t *testing.T) {
		rc := &RabbitMQConsumer{}
		mockPushable := &MockPushable{}
		err := rc.AddPushable(mockPushable)
		if err != nil {
			t.Errorf("Expected no error from AddPushable, got %v", err)
		}
		if rc.pushable != mockPushable {
			t.Errorf("Expected AddPushable to set the Pushable, got %v", rc.pushable)
		}
		// Test when the Pushable is already set
		err = rc.AddPushable(mockPushable)
		if err == nil {
			t.Errorf("Expected error from AddPushable, got nil")
		}
	})
	t.Run("Setup", func(t *testing.T) {
		rc := &RabbitMQConsumer{}
		// Test when the config is not a RabbitMQConsumerConfig
		err := rc.Setup(&MockConfig{})
		if err == nil {
			t.Errorf("Expected error from Setup, got nil")
		}
		if err.Error() != "config is not a RabbitMQConsumerConfig" {
			t.Errorf("Expected error message to be 'config is not a RabbitMQConsumerConfig', got %v", err.Error())
		}
		// Test when the config is a RabbitMQConsumerConfig
		config := &RabbitMQConsumerConfig{}
		err = rc.Setup(config)
		if err != nil {
			t.Errorf("Expected no error from Setup, got %v", err)
		}
		if rc.config != config {
			t.Errorf("Expected config to be %v, got %v", config, rc.config)
		}
		if rc.dial == nil {
			t.Errorf("Expected dial to be set, got nil")
		}
	})
	t.Run("sendRabbitMQMessageDataToPushable", func(t *testing.T) {
		pushable := &MockPushable{}
		msg := rabbitmq.Delivery{Body: []byte(`{"key":"value"}`)}
		err := sendRabbitMQMessageDataToPushable(&msg, pushable)
		if err != nil {
			t.Errorf("Expected no error from sendRabbitMQMessageDataToPushable, got %v", err)
		}
		if pushable.incomingData == nil {
			t.Errorf("Expected data to be set, got nil")
		}
		appData := pushable.incomingData
		if dataMap, ok := appData.data.(map[string]any); !ok {
			t.Errorf("Expected data to be a map, got %T", appData.data)
		} else {
			if !reflect.DeepEqual(dataMap, map[string]any{"key": "value"}) {
				t.Errorf("Expected data to be %v, got %v", map[string]any{"key": "value"}, dataMap)
			}
		}
		// Test when there is an error in convertBytesJSONDataToAppData
		pushable = &MockPushable{}
		msg = rabbitmq.Delivery{Body: []byte(`{"key":"value"`)}
		err = sendRabbitMQMessageDataToPushable(&msg, pushable)
		if err == nil {
			t.Errorf("Expected error from sendRabbitMQMessageDataToPushable, got nil")
		}
		// Test when there is an error in pushable.SendTo
		pushable = &MockPushable{isSendToError: true}
		msg = rabbitmq.Delivery{Body: []byte(`{"key":"value"}`)}
		err = sendRabbitMQMessageDataToPushable(&msg, pushable)
		if err == nil {
			t.Errorf("Expected error from sendRabbitMQMessageDataToPushable, got nil")
		}
	})
	t.Run("sendChannelOfRabbitMQDeliveryToPushable", func(t *testing.T) {
		// Test when there is no error
		pushable := &MockPushableForRabbitMQConsumer{
			incomingData: []*AppData{},
		}
		channel := make(chan rabbitmq.Delivery, 2)
		mockAcknowledger := &MockAcknowledger{}
		msg1 := rabbitmq.Delivery{
			Body:         []byte(`{"key":"value1"}`),
			Acknowledger: mockAcknowledger,
		}
		msg2 := rabbitmq.Delivery{
			Body:         []byte(`{"key":"value2"}`),
			Acknowledger: mockAcknowledger,
		}
		channel <- msg1
		channel <- msg2
		close(channel)
		err := sendChannelOfRabbitMQDeliveryToPushable(channel, pushable)
		if err != nil {
			t.Fatalf("Expected no error from sendChannelOfRabbitMQDeliveryToPushable, got %v", err)
		}
		if len(pushable.incomingData) != 2 {
			t.Fatalf("Expected incomingData to have 2 elements, got %v", len(pushable.incomingData))
		}
		incomingData := pushable.incomingData
		sort.Slice(incomingData, func(i, j int) bool {
			return incomingData[i].data.(map[string]any)["key"].(string) < incomingData[j].data.(map[string]any)["key"].(string)
		})
		for i := 0; i < 2; i++ {
			appData := incomingData[i]
			if dataMap, ok := appData.data.(map[string]any); !ok {
				t.Fatalf("Expected data to be a map, got %T", appData.data)
			} else {
				if !reflect.DeepEqual(dataMap, map[string]any{"key": "value" + strconv.Itoa(i+1)}) {
					t.Fatalf("Expected data to be %v, got %v", map[string]any{"key": "value" + strconv.Itoa(i+1)}, dataMap)
				}
			}
		}
		// Test when there is an error in sendRabbitMQMessageDataToPushable
		pushable = &MockPushableForRabbitMQConsumer{
			incomingData: []*AppData{},
		}
		channel = make(chan rabbitmq.Delivery, 2)
		msg1 = rabbitmq.Delivery{
			Body:         []byte(`{"key":"value1"}`),
			Acknowledger: mockAcknowledger,
		}
		msg2 = rabbitmq.Delivery{
			Body:         []byte(`{"key":"value2"`),
			Acknowledger: mockAcknowledger,
		}
		channel <- msg1
		channel <- msg2
		close(channel)
		err = sendChannelOfRabbitMQDeliveryToPushable(channel, pushable)
		if err == nil {
			t.Fatalf("Expected error from sendChannelOfRabbitMQDeliveryToPushable, got nil")
		}
		// Test when there is an error in Ack
		pushable = &MockPushableForRabbitMQConsumer{
			incomingData: []*AppData{},
		}
		channel = make(chan rabbitmq.Delivery, 2)
		msg2 = rabbitmq.Delivery{
			Body:         []byte(`{"key":"value2"}`),
			Acknowledger: mockAcknowledger,
		}
		mockAcknowledger.isError = true
		channel <- msg1
		channel <- msg2
		close(channel)
		err = sendChannelOfRabbitMQDeliveryToPushable(channel, pushable)
		if err == nil {
			t.Fatalf("Expected error from sendChannelOfRabbitMQDeliveryToPushable, got nil")
		}
		if err.Error() != "acknowledge error" {
			t.Fatalf("Expected error message to be 'acknowledge error', got %v", err.Error())
		}
	})
	t.Run("Serve", func(t *testing.T) {
		rc := &RabbitMQConsumer{}
		// Test when the pushable is not set
		err := rc.Serve()
		if err == nil {
			t.Fatalf("Expected error from Serve, got nil")
		}
		if err.Error() != "Pushable not set" {
			t.Fatalf("Expected error message to be 'Pushable not set', got %v", err.Error())
		}
		// Test when the pushable is set but config is not set
		rc.pushable = &MockPushable{}
		err = rc.Serve()
		if err == nil {
			t.Fatalf("Expected error from Serve, got nil")
		}
		if err.Error() != "Config not set" {
			t.Fatalf("Expected error message to be 'Config not set', got %v", err.Error())
		}
		// Test when the pushable and config are set but the dial is not set
		rc.config = &RabbitMQConsumerConfig{}
		err = rc.Serve()
		if err == nil {
			t.Fatalf("Expected error from Serve, got nil")
		}
		if err.Error() != "Dialer not set" {
			t.Fatalf("Expected error message to be 'Dialer not set', got %v", err.Error())
		}
		// Test when the pushable, config, and dial are set but there is an error in dial
		rc.dial = MockRabbitMQDialWrapper("", nil, true)
		err = rc.Serve()
		if err == nil {
			t.Fatalf("Expected error from Serve, got nil")
		}
		if err.Error() != "dial error" {
			t.Fatalf("Expected error message to be 'dial error', got %v", err.Error())
		}
		// Test when the pushable, config, and dial are set but there is an error in Channel
		rc.dial = MockRabbitMQDialWrapper("", &MockRabbitMQConnection{isChannelError: true}, false)
		err = rc.Serve()
		if err == nil {
			t.Fatalf("Expected error from Serve, got nil")
		}
		if err.Error() != "error getting channel" {
			t.Fatalf("Expected error message to be 'error getting channel', got %v", err.Error())
		}
		// Test when the pushable, config, and dial are set but there is an error in QueueDeclare
		rc.dial = MockRabbitMQDialWrapper("", &MockRabbitMQConnection{channel: &MockRabbitMQChannel{isQueueDeclareError: true}}, false)
		err = rc.Serve()
		if err == nil {
			t.Fatalf("Expected error from Serve, got nil")
		}
		if err.Error() != "error getting queue" {
			t.Fatalf("Expected error message to be 'error getting queue', got %v", err.Error())
		}
		// Test when the pushable, config, and dial are set but there is an error in Consume
		rc.dial = MockRabbitMQDialWrapper("", &MockRabbitMQConnection{channel: &MockRabbitMQChannel{isConsumeError: true}}, false)
		err = rc.Serve()
		if err == nil {
			t.Fatalf("Expected error from Serve, got nil")
		}
		if err.Error() != "error getting consumer" {
			t.Fatalf("Expected error message to be 'error getting consumer', got %v", err.Error())
		}
		// Test when everything is set correctly
		mockPushable := &MockPushableForRabbitMQConsumer{
			incomingData: []*AppData{},
		}
		rc.pushable = mockPushable
		channel := make(chan rabbitmq.Delivery, 2)
		mockRabbitMQConnection := &MockRabbitMQConnection{
			channel: &MockRabbitMQChannel{
				channel: channel,
			},
		}
		rc.dial = MockRabbitMQDialWrapper("", mockRabbitMQConnection, false)
		rc.config = &RabbitMQConsumerConfig{}
		mockAcknowledger := &MockAcknowledger{}
		msg1 := rabbitmq.Delivery{
			Body:         []byte(`{"key":"value1"}`),
			Acknowledger: mockAcknowledger,
		}
		msg2 := rabbitmq.Delivery{
			Body:         []byte(`{"key":"value2"}`),
			Acknowledger: mockAcknowledger,
		}
		channel <- msg1
		channel <- msg2
		close(channel)
		err = rc.Serve()
		if err != nil {
			t.Fatalf("Expected no error from Serve, got %v", err)
		}
		if len(mockPushable.incomingData) != 2 {
			t.Fatalf("Expected incomingData to have 2 elements, got %v", len(mockPushable.incomingData))
		}
		incomingData := mockPushable.incomingData
		sort.Slice(incomingData, func(i, j int) bool {
			return incomingData[i].data.(map[string]any)["key"].(string) < incomingData[j].data.(map[string]any)["key"].(string)
		})
		for i := 0; i < 2; i++ {
			appData := incomingData[i]
			if dataMap, ok := appData.data.(map[string]any); !ok {
				t.Fatalf("Expected data to be a map, got %T", appData.data)
			} else {
				if !reflect.DeepEqual(dataMap, map[string]any{"key": "value" + strconv.Itoa(i+1)}) {
					t.Fatalf("Expected data to be %v, got %v", map[string]any{"key": "value" + strconv.Itoa(i+1)}, dataMap)
				}
			}
		}
		// Tests when there is an error in sendChannelOfRabbitMQDeliveryToPushable
		mockPushable = &MockPushableForRabbitMQConsumer{
			incomingData: []*AppData{},
		}
		rc.pushable = mockPushable
		channel = make(chan rabbitmq.Delivery, 2)
		mockRabbitMQConnection = &MockRabbitMQConnection{
			channel: &MockRabbitMQChannel{
				channel: channel,
			},
		}
		rc.dial = MockRabbitMQDialWrapper("", mockRabbitMQConnection, false)
		msg1 = rabbitmq.Delivery{
			Body:         []byte(`{"key":"value1"}`),
			Acknowledger: mockAcknowledger,
		}
		msg2 = rabbitmq.Delivery{
			Body:         []byte(`{"key":"value2"`),
			Acknowledger: mockAcknowledger,
		}
		channel <- msg1
		channel <- msg2
		close(channel)
		err = rc.Serve()
		if err == nil {
			t.Fatalf("Expected no error from Serve, got %v", err)
		}
	})
}

// Test AMQPOneConsumerConfig
func TestAMQPOneConsumerConfig(t *testing.T) {
	t.Run("ImplementsConfig", func(t *testing.T) {
		ac := &AMQPOneConsumerConfig{}
		// Type assertion to check if AMQPOneConsumerConfig implements Config
		_, ok := interface{}(ac).(Config)
		if !ok {
			t.Errorf("Expected AMQPOneConsumerConfig to implement Config interface")
		}
	})
	t.Run("IngestConfig", func(t *testing.T) {
		ac := &AMQPOneConsumerConfig{}
		// Test when the AMQPOneConsumerConfig has no fields
		err := ac.IngestConfig(map[string]any{})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "invalid Connection - must be a string and must be set" {
			t.Errorf("Expected error message to be 'invalid Connection - must be a string and must be set', got %v", err.Error())
		}
		// Test when the AMQPOneConsumerConfig has the Connection field set but it is not a string
		err = ac.IngestConfig(map[string]any{"Connection": 1})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "invalid Connection - must be a string and must be set" {
			t.Errorf("Expected error message to be 'invalid Connection - must be a string and must be set', got %v", err.Error())
		}
		// Test when the AMQPOneConsumerConfig has a valid Connection field but the Queue field is not set
		err = ac.IngestConfig(map[string]any{"Connection": "test"})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "invalid Queue - must be a string and must be set" {
			t.Errorf("Expected error message to be 'invalid Queue - must be a string and must be set', got %v", err.Error())
		}
		// Tests when the AMQPOneConsumerConfig has a valid Connection and Queue field is set but not a string
		err = ac.IngestConfig(map[string]any{"Connection": "test", "Queue": 1})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "invalid Queue - must be a string and must be set" {
			t.Errorf("Expected error message to be 'invalid Queue - must be a string and must be set', got %v", err.Error())
		}
		// Tests when the AMQPOneConsumerConfig has a valid Connection and Queue field is set but OnValue is not a boolean
		err = ac.IngestConfig(map[string]any{"Connection": "test", "Queue": "test", "OnValue": 1})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "invalid OnValue - must be a boolean" {
			t.Errorf("Expected error message to be 'invalid OnValue - must be a boolean', got %v", err.Error())
		}
		// Test when the AMQPOneConsumerConfig has a valid Connection and Queue and OnValue is not set
		err = ac.IngestConfig(map[string]any{"Connection": "test", "Queue": "test"})
		if err != nil {
			t.Errorf("Expected no error from IngestConfig, got %v", err)
		}
		if ac.Connection != "test" {
			t.Errorf("Expected Connection to be 'test', got %v", ac.Connection)
		}
		if ac.Queue != "test" {
			t.Errorf("Expected Queue to be 'test', got %v", ac.Queue)
		}
		if ac.OnValue {
			t.Errorf("Expected OnValue to be false, got true")
		}
		// Test when the AMQPOneConsumerConfig has a valid Connection, Queue, and OnValue field set
		err = ac.IngestConfig(map[string]any{"Connection": "test", "Queue": "test", "OnValue": true})
		if err != nil {
			t.Errorf("Expected no error from IngestConfig, got %v", err)
		}
		if ac.Connection != "test" {
			t.Errorf("Expected Connection to be 'test', got %v", ac.Connection)
		}
		if ac.Queue != "test" {
			t.Errorf("Expected Queue to be 'test', got %v", ac.Queue)
		}
		if !ac.OnValue {
			t.Errorf("Expected OnValue to be true, got false")
		}
	})
}

// MockPushableForAMQPOneConsumer is a mock implementation of the Pushable interface
type MockPushableForAMQPOneConsumer struct {
	incomingData  chan *AppData
	isSendToError bool
}

func (p *MockPushableForAMQPOneConsumer) SendTo(data *AppData) error {
	if p.isSendToError {
		return errors.New("test error")
	}
	p.incomingData <- data
	return nil
}

// MockAMQPOneReceiver is a mock implementation of the AMQPOneReceiver interface
type MockAMQPOneReceiver struct {
	isReceiveError bool
	isAcceptError  bool
	receiveData   chan *amqp.Message
	cancel     context.CancelFunc
}

func (r *MockAMQPOneReceiver) Receive(ctx context.Context, opts *amqp.ReceiveOptions) (*amqp.Message, error) {
	if r.isReceiveError {
		return nil, errors.New("receive error")
	}
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case data, ok := <-r.receiveData:
			if !ok {
				r.cancel()
				continue
			}
			return data, nil
		}
	}
}

func (r *MockAMQPOneReceiver) Close(ctx context.Context) error {
	return nil
}

func (r *MockAMQPOneReceiver) AcceptMessage(ctx context.Context, msg *amqp.Message) error {
	if r.isAcceptError {
		return errors.New("accept error")
	}
	return nil
}

// MockAMQPOneSession is a mock implementation of the AMQPOneSession interface
type MockAMQPOneConsumerSession struct {
	isNewReceiverError bool
	receiver *MockAMQPOneReceiver
}

func (s *MockAMQPOneConsumerSession) Close(ctx context.Context) error {
	return nil
}

func (s *MockAMQPOneConsumerSession) NewReceiver(ctx context.Context, source string, opts *amqp.ReceiverOptions) (AMQPOneConsumerReceiver, error) {
	if s.isNewReceiverError {
		return nil, errors.New("new receiver error")
	}
	return s.receiver, nil
}

// MockAMQPOneConsumerConnection is a mock implementation of the AMQPOneConsumerConnection interface
type MockAMQPOneConsumerConnection struct {
	isNewSessionError bool
	session *MockAMQPOneConsumerSession
}

func (c *MockAMQPOneConsumerConnection) Close() error {
	return nil
}

func (c *MockAMQPOneConsumerConnection) NewSession(ctx context.Context, opts *amqp.SessionOptions) (AMQPOneConsumerSession, error) {
	if c.isNewSessionError {
		return nil, errors.New("new session error")
	}
	return c.session, nil
}

// MockAMQPOneDialWrapper is a mock implementation of the AMQPOneDialWrapper interface
func MockAMQPOneDialWrapper(mockAMQPOneConsumerConnection *MockAMQPOneConsumerConnection, isError bool) func(ctx context.Context, addr string, opts *amqp.ConnOptions) (AMQPOneConsumerConnection, error) {
	return func(ctx context.Context, addr string, opts *amqp.ConnOptions) (AMQPOneConsumerConnection, error) {
		if isError {
			return nil, errors.New("dial error")
		}
		return mockAMQPOneConsumerConnection, nil
	}
}


// Tests AMQPOneConsumer
func TestAMQPOneConsumer(t *testing.T) {
	t.Run("ImplementsSourceServer", func(t *testing.T) {
		ac := &AMQPOneConsumer{}
		// Type assertion to check if AMQPOneConsumer implements SourceServer
		_, ok := interface{}(ac).(SourceServer)
		if !ok {
			t.Errorf("Expected AMQPOneConsumer to implement SourceServer interface")
		}
	})
	t.Run("Setup", func(t *testing.T) {
		ac := &AMQPOneConsumer{}
		// Test when the config is not a AMQPOneConsumerConfig
		err := ac.Setup(&MockConfig{})
		if err == nil {
			t.Errorf("Expected error from Setup, got nil")
		}
		if err.Error() != "config is not an AMQPOneConsumerConfig" {
			t.Errorf("Expected error message to be 'config is not an AMQPOneConsumerConfig', got %v", err.Error())
		}
		// Test when the config is a AMQPOneConsumerConfig
		config := &AMQPOneConsumerConfig{}
		err = ac.Setup(config)
		if err != nil {
			t.Errorf("Expected no error from Setup, got %v", err)
		}
		if ac.config != config {
			t.Errorf("Expected config to be %v, got %v", config, ac.config)
		}
		if ac.dial == nil {
			t.Errorf("Expected dial to be set, got nil")
		}
	})
	t.Run("AddPushable", func(t *testing.T) {
		ac := &AMQPOneConsumer{}
		mockPushable := &MockPushable{}
		err := ac.AddPushable(mockPushable)
		if err != nil {
			t.Errorf("Expected no error from AddPushable, got %v", err)
		}
		if ac.pushable != mockPushable {
			t.Errorf("Expected AddPushable to set the Pushable, got %v", ac.pushable)
		}
		// Test when the Pushable is already set
		err = ac.AddPushable(mockPushable)
		if err == nil {
			t.Errorf("Expected error from AddPushable, got nil")
		}
	})
	t.Run("Serve", func(t *testing.T) {
		t.Run("Serve", func(t *testing.T) {
			ac := &AMQPOneConsumer{}
			// Test when the pushable is not set
			err := ac.Serve()
			if err == nil {
				t.Fatalf("Expected error from Serve, got nil")
			}
			if err.Error() != "Pushable not set" {
				t.Fatalf("Expected error message to be 'Pushable not set', got %v", err.Error())
			}
			// Test when the pushable is set but config is not set
			ac.pushable = &MockPushable{}
			err = ac.Serve()
			if err == nil {
				t.Fatalf("Expected error from Serve, got nil")
			}
			if err.Error() != "Config not set" {
				t.Fatalf("Expected error message to be 'Config not set', got %v", err.Error())
			}
			// Test when the pushable and config are set but the dial is not set
			ac.config = &AMQPOneConsumerConfig{}
			err = ac.Serve()
			if err == nil {
				t.Fatalf("Expected error from Serve, got nil")
			}
			if err.Error() != "Dialer not set" {
				t.Fatalf("Expected error message to be 'Dialer not set', got %v", err.Error())
			}
			// Test when the pushable, config, and dial are set but there is an error in dial
			ac.dial = MockAMQPOneDialWrapper(nil, true)
			err = ac.Serve()
			if err == nil {
				t.Fatalf("Expected error from Serve, got nil")
			}
			if err.Error() != "dial error" {
				t.Fatalf("Expected error message to be 'dial error', got %v", err.Error())
			}
			// Test when the pushable, config, and dial are set but there is an error in NewSession
			ac.dial = MockAMQPOneDialWrapper(&MockAMQPOneConsumerConnection{isNewSessionError: true}, false)
			err = ac.Serve()
			if err == nil {
				t.Fatalf("Expected error from Serve, got nil")
			}
			if err.Error() != "new session error" {
				t.Fatalf("Expected error message to be 'new session error', got %v", err.Error())
			}
			// Test when the pushable, config, and dial are set but there is an error in NewReceiver
			ac.dial = MockAMQPOneDialWrapper(&MockAMQPOneConsumerConnection{session: &MockAMQPOneConsumerSession{isNewReceiverError: true}}, false)
			err = ac.Serve()
			if err == nil {
				t.Fatalf("Expected error from Serve, got nil")
			}
			if err.Error() != "new receiver error" {
				t.Fatalf("Expected error message to be 'new receiver error', got %v", err.Error())
			}
			// Test when there is an error in Receive
			ac.dial = MockAMQPOneDialWrapper(&MockAMQPOneConsumerConnection{session: &MockAMQPOneConsumerSession{receiver: &MockAMQPOneReceiver{isReceiveError: true}}}, false)
			err = ac.Serve()
			if err == nil {
				t.Fatalf("Expected error from Serve, got nil")
			}
			if err.Error() != "receive error" {
				t.Fatalf("Expected error message to be 'test error', got %v", err.Error())
			}
			// Test when there is an error in convertAndSendAMQPMessageToPushable 
			pushable := &MockPushable{isSendToError: true}
			receiverSendChan := make(chan *amqp.Message, 1)
			ac.dial = MockAMQPOneDialWrapper(&MockAMQPOneConsumerConnection{session: &MockAMQPOneConsumerSession{receiver: &MockAMQPOneReceiver{receiveData: receiverSendChan}}}, false)
			ac.pushable = pushable
			ac.config = &AMQPOneConsumerConfig{}
			receiverSendChan <- amqp.NewMessage([]byte(`{"key":"value"}`))
			err = ac.Serve()
			if err == nil {
				t.Fatalf("Expected error from Serve, got nil")
			}
			if err.Error() != "test error" {
				t.Fatalf("Expected error message to be 'test error', got %v", err.Error())
			}
			close(receiverSendChan)
			hangingData := <-receiverSendChan
			if hangingData != nil {
				t.Fatalf("Expected hangingData to be nil, got %v", hangingData)
			}
			// Test when there is an error in AcceptMessage
			pushable = &MockPushable{}
			receiverSendChan = make(chan *amqp.Message, 1)
			receiverSendChan <- amqp.NewMessage([]byte(`{"key":"value"}`))
			ac.dial = MockAMQPOneDialWrapper(&MockAMQPOneConsumerConnection{session: &MockAMQPOneConsumerSession{receiver: &MockAMQPOneReceiver{isAcceptError: true, receiveData: receiverSendChan}},}, false)
			ac.pushable = pushable
			err = ac.Serve()
			if err == nil {
				t.Fatalf("Expected error from Serve, got nil")
			}
			if err.Error() != "accept error" {
				t.Fatalf("Expected error message to be 'accept error', got %v", err.Error())
			}
			if pushable.incomingData == nil {
				t.Fatalf("Expected incomingData to be set, got nil")
			}
			gotAppData, err := pushable.incomingData.GetData()
			if err != nil {
				t.Fatalf("Expected no error from GetData, got %v", err)
			}
			if dataMap, ok := gotAppData.(map[string]any); !ok {
				t.Fatalf("Expected data to be a map, got %T", gotAppData)
			} else {
				if !reflect.DeepEqual(dataMap, map[string]any{"key": "value"}) {
					t.Fatalf("Expected data to be %v, got %v", map[string]any{"key": "value"}, dataMap)
				}
			}
			close(receiverSendChan)
			dataGone := <-receiverSendChan
			if dataGone != nil {
				t.Fatalf("Expected dataGone to be nil, got %v", dataGone)
			}
			// Test when everything is set correctly
			pushableAMQPOneConsumer := &MockPushableForAMQPOneConsumer{
				incomingData: make(chan *AppData, 2),
			}
			receiverSendChan = make(chan *amqp.Message, 2)
			finishCtx, cancel := context.WithCancel(context.Background())
			ac.finishCtx = finishCtx
			ac.dial = MockAMQPOneDialWrapper(&MockAMQPOneConsumerConnection{session: &MockAMQPOneConsumerSession{receiver: &MockAMQPOneReceiver{receiveData: receiverSendChan, cancel: cancel}},}, false)
			ac.pushable = pushableAMQPOneConsumer
			ac.config = &AMQPOneConsumerConfig{}
			receiverSendChan <- amqp.NewMessage([]byte(`{"key":"value1"}`))
			receiverSendChan <- amqp.NewMessage([]byte(`{"key":"value2"}`))
			close(receiverSendChan)
			err = ac.Serve()
			if err != nil {
				t.Fatalf("Expected no error from Serve, got %v", err)
			}
			close(pushableAMQPOneConsumer.incomingData)
			counter := 0
			for data := range pushableAMQPOneConsumer.incomingData {
				if data == nil {
					t.Fatalf("Expected data to be non-nil")
				}
				gotAppData, err := data.GetData()
				if err != nil {
					t.Fatalf("Expected no error from GetData, got %v", err)
				}
				if dataMap, ok := gotAppData.(map[string]any); !ok {
					t.Fatalf("Expected data to be a map, got %T", gotAppData)
				} else {
					if !reflect.DeepEqual(dataMap, map[string]any{"key": "value" + strconv.Itoa(counter+1)}) {
						t.Fatalf("Expected data to be %v, got %v", map[string]any{"key": "value" + strconv.Itoa(counter+1)}, dataMap)
					}
				}
				counter++
			}
			if counter != 2 {
				t.Fatalf("Expected counter to be 2, got %v", counter)
			}
			// Test when config indicates to extraxt data from Value field of AMQP message
			pushableAMQPOneConsumer = &MockPushableForAMQPOneConsumer{
				incomingData: make(chan *AppData, 2),
			}
			receiverSendChan = make(chan *amqp.Message, 2)
			finishCtx, cancel = context.WithCancel(context.Background())
			ac.finishCtx = finishCtx
			ac.dial = MockAMQPOneDialWrapper(&MockAMQPOneConsumerConnection{session: &MockAMQPOneConsumerSession{receiver: &MockAMQPOneReceiver{receiveData: receiverSendChan, cancel: cancel}},}, false)
			ac.pushable = pushableAMQPOneConsumer
			ac.config = &AMQPOneConsumerConfig{OnValue: true}
			msg1 := &amqp.Message{Value: `{"key":"value1"}`}
			msg2 := &amqp.Message{Value: `{"key":"value2"}`}
			receiverSendChan <- msg1
			receiverSendChan <- msg2
			close(receiverSendChan)
			err = ac.Serve()
			if err != nil {
				t.Fatalf("Expected no error from Serve, got %v", err)
			}
			close(pushableAMQPOneConsumer.incomingData)
			counter = 0
			for data := range pushableAMQPOneConsumer.incomingData {
				if data == nil {
					t.Fatalf("Expected data to be non-nil")
				}
				gotAppData, err := data.GetData()
				if err != nil {
					t.Fatalf("Expected no error from GetData, got %v", err)
				}
				if dataMap, ok := gotAppData.(map[string]any); !ok {
					t.Fatalf("Expected data to be a map, got %T", gotAppData)
				} else {
					if !reflect.DeepEqual(dataMap, map[string]any{"key": "value" + strconv.Itoa(counter+1)}) {
						t.Fatalf("Expected data to be %v, got %v", map[string]any{"key": "value" + strconv.Itoa(counter+1)}, dataMap)
					}
				}
				counter++
			}
			if counter != 2 {
				t.Fatalf("Expected counter to be 2, got %v", counter)
			}
		})
	})
}

// Tests amqpBytesDataConverter
func TestAMQPBytesDataConverter(t *testing.T) {
	// Test when message is nil
	_, err := amqpBytesDataConverter(nil)
	if err == nil {
		t.Errorf("Expected error from amqpBytesDataConverter, got nil")
	}
	if err.Error() != "message is nil" {
		t.Errorf("Expected error message to be 'message is nil', got %v", err.Error())
	}
	// Test when there is an error from convertBytesJSONDataToAppData
	msg := amqp.NewMessage([]byte(`{"key":"value"`))
	_, err = amqpBytesDataConverter(msg)
	if err == nil {
		t.Errorf("Expected error from amqpBytesDataConverter, got nil")
	}
	if err.Error() != "Bytes data is not a JSON map or an array" {
		t.Errorf("Expected error message to be 'Bytes data is not a JSON map or an array', got %v", err.Error())
	}
	// Test when there is no error
	msg = amqp.NewMessage([]byte(`{"key":"value"}`))
	appData, err := amqpBytesDataConverter(msg)
	if err != nil {
		t.Errorf("Expected no error from amqpBytesDataConverter, got %v", err)
	}
	if dataMap, ok := appData.data.(map[string]any); !ok {
		t.Errorf("Expected data to be a map, got %T", appData.data)
	} else {
		if !reflect.DeepEqual(dataMap, map[string]any{"key": "value"}) {
			t.Errorf("Expected data to be %v, got %v", map[string]any{"key": "value"}, dataMap)
		}
	}
}

// Tests amqpStringValueConverter
func TestAMQPStringValueConverter(t *testing.T) {
	// Test when message is nil
	_, err := amqpStringValueConverter(nil)
	if err == nil {
		t.Errorf("Expected error from amqpStringValueConverter, got nil")
	}
	if err.Error() != "message is nil" {
		t.Errorf("Expected error message to be 'message is nil', got %v", err.Error())
	}
	// Test when the value is nil
	msg := &amqp.Message{}
	_, err = amqpStringValueConverter(msg)
	if err == nil {
		t.Errorf("Expected error from amqpStringValueConverter, got nil")
	}
	if err.Error() != "value is nil" {
		t.Errorf("Expected error message to be 'value is nil', got %v", err.Error())
	}
	// Test when the value is not a string
	msg = &amqp.Message{Value: 1}
	_, err = amqpStringValueConverter(msg)
	if err == nil {
		t.Errorf("Expected error from amqpStringValueConverter, got nil")
	}
	if err.Error() != "value is not a string" {
		t.Errorf("Expected error message to be 'value is not a string', got %v", err.Error())
	}
	// Test when there is an error in convertStringJSONDataToAppData
	msg = &amqp.Message{Value: "test"}
	_, err = amqpStringValueConverter(msg)
	if err == nil {
		t.Errorf("Expected no error from amqpStringValueConverter, got %v", err)
	}
	if err.Error() != "String data is not a JSON map or an array" {
		t.Errorf("Expected error message to be 'String data is not a JSON map or an array', got %v", err.Error())
	}
	// Test when there is no error
	msg = &amqp.Message{Value: `{"key":"value"}`}
	appData, err := amqpStringValueConverter(msg)
	if err != nil {
		t.Errorf("Expected no error from amqpStringValueConverter, got %v", err)
	}
	if dataMap, ok := appData.data.(map[string]any); !ok {
		t.Errorf("Expected data to be a map, got %T", appData.data)
	} else {
		if !reflect.DeepEqual(dataMap, map[string]any{"key": "value"}) {
			t.Errorf("Expected data to be %v, got %v", map[string]any{"key": "value"}, dataMap)
		}
	}
}

// mockAMQPMessageConverter is a mock implementation of the AMQPMessageConverter function type
type mockAMQPMessageConverter struct {
	isError bool
}

func (c *mockAMQPMessageConverter) Convert(msg *amqp.Message) (*AppData, error) {
	if c.isError {
		return nil, errors.New("convert error")
	}
	return &AppData{data: map[string]any{"key": "value"}}, nil
}



// Tests convertAndSendAMQPDataToPushable
func TestConvertAndSendAMQPMessageToPushable(t *testing.T) {
	pushable := &MockPushable{}
	// Test when converter returns an error
	converter := &mockAMQPMessageConverter{isError: true}
	msg := &amqp.Message{}
	err := convertAndSendAMQPMessageToPushable(msg, pushable, converter.Convert)
	if err == nil {
		t.Errorf("Expected error from convertAndSendAMQPDataToPushable, got nil")
	}
	if err.Error() != "convert error" {
		t.Errorf("Expected error message to be 'convert error', got %v", err.Error())
	}
	// Tests when SendTo returns an error
	converter = &mockAMQPMessageConverter{}
	pushable = &MockPushable{isSendToError: true}
	err = convertAndSendAMQPMessageToPushable(msg, pushable, converter.Convert)
	if err == nil {
		t.Errorf("Expected error from convertAndSendAMQPDataToPushable, got nil")
	}
	if err.Error() != "test error" {
		t.Errorf("Expected error message to be 'test error', got %v", err.Error())
	}
	// Test when everything is set correctly
	pushable = &MockPushable{}
	err = convertAndSendAMQPMessageToPushable(msg, pushable, converter.Convert)
	if err != nil {
		t.Errorf("Expected no error from convertAndSendAMQPDataToPushable, got %v", err)
	}
	if pushable.incomingData == nil {
		t.Errorf("Expected incomingData to be set, got nil")
	}
	gotAppData, err := pushable.incomingData.GetData()
	if err != nil {
		t.Errorf("Expected no error from GetData, got %v", err)
	}
	if dataMap, ok := gotAppData.(map[string]any); !ok {
		t.Errorf("Expected data to be a map, got %T", gotAppData)
	} else {
		if !reflect.DeepEqual(dataMap, map[string]any{"key": "value"}) {
			t.Errorf("Expected data to be %v, got %v", map[string]any{"key": "value"}, dataMap)
		}
	}
}