package Server

import (
	"reflect"
	"testing"
)

// Tests for AppData
func TestAppData(t *testing.T) {
	t.Run("GetData", func(t *testing.T) {
		testData := []byte("test data")
		appData := &AppData{data: testData}
		data, err := appData.GetData()
		if err != nil {
			t.Errorf("Expected no error from GetData, got %v", err)
		}
		if !reflect.DeepEqual(data, testData) {
			t.Errorf("Expected data to be '%v', got '%v'", testData, data)
		}
		// Test for nil data
		appData = &AppData{}
		_, err = appData.GetData()
		if err == nil {
			t.Errorf("Expected error from GetData, got nil")
		}
	})
	t.Run("GetRoutingKey", func(t *testing.T) {
		testRoutingKey := "test key"
		appData := &AppData{routingKey: testRoutingKey}
		routingKey, err := appData.GetRoutingKey()
		if err != nil {
			t.Errorf("Expected no error from GetRoutingKey, got %v", err)
		}
		if routingKey != testRoutingKey {
			t.Errorf("Expected routingKey to be '%v', got '%v'", testRoutingKey, routingKey)
		}
		// Test for empty routing key
		appData = &AppData{}
		routingKey, err = appData.GetRoutingKey()
		if err == nil {
			t.Errorf("Expected error from GetRoutingKey, got nil")
		}
		if routingKey != "" {
			t.Errorf("Expected routingKey to be '', got '%v'", routingKey)
		}
	})
}

// Tests for convertBytesJSONDataToAppData
func TestConvertBytesJSONDataToAppData(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// check when JSON is a map
		message := []byte(`{"key":"value"}`)

		appData, err := convertBytesJSONDataToAppData(message)

		if err != nil {
			t.Errorf("Expected no error from convertBytesJSONDataToAppData, got %v", err)
		}
		if appData.data == nil {
			t.Errorf("Expected data to be set, got nil")
		}
		if !reflect.DeepEqual(appData.data, message) {
			t.Errorf("Expected data to be '%v', got '%v'", message, appData.data)
		}
	})
	t.Run("Error", func(t *testing.T) {
		message := []byte(`{"key":"value"`)

		_, err := convertBytesJSONDataToAppData(message)

		if err == nil {
			t.Errorf("Expected error from convertBytesJSONDataToAppData, got nil")
		}
		if _, ok := err.(*InvalidError); !ok {
			t.Errorf("Expected error to be of type InvalidError, got %T", err)
		}
	})
}

// Tests for convertStringJSONDataToAppData
func TestConvertStringJSONDataToAppData(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// check when JSON is a map
		message := `{"key":"value"}`

		appData, err := convertStringJSONDataToAppData(message)

		if err != nil {
			t.Errorf("Expected no error from convertStringJSONDataToAppData, got %v", err)
		}
		if appData.data == nil {
			t.Errorf("Expected data to be set, got nil")
		}
		if !reflect.DeepEqual(appData.data, []byte(message)) {
			t.Errorf("Expected data to be '%v', got '%v'", []byte(message), appData.data)
		}
	})
	t.Run("Error", func(t *testing.T) {
		message := `{"key":"value"`

		_, err := convertStringJSONDataToAppData(message)

		if err == nil {
			t.Errorf("Expected error from convertStringJSONDataToAppData, got nil")
		}
		if _, ok := err.(*InvalidError); !ok {
			t.Errorf("Expected error to be of type InvalidError, got %T", err)
		}
	})
}
