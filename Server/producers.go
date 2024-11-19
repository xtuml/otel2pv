package Server

import "errors"

// ProducerConfig is an interface that represents a component capable
// of ingesting configuration data.
type ProducerConfig interface {
	IngestConfig(map[string]any) error
}

// Producer is an interface that represents a component capable
// of sending data to a location based on the setup.
// Setup is used to configure the producer using a ProducerConfig
// interface, and it has a SinkServer interface that combines the
// Server and Pushable interfaces.
type Producer interface {
	Setup(ProducerConfig) error
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
		return errors.New("invalid URL")
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
