package Server

import (
	"strings"
	"testing"
)

// TestAppConfig tests the IngestConfig method of the AppConfig struct
func TestAppConfig(t *testing.T) {
	t.Run("IngestConfig", func(t *testing.T) {
		appConfig := AppConfig{}
		// Test when the PipeServerConfig is not set
		err := appConfig.IngestConfig(map[string]interface{}{}, PRODUCERCONFIGMAP, CONSUMERCONFIGMAP)
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "PipeServerConfig must be set before calling IngestConfig" {
			t.Errorf("Expected error message to be 'PipeServerConfig must be set before calling IngestConfig', got %v", err.Error())
		}
		// Test when the PipeServerConfig is set but the AppConfig map field is not a map
		appConfig.PipeServerConfig = &MockConfig{}
		err = appConfig.IngestConfig(map[string]interface{}{"AppConfig": 1}, PRODUCERCONFIGMAP, CONSUMERCONFIGMAP)
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "AppConfig not set correctly - must map fields to values" {
			t.Errorf("Expected error message to be 'AppConfig not set correctly - must map fields to values', got %v", err.Error())
		}
		// Test when the PipeServerConfig is set and the AppConfig map field is a map but there is an error in the IngestConfig method of the PipeServerConfig
		appConfig.PipeServerConfig = &MockConfig{isError: true}
		err = appConfig.IngestConfig(map[string]interface{}{"AppConfig": map[string]interface{}{}}, PRODUCERCONFIGMAP, CONSUMERCONFIGMAP)
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		// Test when the PipeServerConfig is set and the AppConfig map field is a map and there is no error in the IngestConfig method of the PipeServerConfig
		// but the ProducersSetup field is not set correctly
		appConfig.PipeServerConfig = &MockConfig{}
		err = appConfig.IngestConfig(map[string]interface{}{"AppConfig": map[string]interface{}{}, "ProducersSetup": 1}, PRODUCERCONFIGMAP, CONSUMERCONFIGMAP)
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if err.Error() != "ProducersSetup field not set correctly - must map fields to values" {
			t.Errorf("Expected error message to be 'ProducersSetup field not set correctly - must map fields to values', got %v", err.Error())
		}
		// Test when the PipeServerConfig is set and the AppConfig map field is a map and there is no error in the IngestConfig method of the PipeServerConfig
		// and the ProducersSetup field is set correctly but there is an error in the IngestConfig method of the SetupProducersConfig
		err = appConfig.IngestConfig(
			map[string]interface{}{"AppConfig": map[string]interface{}{},
			"ProducersSetup": map[string]interface{}{}}, PRODUCERCONFIGMAP, CONSUMERCONFIGMAP,
		)
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
			"ConsumersSetup": 1}, PRODUCERCONFIGMAP, CONSUMERCONFIGMAP)
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
			"ConsumersSetup": map[string]interface{}{}}, PRODUCERCONFIGMAP, CONSUMERCONFIGMAP)
		if err == nil {
			t.Errorf("Expected error from IngestConfig, got nil")
		}
		if !strings.Contains(err.Error(), "ConsumersSetup subfields incorrect:") {
			t.Errorf("Expected error message to contain 'ConsumersSetup subfields incorrect', got %v", err.Error())
		}
		// Test when everything is set correctly
		err = appConfig.IngestConfig(map[string]interface{}{
			"AppConfig": map[string]interface{}{},
			"ProducersSetup": map[string]interface{}{
				"ProducerConfigs": []map[string]interface{}{
					{"Type": "HTTP", "ProducerConfig": map[string]interface{}{"URL": "http://localhost:8080"}},
				}},
			"ConsumersSetup": map[string]interface{}{
				"ConsumerConfigs": []map[string]interface{}{
					{"Type": "RabbitMQ", "ConsumerConfig": map[string]interface{}{
						"Connection": "amqp://localhost:5672",
						"Queue":      "test",
					}},
				}}}, PRODUCERCONFIGMAP, CONSUMERCONFIGMAP)
		if err != nil {
			t.Errorf("Expected no error from IngestConfig, got %v", err)
		}
	})
}

