package Server

import (
	"context"
	"errors"
	"sync"
	"time"

	amqp "github.com/Azure/go-amqp"
	rabbitmq "github.com/rabbitmq/amqp091-go"
)

// CONSUMERMAP is a map that maps a string to a Consumer.
var CONSUMERMAP = map[string]func() SourceServer{
	"RabbitMQ": func() SourceServer { return &RabbitMQConsumer{} },
    "AMQPOne":  func() SourceServer { return &AMQPOneConsumer{} },
}

// CONSUMERCONFIGMAP is a map that maps a string to a Config.
var CONSUMERCONFIGMAP = map[string]func() Config{
	"RabbitMQ": func() Config { return &RabbitMQConsumerConfig{} },
	"AMQPOne":  func() Config { return &AMQPOneConsumerConfig{} },
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

// sendBytesJSONDataToPushable is a function that sends bytes JSON data to an AppData.
//
// Args:
//
// 1. data: []byte. The data to send.
//
// 2. pushable: Pushable. The Pushable to send to.
//
// Returns:
//
// 1. error. An error if the process fails.
func sendBytesJSONDataToPushable(data []byte, pushable Pushable) error {
	appData, err := convertBytesJSONDataToAppData(data)
	if err != nil {
		return err
	}
	err = pushable.SendTo(appData)
	if err != nil {
		return err
	}
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
	return sendBytesJSONDataToPushable(msg.Body, pushable)
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
    defer cancel(nil)
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

// AMQPOneConsumerConfig is a struct that represents the configuration
// for an AMQP one consumer.
// It has the following fields:
//
// 1. Connection: string. The connection string for the AMQP server.
//
// 2. Queue: string. The name of the queue to consume from.
//
// 3. OnValue: bool. Whether to use the Value field of the AMQP message. Defaults to false.
//
// 4. MaxConcurrentMessages: int. The maximum number of concurrent messages to process. Defaults to 1.
type AMQPOneConsumerConfig struct {
	Connection string
	Queue      string
	OnValue bool
	MaxConcurrentMessages int
}

// IngestConfig is a method that will ingest the configuration
// for the AMQPOneConsumerConfig.
//
// Args:
//
// 1. config: map[string]any. The raw configuration for the AMQP one consumer.
//
// Returns:
//
// 1. error. An error if the process fails.
func (a *AMQPOneConsumerConfig) IngestConfig(config map[string]any) error {
	connection, ok := config["Connection"].(string)
	if !ok {
		return errors.New("invalid Connection - must be a string and must be set")
	}
	a.Connection = connection
	queue, ok := config["Queue"].(string)
	if !ok {
		return errors.New("invalid Queue - must be a string and must be set")
	}
	a.Queue = queue
	onValue, ok := config["OnValue"]
	if ok {
		if onValue, ok := onValue.(bool); !ok {
			return errors.New("invalid OnValue - must be a boolean")
		} else {
			a.OnValue = onValue
		}
	}
	maxConcurrentMessages, ok := config["MaxConcurrentMessages"]
	if ok {
		if maxConcurrentMessagesFloat, ok := maxConcurrentMessages.(float64); ok {
			a.MaxConcurrentMessages = int(maxConcurrentMessagesFloat)
		} else {
			if maxConcurrentMessages, ok := maxConcurrentMessages.(int); !ok {
				return errors.New("invalid MaxConcurrentMessages - must be a number")
			} else {
				a.MaxConcurrentMessages = maxConcurrentMessages
			}
		}
	} else {
		a.MaxConcurrentMessages = 1
	}
	return nil
}

// AMQPOneConsumerReceiver is an interface that represents a receiver for an AMQP one consumer.
type AMQPOneConsumerReceiver interface {
	Receive(ctx context.Context, opts *amqp.ReceiveOptions) (*amqp.Message, error)
	Close(ctx context.Context) error
	AcceptMessage(ctx context.Context, msg *amqp.Message) error
}

// AMQPOneConsumerSession is an interface that represents a session for an AMQP producer.
type AMQPOneConsumerSession interface {
	Close(ctx context.Context) error
	NewReceiver(ctx context.Context, source string, opts *amqp.ReceiverOptions) (AMQPOneConsumerReceiver, error)
}

// AMQPOneConsumerConnection is an interface that represents a connection for an AMQP producer.
type AMQPOneConsumerConnection interface {
	Close() error
	NewSession(ctx context.Context, opts *amqp.SessionOptions) (AMQPOneConsumerSession, error)
}

// AMQPOneConsumerDial is a function type that represents a dialer for an AMQP producer.
type AMQPOneConsumerDial func(ctx context.Context, addr string, opts *amqp.ConnOptions) (AMQPOneConsumerConnection, error)

// AMQPOneConsumerDialWrapper is a function that wraps the AMQP dial function.
// It takes in a context.Context, a string, and an *amqp.ConnOptions and returns
// an AMQPOneConsumerConnection and an error.
func AMQPOneConsumerDialWrapper(ctx context.Context, addr string, opts *amqp.ConnOptions) (AMQPOneConsumerConnection, error) {
	conn, err := amqp.Dial(ctx, addr, opts)
	if err != nil {
		return nil, err
	}
	return &AMQPOneConsumerConnectionWrapper{conn: conn}, nil
}

// AMQPOneConsumerConnectionWrapper is a struct that wraps an AMQP connection.
type AMQPOneConsumerConnectionWrapper struct {
	conn *amqp.Conn
}

// Close is a method that will close the AMQP connection.
func (a AMQPOneConsumerConnectionWrapper) Close() error {
	return a.conn.Close()
}

// NewSession is a method that will return a new AMQP session.
func (a AMQPOneConsumerConnectionWrapper) NewSession(ctx context.Context, opts *amqp.SessionOptions) (AMQPOneConsumerSession, error) {
	session, err := a.conn.NewSession(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &AMQPOneConsumerSessionWrapper{session: session}, nil
}

// AMQPOneConsumerSessionWrapper is a struct that wraps an AMQP session.
type AMQPOneConsumerSessionWrapper struct {
	session *amqp.Session
}

// Close is a method that will close the AMQP session.
func (a AMQPOneConsumerSessionWrapper) Close(ctx context.Context) error {
	return a.session.Close(ctx)
}

// NewReceiver is a method that will return a new AMQP receiver.
func (a AMQPOneConsumerSessionWrapper) NewReceiver(ctx context.Context, source string, opts *amqp.ReceiverOptions) (AMQPOneConsumerReceiver, error) {
	return a.session.NewReceiver(ctx, source, opts)
}

// AMQPOneConsumer is a struct that represents an AMQP one consumer.
// It has the following fields:
//
// 1. config: *AMQPOneConsumerConfig. The configuration for the AMQP one consumer.
//
// 2. pushable: Pushable. The pushable to push data to.
//
// 3. dial: AMQPOneConsumerDial. The dialer for the AMQP one consumer.
//
// 4. finishCtx: context.Context. The context for finishing the AMQP one consumer.
type AMQPOneConsumer struct {
	config   *AMQPOneConsumerConfig
	pushable Pushable
	dial     AMQPOneConsumerDial
    finishCtx context.Context
}

// Setup is a method that will set up the AMQP one consumer.
//
// Args:
//
// 1. config: Config. The configuration for the AMQP one consumer.
//
// Returns:
//
// 1. error. An error if the process fails.
func (a *AMQPOneConsumer) Setup(config Config) error {
	c, ok := config.(*AMQPOneConsumerConfig)
	if !ok {
		return errors.New("config is not an AMQPOneConsumerConfig")
	}
	a.config = c
	a.dial = AMQPOneConsumerDialWrapper
	return nil
}

// AddPushable is a method that will add a Pushable to the AMQP one consumer.
//
// Args:
//
// 1. pushable: Pushable. The Pushable to add.
//
// Returns:
//
// 1. error. An error if the process fails.
func (a *AMQPOneConsumer) AddPushable(pushable Pushable) error {
	if a.pushable != nil {
		return errors.New("Pushable already set")
	}
	a.pushable = pushable
	return nil
}

// Serve is a method that will start the AMQP one consumer and begin consuming messages.
// and sending them to the pushable, after which confirmation will be sent.
//
// Returns:
//
// 1. error. An error if the process fails.
func (a *AMQPOneConsumer) Serve() error {
	if a.pushable == nil {
		return errors.New("Pushable not set")
	}
	if a.config == nil {
		return errors.New("Config not set")
	}
	if a.dial == nil {
		return errors.New("Dialer not set")
	}
	var msgConvert amqpMessageConverter
	if a.config.OnValue {
		msgConvert = amqpConverters["StringValue"]
	} else {
		msgConvert = amqpConverters["BytesData"]
	}

	if a.finishCtx == nil {
        a.finishCtx = context.Background()
    }
	ctx := context.Background()
	conn, err := a.dial(ctx, a.config.Connection, nil)
	if err != nil {
		return err
	}
	defer conn.Close()
	session, err := conn.NewSession(ctx, nil)
	if err != nil {
		return err
	}
	defer session.Close(ctx)
	receiverOptions := &amqp.ReceiverOptions{
		Credit: int32(a.config.MaxConcurrentMessages),
	}
	receiver, err := session.NewReceiver(ctx, a.config.Queue, receiverOptions)
	if err != nil {
		return err
	}
	defer receiver.Close(ctx)
	sendOnErr, sendOnCancel := context.WithCancelCause(ctx)
	defer sendOnCancel(nil)
	wg := &sync.WaitGroup{}
	BREAK:
	for {
		select {
		case <-a.finishCtx.Done():
			break BREAK
		case <-sendOnErr.Done():
			return context.Cause(sendOnErr)
		default:
			timeoutCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
			msg, err := receiver.Receive(timeoutCtx, nil)
			cancel()
			if err != nil {
                if err == context.DeadlineExceeded {
                    continue
                }
				return err 
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := convertAndSendAMQPMessageToPushable(msg, a.pushable, msgConvert)
				if err != nil {
					sendOnCancel(err)
					return
				}
				err = receiver.AcceptMessage(ctx, msg)
				if err != nil {
					sendOnCancel(err)
					return
				}	
			}()

		}
	}
	wg.Wait()
    sendOnCancel(nil)
    // check if there was an error in the sendOnErr context
    // that occurred between receiving a finished signal and sending the message
    <- sendOnErr.Done()
    if context.Cause(sendOnErr) != context.Canceled {
        return context.Cause(sendOnErr)
    }
	return nil
}

// amqpMessageConverter is a function type that converts an AMQP message to an AppData.
//
// Args:
//
// 1. msg: *amqp.Message. The message to convert.
//
// Returns:
//
// 1. *AppData. The AppData that was converted.
//
// 2. error. An error if the process fails.
type amqpMessageConverter func(msg *amqp.Message) (*AppData, error)

// amqpBytesDataConverter is a function that converts an AMQP message to an AppData.
//
// Args:
//
// 1. msg: *amqp.Message. The message to convert.
//
// Returns:
//
// 1. *AppData. The AppData that was converted.
//
// 2. error. An error if the process fails.
func amqpBytesDataConverter(msg *amqp.Message) (*AppData, error) {
	if msg == nil {
		return nil, errors.New("message is nil")
	}
	appData, err := convertBytesJSONDataToAppData(msg.GetData())
	if err != nil {
		return nil, err
	}
	return appData, nil
}

// amqpStringValueConverter is a function that converts the Value field of an AMQP message to an AppData.
//
// Args:
//
// 1. msg: *amqp.Message. The message to convert.
//
// Returns:
//
// 1. *AppData. The AppData that was converted.
//
// 2. error. An error if the process fails.
func amqpStringValueConverter(msg *amqp.Message) (*AppData, error) {
	if msg == nil {
		return nil, errors.New("message is nil")
	}
	rawValue := msg.Value
	if rawValue == nil {
		return nil, errors.New("value is nil")
	}
	value, ok := rawValue.(string)
	if !ok {
		return nil, errors.New("value is not a string")
	}
	appData, err := convertStringJSONDataToAppData(value)
	if err != nil {
		return nil, err
	}
	return appData, nil
}

// amqpConverters is a map that maps a string to an amqpMessageConverter.
var amqpConverters = map[string]amqpMessageConverter{
	"BytesData": amqpBytesDataConverter,
	"StringValue": amqpStringValueConverter,
}

// convertAndSendAMQPMessageToPushable is a function that converts an AMQP message to an AppData
// and sends it to a Pushable.
//
// Args:
//
// 1. msg: *amqp.Message. The message to convert.
//
// 2. pushable: Pushable. The Pushable to send to.
//
// 3. converter: amqpMessageConverter. The converter to use.
//
// Returns:
//
// 1. error. An error if the process fails.
func convertAndSendAMQPMessageToPushable(msg *amqp.Message, pushable Pushable, converter amqpMessageConverter) error {
	appData, err := converter(msg)
	if err != nil {
		return err
	}
	err = pushable.SendTo(appData)
	if err != nil {
		return err
	}
	return nil
}