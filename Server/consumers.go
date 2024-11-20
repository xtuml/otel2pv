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
var CONSUMERCONFIGMAP = map[string]func() Config{}

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
	s.Type = consumerType
	consumerConfig, ok := config["ConsumerConfig"].(map[string]any)
	if !ok {
		return errors.New("Consumer config not set correctly")
	}
	configMap, ok := CONSUMERCONFIGMAP[s.Type]
	if !ok {
		return errors.New("invalid consumer type: " + s.Type)
	}
	consumerConfigStruct := configMap()

	err := consumerConfigStruct.IngestConfig(consumerConfig)
	if err != nil {
		return err
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
