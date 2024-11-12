package jqextractor

import (
	"errors"

	"github.com/itchyny/gojq"

	"github.com/SmartDCSITlimited/CDS-OTel-To-PV/Server"
)

type JQTransformer struct {
	// JQTransformer is a struct that contains the required fields for the JQTransformer
	// It has the following fields:
	// 1. jqProgram: Code. It is the JQ program that will be used to transform the input JSON
	// 2. inReceiver: Receiver. It is the receiver that will be used to receive data
	// 3. outReceiver: Receiver. It is the receiver that will be used to send data
	jqProgram   *gojq.Code
	inReceiver  Server.Receiver
	outReceiver Server.Receiver
}

func (jqt *JQTransformer) GetReceiver() (Server.Receiver, error) {
	// GetReceiver is a method that returns the receiver that will be used to receive data
	// It returns an error if the in receiver is not set
	if jqt.inReceiver == nil {
		return nil, errors.New("receiver not set")
	}
	return jqt.inReceiver, nil
}

func (jqt *JQTransformer) AddReceiver(receiver Server.Receiver) error {
	// AddReceiver is a method that sets the receiver that will be used to receive data
	// It returns an error if the out receiver is already set
	if jqt.outReceiver != nil {
		return errors.New("receiver already set")
	}
	jqt.outReceiver = receiver
	return nil
}