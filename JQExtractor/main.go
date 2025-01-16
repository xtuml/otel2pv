package main

import (
	"JQExtractor/jqextractor"
	"log/slog"

	"github.com/SmartDCSITlimited/CDS-OTel-To-PV/Server"
)

func main() {
	// Main function to run the JQExtractor
	// It will run the application with the JQExtractor
	// as the pipe server
	jqTransformer := &jqextractor.JQTransformer{}
	jqTransformerConfig := &jqextractor.JQTransformerConfig{}
	err := Server.RunApp(
		jqTransformer, jqTransformerConfig,
		Server.PRODUCERCONFIGMAP, Server.CONSUMERCONFIGMAP,
		Server.PRODUCERMAP, Server.CONSUMERMAP,
	)
	if err != nil {
		slog.Error("Error running JQExtractor", "details", err.Error())
	} else {
		slog.Info("JQExtractor ran successfully")
	}
}
