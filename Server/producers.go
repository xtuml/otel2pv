package Server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	rabbitmq "github.com/rabbitmq/amqp091-go"

	amqp "github.com/Azure/go-amqp"

	backoff "github.com/cenkalti/backoff/v4"
)

// HTTPProducerConfig is a struct that represents the configuration
// for an HTTPProducer.
// It has the following fields:
//
// 1. URL: string. The URL to send the data to.
//
// 2. numRetries: int. The number of times to retry sending the data.
// This defaults to 3
//
// 3. timeout: int. The timeout for sending the data. This defaults to 10
//
// 4. initialRetryInterval: time.Duration. The initial retry interval for sending the data.
// This defaults to 1 second.
//
// 5. retryIntervalmultiplier: float64. The multiplier for the retry interval.
// This defaults to 1.0.
type HTTPProducerConfig struct {
	URL        string
	numRetries int
	timeout    int
	initialRetryInterval time.Duration
	retryIntervalmultiplier float64
}

// IngestConfig is a method that will ingest the configuration
// for the HTTPProducer.
//
// Its args are:
//
// 1. config: map[string]any. The raw configuration for the producer.
//
// It returns:
//
// 1. error. An error if the process fails.
func (h *HTTPProducerConfig) IngestConfig(config map[string]any) error {
	url, ok := config["URL"].(string)
	if !ok {
		return errors.New("invalid URL - must be a string and must be set")
	}
	h.URL = url
	numRetries, ok := config["numRetries"]
	if !ok {
		h.numRetries = 3
	} else {
        if numRetriesInt, ok := numRetries.(int); ok {
            h.numRetries = numRetriesInt
        } else {
            numRetriesFloat, ok := numRetries.(float64)
            if !ok {
                return errors.New("invalid numRetries - must be an integer")
            }
            numRetries := int(numRetriesFloat)
            h.numRetries = numRetries
        }
	}
	timeout, ok := config["timeout"]
	if !ok {
		h.timeout = 10
	} else {
		if timeoutInt, ok := timeout.(int); ok {
            h.timeout = timeoutInt
        } else {
            timeoutFloat, ok := timeout.(float64)
            if !ok {
                return errors.New("invalid timeout - must be an integer")
            }
            h.timeout = int(timeoutFloat)
        }
	}
	initialRetryInterval, ok := config["initialRetryInterval"]
	if !ok {
		h.initialRetryInterval = 1 * time.Second
	} else {
		initialRetryInterval, ok := initialRetryInterval.(float64)
		if !ok {
			return errors.New("invalid initialRetryInterval - must be an integer")
		}
		h.initialRetryInterval = time.Duration(initialRetryInterval) * time.Second
	}
	retryIntervalmultiplier, ok := config["retryIntervalmultiplier"]
	if !ok {
		h.retryIntervalmultiplier = 1.0
	} else {
		retryIntervalmultiplier, ok := retryIntervalmultiplier.(float64)
		if !ok {
			return errors.New("invalid retryIntervalmultiplier - must be an integer")
		}
		h.retryIntervalmultiplier = retryIntervalmultiplier
	}
	return nil
}

// HTTPProducer is a struct that represents an HTTP producer.
// It has the following fields:
//
// 1. config: *HTTPProducerConfig. A pointer to the configuration for the producer.
//
// 2. client: *http.Client. A pointer to the HTTP client that will be used to send the data.
type HTTPProducer struct {
	config *HTTPProducerConfig
	client *http.Client
}

// Setup is a method that will set up the HTTPProducer.
//
// Its args are:
//
// 1. config: Config. The configuration for the HTTPProducer.
//
// It returns:
//
// 1. error. An error if the process fails.
func (h *HTTPProducer) Setup(config Config) error {
	c, ok := config.(*HTTPProducerConfig)
	if !ok {
		return errors.New("invalid config")
	}
	h.config = c
	h.client = &http.Client{
		Timeout: time.Duration(h.config.timeout) * time.Second,
	}
	return nil
}

