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
		Server.Logger.Error("Error running Sequencer", slog.String("err", err.Error()))
	} else {
		Server.Logger.Info("Sequencer ran successfully")
	}
}
