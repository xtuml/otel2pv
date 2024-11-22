package Server

import (
	"reflect"
	"testing"
)

// TestConvertBytesToMap is a function that tests the convertBytesToMap function.
func TestConvertBytesToMap(t *testing.T) {
	data := []byte(`{"key":"value"}`)
	expected := map[string]any{"key": "value"}
	actual, err := convertBytesToMap(data)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected: %v, Actual: %v", expected, actual)
	}
	// Tests case where JSON is not a map
	data = []byte(`["value"]`)
	_, err = convertBytesToMap(data)
	if err == nil {
		t.Errorf("Expected error")
	}
	// Tests case where JSON is not valid
	data = []byte(`{"key":"value"`)
	_, err = convertBytesToMap(data)
	if err == nil {
		t.Errorf("Expected error")
	}
}

// TestConvertBytesToArray is a function that tests the convertBytesToArray function.
func TestConvertBytesToArray(t *testing.T) {
	data := []byte(`["value"]`)
	expected := []any{"value"}
	actual, err := convertBytesToArray(data)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected: %v, Actual: %v", expected, actual)
	}
	// Tests case where JSON is not an array
	data = []byte(`{"key":"value"}`)
	_, err = convertBytesToArray(data)
	if err == nil {
		t.Errorf("Expected error")
	}
	// Tests case where JSON is not valid
	data = []byte(`["value"`)
	_, err = convertBytesToArray(data)
	if err == nil {
		t.Errorf("Expected error")
	}
}