// Serve is a method that will start the HTTPProducer.
// It will return an error if the producer fails to send the data.
func (h *HTTPProducer) Serve() error {
	return nil
}

// SendTo is a method that will send data to the HTTPProducer to be sent
// to the configured URL as an HTTP JSON POST request.
// It takes in an *AppData and returns an error if the data is invalid.
func (h *HTTPProducer) SendTo(data *AppData) error {
	var err error
	gotData, err := data.GetData()
	if err != nil {
		return err
	}
	if !json.Valid(gotData) {
		return errors.New("data is not valid JSON")
	}
	req, err := http.NewRequest("POST", h.config.URL, bytes.NewBuffer(gotData))
	if err != nil {
		return NewSendErrorFromError(err)
	}
	req.Header.Set("Content-Type", "application/json")
	var resp *http.Response
	var backoffClient backoff.ExponentialBackOff
	for i := 0; i < h.config.numRetries; i++ {
		// try to send the data
		resp, err = h.client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			Logger.Info("Successfully sent message", slog.Group(
				"details", slog.String("producer", "HTTP"), slog.String("url", h.config.URL),
			))
			return nil
		}
		// backoff
		if i == 0 {
			backoffClient = backoff.ExponentialBackOff{
				InitialInterval: h.config.initialRetryInterval,
				Multiplier: 	h.config.retryIntervalmultiplier,
			}
		}
		backOff := backoffClient.NextBackOff()
		time.Sleep(backOff)
	}
	if resp == nil {
		return NewSendError(fmt.Sprintf("Failed to send data via http to %s, with no response", h.config.URL))
	}
	return NewSendError(fmt.Sprintf("Failed to send data via http to %s, with response %s", h.config.URL, resp.Status))
}

// ProducerMap is a map that contains functions to get Producers.
var PRODUCERMAP = map[string]func() SinkServer{
	"HTTP": func() SinkServer {
		return &HTTPProducer{}
	},
	"RabbitMQ": func() SinkServer {
		return &RabbitMQProducer{}
	},
	"AMQPOne": func() SinkServer {
		return &AMQPOneProducer{}
	},
}

// ProducerConfigMap is a map that contains functions to get ProducerConfigs.
var PRODUCERCONFIGMAP = map[string]func() Config{
	"HTTP": func() Config {
		return &HTTPProducerConfig{}
	},
	"RabbitMQ": func() Config {
		return &RabbitMQProducerConfig{}
	},
	"AMQPOne": func() Config {
		return &AMQPOneProducerConfig{}
	},
}

// SelectProducerConfig is a struct that contains the configuration
// for selecting a producer.
// It has the following fields:
//
// 1. Type: string. The type of producer to select.
//
// 2. Map: string. An optional string that is used to map data to a producer.
//
// 3. ProducerConfig: Config. Configuration for the producer.
type SelectProducerConfig struct {
	Type           string
	Map            string
	ProducerConfig Config
}

// IngestConfig is a method that will ingest the configuration for the
// SelectProducerConfig.
// Its args are:
// 1. config: map[string]any. The raw configuration for the producer.
// 2. configMap: map[string]func() Config. A map that maps a string to a func that produces Config.
// It returns:
// 1. error. An error if the process fails.
func (s *SelectProducerConfig) IngestConfig(config map[string]any, configMap map[string]func() Config) error {
	producerType, ok := config["Type"].(string)
	if !ok {
		return errors.New("invalid producer type")
	}
	s.Type = producerType
	configMapFunc, ok := configMap[s.Type]
	if !ok {
		return errors.New("invalid producer type: " + s.Type)
	}
	producerConfig, ok := config["ProducerConfig"].(map[string]any)
	if !ok {
		return errors.New("Producer config not set correctly")
	}
	producerConfigStruct := configMapFunc()

	err := producerConfigStruct.IngestConfig(producerConfig)
	if err != nil {
		return err
	}
	s.ProducerConfig = producerConfigStruct
	if producerMap, ok := config["Map"].(string); ok {
		s.Map = producerMap
	}
	return nil
}

