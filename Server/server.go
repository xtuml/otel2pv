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

// ServersRun is a function that will start the servers
// and returns an error if any of them fail at any point.
// It takes in a SourceServer, PipeServer and SinkServer
// as arguments and adds the PipeServer to the SourceServer
// and the SinkServer to the PipeServer.
func ServersRun(sourceServer SourceServer, pipeServer PipeServer, sinkServer SinkServer) error {
	err := sourceServer.AddPushable(pipeServer)
	if err != nil {
		return err
	}
	err = pipeServer.AddPushable(sinkServer)
	if err != nil {
		return err
	}
	errChan := make(chan error)
	go func() {
		errChan <- sourceServer.Serve()
	}()
	go func() {
		errChan <- pipeServer.Serve()
	}()
	go func() {
		errChan <- sinkServer.Serve()
	}()
	counter := 0
	for err := range errChan {
		if err != nil {
			return err
		}
		counter++
		if counter == 3 {
			close(errChan)
			break
		}
	}
	return nil
}

// MapSinkServer is a struct that contains the required fields for the MapSinkServer
// It has the following fields:
// 1. sinkServerMap: map[string]SinkServer. A map that contains the SinkServers
type MapSinkServer struct {
	sinkServerMap map[string]SinkServer
}