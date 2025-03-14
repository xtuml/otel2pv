package Server

import (
	"encoding/json"
	"errors"
)

// AppData is a struct that holds data and a CompletionHandler for a task.
type AppData struct {
	data       []byte
	routingKey string
}

// GetData returns the data stored in the AppData struct.
//
// It returns:
//
// 1. any. The data.
//
// 2. error. An error if the data is not set.
func (a *AppData) GetData() ([]byte, error) {
	if a.data == nil {
		return nil, errors.New("data is not set")
	}
	return a.data, nil
}

// GetRoutingKey returns the routing key stored in the AppData struct.
//
// It returns:
//
// 1. string. The routing key.
//
// 2. error. An error if the routing key is not set.
func (a *AppData) GetRoutingKey() (string, error) {
	if a.routingKey == "" {
		return "", errors.New("routing key is not set")
	}
	return a.routingKey, nil
}

// NewAppData is a function that creates a new AppData struct.
//
// It takes as args:
//
// 1. data: []byte. The data to be stored in the AppData struct.
//
// 2. routingKey: string. The routing key to be stored in the AppData struct, if any.
//
// It returns:
//
// 1. *AppData. A pointer to the AppData struct.
func NewAppData(data []byte, routingKey string) *AppData {
	return &AppData{
		data:       data,
		routingKey: routingKey,
	}
}

// convertBytesJSONDataToAppData is a function that converts a byte slice to an AppData struct.
//
// It takes as args:
//
// 1. message: []byte. The byte slice to convert.
//
// It returns:
//
// 1. *AppData. A pointer to the AppData struct.
//
// 2. error. An error if the conversion fails.
func convertBytesJSONDataToAppData(message []byte) (*AppData, error) {
	if !json.Valid(message) {
		return nil, NewInvalidError("Bytes data is not a valid JSON")
	}
	return &AppData{
		data: message,
	}, nil
}

// convertStringJSONDataToAppData is a function that converts a string to an AppData struct.
//
// Args:
//
// 1. message: string. The string to convert.
//
// Returns:
//
// 1. *AppData. A pointer to the AppData struct.
//
// 2. error. An error if the conversion fails.
func convertStringJSONDataToAppData(message string) (*AppData, error) {
	messageJSON := []byte(message)
	if !json.Valid(messageJSON) {
		return nil, NewInvalidError("String data is not a valid JSON")
	}
	return &AppData{
		data: messageJSON,
	}, nil
}
