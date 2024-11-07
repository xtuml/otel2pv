package Server

// CompletionHandler is an interface for handling actions upon task completion.
// The Complete method is called when a task finishes, taking in the resulting data
// and an error, if one occurred, and returns an error if the handling fails.
type CompletionHandler interface {
	Complete(data any, err error) error
}

// AppData is a struct that holds data and a CompletionHandler for a task.
type AppData struct {
	data    any
	handler CompletionHandler
}

// GetData returns the data stored in the AppData struct.
func (a *AppData) GetData() any {
	return a.data
}