// TestSetupProducerWithTypeConfigAndMap tests the setupProducerWithTypeConfigAndMap function
func TestSetupProducerWithTypeConfigAndMap(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		_, err := setupProducerWithTypeConfigAndMap(&MockConfig{}, "MockSink", map[string]func() SinkServer{"MockSink": func() SinkServer { return &MockSinkServer{} }})
		if err != nil {
			t.Errorf("Expected no error from setupProducerWithTypeConfigAndMap, got %v", err)
		}
	})
	t.Run("Error", func(t *testing.T) {
		// Test when the producer type is not found
		_, err := setupProducerWithTypeConfigAndMap(&MockConfig{}, "Invalid", map[string]func() SinkServer{})
		if err == nil {
			t.Errorf("Expected error from setupProducerWithTypeConfigAndMap, got nil")
		}
		if err.Error() != "Producer type not found: Invalid" {
			t.Errorf("Expected error message to be 'Producer type not found: Invalid', got %v", err.Error())
		}
		// Test when there is an error in the Setup method of the SinkServer
		_, err = setupProducerWithTypeConfigAndMap(&MockConfig{}, "MockSink", map[string]func() SinkServer{"MockSink": func() SinkServer { return &MockSinkServer{isSetupError: true} }})
		if err == nil {
			t.Errorf("Expected error from setupProducerWithTypeConfigAndMap, got nil")
		}
		// Test case where everything is set correctly
		sinkServer := &MockSinkServer{}
		outSinkServer, err := setupProducerWithTypeConfigAndMap(&MockConfig{}, "MockSink", map[string]func() SinkServer{"MockSink": func() SinkServer { return sinkServer }})
		if err != nil {
			t.Errorf("Expected no error from setupProducerWithTypeConfigAndMap, got %v", err)
		}
		if outSinkServer != sinkServer {
			t.Errorf("Expected outSinkServer to be the same as sinkServer")
		}
	})
}

// TestSetupConsumerWithTypeConfigAndMap tests the setupConsumerWithTypeConfigAndMap function
func TestSetupConsumerWithTypeConfigAndMap(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		_, err := setupConsumerWithTypeConfigAndMap(&MockConfig{}, "MockSource", map[string]func() SourceServer{"MockSource": func() SourceServer { return &MockSourceServer{} }})
		if err != nil {
			t.Errorf("Expected no error from setupConsumerWithTypeConfigAndMap, got %v", err)
		}
	})
	t.Run("Error", func(t *testing.T) {
		// Test when the consumer type is not found
		_, err := setupConsumerWithTypeConfigAndMap(&MockConfig{}, "Invalid", map[string]func() SourceServer{})
		if err == nil {
			t.Errorf("Expected error from setupConsumerWithTypeConfigAndMap, got nil")
		}
		if err.Error() != "Consumer type not found: Invalid" {
			t.Errorf("Expected error message to be 'Consumer type not found: Invalid', got %v", err.Error())
		}
		// Test when there is an error in the Setup method of the SourceServer
		_, err = setupConsumerWithTypeConfigAndMap(&MockConfig{}, "MockSource", map[string]func() SourceServer{"MockSource": func() SourceServer { return &MockSourceServer{isSetupError: true} }})
		if err == nil {
			t.Errorf("Expected error from setupConsumerWithTypeConfigAndMap, got nil")
		}
		// Test case where everything is set correctly
		sourceServer := &MockSourceServer{}
		outSourceServer, err := setupConsumerWithTypeConfigAndMap(&MockConfig{}, "MockSource", map[string]func() SourceServer{"MockSource": func() SourceServer { return sourceServer }})
		if err != nil {
			t.Errorf("Expected no error from setupConsumerWithTypeConfigAndMap, got %v", err)
		}
		if outSourceServer != sourceServer {
			t.Errorf("Expected outSourceServer to be the same as sourceServer")
		}
	})
}

