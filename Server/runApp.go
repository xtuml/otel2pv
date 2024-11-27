package Server

import (
	"flag"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "", "path to config file")
}

// RunApp is a function that runs the application.
//
// It takes as args:
//
// 1. pipeServer: PipeServer. A pointer to the server that will handle the data.
//
// 2. pipeServerConfig: Config. The configuration struct for the pipe server.
//
// 3. producerConfigMap: map[string]func() Config. A map that maps a string to a func that produces Config.
//
// 4. consumerConfigMap: map[string]func() Config. A map that maps a string to a func that produces Config.
//
// 5. producerMap: map[string]func() SinkServer. A map that maps a string to a func that produces SinkServer.
//
// 6. consumerMap: map[string]func() SourceServer. A map that maps a string to a func that produces SourceServer.
//
// It returns:
//
// 1. error. An error if ingestion fails.
func RunApp(
	pipeServer PipeServer, pipeServerConfig Config,
	producerConfigMap map[string]func() Config, consumerConfigMap map[string]func() Config,
	producerMap map[string]func() SinkServer, consumerMap map[string]func() SourceServer,
) error {
	// Parse the flags
	flag.Parse()
	return RunAppFromConfigPath(
		configPath, pipeServer, pipeServerConfig,
		producerConfigMap, consumerConfigMap,
		producerMap, consumerMap,
	)
}