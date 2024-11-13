package jqextractor

import (
	"errors"

	"github.com/itchyny/gojq"

	"github.com/SmartDCSITlimited/CDS-OTel-To-PV/Server"
)

func getDataFromJQIterator(iter *gojq.Iter) (any, error) {
	// getDataFromIterator is a helper function that will get the data from the iterator
	// It returns the data and an error if the data cannot be retrieved
	counter	:= 0
	iterator := *iter
	var data any
	for {
		if v, ok := iterator.Next(); ok {
			data = v
		} else {
			break
		}
		if counter > 0 {
			return nil, errors.New("more than one data")
		}
		counter++
	}
	if data == nil {
		return nil, errors.New("no data")
	}
	return data, nil
}

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

func (jqt *JQTransformer) HandleIncomingData(data *Server.AppData) error {
	// HandleIncomingData is a method that will handle incoming data
	// It will transform the data using the JQ program and send it to the out receiver
	// It returns an error if the transformation fails
	if jqt.jqProgram == nil {
		return errors.New("jq program not set")
	}
	if jqt.outReceiver == nil {
		return errors.New("out receiver not set")
	}
	iter := jqt.jqProgram.Run(data.GetData())
	outData, err := getDataFromJQIterator(&iter)
	if err != nil {
		return err
	}
	handler, err := data.GetHandler()
	if err != nil {
		return err
	}
	appData := Server.NewAppData(outData, handler)
	err = jqt.outReceiver.SendTo(appData)
	if err != nil {
		return err
	}
	return nil
}