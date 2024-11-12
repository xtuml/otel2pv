package Server

// ProducerConfig is an interface that represents a component capable
// of ingesting configuration data.
type ProducerConfig interface {
	IngestConfig(any) error
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