// SetupProducersConfig is a struct that contains the configuration
// for setting up producers.
// It has the following fields:
//
// 1. IsMapping: bool. A flag that indicates whether the producers will have data
// mapped to them. Defaults to false.
//
// 2. SelectProducerConfigs: []*SelectProducerConfig. A slice of SelectProducerConfigs
// that will be used to set up the producers.
type SetupProducersConfig struct {
	IsMapping             bool
	SelectProducerConfigs []*SelectProducerConfig
}

// IngestConfig is a method that will ingest the configuration for the
// SetupProducersConfig.
// Its args are:
// 1. config: map[string]any. The raw configuration for the producers.
// 2. configMap: map[string]func() Config. A map that maps a string to a func that produces Config.
// It returns:
// 1. error. An error if the process fails.
func (s *SetupProducersConfig) IngestConfig(config map[string]any, configMap map[string]func() Config) error {
	isMapping, ok := config["IsMapping"]
	if ok {
		isMapping, ok := isMapping.(bool)
		if !ok {
			return errors.New("invalid IsMapping - must be a boolean")
		}
		s.IsMapping = isMapping
	}
	selectProducerConfigs, ok := config["ProducerConfigs"].([]any)
	if !ok {
		return errors.New("ProducerConfigs not set correctly")
	}
	if len(selectProducerConfigs) == 0 {
		return errors.New("ProducerConfigs is empty")
	}
	s.SelectProducerConfigs = []*SelectProducerConfig{}
	for _, selectProducerConfig := range selectProducerConfigs {
		selectProducerConfig, ok := selectProducerConfig.(map[string]any)
		if !ok {
			return errors.New("ProducerConfig not set correctly")
		}
		selectProducerConfigStruct := &SelectProducerConfig{}
		err := selectProducerConfigStruct.IngestConfig(selectProducerConfig, configMap)
		if err != nil {
			return err
		}
		s.SelectProducerConfigs = append(s.SelectProducerConfigs, selectProducerConfigStruct)
	}
	return nil
}

// RabbitMQProducerConfig is a struct that represents the configuration
// for a RabbitMQProducer.
// It has the following fields:
//
// 1. Connection: string. The connection string for the RabbitMQ server.
//
//  2. Exchange: string. The name of the exchange to send the data to.
//     This defaults to "".
//
// 3. RoutingKey: string. The routing key for the exchange.
type RabbitMQProducerConfig struct {
	Connection string
	Exchange   string
	RoutingKey string
}

// IngestConfig is a method that will ingest the configuration
// for the RabbitMQProducerConfig.
//
// Its args are:
//
// 1. config: map[string]any. The raw configuration for the RabbitMQ producer.
//
// It returns:
//
// 1. error. An error if the process fails.
func (r *RabbitMQProducerConfig) IngestConfig(config map[string]any) error {
	connection, ok := config["Connection"].(string)
	if !ok {
		return errors.New("invalid Connection - must be a string and must be set")
	}
	r.Connection = connection
	exchange, ok := config["Exchange"]
	if !ok {
		r.Exchange = ""
	} else {
		exchange, ok := exchange.(string)
		if !ok {
			return errors.New("invalid Exchange - must be a string")
		}
		r.Exchange = exchange
	}
	routingKey, ok := config["Queue"].(string)
	if !ok {
		return errors.New("invalid Queue - must be a string and must be set")
	}
	r.RoutingKey = routingKey
	return nil
}

// RabbitMQProducerChannel is an interface that represents a RabbitMQ channel.
type RabbitMQProducerChannel interface {
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args rabbitmq.Table) (rabbitmq.Queue, error)
	Publish(exchange, key string, mandatory, immediate bool, msg rabbitmq.Publishing) error
	Close() error
}

// RabbitMQProducerConnection is an interface that represents a RabbitMQ connection.
type RabbitMQProducerConnection interface {
	Channel() (RabbitMQProducerChannel, error)
	Close() error
}

// RabbitMQProducerDial is a function type that represents a dialer for RabbitMQ.
type RabbitMQProducerDial func(url string) (RabbitMQProducerConnection, error)

