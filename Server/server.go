package Server

// Server is an interface for a server that will
// do some task/s. It has a Serve method that will
// start the server and return an error if this fails
type Server interface {
	Serve() error
}


// Pullable is an interface that can
// send data to and added receiver. It has an AddReceiver
// method that will add a receiver to the server
type Pullable interface {
	AddReceiver(receiver Receiver) error
}