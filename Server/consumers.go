package Server

import "errors"

// Consumer is an interface that represents a component capable
// of receiving data from a location based on the setup.
// Setup is used to configure the consumer, and it has a SourceServer
// interface that combines the Server and Pullable interfaces.
type Consumer interface {
	SourceServer
}

// CONSUMERMAP is a map that maps a string to a Consumer.
var CONSUMERMAP = map[string]func() Consumer{}

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
