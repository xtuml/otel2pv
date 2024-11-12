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
// HandleIncomingData is a method that will handle incoming data
// and return an error if it fails.
type Pushable interface {
	GetReceiver() (Receiver, error)
	HandleIncomingData(data AppData) error
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

// SinkServer is an interface that combines the Server and Pushable interfaces.
// It has a Serve method that will start the server and return an error if this fails,
// and a GetReceiver method that will return the Receiver instance for handling incoming
// data, or an error if it fails.
type SinkServer interface {
	Pushable
	Server
}

// PipeServer is an interface that combines the Server, Pullable and Pushable
// interfaces.
// It has a Serve method that will start the server and return an error if this fails,
// an AddReceiver method that will add a receiver to the server and return an error if it fails,
// and a GetReceiver method that will return the Receiver instance for handling incoming data,
// or an error if it fails.
type PipeServer interface {
	Pullable
	Pushable
	Server
}

// HandlePushableDataReceipt will handle the data in the
// Recevier instance and call the HandleIncomingData method
func HandlePushableDataReceipt(p Pushable) error {
	receiver, err := p.GetReceiver()
	if err != nil {
		return err
	}
	outChan, err := receiver.GetOutChan()
	if err != nil {
		return err
	}
	for data := range outChan {
		err := p.HandleIncomingData(data)
		if err != nil {
			return err
		}
	}
	return nil
}
