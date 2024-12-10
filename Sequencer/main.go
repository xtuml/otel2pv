package main

import (
	"Sequencer/sequencer"
	"log/slog"

	"github.com/SmartDCSITlimited/CDS-OTel-To-PV/Server"
)

func main() {
	// Main function to run the Sequencer
	// It will run the application with the Sequencer
	// as the pipe server
	sequencerPipe := &sequencer.Sequencer{}
	sequencerConfig := &sequencer.SequencerConfig{}
	err := Server.RunApp(
		sequencerPipe, sequencerConfig,
		Server.PRODUCERCONFIGMAP, Server.CONSUMERCONFIGMAP,
		Server.PRODUCERMAP, Server.CONSUMERMAP,
	)
	if err != nil {
		slog.Error("Error running Sequencer", "details", err.Error())
	} else {
		slog.Info("Sequencer ran successfully")
	}
}