package main

import (
	"GroupAndVerify/groupandverify"
	"log/slog"

	"github.com/SmartDCSITlimited/CDS-OTel-To-PV/Server"
)

func main() {
	// Main function to run the GroupAndVerify app
	// It will run the application with the GroupAndVerify struct
	// as the pipe server
	groupAndVerify := &groupandverify.GroupAndVerify{}
	groupAndVerifyConfig := &groupandverify.GroupAndVerifyConfig{}
	err := Server.RunApp(
		groupAndVerify, groupAndVerifyConfig,
		Server.PRODUCERCONFIGMAP, Server.CONSUMERCONFIGMAP,
		Server.PRODUCERMAP, Server.CONSUMERMAP,
	)
	if err != nil {
		slog.Error("Error running GroupAndVerify", "details", err.Error())
	} else {
		slog.Info("GroupAndVerify ran successfully")
	}
}