// TestSetupProducerServer tests the setupProducerServer function
func TestSetupProducerServer(t *testing.T) {
	t.Run("IsMappingTrue", func(t *testing.T) {
		// Test when there are no producer configs
		setupProducersConfig := SetupProducersConfig{
			IsMapping:             true,
			SelectProducerConfigs: []*SelectProducerConfig{},
		}
		_, err := setupProducerServer(setupProducersConfig, map[string]func() SinkServer{})
		if err == nil {
			t.Errorf("Expected error from setupProducerServer, got nil")
		}
		if err.Error() != "At least one producer must be set up when IsMapping is true" {
			t.Errorf("Expected error message to be 'At least one producer config must be set up when IsMapping is true', got %v", err.Error())
		}
		// Test case where Map field is not set in a producer config
		setupProducersConfig = SetupProducersConfig{
			IsMapping: true,
			SelectProducerConfigs: []*SelectProducerConfig{
				{},
			},
		}
		_, err = setupProducerServer(setupProducersConfig, map[string]func() SinkServer{})
		if err == nil {
			t.Errorf("Expected error from setupProducerServer, got nil")
		}
		if err.Error() != "Producer map not set in producer config when IsMapping is true" {
			t.Errorf("Expected error message to be 'Producer map not set in producer config when IsMapping is true', got %v", err.Error())
		}
		// Tests when there is an error in the setupProducerWithTypeConfigAndMap function
		setupProducersConfig = SetupProducersConfig{
			IsMapping: true,
			SelectProducerConfigs: []*SelectProducerConfig{
				{
					Map:  "A",
					Type: "MockSink",
				},
			},
		}
		_, err = setupProducerServer(setupProducersConfig, map[string]func() SinkServer{})
		if err == nil {
			t.Errorf("Expected error from setupProducerServer, got nil")
		}
		if err.Error() != "Producer type not found: MockSink" {
			t.Errorf("Expected error message to be 'Producer type not found: MockSink', got %v", err.Error())
		}
		// Test case when there is a duplicate map
		setupProducersConfig = SetupProducersConfig{
			IsMapping: true,
			SelectProducerConfigs: []*SelectProducerConfig{
				{
					Map:  "A",
					Type: "MockSink",
				},
				{
					Map:  "A",
					Type: "MockSink",
				},
			},
		}
		_, err = setupProducerServer(setupProducersConfig, map[string]func() SinkServer{"MockSink": func() SinkServer { return &MockSinkServer{} }})
		if err == nil {
			t.Errorf("Expected error from setupProducerServer, got nil")
		}
		if err.Error() != "Duplicate Map found for Producers: A" {
			t.Errorf("Expected error message to be 'Duplicate Map found for Producers: A', got %v", err.Error())
		}
		// Test case where everything is set correctly
		setupProducersConfig = SetupProducersConfig{
			IsMapping: true,
			SelectProducerConfigs: []*SelectProducerConfig{
				{
					Map:  "A",
					Type: "MockSink",
				},
				{
					Map:  "B",
					Type: "MockSink",
				},
			},
		}
		mapProducerServer, err := setupProducerServer(setupProducersConfig, map[string]func() SinkServer{
			"MockSink": func() SinkServer { return &MockSinkServer{} },
		})
		if err != nil {
			t.Errorf("Expected no error from setupProducerServer, got %v", err)
		}
		if _, ok := mapProducerServer.(*MapSinkServer); !ok {
			t.Errorf("Expected mapProducerServer to be of type *MapSinkServer")
		}
		if len(mapProducerServer.(*MapSinkServer).sinkServerMap) != 2 {
			t.Errorf("Expected mapProducerServer to have 2 keys")
		}
		for _, key := range []string{"A", "B"} {
			sinkServer, ok := mapProducerServer.(*MapSinkServer).sinkServerMap[key]
			if !ok {
				t.Errorf("Expected mapProducerServer to have key %s", key)
			}
			if _, ok := sinkServer.(*MockSinkServer); !ok {
				t.Errorf("Expected sinkServer to be of type *MockSinkServer")
			}
		}
	})
	t.Run("IsMappingFalse", func(t *testing.T) {
		// Test when there are no producer configs
		setupProducersConfig := SetupProducersConfig{
			IsMapping:             false,
			SelectProducerConfigs: []*SelectProducerConfig{},
		}
		_, err := setupProducerServer(setupProducersConfig, map[string]func() SinkServer{})
		if err == nil {
			t.Errorf("Expected error from setupProducerServer, got nil")
		}
		if err.Error() != "Only one producer can be set up when IsMapping is false" {
			t.Errorf("Expected error message to be 'Only one producer can be set up when IsMapping is false', got %v", err.Error())
		}
		// Tests case when there are multiple producer configs
		setupProducersConfig = SetupProducersConfig{
			IsMapping: false,
			SelectProducerConfigs: []*SelectProducerConfig{
				{}, {},
			},
		}
		_, err = setupProducerServer(setupProducersConfig, map[string]func() SinkServer{})
		if err == nil {
			t.Errorf("Expected error from setupProducerServer, got nil")
		}
		if err.Error() != "Only one producer can be set up when IsMapping is false" {
			t.Errorf("Expected error message to be 'Only one producer can be set up when IsMapping is false', got %v", err.Error())
		}
		// Tests case where setupProducerWithTypeConfigAndMap returns an error
		setupProducersConfig = SetupProducersConfig{
			IsMapping: false,
			SelectProducerConfigs: []*SelectProducerConfig{
				{
					Type: "MockSink",
				},
			},
		}
		_, err = setupProducerServer(setupProducersConfig, map[string]func() SinkServer{})
		if err == nil {
			t.Errorf("Expected error from setupProducerServer, got nil")
		}
		if err.Error() != "Producer type not found: MockSink" {
			t.Errorf("Expected error message to be 'Producer type not found: MockSink', got %v", err.Error())
		}
		// Test case where everything is set correctly
		setupProducersConfig = SetupProducersConfig{
			IsMapping: false,
			SelectProducerConfigs: []*SelectProducerConfig{
				{
					Type: "MockSink",
				},
			},
		}
		producerServer, err := setupProducerServer(setupProducersConfig, map[string]func() SinkServer{
			"MockSink": func() SinkServer { return &MockSinkServer{} },
		})
		if err != nil {
			t.Errorf("Expected no error from setupProducerServer, got %v", err)
		}
		if _, ok := producerServer.(*MockSinkServer); !ok {
			t.Errorf("Expected producerServer to be of type *MockSinkServer")
		}
	})
}

