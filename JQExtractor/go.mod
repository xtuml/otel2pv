module JQExtractor

go 1.23.2

replace github.com/SmartDCSITlimited/CDS-OTel-To-PV/Server => ../Server

require (
	github.com/SmartDCSITlimited/CDS-OTel-To-PV/Server v0.0.0-00010101000000-000000000000
	github.com/itchyny/gojq v0.12.16
)

require (
	github.com/itchyny/timefmt-go v0.1.6 // indirect
	github.com/rabbitmq/amqp091-go v1.10.0 // indirect
)
