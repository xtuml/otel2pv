module Sequencer

go 1.23.6

replace github.com/SmartDCSITlimited/CDS-OTel-To-PV/Server => ../Server

require github.com/SmartDCSITlimited/CDS-OTel-To-PV/Server v0.0.0-00010101000000-000000000000

require (
	github.com/Azure/go-amqp v1.3.0 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/rabbitmq/amqp091-go v1.10.0 // indirect
)
