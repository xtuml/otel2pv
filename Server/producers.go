package Server

// ProducerConfig is an interface that represents a component capable
// of ingesting configuration data.
type ProducerConfig interface {
	IngestConfig(any) error
}

// Producer is an interface that represents a component capable
// of sending data to a location based on the setup.
// Setup is used to configure the producer, Start is used to start
// the producer, and Send is used to send data to the location.
type Producer interface {
	Setup(ProducerConfig) error
	Start() error
	Send(any) error
}