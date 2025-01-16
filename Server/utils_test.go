package Server

import (
	"os"
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

// TestConvertStringToMap is a function that tests the convertStringToMap function.
func TestConvertStringToMap(t *testing.T) {
	data := `{"key":"value"}`
	expected := map[string]any{"key": "value"}
	actual, err := convertStringToMap(data)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected: %v, Actual: %v", expected, actual)
	}
	// Tests case where JSON is not a map
	data = `["value"]`
	_, err = convertStringToMap(data)
	if err == nil {
		t.Errorf("Expected error")
	}
	// Tests case where JSON is not valid
	data = `{"key":"value"`
	_, err = convertStringToMap(data)
	if err == nil {
		t.Errorf("Expected error")
	}
}

// TestConvertStringToArray is a function that tests the convertStringToArray function.
func TestConvertStringToArray(t *testing.T) {
	data := `["value"]`
	expected := []any{"value"}
	actual, err := convertStringToArray(data)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected: %v, Actual: %v", expected, actual)
	}
	// Tests case where JSON is not an array
	data = `{"key":"value"}`
	_, err = convertStringToArray(data)
	if err == nil {
		t.Errorf("Expected error")
	}
	// Tests case where JSON is not valid
	data = `["value"`
	_, err = convertStringToArray(data)
	if err == nil {
		t.Errorf("Expected error")
	}
}

// TestReadConfigJSON is a function that tests the ReadConfigJSON function.
func TestReadConfigJSON(t *testing.T) {
	// Tests case where file is read successfully
	tmpFile1, err := os.CreateTemp("", "test.json")
	if err != nil {
		t.Fatalf("Error: %s", err)
	}
	defer os.Remove(tmpFile1.Name())
	data := []byte(`{"key":"value"}`)
	err = os.WriteFile(tmpFile1.Name(), data, 0644)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	expected := map[string]any{"key": "value"}
	actual, err := ReadConfigJSON(tmpFile1.Name())
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected: %v, Actual: %v", expected, actual)
	}
	// Tests case where file is not found
	tmpFile2, err := os.CreateTemp("", "notfound.json")
	if err != nil {
		t.Fatalf("Error: %s", err)
	}
	defer os.Remove(tmpFile2.Name())
	_, err = ReadConfigJSON(tmpFile2.Name())
	if err == nil {
		t.Errorf("Expected error")
	}
	// Tests case where file is not valid
	tmpFile3, err := os.CreateTemp("", "invalid.json")
	if err != nil {
		t.Fatalf("Error: %s", err)
	}
	defer os.Remove(tmpFile3.Name())
	data = []byte(`{"key":"value"`)
	err = os.WriteFile(tmpFile3.Name(), data, 0644)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	_, err = ReadConfigJSON(tmpFile3.Name())
	if err == nil {
		t.Errorf("Expected error")
	}
}
