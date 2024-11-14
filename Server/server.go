package Server


// Server is an interface for a server that will
// do some task/s. It has a Serve method that will
// start the server and return an error if this fails
type Server interface {
	Serve() error
}

// Pushable is an interface that represents a component capable
// of being sent data.
// GetReceiver returns the Receiver instance for handling incoming
// data, or an error if it fails.
// SendTo is a method that will handle incoming data
// returning an error if it fails.
type Pushable interface {
	SendTo(data *AppData) error
}

// Pullable is an interface that can
// send data to a Pushable. It has an AddPushable
// method that will add a Pushable to the Pullable
type Pullable interface {
	AddPushable(pushable Pushable) error
}

// SourceServer is an interface that combines the Server and Pullable interfaces.
// It has a Serve method that will start the server and return an error if this fails,
// and an AddPushable method that will add a Pushable to the server and
// return an error if it fails.
type SourceServer interface {
	Pullable
	Server
}

// SinkServer is an interface that combines the Server and Pushable interfaces.
// It has a Serve method that will start the server and return an error if this fails,
// and a SendTo method that provides a way to send data to it
type SinkServer interface {
	Pushable
	Server
}

// PipeServer is an interface that combines the Server, Pullable and Pushable
// interfaces.
// It has a Serve method that will start the server and return an error if this fails,
// an AddPushable method that will add a Pushable to the server and return an error if it fails,
// and a SendTo method that provides a way to send data to it
type PipeServer interface {
	Pullable
	Pushable
	Server
}
