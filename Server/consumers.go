package Server

import (
	"context"
	"errors"
	"sync"
	"time"

	rabbitmq "github.com/rabbitmq/amqp091-go"
)

// CONSUMERMAP is a map that maps a string to a Consumer.
var CONSUMERMAP = map[string]func() SourceServer{
	"RabbitMQ": func() SourceServer { return &RabbitMQConsumer{} },
}

// CONSUMERCONFIGMAP is a map that maps a string to a Config.
var CONSUMERCONFIGMAP = map[string]func() Config{
	"RabbitMQ": func() Config { return &RabbitMQConsumerConfig{} },
}

// SelectConsumerConfig is a struct that represents the configuration
// for selecting a consumer.
// It has the following fields:
//
// 1. Type: string. The type of consumer to select.
//
// 2. ConsumerConfig: Config. The configuration for the consumer.
type SelectConsumerConfig struct {
	Type           string
	ConsumerConfig Config
}

// IngestConfig is a method that will ingest the configuration
// for the SelectConsumerConfig.
// Its args are:
//
// 1. config: map[string]any. The raw configuration for the consumer.
//
// 2. configMap: map[string]func() Config. A map that maps a string to a func that produces Config.
//
// It returns:\
//
// 1. error. An error if the process fails.
func (s *SelectConsumerConfig) IngestConfig(config map[string]any, configMap map[string]func() Config) error {
	consumerType, ok := config["Type"].(string)
	if !ok {
		return errors.New("invalid Type - must be a string and must be set")
	}
	configMapFunc, ok := configMap[consumerType]
	if !ok {
		return errors.New("invalid consumer type: " + consumerType)
	}
	s.Type = consumerType
	consumerConfigStruct := configMapFunc()
	consumerConfig, ok := config["ConsumerConfig"].(map[string]any)
	if !ok {
		return errors.New("Consumer config not set correctly")
	}
	err := consumerConfigStruct.IngestConfig(consumerConfig)
	if err != nil {
		return errors.New("Consumer config not set correctly:\n" + err.Error())
	}
	s.ConsumerConfig = consumerConfigStruct
	return nil
}

// SetupConsumersConfig is a struct that represents the configuration
// for setting up consumers.
// It has the following fields:
//
// 1. Consumers: []*SelectConsumerConfig. A list of pointers to SelectConsumerConfigs
type SetupConsumersConfig struct {
	SelectConsumerConfigs []*SelectConsumerConfig
}

// IngestConfig is a method that will ingest the configuration
// for the SetupConsumersConfig.
//
// Its args are:
//
// 1. config: map[string]any. The raw configuration for the consumers.
//
// 2. configMap: map[string]func() Config. A map that maps a string to a func that produces Config.
//
// It returns:
//
// 1. error. An error if the process fails.
func (s *SetupConsumersConfig) IngestConfig(config map[string]any, configMap map[string]func() Config) error {
	selectConsumerConfigs, ok := config["ConsumerConfigs"].([]any)
	if !ok {
		return errors.New("ConsumerConfigs not set correctly")
	}
	if len(selectConsumerConfigs) == 0 {
		return errors.New("ConsumerConfigs is empty")
	}
	s.SelectConsumerConfigs = []*SelectConsumerConfig{}
	for _, selectConsumerConfig := range selectConsumerConfigs {
		selectConsumerConfig, ok := selectConsumerConfig.(map[string]any)
		if !ok {
			return errors.New("ConsumerConfig not set correctly")
		}
		selectConsumerConfigStruct := &SelectConsumerConfig{}
		err := selectConsumerConfigStruct.IngestConfig(selectConsumerConfig, configMap)
		if err != nil {
			return err
		}
		s.SelectConsumerConfigs = append(s.SelectConsumerConfigs, selectConsumerConfigStruct)
	}
	return nil
}

// RabbitMQConsumerConfig is a struct that represents the configuration
// for a RabbitMQ consumer.
// It has the following fields:
//
// 1. Connection: string. The connection string for the RabbitMQ server.
//
// 2. Queue: string. The name of the queue to consume from.
//
// 3. ConsumerTag: string. The consumer tag for the consumer. Defaults to "".
type RabbitMQConsumerConfig struct {
	Connection  string
	Queue       string
	ConsumerTag string
}

// IngestConfig is a method that will ingest the configuration
// for the RabbitMQConsumerConfig.
//
// Its args are:
//
// 1. config: map[string]any. The raw configuration for the RabbitMQ consumer.
//
// It returns:
//
// 1. error. An error if the process fails.
func (r *RabbitMQConsumerConfig) IngestConfig(config map[string]any) error {
	connection, ok := config["Connection"].(string)
	if !ok {
		return errors.New("invalid Connection - must be a string and must be set")
	}
	r.Connection = connection
	queue, ok := config["Queue"].(string)
	if !ok {
		return errors.New("invalid Queue - must be a string and must be set")
	}
	r.Queue = queue
	consumerTag, ok := config["ConsumerTag"]
	if !ok {
		r.ConsumerTag = ""
	} else {
		if consumerTag, ok := consumerTag.(string); ok {
			r.ConsumerTag = consumerTag
		} else {
			return errors.New("invalid ConsumerTag - must be a string")
		}
	}
	return nil
}

