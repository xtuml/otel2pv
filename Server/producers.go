package Server

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// Producer is an interface that represents a component capable
// of sending data to a location based on the setup.
// Setup is used to configure the producer using a ProducerConfig
// interface, and it has a SinkServer interface that combines the
// Server and Pushable interfaces.
type Producer interface {
	SinkServer
}

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
// It takes in a map[string]any and returns an error if the
// configuration is invalid.
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
// It takes in a ProducerConfig and returns an error if the setup fails.
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
	completeHandler, err := data.GetHandler()
	if err != nil {
		return err
	}
	dataForHandler := data.GetData()
	defer func() {
		errHandler := completeHandler.Complete(dataForHandler, err)
		if errHandler != nil {
			panic(errHandler)
		}
	}()
	gotData, ok := data.GetData().(map[string]any)
	if !ok {
		err = errors.New("invalid data")
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
var PRODUCERMAP = map[string]func() Producer{
	"HTTP": func() Producer {
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
// It takes in a map[string]any and returns an error if the configuration is invalid.
func (s *SelectProducerConfig) IngestConfig(config map[string]any) error {
	producerType, ok := config["Type"].(string)
	if !ok {
		return errors.New("invalid producer type")
	}
	s.Type = producerType
	configMap, ok := PRODUCERCONFIGMAP[s.Type]
	if !ok {
		return errors.New("invalid producer type: " + s.Type)
	}
	producerConfig, ok := config["ProducerConfig"].(map[string]any)
	if !ok {
		return errors.New("Producer config not set correctly")
	}
	producerConfigStruct := configMap()

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
