package Server

import (
	"errors"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	rabbitmq "github.com/rabbitmq/amqp091-go"
)

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

// Tests for RabbitMQCompletionHandler
func TestRabbitMQCompletionHandler(t *testing.T) {
	t.Run("ImplementsCompletionHandler", func(t *testing.T) {
		rch := &RabbitMQCompletionHandler{}
		// Type assertion to check if RabbitMQCompletionHandler implements CompletionHandler
		_, ok := interface{}(rch).(CompletionHandler)
		if !ok {
			t.Errorf("Expected RabbitMQCompletionHandler to implement CompletionHandler interface")
		}
	})
	t.Run("Complete", func(t *testing.T) {
		rch := &RabbitMQCompletionHandler{}
		// Tests when the message is not set
		err := rch.Complete(nil, nil)
		if err == nil {
			t.Errorf("Expected error from Complete, got nil")
		}
		if err.Error() != "message not set" {
			t.Errorf("Expected error message to be 'message not set', got %v", err.Error())
		}
		// Test when there is an error but there is an error in Nack
		mockAcknowledger := &MockAcknowledger{
			isNackError: true,
		}
		msg := rabbitmq.Delivery{Acknowledger: mockAcknowledger}
		rch.message = &msg
		err = rch.Complete(nil, errors.New("test error"))
		if err == nil {
			t.Errorf("Expected error from Complete, got nil")
		}
		if err.Error() != "nack error" {
			t.Errorf("Expected error message to be 'nack error', got %v", err.Error())
		}
		// Test when there is an error but there is no error in Nack
		mockAcknowledger = &MockAcknowledger{}
		msg = rabbitmq.Delivery{Acknowledger: mockAcknowledger}
		rch.message = &msg
		err = rch.Complete(nil, errors.New("test error"))
		if err != nil {
			t.Errorf("Expected no error from Complete, got %v", err)
		}
		if !mockAcknowledger.hasNacknowledged {
			t.Errorf("Expected message to be nacked")
		}
		// Test when there is no error but there is an error in Ack
		mockAcknowledger = &MockAcknowledger{
			isError: true,
		}
		msg = rabbitmq.Delivery{Acknowledger: mockAcknowledger}
		rch.message = &msg
		err = rch.Complete(nil, nil)
		if err == nil {
			t.Errorf("Expected error from Complete, got nil")
		}
		if err.Error() != "acknowledge error" {
			t.Errorf("Expected error message to be 'acknowledge error', got %v", err.Error())
		}
		// Test when there is no error and there is no error in Ack
		mockAcknowledger = &MockAcknowledger{}
		msg = rabbitmq.Delivery{Acknowledger: mockAcknowledger}
		rch.message = &msg
		err = rch.Complete(nil, nil)
		if err != nil {
			t.Errorf("Expected no error from Complete, got %v", err)
		}
		if !mockAcknowledger.hasAcknowledged {
			t.Errorf("Expected message to be acknowledged")
		}
	})
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
		if appData.handler == nil {
			t.Errorf("Expected handler to be set, got nil")
		}
		if appData.handler.(*RabbitMQCompletionHandler).message != &msg {
			t.Errorf("Expected handler to be %v, got %v", &msg, appData.handler.(*RabbitMQCompletionHandler).message)
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
		msgs := []rabbitmq.Delivery{msg1, msg2}
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
			if appData.handler == nil {
				t.Fatalf("Expected handler to be set, got nil")
			}
			if appData.handler.(*RabbitMQCompletionHandler).message.Acknowledger != msgs[i].Acknowledger {
				t.Fatalf("Expected handler to be %v, got %v", msgs[i], appData.handler.(*RabbitMQCompletionHandler).message)
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
		msgs := []rabbitmq.Delivery{msg1, msg2}
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
			if appData.handler == nil {
				t.Fatalf("Expected handler to be set, got nil")
			}
			if appData.handler.(*RabbitMQCompletionHandler).message.Acknowledger != msgs[i].Acknowledger {
				t.Fatalf("Expected handler to be %v, got %v", msgs[i], appData.handler.(*RabbitMQCompletionHandler).message)
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