// TestSetupConsumerServer tests the setupConsumerServer function
func TestSetupConsumerServer(t *testing.T) {
	// Tests case when there are no consumer configs
	setupConsumersConfig := SetupConsumersConfig{
		SelectConsumerConfigs: []*SelectConsumerConfig{},
	}
	_, err := setupConsumerServer(setupConsumersConfig, map[string]func() SourceServer{})
	if err == nil {
		t.Errorf("Expected error from setupConsumerServer, got nil")
	}
	if err.Error() != "Only one consumer can be set up" {
		t.Errorf("Expected error message to be 'Only one consumer can be set up', got %v", err.Error())
	}
	// Tests case when there are multiple consumer configs
	setupConsumersConfig = SetupConsumersConfig{
		SelectConsumerConfigs: []*SelectConsumerConfig{
			{}, {},
		},
	}
	_, err = setupConsumerServer(setupConsumersConfig, map[string]func() SourceServer{})
	if err == nil {
		t.Errorf("Expected error from setupConsumerServer, got nil")
	}
	if err.Error() != "Only one consumer can be set up" {
		t.Errorf("Expected error message to be 'Only one consumer can be set up', got %v", err.Error())
	}
	// Tests case when there is an error in the setupConsumerWithTypeConfigAndMap function
	setupConsumersConfig = SetupConsumersConfig{
		SelectConsumerConfigs: []*SelectConsumerConfig{
			{
				Type: "Invalid",
			},
		},
	}
	_, err = setupConsumerServer(setupConsumersConfig, map[string]func() SourceServer{})
	if err == nil {
		t.Errorf("Expected error from setupConsumerServer, got nil")
	}
	if err.Error() != "Consumer type not found: Invalid" {
		t.Errorf("Expected error message to be 'Consumer type not found: Invalid', got %v", err.Error())
	}
	// Test case where everything is set correctly
	setupConsumersConfig = SetupConsumersConfig{
		SelectConsumerConfigs: []*SelectConsumerConfig{
			{
				Type: "MockSource",
			},
		},
	}
	sourceServer, err := setupConsumerServer(setupConsumersConfig, map[string]func() SourceServer{
		"MockSource": func() SourceServer { return &MockSourceServer{} },
	})
	if err != nil {
		t.Errorf("Expected no error from setupConsumerServer, got %v", err)
	}
	if _, ok := sourceServer.(*MockSourceServer); !ok {
		t.Errorf("Expected sourceServer to be of type *MockSourceServer")
	}
}