// RabbitMQProducerDialWrapper is a function that wraps the RabbitMQ dial function.
// It takes in a string and returns a RabbitMQProducerConnection and an error.
func RabbitMQProducerDialWrapper(url string) (RabbitMQProducerConnection, error) {
	conn, err := rabbitmq.Dial(url)
	if err != nil {
		return nil, err
	}
	return &RabbitMQProducerConnectionWrapper{conn: conn}, nil
}

// RabbitMQConnectionWrapper is a struct that wraps a RabbitMQ connection.
type RabbitMQProducerConnectionWrapper struct {
	conn *rabbitmq.Connection
}

// Channel is a method that will return a RabbitMQ channel.
func (r RabbitMQProducerConnectionWrapper) Channel() (RabbitMQProducerChannel, error) {
	return r.conn.Channel()
}

// Close is a method that will close the RabbitMQ connection.
func (r RabbitMQProducerConnectionWrapper) Close() error {
	return r.conn.Close()
}

// RabbitMQProducer is a struct that represents a RabbitMQ producer.
// It has the following fields:
//
// 1. config: *RabbitMQProducerConfig. A pointer to the configuration for the producer.
//
// 2. dial: RabbitMQProducerDial. The dialer for the RabbitMQ producer.
//
// 3. channel: RabbitMQProducerChannel. The channel for the RabbitMQ producer.
//
// 4. ctx: context.Context. The context for the RabbitMQ producer. Gives an appropriate
// way to cancel the producer.
//
// 5. cancel: context.CancelCauseFunc. The cancel function for the RabbitMQ producer. Gives
// an appropriate way to cancel the producer.
type RabbitMQProducer struct {
	config  *RabbitMQProducerConfig
	dial    RabbitMQProducerDial
	channel RabbitMQProducerChannel
	ctx     context.Context
	cancel  context.CancelCauseFunc
}

// Setup is a method that will set up the RabbitMQProducer.
//
// Its args are:
//
// 1. config: Config. The configuration for the RabbitMQ producer.
//
// It returns:
//
// 1. error. An error if the process fails.
func (r *RabbitMQProducer) Setup(config Config) error {
	c, ok := config.(*RabbitMQProducerConfig)
	if !ok {
		return errors.New("config is not a RabbitMQProducerConfig")
	}
	r.config = c
	r.dial = RabbitMQProducerDialWrapper
	ctx, cancel := context.WithCancelCause(context.Background())
	r.ctx = ctx
	r.cancel = cancel
	return nil
}

// Serve is a method that will start the RabbitMQProducer.
//
// It returns:
//
// 1. error. An error if the producer fails to send the data.
func (r *RabbitMQProducer) Serve() error {
	if r.config == nil {
		return errors.New("config not set")
	}
	if r.dial == nil {
		return errors.New("dial not set")
	}
	if r.ctx == nil {
		return errors.New("context not set")
	}
	if r.cancel == nil {
		return errors.New("context cancel not set")
	}
	conn, err := r.dial(r.config.Connection)
	if err != nil {
		return err
	}
	defer conn.Close()
	r.channel, err = conn.Channel()
	if err != nil {
		return err
	}
	defer r.channel.Close()
	_, err = r.channel.QueueDeclare(r.config.RoutingKey, true, false, false, false, nil)
	if err != nil {
		return err
	}

	<-r.ctx.Done()
	if context.Cause(r.ctx) != nil {
		if context.Cause(r.ctx).Error() == "context canceled" {
			return nil
		}
		return context.Cause(r.ctx)
	}
	return nil
}