// RabbitMQChannel is an interface that represents a RabbitMQ channel.
type RabbitMQChannel interface {
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args rabbitmq.Table) (rabbitmq.Queue, error)
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args rabbitmq.Table) (<-chan rabbitmq.Delivery, error)
	Close() error
}

// RabbitMQConnection is an interface that represents a RabbitMQ connection.
type RabbitMQConnection interface {
	Channel() (RabbitMQChannel, error)
	Close() error
}

// RabbitMQDial is a function type that represents a dialer for RabbitMQ.
type RabbitMQDial func(url string) (RabbitMQConnection, error)

// RabbitMQDialWrapper is a function that wraps the RabbitMQ dial function.
// It takes in a string and returns a RabbitMQConnection and an error.
func RabbitMQDialWrapper(url string) (RabbitMQConnection, error) {
	conn, err := rabbitmq.Dial(url)
	if err != nil {
		return nil, err
	}
	return &RabbitMQConnectionWrapper{conn: conn}, nil
}

// RabbitMQConnectionWrapper is a struct that wraps a RabbitMQ connection.
type RabbitMQConnectionWrapper struct {
	conn *rabbitmq.Connection
}

// Channel is a method that will return a RabbitMQ channel.
func (r RabbitMQConnectionWrapper) Channel() (RabbitMQChannel, error) {
	return r.conn.Channel()
}

// Close is a method that will close the RabbitMQ connection.
func (r RabbitMQConnectionWrapper) Close() error {
	return r.conn.Close()
}

// RabbitMQConsumer is a struct that represents a RabbitMQ consumer.
// It has the following fields:
//
// 1. config: *RabbitMQConsumerConfig. The configuration for the RabbitMQ consumer.
//
// 2. pushable: Pushable. The pushable to push data to.
type RabbitMQConsumer struct {
	config   *RabbitMQConsumerConfig
	pushable Pushable
	dial     RabbitMQDial
}

// Setup is a method that will set up the RabbitMQ consumer.
//
// Its args are:
//
// 1. config: Config. The configuration for the RabbitMQ consumer.
//
// It returns:
//
// 1. error. An error if the process fails.
func (r *RabbitMQConsumer) Setup(config Config) error {
	c, ok := config.(*RabbitMQConsumerConfig)
	if !ok {
		return errors.New("config is not a RabbitMQConsumerConfig")
	}
	r.config = c
	r.dial = RabbitMQDialWrapper
	return nil
}

// AddPushable is a method that will add a Pushable to the RabbitMQ consumer.
// Its args are:
// 1. p: Pushable. The Pushable to add.
// It returns:
// 1. error. An error if the process fails.
func (r *RabbitMQConsumer) AddPushable(p Pushable) error {
	if r.pushable != nil {
		return errors.New("Pushable already set")
	}
	r.pushable = p
	return nil
}

// sendRabbitMQMessageDataToPushable is a function that sends RabbitMQ message data to a Pushable.
//
// Its args are:
//
// 1. msg: *rabbitmq.Delivery. The message to send.
//
// 2. pushable: Pushable. The Pushable to send to.
//
// It returns:
//
// 1. error. An error if the process fails.
func sendRabbitMQMessageDataToPushable(msg *rabbitmq.Delivery, pushable Pushable) error {
	appData, err := convertBytesJSONDataToAppData(msg.Body)
	if err != nil {
		return err
	}
	err = pushable.SendTo(appData)
	if err != nil {
		return err
	}
	return nil
}

// sendChannelOfRabbitMQDeliveryToPushable is a function that sends a channel of RabbitMQ deliveries to a Pushable.
//
// Its args are:
//
// 1. channel: <-chan rabbitmq.Delivery. The channel of RabbitMQ deliveries to send.
//
// 2. pushable: Pushable. The Pushable to send to.
//
// It returns:
//
// 1. error. An error if the process fails.
func sendChannelOfRabbitMQDeliveryToPushable(channel <-chan rabbitmq.Delivery, pushable Pushable) error {
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancelCause(context.Background())
	BREAK:
	for {
		select {
		case <-ctx.Done():
			break BREAK
		case msg, ok := <-channel:
			if !ok {
				break BREAK
			}
			wg.Add(1)
			go func(msg *rabbitmq.Delivery) {
				defer wg.Done()
				err := sendRabbitMQMessageDataToPushable(msg, pushable)
				if err != nil {
					cancel(err)
				}
				ackErr := msg.Ack(false)
				if ackErr != nil {
					cancel(ackErr)
				}
			}(&msg)
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
	wg.Wait()
	if ctx.Err() != nil {
		return context.Cause(ctx)
	}
	return nil
}

// Serve is a method that will start the RabbitMQ consumer and begin consuming messages.
//
// It returns:
//
// 1. error. An error if the process fails.
func (r *RabbitMQConsumer) Serve() error {
	if r.pushable == nil {
		return errors.New("Pushable not set")
	}
	if r.config == nil {
		return errors.New("Config not set")
	}
	if r.dial == nil {
		return errors.New("Dialer not set")
	}
	conn, err := r.dial(r.config.Connection)
	if err != nil {
		return err
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	q, err := ch.QueueDeclare(r.config.Queue, true, false, false, false, nil)
	if err != nil {
		return err
	}
	msgs, err := ch.Consume(q.Name, r.config.ConsumerTag, false, false, false, false, nil)
	if err != nil {
		return err
	}
	return sendChannelOfRabbitMQDeliveryToPushable(msgs, r.pushable)
}