// TestSetupApp
func TestSetupApp(t *testing.T) {
	// Tests case where the pipeServer Setup method returns an error
	pipeServer := &MockPipeServer{isSetupError: true}
	_, _, err := SetupApp(AppConfig{PipeServerConfig: &MockConfig{}}, pipeServer, map[string]func() SourceServer{}, map[string]func() SinkServer{})
	if err == nil {
		t.Errorf("Expected error from SetupApp, got nil")
	}
	if err.Error() != "test error" {
		t.Errorf("Expected error message to be 'test error', got %v", err.Error())
	}
	// Tests case where the setupProducerServer function returns an error
	pipeServer = &MockPipeServer{}
	_, _, err = SetupApp(AppConfig{
		PipeServerConfig: &MockConfig{},
		SetupProducersConfig: SetupProducersConfig{
			IsMapping:             false,
			SelectProducerConfigs: []*SelectProducerConfig{},
		},
		SetupConsumersConfig: SetupConsumersConfig{},
	}, pipeServer, map[string]func() SourceServer{}, map[string]func() SinkServer{})
	if err == nil {
		t.Errorf("Expected error from SetupApp, got nil")
	}
	if err.Error() != "Only one producer can be set up when IsMapping is false" {
		t.Errorf("Expected error message to be 'At least one producer must be set up when IsMapping is false', got %v", err.Error())
	}
	// Tests case where the setupConsumerServer function returns an error
	pipeServer = &MockPipeServer{}
	_, _, err = SetupApp(
		AppConfig{
			PipeServerConfig: &MockConfig{},
			SetupProducersConfig: SetupProducersConfig{
				IsMapping: false,
				SelectProducerConfigs: []*SelectProducerConfig{
					{
						Type:           "MockSink",
						ProducerConfig: &MockConfig{},
					},
				},
			},
			SetupConsumersConfig: SetupConsumersConfig{
				SelectConsumerConfigs: []*SelectConsumerConfig{},
			},
		},
		pipeServer,
		map[string]func() SourceServer{},
		map[string]func() SinkServer{"MockSink": func() SinkServer { return &MockSinkServer{} }},
	)
	if err == nil {
		t.Errorf("Expected error from SetupApp, got nil")
	}
	if err.Error() != "Only one consumer can be set up" {
		t.Errorf("Expected error message to be 'Only one consumer can be set up', got %v", err.Error())
	}
	// Tests case where everything is set correctly
	pipeServer = &MockPipeServer{}
	mockSourceServer := &MockSourceServer{}
	MockSinkServer := &MockSinkServer{}
	sourceServer, sinkServer, err := SetupApp(
		AppConfig{
			PipeServerConfig: &MockConfig{},
			SetupProducersConfig: SetupProducersConfig{
				IsMapping: false,
				SelectProducerConfigs: []*SelectProducerConfig{
					{
						Type:           "MockSink",
						ProducerConfig: &MockConfig{},
					},
				},
			},
			SetupConsumersConfig: SetupConsumersConfig{
				SelectConsumerConfigs: []*SelectConsumerConfig{
					{
						Type:           "MockSource",
						ConsumerConfig: &MockConfig{},
					},
				},
			},
		},
		pipeServer,
		map[string]func() SourceServer{"MockSource": func() SourceServer { return mockSourceServer }},
		map[string]func() SinkServer{"MockSink": func() SinkServer { return MockSinkServer }},
	)
	if err != nil {
		t.Errorf("Expected no error from SetupApp, got %v", err)
	}
	if sourceServer != mockSourceServer {
		t.Errorf("Expected sourceServer to be the same as mockSourceServer")
	}
	if sinkServer != MockSinkServer {
		t.Errorf("Expected sinkServer to be the same as MockSinkServer")
	}
}

