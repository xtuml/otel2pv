package Server

import (
	"flag"
	"os"
	"testing"
)

// TestRunApp is a function that tests the RunApp function.
func TestRunApp(t *testing.T) {
	// Tests case where config flag is set but the file does not exist
	err := flag.Set("config", "config.json")
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	err = RunApp(&MockPipeServer{}, &MockConfig{}, map[string]func() Config{}, map[string]func() Config{}, map[string]func() SinkServer{}, map[string]func() SourceServer{})
	if err == nil {
		t.Errorf("Error: %s", err)
	}
	// Tests case when config file exists and everything works correctly
	err = flag.Set("config", "config_test.json")
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	tmpFile, err := os.Create("config_test.json")
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	defer os.Remove(tmpFile.Name())
	data := []byte(`{"AppConfig":{},"ProducersSetup":{"ProducerConfigs":[{"Type":"MockSink","ProducerConfig":{}}]},"ConsumersSetup":{"ConsumerConfigs":[{"Type":"MockSource","ConsumerConfig":{}}]}}`)
	err = os.WriteFile(tmpFile.Name(), data, 0644)
	if err != nil {
		t.Errorf("Error writing to temp file: %v", err)
	}
	err = RunApp(
		&MockPipeServer{}, &MockConfig{},
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
		t.Errorf("Error: %s", err)
	}
}
