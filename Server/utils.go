package Server

import "encoding/json"

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
