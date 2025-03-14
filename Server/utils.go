package Server

import (
	"encoding/json"
	"os"
	"strings"
)

// convertBytesToMap is a function that converts a byte slice to a map.
// It takes in a byte slice and returns a map[string]any and an error.
func convertBytesToMap(data []byte) (map[string]any, error) {
	var jsonData = map[string]any{}
	err := json.Unmarshal(data, &jsonData)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

// convertBytesToArray is a function that converts a byte slice to an array.
// It takes in a byte slice and returns an array of any and an error.
func convertBytesToArray(data []byte) ([]any, error) {
	var jsonData = []any{}
	err := json.Unmarshal(data, &jsonData)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

// convertStringToMap is a function that converts a string to a map.
// It takes in a string and returns a map[string]any and an error.
func convertStringToMap(data string) (map[string]any, error) {
	var jsonData = map[string]any{}
	err := json.NewDecoder(strings.NewReader(data)).Decode(&jsonData)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

// convertStringToArray is a function that converts a string to an array.
// It takes in a string and returns an array of any and an error.
func convertStringToArray(data string) ([]any, error) {
	var jsonData = []any{}
	err := json.NewDecoder(strings.NewReader(data)).Decode(&jsonData)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

// ReadConfigJSON is a function that reads a JSON file and returns a map.
//
// It takes as args:
//
// 1. path: string. The path to the JSON file.
//
// It returns:
//
// 1. map[string]any. The map that was read from the JSON file.
//
// 2. error. An error if ingestion fails.
func ReadConfigJSON(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return convertBytesToMap(data)
}

// Validator is an interface that defines a Validate method.
type Validator interface {
	Validate(v any) error
}

// DummyValidator is a struct that implements the Validator interface.
type DummyValidator struct{}

// Validate is a method that validates a value.
func (d DummyValidator) Validate(v any) error {
	return nil
}