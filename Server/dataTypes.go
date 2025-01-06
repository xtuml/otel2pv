package Server

import (
	"errors"
)

// AppData is a struct that holds data and a CompletionHandler for a task.
type AppData struct {
	data    any
	routingKey string
}

// GetData returns the data stored in the AppData struct.
//
// It returns:
//
// 1. any. The data.
//
// 2. error. An error if the data is not set.
func (a *AppData) GetData() (any, error) {
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
// 1. data: any. The data to be stored in the AppData struct.
//
// 2. routingKey: string. The routing key to be stored in the AppData struct, if any.
//
// It returns:
//
// 1. *AppData. A pointer to the AppData struct.
func NewAppData(data any, routingKey string) *AppData {
	return &AppData{
		data:    data,
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
	if jsonDataMap, err := convertBytesToMap(message); err == nil {
		appData := &AppData{
			data:    jsonDataMap,
		}
		return appData, nil
	}
	if jsonDataArray, err := convertBytesToArray(message); err == nil {
		appData := &AppData{
			data:    jsonDataArray,
		}
		return appData, nil
	}
	return nil, errors.New("Bytes data is not a JSON map or an array")
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
	if jsonDataMap, err := convertStringToMap(message); err == nil {
		appData := &AppData{
			data:    jsonDataMap,
		}
		return appData, nil
	}
	if jsonDataArray, err := convertStringToArray(message); err == nil {
		appData := &AppData{
			data:    jsonDataArray,
		}
		return appData, nil
	}
	return nil, errors.New("String data is not a JSON map or an array")
}