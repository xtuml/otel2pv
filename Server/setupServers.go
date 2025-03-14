package Server

import (
	"errors"
	"os"
	"strings"

	"log/slog"
)
var lvl *slog.LevelVar
var handlerOptions *slog.HandlerOptions
var Logger *slog.Logger
var handler slog.Handler
var serverLogger slog.Logger

func init() {
	envLogLevel, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		envLogLevel = "DEBUG"
	}
	logLevel, err := getLogLevel(envLogLevel)
	if err != nil {
		panic(err)
	}
	lvl = new(slog.LevelVar)
	lvl.Set(logLevel)
	handlerOptions = &slog.HandlerOptions{
		Level: lvl,
	}
	envLogFile, ok := os.LookupEnv("LOG_FILE")
	if ok {
		file, err := os.OpenFile(envLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
		handler = slog.NewJSONHandler(file, handlerOptions)
	} else {
		handler = slog.NewJSONHandler(os.Stderr, handlerOptions)
	}
	Logger = slog.New(handler)
	serverLogger = *Logger.With(
		slog.Group(
			"logger_info",
			slog.String("logger_name", "Server"),
		),
	)
}




// getLogLevel is a function that will return the log level
// based on the log level string.
//
// Args:
//
// 1. logLevelStr: string. The log level string.
//
// Returns:
//
// 1. slog.Level. The log level.
//
// 2. error. An error if the log level string is invalid.
func getLogLevel(logLevelStr string) (slog.Level, error) {
	logLevelStr = strings.ToUpper(logLevelStr)
	switch logLevelStr {
	case "DEBUG":
		return slog.LevelDebug, nil
	case "INFO":
		return slog.LevelInfo, nil
	case "WARN":
		return slog.LevelWarn, nil
	case "ERROR":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, errors.New("Invalid log level: " + logLevelStr)
	}
}

type AppConfig struct {
	PipeServerConfig     Config
	SetupProducersConfig SetupProducersConfig
	SetupConsumersConfig SetupConsumersConfig
}

// IngestConfig is a method that will ingest the configuration
// for the AppConfig.
// Its args are:
//
// 1. config: map[string]any. The raw configuration for the application.
//
// 2. producerConfigMap: map[string]func() Config. A map that maps a string to a func that produces Config.
//
// 3. consumerConfigMap: map[string]func() Config. A map that maps a string to a func that produces Config.
//
// It returns:
//
// 1. error. An error if the process fails
func (a *AppConfig) IngestConfig(
	config map[string]any,
	producerConfigMap map[string]func() Config,
	consumerConfigMap map[string]func() Config,
) error {
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
	err = producersConfigStruct.IngestConfig(producersConfig, producerConfigMap)
	if err != nil {
		return errors.New("ProducersSetup subfields incorrect:\n" + err.Error())
	}
	a.SetupProducersConfig = producersConfigStruct

	consumersConfig, ok := config["ConsumersSetup"].(map[string]any)
	if !ok {
		return errors.New("ConsumersSetup field not set correctly - must map fields to values")
	}
	consumersConfigStruct := SetupConsumersConfig{}
	err = consumersConfigStruct.IngestConfig(consumersConfig, consumerConfigMap)
	if err != nil {
		return errors.New("ConsumersSetup subfields incorrect:\n" + err.Error())
	}
	a.SetupConsumersConfig = consumersConfigStruct
	return nil
}