// SendTo is a method that will send data to the RabbitMQProducer to be sent
// to the configured exchange and routing key.
//
// Its args are:
//
// 1. data: *AppData. The data to send.
//
// It returns:
//
// 1. error. An error if the data is invalid.
func (r *RabbitMQProducer) SendTo(data *AppData) error {
	if r.config == nil {
		return errors.New("config not set")
	}
	if r.ctx == nil {
		return errors.New("context not set")
	}
	if r.cancel == nil {
		return errors.New("context cancel not set")
	}
	if r.channel == nil {
		return errors.New("channel not set")
	}

	var err error
	gotData, err := data.GetData()
	if err != nil {
		return err
	}
	if !json.Valid(gotData) {
		return errors.New("data is not valid JSON")
	}
	err = r.channel.Publish(r.config.Exchange, r.config.RoutingKey, false, false, rabbitmq.Publishing{
		ContentType: "application/json",
		Body:        gotData,
	})
	if err != nil {
		err = NewSendErrorFromError(err)
		r.cancel(err)
		return err
	}
	Logger.Info("Successfully sent message", slog.Group(
		"details", slog.String("producer", "RabbitMQ"), slog.String("exchange", r.config.Exchange), slog.String("routing_key", r.config.RoutingKey),
	))
	return nil
}

// AMQPOneProducerConfig is a struct that represents the configuration
// for an AMQPOneProducer.
// It has the following fields:
//
// 1. Connection: string. The connection string for the AMQP server.
//
// 2. Queue: string. The name of the queue to send the data to.
type AMQPOneProducerConfig struct {
	Connection string
	Queue      string
}

// IngestConfig is a method that will ingest the configuration for the
// AMQPOneProducerConfig.
//
// Args:
//
// 1. config: map[string]any. The raw configuration for the producer.
//
// Returns:
//
// 1. error. An error if the process fails.
func (a *AMQPOneProducerConfig) IngestConfig(config map[string]any) error {
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
	return nil
}

// AMQPOneProducerSender is an interface that represents a sender for an AMQP producer.
type AMQPOneProducerSender interface {
	Close(ctx context.Context) error
	Send(ctx context.Context, msg *amqp.Message, opts *amqp.SendOptions) error
}

// AMQPOneProducerSession is an interface that represents a session for an AMQP producer.
type AMQPOneProducerSession interface {
	Close(ctx context.Context) error
	NewSender(ctx context.Context, target string, opts *amqp.SenderOptions) (AMQPOneProducerSender, error)
}

// AMQPOneProducerConnection is an interface that represents a connection for an AMQP producer.
type AMQPOneProducerConnection interface {
	Close() error
	NewSession(ctx context.Context, opts *amqp.SessionOptions) (AMQPOneProducerSession, error)
}

// AMQPOneProducerDial is a function type that represents a dialer for an AMQP producer.
type AMQPOneProducerDial func(ctx context.Context, addr string, opts *amqp.ConnOptions) (AMQPOneProducerConnection, error)

// AMQPOneProducerDialWrapper is a function that wraps the AMQP dial function.
// It takes in a context.Context, a string, and an *amqp.ConnOptions and returns
// an AMQPOneProducerConnection and an error.
func AMQPOneProducerDialWrapper(ctx context.Context, addr string, opts *amqp.ConnOptions) (AMQPOneProducerConnection, error) {
	conn, err := amqp.Dial(ctx, addr, opts)
	if err != nil {
		return nil, err
	}
	return &AMQPOneProducerConnectionWrapper{conn: conn}, nil
}

// AMQPOneProducerConnectionWrapper is a struct that wraps an AMQP connection.
type AMQPOneProducerConnectionWrapper struct {
	conn *amqp.Conn
}

// Close is a method that will close the AMQP connection.
func (a AMQPOneProducerConnectionWrapper) Close() error {
	return a.conn.Close()
}

