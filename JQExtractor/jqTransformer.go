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
	// 2. pushable: Pushable. Pushable to send data onto the next stage
	jqProgram   *gojq.Code
	pushable  Server.Pushable
}

func (jqt *JQTransformer) AddPushable(pushable Server.Pushable) error {
	// AddPushable is a method that sets the Pushable that will be used to send data to the next stage
	// It returns an error if the pushable field is already set
	if jqt.pushable != nil {
		return errors.New("pushable already set")
	}
	jqt.pushable = pushable
	return nil
}

func (jqt *JQTransformer) SendTo(data *Server.AppData) error {
	// SendTo is a method that will handle incoming data
	// It will transform the data using the JQ program and send it to the next stage
	// It returns an error if the passing of data fails
	if jqt.jqProgram == nil {
		return errors.New("jq program not set")
	}
	if jqt.pushable == nil {
		return errors.New("pushable not set")
	}
	handler, err := data.GetHandler()
	if err != nil {
		return err
	}
	iter := jqt.jqProgram.Run(data.GetData())
	outData, err := getDataFromJQIterator(&iter)
	if err != nil {
		return err
	}
	appData := Server.NewAppData(outData, handler)
	err = jqt.pushable.SendTo(appData)
	if err != nil {
		return err
	}
	return nil
}

func (jqt *JQTransformer) Serve() error {
	// Serve is a method that will start the JQTransformer
	// It will return an error if the pushable field is not set
	// or if the jq program is not set
	if jqt.pushable == nil {
		return errors.New("pushable not set")
	}
	if jqt.jqProgram == nil {
		return errors.New("jq program not set")
	}
	return nil
}