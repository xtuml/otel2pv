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