// setupProducerWithConfigTypeAndMap is a function that will setup the producer server.
//
// It takes as args:
//
// 1. config: Config. The configuration for the producer.
//
// 2. configType: string. The type of the producer.
//
// 3. serverMap: map[string]func() SinkServer. A map that maps a string to a func that produces SinkServer.
//
// It returns:
//
// 1. SinkServer. The server that will handle the outgoing data.
//
// 2. error. An error if the setup fails.
func setupProducerWithTypeConfigAndMap(
	config Config,
	configType string,
	serverMap map[string]func() SinkServer,
) (SinkServer, error) {
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

// setupConsumerWithConfigTypeAndMap is a function that will setup the consumer server.
//
// It takes as args:
//
// 1. config: Config. The configuration for the consumer.
//
// 2. configType: string. The type of the consumer.
//
// 3. serverMap: map[string]func() SourceServer. A map that maps a string to a func that produces SourceServer.
//
// It returns:
//
// 1. SourceServer. The server that will handle the incoming data.
//
// 2. error. An error if the setup fails.
func setupConsumerWithTypeConfigAndMap(
	config Config,
	configType string,
	serverMap map[string]func() SourceServer,
) (SourceServer, error) {
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
//
// It takes as args:
//
// 1. producersConfig: SetupProducersConfig. The configuration for the producers.
//
// 2. producerMap: map[string]func() SinkServer. A map that maps a string to a func that produces SinkServer.
//
// It returns:
//
// 1. SinkServer. The server that will handle the outgoing data.
//
// 2. error. An error if the setup fails.
func setupProducerServer(
	producersConfig SetupProducersConfig,
	producerMap map[string]func() SinkServer,
) (SinkServer, error) {
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
//
// It takes as args:
//
// 1. consumersConfig: SetupConsumersConfig. The configuration for the consumers.
//
// 2. consumerMap: map[string]func() SourceServer. A map that maps a string to a func that produces SourceServer.
//
// It returns:
//
// 1. SourceServer. The server that will handle the incoming data.
//
// 2. error. An error if the setup fails.
func setupConsumerServer(
	consumersConfig SetupConsumersConfig,
	consumerMap map[string]func() SourceServer,
) (SourceServer, error) {
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
//
// It takes as args:
//
// 1. appConfig: AppConfig. The configuration for the application.
//
// 2. pipeServer: *PipeServer. A pointer to the server that will handle the data.
//
// 3. consumerMap: map[string]func() SourceServer. A map that maps a string to a func that produces SourceServer.
//
// 4. producerMap: map[string]func() SinkServer. A map that maps a string to a func that produces StringServer.
//
// It returns:
//
// 1. SourceServer. The server that will handle the incoming data.
//
// 2. SinkServer. The server that will handle the outgoing data.
//
// 3. error. An error if the setup fails.
func SetupApp(
	appConfig AppConfig,
	pipeServer PipeServer,
	consumerMap map[string]func() SourceServer,
	producerMap map[string]func() SinkServer,
) (SourceServer, SinkServer, error) {
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

// SetupAndRunApp is a method that will setup and run the application
// based on raw configuration and provided pipeserver, pipeserverconfig,
// config maps and producer and consumer maps.
//
// It takes as args:
//
// 1. config: map[string]any. The raw configuration for the application.
//
// 2. pipeServer: *PipeServer. A pointer to the server that will handle the data.
//
// 3. pipeServerConfig: Config. The configuration for the pipe server.
//
// 4. producerConfigMap: map[string]func() Config. A map that maps a string to a func that produces Config.
//
// 5. consumerConfigMap: map[string]func() Config. A map that maps a string to a func that produces Config.
//
// 6. producerMap: map[string]func() SinkServer. A map that maps a string to a func that produces SinkServer.
//
// 7. consumerMap: map[string]func() SourceServer. A map that maps a string to a func that produces SourceServer.
//
// It returns:
//
// 1. error. An error if anything fails.
func SetupAndRunApp(
	config map[string]any,
	pipeServer PipeServer,
	pipeServerConfig Config,
	producerConfigMap map[string]func() Config,
	consumerConfigMap map[string]func() Config,
	producerMap map[string]func() SinkServer,
	consumerMap map[string]func() SourceServer,
) error {
	appConfig := AppConfig{
		PipeServerConfig:     pipeServerConfig,
		SetupProducersConfig: SetupProducersConfig{},
		SetupConsumersConfig: SetupConsumersConfig{},
	}
	err := appConfig.IngestConfig(config, producerConfigMap, consumerConfigMap)
	if err != nil {
		return errors.New("Error ingesting config: " + err.Error())
	}
	consumerServer, producerServer, err := SetupApp(appConfig, pipeServer, consumerMap, producerMap)
	if err != nil {
		return errors.New("Error setting up app: " + err.Error())
	}
	err = ServersRun(consumerServer, pipeServer, producerServer)
	if err != nil {
		return errors.New("Error running app: " + err.Error())
	}
	return nil
}

// RunAppCLI is a function that will run the application with
// a flag --config that points to the configuration file.
//
// It takes as args:
// 1. configPath: string. The path to the configuration file.
//
// 2. pipeServer: PipeServer. A pointer to the server that will handle the data.
//
// 3. pipeServerConfig: Config. The configuration struct for the pipe server.
//
// 4. producerConfigMap: map[string]func() Config. A map that maps a string to a func that produces Config.
//
// 5. consumerConfigMap: map[string]func() Config. A map that maps a string to a func that produces Config.
//
// 6. producerMap: map[string]func() SinkServer. A map that maps a string to a func that produces SinkServer.
//
// 7. consumerMap: map[string]func() SourceServer. A map that maps a string to a func that produces SourceServer.
//
// It returns:
//
// 1. error. An error if anything fails.
func RunAppFromConfigPath(
	configPath string,
	pipeServer PipeServer,
	pipeServerConfig Config,
	producerConfigMap map[string]func() Config,
	consumerConfigMap map[string]func() Config,
	producerMap map[string]func() SinkServer,
	consumerMap map[string]func() SourceServer,
) error {
	config, err := ReadConfigJSON(configPath)
	if err != nil {
		return errors.New("Error reading config file: " + err.Error())
	}
	err = SetupAndRunApp(
		config,
		pipeServer,
		pipeServerConfig,
		producerConfigMap,
		consumerConfigMap,
		producerMap,
		consumerMap,
	)
	if err != nil {
		return err
	}
	return nil
}
