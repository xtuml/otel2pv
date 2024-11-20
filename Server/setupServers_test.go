package Server

import (
	"strings"
	"testing"
)

// TestAppConfig tests the IngestConfig method of the AppConfig struct
func TestAppConfig(t *testing.T) {
	t.Run("ImplementsConfig", func(t *testing.T) {
		ac := &AppConfig{}
		// Type assertion to check if AppConfig implements Config
		_, ok := interface{}(ac).(Config)
		if !ok {
			t.Errorf("Expected AppConfig to implement Config interface")
		}
	})
	t.Run("IngestConfig", func(t *testing.T) {
		appConfig := AppConfig{}
		// Test when the PipeServerConfig is not set
		err := appConfig.IngestConfig(map[string]interface{}{})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "PipeServerConfig must be set before calling IngestConfig" {
			t.Errorf("Expected error message to be 'PipeServerConfig must be set before calling IngestConfig', got %v", err.Error())
		}
		// Test when the PipeServerConfig is set but the AppConfig map field is not a map
		appConfig.PipeServerConfig = &MockConfig{}
		err = appConfig.IngestConfig(map[string]interface{}{"AppConfig": 1})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "AppConfig not set correctly - must map fields to values" {
			t.Errorf("Expected error message to be 'AppConfig not set correctly - must map fields to values', got %v", err.Error())
		}
		// Test when the PipeServerConfig is set and the AppConfig map field is a map but there is an error in the IngestConfig method of the PipeServerConfig
		appConfig.PipeServerConfig = &MockConfig{isError: true}
		err = appConfig.IngestConfig(map[string]interface{}{"AppConfig": map[string]interface{}{}})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		// Test when the PipeServerConfig is set and the AppConfig map field is a map and there is no error in the IngestConfig method of the PipeServerConfig
		// but the ProducersSetup field is not set correctly
		appConfig.PipeServerConfig = &MockConfig{}
		err = appConfig.IngestConfig(map[string]interface{}{"AppConfig": map[string]interface{}{}, "ProducersSetup": 1})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "ProducersSetup field not set correctly - must map fields to values" {
			t.Errorf("Expected error message to be 'ProducersSetup field not set correctly - must map fields to values', got %v", err.Error())
		}
		// Test when the PipeServerConfig is set and the AppConfig map field is a map and there is no error in the IngestConfig method of the PipeServerConfig
		// and the ProducersSetup field is set correctly but there is an error in the IngestConfig method of the SetupProducersConfig
		err = appConfig.IngestConfig(map[string]interface{}{"AppConfig": map[string]interface{}{}, "ProducersSetup": map[string]interface{}{}})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		// Test when the PipeServerConfig is set and the AppConfig map field is a map and there is no error in the IngestConfig method of the PipeServerConfig
		// and the ProducersSetup field is set correctly and there is no error in the IngestConfig method of the SetupProducersConfig
		// but the ConsumersSetup field is not set correctly
		err = appConfig.IngestConfig(map[string]interface{}{
			"AppConfig": map[string]interface{}{},
			"ProducersSetup": map[string]interface{}{
				"ProducerConfigs": []map[string]interface{}{
					{"Type": "HTTP", "ProducerConfig": map[string]interface{}{"URL": "http://localhost:8080"}},
				}},
			"ConsumersSetup": 1})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "ConsumersSetup field not set correctly - must map fields to values" {
			t.Errorf("Expected error message to be 'ConsumersSetup field not set correctly - must map fields to values', got %v", err.Error())
		}
		// Test when the PipeServerConfig is set and the AppConfig map field is a map and there is no error in the IngestConfig method of the PipeServerConfig
		// and the ProducersSetup field is set correctly and there is no error in the IngestConfig method of the SetupProducersConfig
		// and the ConsumersSetup field is set correctly but there is an error in the IngestConfig method of the SetupConsumersConfig
		err = appConfig.IngestConfig(map[string]interface{}{
			"AppConfig": map[string]interface{}{},
			"ProducersSetup": map[string]interface{}{
				"ProducerConfigs": []map[string]interface{}{
					{"Type": "HTTP", "ProducerConfig": map[string]interface{}{"URL": "http://localhost:8080"}},
				}},
			"ConsumersSetup": map[string]interface{}{}})
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if !strings.Contains(err.Error(), "ConsumersSetup subfields incorrect:") {
			t.Errorf("Expected error message to contain 'ConsumersSetup subfields incorrect', got %v", err.Error())
		}
		// Test when everything is set correctly
		// TODO: update this line once Consumers have been added
		// err = appConfig.IngestConfig(map[string]interface{}{
		// 	"AppConfig": map[string]interface{}{},
		// 	"ProducersSetup": map[string]interface{}{
		// 		"ProducerConfigs": []map[string]interface{}{
		// 			{"Type": "HTTP", "ProducerConfig": map[string]interface{}{"URL": "http://localhost:8080"}},
		// 		}},
		// 	"ConsumersSetup": map[string]interface{}{
		// 		"ConsumerConfigs": []map[string]interface{}{
		// 			{"Type": "HTTP", "ConsumerConfig": map[string]interface{}{"URL": "http://localhost:8080"}},
		// 		}}})
		// if err != nil {
		// 	t.Errorf("Expected no error from IngestConfig, got %v", err)
		// }
	})
}
