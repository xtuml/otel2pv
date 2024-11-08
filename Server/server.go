package Server

// Server is an interface for a server that will
// do some task/s. It has a Serve method that will
// start the server and return an error if this fails
type Server interface {
	Serve() error
}

// Pushable is an interface that represents a component capable
// of exposing a Receiver for data handling.
// GetReceiver returns the Receiver instance for handling incoming
// data, or an error if it fails.
type Pushable interface {
	GetReceiver() (Receiver, error)
}

// Pullable is an interface that can
// send data to and added receiver. It has an AddReceiver
// method that will add a receiver to the server
type Pullable interface {
	AddReceiver(receiver Receiver) error
}

// SourceServer is an interface that combines the Server and Pullable interfaces.
// It has a Serve method that will start the server and return an error if this fails,
// and an AddReceiver method that will add a receiver to the server and
// return an error if it fails.
type SourceServer interface {
	Pullable
	Server
}
