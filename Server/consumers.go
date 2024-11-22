package Server

import (
	"context"
	"errors"
	"sync"

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
// It takes in a map[string]any and returns an error if the
// configuration is invalid.
func (s *SelectConsumerConfig) IngestConfig(config map[string]any) error {
	consumerType, ok := config["Type"].(string)
	if !ok {
		return errors.New("invalid Type - must be a string and must be set")
	}
	configMap, ok := CONSUMERCONFIGMAP[consumerType]
	if !ok {
		return errors.New("invalid consumer type: " + consumerType)
	}
	s.Type = consumerType
	consumerConfigStruct := configMap()
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
// It takes in a map[string]any and returns an error if the configuration
// is invalid.
func (s *SetupConsumersConfig) IngestConfig(config map[string]any) error {
	selectConsumerConfigs, ok := config["ConsumerConfigs"].([]map[string]any)
	if !ok {
		return errors.New("ConsumerConfigs not set correctly")
	}
	if len(selectConsumerConfigs) == 0 {
		return errors.New("ConsumerConfigs is empty")
	}
	s.SelectConsumerConfigs = []*SelectConsumerConfig{}
	for _, selectConsumerConfig := range selectConsumerConfigs {
		selectConsumerConfigStruct := &SelectConsumerConfig{}
		err := selectConsumerConfigStruct.IngestConfig(selectConsumerConfig)
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
// It takes in a map[string]any and returns an error if the configuration
// is invalid.
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

// RabbitMQCompletionHandler is a struct that represents a RabbitMQ completion handler.
// It has the following fields:
//
// 1. message: *rabbitmq.Delivery. The message to handle.
type RabbitMQCompletionHandler struct {
	message *rabbitmq.Delivery
}

// Complete is a method that will complete the message.
func (r *RabbitMQCompletionHandler) Complete(data any, err error) error {
	if r.message == nil {
		return errors.New("message not set")
	}
	if err != nil {
		nackErr := r.message.Nack(false, false)
		if nackErr != nil {
			return nackErr
		}
		return nil
	}
	ackErr := r.message.Ack(false)
	if ackErr != nil {
		return ackErr
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
// It takes in a Config and returns an error if the setup fails.
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
// It takes in a Pushable and returns an error if the Pushable is already set.
func (r *RabbitMQConsumer) AddPushable(p Pushable) error {
	if r.pushable != nil {
		return errors.New("Pushable already set")
	}
	r.pushable = p
	return nil
}

// sendRabbitMQMessageDataToPushable is a function that sends RabbitMQ message data to a Pushable.
// It takes in a *rabbitmq.Delivery and a Pushable and returns an error if the conversion to
// AppData fails or if the Pushable fails to send the data.
func sendRabbitMQMessageDataToPushable(msg *rabbitmq.Delivery, pushable Pushable) error {
	completionHandler := &RabbitMQCompletionHandler{message: msg}

	appData, err := convertBytesJSONDataToAppData(msg.Body, completionHandler)
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
func sendChannelOfRabbitMQDeliveryToPushable(channel <-chan rabbitmq.Delivery, pushable Pushable) error {
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancelCause(context.Background())
	for msg := range channel {
		if ctx.Err() != nil {
			break
		}
		wg.Add(1)
		go func(msg *rabbitmq.Delivery) {
			defer wg.Done()
			err := sendRabbitMQMessageDataToPushable(msg, pushable)
			if err != nil {
				cancel(err)
			}
		}(&msg)
	}
	wg.Wait()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return nil
}

// Serve is a method that will start the RabbitMQ consumer and begin consuming messages.
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
