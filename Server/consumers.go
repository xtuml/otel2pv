package Server

// Consumer is an interface that represents a component capable
// of receiving data from a location based on the setup.
// Setup is used to configure the consumer, and it has a SourceServer
// interface that combines the Server and Pullable interfaces.
type Consumer interface {
	SourceServer
	Setup(ConsumerConfig) error
}

// ConsumerConfig is a base interface that represents
// configuration data for a consumer.
type ConsumerConfig interface {
	IngestConfig(map[string]any) error
}