// Test SetupAndRunApp
func TestSetupAndRunApp(t *testing.T) {
	// Tests the case where the appConfig IngestConfig method returns an error
	config := map[string]any{}
	err := SetupAndRunApp(
		config, &MockPipeServer{}, &MockConfig{},
		map[string]func() Config{}, map[string]func() Config{},
		map[string]func() SinkServer{}, map[string]func() SourceServer{},
	)
	if err == nil {
		t.Errorf("Expected error from SetupAndRunApp, got nil")
	}
	if err.Error() != "AppConfig not set correctly - must map fields to values" {
		t.Errorf("Expected error message to be 'AppConfig not set correctly - must map fields to values', got %v", err.Error())
	}
	// tests the case where the SetupApp method returns an error
	config = map[string]any{
		"AppConfig": map[string]any{},
		"ProducersSetup": map[string]interface{}{
				"ProducerConfigs": []map[string]interface{}{
					{"Type": "HTTP", "ProducerConfig": map[string]interface{}{"URL": "http://localhost:8080"}},
				}},
		"ConsumersSetup": map[string]interface{}{
			"ConsumerConfigs": []map[string]interface{}{
				{"Type": "RabbitMQ", "ConsumerConfig": map[string]interface{}{
					"Connection": "amqp://localhost:5672",
					"Queue":      "test",
				}},
		}},
	}
	err = SetupAndRunApp(
		config, &MockPipeServer{isSetupError: true}, &MockConfig{},
		PRODUCERCONFIGMAP, CONSUMERCONFIGMAP,
		map[string]func() SinkServer{}, map[string]func() SourceServer{},
	)
	if err == nil {
		t.Errorf("Expected error from SetupAndRunApp, got nil")
	}
	if err.Error() != "test error" {
		t.Errorf("Expected error message to be 'test error', got %v", err.Error())
	}
	// Tests the case when ServersRun returns an error
	err = SetupAndRunApp(
		config, &MockPipeServer{isAddError: true}, &MockConfig{},
		PRODUCERCONFIGMAP, CONSUMERCONFIGMAP,
		PRODUCERMAP, CONSUMERMAP,
	)
	if err == nil {
		t.Errorf("Expected error from SetupAndRunApp, got nil")
	}
	if err.Error() != "test error" {
		t.Errorf("Expected error message to be 'test error', got %v", err.Error())
	}
	// test case when everything runs correctly
	config = map[string]any{
		"AppConfig": map[string]any{},
		"ProducersSetup": map[string]interface{}{
			"ProducerConfigs": []map[string]interface{}{
				{"Type": "MockSink", "ProducerConfig": map[string]interface{}{}},
			}},
		"ConsumersSetup": map[string]interface{}{
			"ConsumerConfigs": []map[string]interface{}{
				{"Type": "MockSource", "ConsumerConfig": map[string]interface{}{}},
			}},
	}
	err = SetupAndRunApp(
		config, &MockPipeServer{}, &MockConfig{},
		map[string]func() Config{
			"MockSink": func() Config { return &MockConfig{} },
		}, map[string]func() Config{
			"MockSource": func() Config { return &MockConfig{} },
		},
		map[string]func() SinkServer{
			"MockSink": func() SinkServer { return &MockSinkServer{} },
		}, map[string]func() SourceServer{
			"MockSource": func() SourceServer { return &MockSourceServer{} },
		},
	)
	if err != nil {
		t.Errorf("Expected no error from SetupAndRunApp, got %v", err)
	}
}