// NewSession is a method that will return a new AMQP session.
func (a AMQPOneProducerConnectionWrapper) NewSession(ctx context.Context, opts *amqp.SessionOptions) (AMQPOneProducerSession, error) {
	session, err := a.conn.NewSession(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &AMQPOneProducerSessionWrapper{session: session}, nil
}

// AMQPOneProducerSessionWrapper is a struct that wraps an AMQP session.
type AMQPOneProducerSessionWrapper struct {
	session *amqp.Session
}

// Close is a method that will close the AMQP session.
func (a AMQPOneProducerSessionWrapper) Close(ctx context.Context) error {
	return a.session.Close(ctx)
}

// NewSender is a method that will return a new AMQP sender.
func (a AMQPOneProducerSessionWrapper) NewSender(ctx context.Context, target string, opts *amqp.SenderOptions) (AMQPOneProducerSender, error) {
	return a.session.NewSender(ctx, target, opts)
}

// AMQPOneProducer is a struct that represents an AMQP producer.
// It has the following fields:
//
// 1. config: *AMQPOneProducerConfig. A pointer to the configuration for the producer.
//
// 2. sender: AMQPOneProducerSender. The sender for the AMQP producer.
//
// 3. dial: AMQPOneProducerDial. The dialer for the AMQP producer.
//
// 4. ctx: context.Context. The context for the AMQP producer. Gives an appropriate
// way to cancel the producer.
//
// 5. cancel: context.CancelCauseFunc. The cancel function for the AMQP producer. Gives
// an appropriate way to cancel the producer.
type AMQPOneProducer struct {
	config *AMQPOneProducerConfig
	sender AMQPOneProducerSender
	dial   AMQPOneProducerDial
	ctx    context.Context
	cancel context.CancelCauseFunc
}

// Setup is a method that will set up the AMQPOneProducer.
//
// Args:
//
// 1. config: Config. The configuration for the AMQP producer.
//
// Returns:
//
// 1. error. An error if the process fails.
func (a *AMQPOneProducer) Setup(config Config) error {
	c, ok := config.(*AMQPOneProducerConfig)
	if !ok {
		return errors.New("config is not an AMQPOneProducerConfig")
	}
	a.config = c
	a.dial = AMQPOneProducerDialWrapper
	ctx, cancel := context.WithCancelCause(context.Background())
	a.ctx = ctx
	a.cancel = cancel
	return nil
}

// Serve is a method that will start the AMQPOneProducer.
//
// Returns:
//
// 1. error. An error if the producer fails to send the data.
func (a *AMQPOneProducer) Serve() error {
	if a.config == nil {
		return errors.New("config not set")
	}
	if a.dial == nil {
		return errors.New("dial not set")
	}
	if a.ctx == nil {
		return errors.New("context not set")
	}
	if a.cancel == nil {
		return errors.New("context cancel not set")
	}
	conn, err := a.dial(a.ctx, a.config.Connection, nil)
	if err != nil {
		return err
	}
	defer conn.Close()
	session, err := conn.NewSession(a.ctx, nil)
	if err != nil {
		return err
	}
	defer session.Close(a.ctx)
	sender, err := session.NewSender(a.ctx, a.config.Queue, nil)
	if err != nil {
		return err
	}
	defer sender.Close(a.ctx)
	a.sender = sender
	<-a.ctx.Done()
	if context.Cause(a.ctx) != nil {
		if context.Cause(a.ctx).Error() == "context canceled" {
			return nil
		}
		return context.Cause(a.ctx)
	}
	return nil
}

var contentType = "application/json"

// SendTo is a method that will send data to the AMQPOneProducer to be sent
// to the configured queue.
//
// Args:
//
// 1. data: *AppData. The data to send.
//
// Returns:
//
// 1. error. An error if the data is invalid.
func (a *AMQPOneProducer) SendTo(data *AppData) error {
	if a.config == nil {
		return errors.New("config not set")
	}
	if a.ctx == nil {
		return errors.New("context not set")
	}
	if a.cancel == nil {
		return errors.New("context cancel not set")
	}
	if a.sender == nil {
		return errors.New("sender not set")
	}

	var err error
	gotData, err := data.GetData()
	if err != nil {
		return err
	}
	if !json.Valid(gotData) {
		return errors.New("data is not valid JSON")
	}
	msg := amqp.NewMessage(gotData)
	msg.Properties = &amqp.MessageProperties{
		ContentType: &contentType,
	}
	err = a.sender.Send(a.ctx, msg, nil)
	if err != nil {
		err = NewSendErrorFromError(err)
		a.cancel(err)
		return err
	}
	Logger.Info("Successfully sent message", slog.Group(
		"details", slog.String("producer", "AMQP1.0"), slog.String("connection", a.config.Connection),  slog.String("queue", a.config.Queue),
	))
	return nil
}
