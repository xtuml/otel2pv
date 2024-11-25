package Server

import "errors"

type AppConfig struct {
	PipeServerConfig     Config
	SetupProducersConfig SetupProducersConfig
	SetupConsumersConfig SetupConsumersConfig
}

// IngestConfig is a method that will ingest the configuration
// for the AppConfig.
// It takes in a map[string]any and returns an error if the
// configuration is invalid.
func (a *AppConfig) IngestConfig(config map[string]any) error {
	// make sure PipeServerConfig is already set
	if a.PipeServerConfig == nil {
		return errors.New("PipeServerConfig must be set before calling IngestConfig")
	}
	pipeServerConfig, ok := config["AppConfig"].(map[string]any)
	if !ok {
		return errors.New("AppConfig not set correctly - must map fields to values")
	}
	err := a.PipeServerConfig.IngestConfig(pipeServerConfig)
	if err != nil {
		return err
	}

	producersConfig, ok := config["ProducersSetup"].(map[string]any)
	if !ok {
		return errors.New("ProducersSetup field not set correctly - must map fields to values")
	}
	producersConfigStruct := SetupProducersConfig{}
	err = producersConfigStruct.IngestConfig(producersConfig)
	if err != nil {
		return errors.New("ProducersSetup subfields incorrect:\n" + err.Error())
	}
	a.SetupProducersConfig = producersConfigStruct

	consumersConfig, ok := config["ConsumersSetup"].(map[string]any)
	if !ok {
		return errors.New("ConsumersSetup field not set correctly - must map fields to values")
	}
	consumersConfigStruct := SetupConsumersConfig{}
	err = consumersConfigStruct.IngestConfig(consumersConfig)
	if err != nil {
		return errors.New("ConsumersSetup subfields incorrect:\n" + err.Error())
	}
	a.SetupConsumersConfig = consumersConfigStruct
	return nil
}

// setupProducerWithConfigTypeAndMap
func setupProducerWithTypeConfigAndMap(config Config, configType string, serverMap map[string]func() SinkServer) (SinkServer, error) {
	serverGetFunc, ok := serverMap[configType]
	if !ok {
		return nil, errors.New("Producer type not found: " + configType)
	}
	server := serverGetFunc()
	err := server.Setup(config)
	if err != nil {
		return nil, err
	}
	return server, nil
}

// setupConsumerWithConfigTypeAndMap
func setupConsumerWithTypeConfigAndMap(config Config, configType string, serverMap map[string]func() SourceServer) (SourceServer, error) {
	serverGetFunc, ok := serverMap[configType]
	if !ok {
		return nil, errors.New("Consumer type not found: " + configType)
	}
	server := serverGetFunc()
	err := server.Setup(config)
	if err != nil {
		return nil, err
	}
	return server, nil
}

// setupProducerServer is a function that will setup the producer server.
// It takes as args:
// 1. producersConfig: SetupProducersConfig. The configuration for the producers.
// 2. producerMap: map[string]func() SinkServer. A map that maps a string to a func that produces SinkServer.
func setupProducerServer(producersConfig SetupProducersConfig, producerMap map[string]func() SinkServer) (SinkServer, error) {
	var producerServer SinkServer
	var err error
	if producersConfig.IsMapping {
		producerMapForMapSinkServer := map[string]SinkServer{}
		if len(producersConfig.SelectProducerConfigs) == 0 {
			return nil, errors.New("At least one producer must be set up when IsMapping is true")
		}
		for _, producerConfig := range producersConfig.SelectProducerConfigs {
			if producerConfig.Map == "" {
				return nil, errors.New("Producer map not set in producer config when IsMapping is true")
			}

			producerSubServer, err := setupProducerWithTypeConfigAndMap(
				producerConfig.ProducerConfig,
				producerConfig.Type,
				producerMap,
			)
			if err != nil {
				return nil, err
			}
			if _, ok := producerMapForMapSinkServer[producerConfig.Map]; ok {
				return nil, errors.New("Duplicate Map found for Producers: " + producerConfig.Map)
			}
			producerMapForMapSinkServer[producerConfig.Map] = producerSubServer
		}
		producerServer = &MapSinkServer{
			sinkServerMap: producerMapForMapSinkServer,
		}
	} else {
		if len(producersConfig.SelectProducerConfigs) != 1 {
			return nil, errors.New("Only one producer can be set up when IsMapping is false")
		}
		producerConfig := producersConfig.SelectProducerConfigs[0]
		producerServer, err = setupProducerWithTypeConfigAndMap(
			producerConfig.ProducerConfig,
			producerConfig.Type,
			producerMap,
		)
		if err != nil {
			return nil, err
		}
	}
	return producerServer, nil
}

// setupConsumerServer is a function that will setup the consumer server.
// It takes as args:
// 1. consumersConfig: SetupConsumersConfig. The configuration for the consumers.
// 2. consumerMap: map[string]func() SourceServer. A map that maps a string to a func that produces SourceServer.
func setupConsumerServer(consumersConfig SetupConsumersConfig, consumerMap map[string]func() SourceServer) (SourceServer, error) {
	if len(consumersConfig.SelectConsumerConfigs) != 1 {
		return nil, errors.New("Only one consumer can be set up")
	}
	consumerConfig := consumersConfig.SelectConsumerConfigs[0]
	consumerServer, err := setupConsumerWithTypeConfigAndMap(
		consumerConfig.ConsumerConfig,
		consumerConfig.Type,
		consumerMap,
	)
	if err != nil {
		return nil, err
	}
	return consumerServer, nil
}

// SetupApp is a method that will setup the application
// based on the configuration in the AppConfig.
// It takes as args:
// 1. appConfig: AppConfig. The configuration for the application.
// 2. pipeServer: *PipeServer. A pointer to the server that will handle the data.
// 3. consumerMap: map[string]func() SourceServer. A map that maps a string to a func that produces SourceServer.
// 4. producerMap: map[string]func() SinkServer. A map that maps a string to a func that produces StringServer.
// It returns:
// 1. SourceServer. The server that will handle the incoming data.
// 2. SinkServer. The server that will handle the outgoing data.
// 3. error. An error if the setup fails.
func SetupApp(appConfig AppConfig, pipeServer PipeServer, consumerMap map[string]func() SourceServer, producerMap map[string]func() SinkServer) (SourceServer, SinkServer, error) {
	// setup the pipe server
	err := pipeServer.Setup(appConfig.PipeServerConfig)
	if err != nil {
		return nil, nil, err
	}
	// setup the producer/s
	producerServer, err := setupProducerServer(appConfig.SetupProducersConfig, producerMap)
	if err != nil {
		return nil, nil, err
	}
	// setup the consumer
	consumerServer, err := setupConsumerServer(appConfig.SetupConsumersConfig, consumerMap)
	if err != nil {
		return nil, nil, err
	}
	// return the first producer and the first consumer
	return consumerServer, producerServer, nil
}
