package Server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	rabbitmq "github.com/rabbitmq/amqp091-go"
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
type HTTPProducerConfig struct {
	URL        string
	numRetries int
	timeout    int
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
		numRetries, ok := numRetries.(int)
		if !ok {
			return errors.New("invalid numRetries - must be an integer")
		}
		h.numRetries = numRetries
	}
	timeout, ok := config["timeout"]
	if !ok {
		h.timeout = 10
	} else {
		timeout, ok := timeout.(int)
		if !ok {
			return errors.New("invalid timeout - must be an integer")
		}
		h.timeout = timeout
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
	jsonData, err := json.Marshal(gotData)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", h.config.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for i := 0; i < h.config.numRetries; i++ {
		// try to send the data
		resp, err := h.client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	err = errors.New("failed to send data")
	return err
}

// ProducerMap is a map that contains functions to get Producers.
var PRODUCERMAP = map[string]func() SinkServer{
	"HTTP": func() SinkServer {
		return &HTTPProducer{}
	},
}

// ProducerConfigMap is a map that contains functions to get ProducerConfigs.
var PRODUCERCONFIGMAP = map[string]func() Config{
	"HTTP": func() Config {
		return &HTTPProducerConfig{}
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
	ctx    context.Context
	cancel context.CancelCauseFunc
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
	jsonData, err := json.Marshal(gotData)
	if err != nil {
		return err
	}
	err = r.channel.Publish(r.config.Exchange, r.config.RoutingKey, false, false, rabbitmq.Publishing{
		ContentType: "application/json",
		Body:        jsonData,
	})
	if err != nil {
		r.cancel(err)
		return err
	}
	return nil
}
