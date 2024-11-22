package Server

import (
	"errors"
	"sync"
)

// CompletionHandler is an interface for handling actions upon task completion.
// The Complete method is called when a task finishes, taking in the resulting data
// and an error, if one occurred, and returns an error if the handling fails.
type CompletionHandler interface {
	Complete(data any, err error) error
}

type WaitGroupCompletionHandler struct {
	// WaitGroupCompletionHandler is a struct that provides a way to wait for a task to complete
	// It has the following fields:
	// 1. wg: *sync.WaitGroup. A pointer to a WaitGroup that will be used to wait for the task to complete
	// 2. data: any. The data that will be returned when the task completes
	// 3. err: error. The error that will be returned when the task completes
	// 4. isComplete: bool. A flag that indicates whether the task has completed
	wg         *sync.WaitGroup
	data       any
	err        error
	isComplete bool
}

func (w *WaitGroupCompletionHandler) Complete(data any, err error) error {
	// Complete is a method that will be called when the task completes
	// It will set the data and error fields and decrement the WaitGroup
	// It returns an error if the WaitGroup is nil
	if w.wg == nil {
		return errors.New("WaitGroup not set")
	}
	if w.isComplete {
		return errors.New("task already complete")
	}
	w.data = data
	w.err = err
	w.isComplete = true
	w.wg.Done()
	return nil
}

func (w *WaitGroupCompletionHandler) GetData() (any, error) {
	// GetData is a method that returns the data field
	if !w.isComplete {
		return nil, errors.New("task not complete")
	}
	return w.data, nil
}

func (w *WaitGroupCompletionHandler) GetError() (error, error) {
	// GetError is a method that returns the error field
	// It returns an error if the task is not complete
	if !w.isComplete {
		return nil, errors.New("task not complete")
	}
	return w.err, nil
}

func NewWaitGroupCompletionHandler(wg *sync.WaitGroup) *WaitGroupCompletionHandler {
	// NewWaitGroupCompletionHandler is a function that creates a new WaitGroupCompletionHandler
	// It takes in a pointer to a WaitGroup and returns a pointer to a WaitGroupCompletionHandler
	// It adds 1 to the WaitGroup
	wg.Add(1)
	return &WaitGroupCompletionHandler{
		wg:         wg,
		isComplete: false,
	}
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

// GetHandler returns the CompletionHandler stored in the AppData struct.
func (a *AppData) GetHandler() (CompletionHandler, error) {
	if a.handler == nil {
		return nil, errors.New("handler not set")
	}
	return a.handler, nil
}

// NewAppData is a function that creates a new AppData struct.
func NewAppData(data any, handler CompletionHandler) *AppData {
	return &AppData{
		data:    data,
		handler: handler,
	}
}

// convertBytesJSONDataToAppData is a function that converts a byte slice to an AppData struct.
// It takes in a byte slice and a CompletionHandler and returns an AppData struct and an error.
func convertBytesJSONDataToAppData(message []byte, completionHandler CompletionHandler) (*AppData, error) {
	if jsonDataMap, err := convertBytesToMap(message); err == nil {
		appData := &AppData{
			data:    jsonDataMap,
			handler: completionHandler,
		}
		return appData, nil
	}
	if jsonDataArray, err := convertBytesToArray(message); err == nil {
		appData := &AppData{
			data:    jsonDataArray,
			handler: completionHandler,
		}
		return appData, nil
	}
	return nil, errors.New("Bytes data is not a JSON map or an array")
}
