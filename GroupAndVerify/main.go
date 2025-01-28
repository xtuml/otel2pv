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
		Server.Logger.Error("Error running GroupAndVerify", slog.String("err", err.Error()))
	} else {
		Server.Logger.Info("GroupAndVerify ran successfully")
	}
}
