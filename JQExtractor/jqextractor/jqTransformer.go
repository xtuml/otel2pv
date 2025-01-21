package jqextractor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/itchyny/gojq"

	"github.com/SmartDCSITlimited/CDS-OTel-To-PV/Server"
)

// jqQueryInputType is a type that is used to define the input type for the JQTransformer
// It has the following values:
//
// 1. jqQueryInputTypeString: jqQueryInputType. A string that is used to define the input type as a string
//
// 2. jqQueryInputTypeFile: jqQueryInputType. A string that is used to define the input type as a file
type jqQueryInputType string

const (
	jqQueryInputTypeString jqQueryInputType = "string"
	jqQueryInputTypeFile   jqQueryInputType = "file"
)

type JQTransformerConfig struct {
	// JQTransformerConfig is a struct that contains the required fields for the JQTransformerConfig
	// It has the following fields:
	// 1. JQQueryStrings: map[string]string. A map that contains the JQ programs
	JQQueryStrings map[string]string
}

// getJQStringFromInput is a helper function that will handle the input type for the JQTransformer
//
// Args:
//
// 1. jqStringDataMapRaw: any. A map that contains the JQ program and the input type
//
// Returns:
//
// 1. string. The JQ program
//
// 2. error. An error if the input type is invalid
func getJQStringFromInput(jqStringDataMapRaw any) (string, error) {
	jqStringDataMap, ok := jqStringDataMapRaw.(map[string]any)
	if !ok {
		return "", errors.New("invalid JQQueryStrings map in config. Must map a string identifier to a string")
	}
	jqString, ok := jqStringDataMap["jq"].(string)
	if !ok {
		return "", errors.New("invalid JQQueryStrings map in config. Must include field: \"jq\" and this field must be a string")
	}
	jqStringTypeRaw, ok := jqStringDataMap["type"]
	var jqStringType string
	if ok {
		jqStringType, ok = jqStringTypeRaw.(string)
		if !ok {
			return "", errors.New("invalid JQQueryStrings map in config. If field: \"type\" is present it must be a string")
		}
	} else {
		jqStringType = string(jqQueryInputTypeString)
	}
	switch jqQueryInputType(jqStringType) {
	case jqQueryInputTypeString:
		return jqString, nil
	case jqQueryInputTypeFile:
		jqStringBytes, err := os.ReadFile(jqString)
		if err != nil {
			return "", fmt.Errorf("invalid JQQueryStrings map in config. %s", err.Error())
		}
		return string(jqStringBytes), nil
	default:
		return "", fmt.Errorf("invalid JQQueryStrings map in config. Invalid field \"type\": %s", jqStringType)
	}
}

func (jqt *JQTransformerConfig) IngestConfig(config map[string]any) error {
	// IngestConfig is a method that will ingest the configuration for the JQTransformer
	// It takes in a map[string]any and returns an error if the configuration is invalid
	jqQueryStrings, ok := config["JQQueryStrings"].(map[string]any)
	if !ok {
		return errors.New("invalid JQQueryStrings map in config. Must map a string identifier to a string")
	}
	if len(jqQueryStrings) == 0 {
		return errors.New("JQQueryStrings map is empty")
	}
	jqQueryStringsMap := make(map[string]string)
	for key, value := range jqQueryStrings {
		jqString, err := getJQStringFromInput(value)
		if err != nil {
			return fmt.Errorf("%s. Key: %s", err.Error(), key)
		}
		jqQueryStringsMap[key] = jqString
	}
	jqt.JQQueryStrings = jqQueryStringsMap
	return nil
}

func getDataFromJQIterator(iter *gojq.Iter) (map[string][]any, error) {
	// getDataFromIterator is a helper function that will get the data from the iterator
	// It returns the data and an error if the data cannot be retrieved
	counter := 0
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
	mapDataArray, ok := data.(map[string][]any)
	if !ok {
		mapData, ok := data.(map[string]any)
		if !ok {
			return nil, errors.New("data is not a map of strings to arrays")
		}
		mapDataArray := make(map[string][]any)
		for key, value := range mapData {
			arrayData, ok := value.([]any)
			if !ok {
				return nil, errors.New("data is not a map of strings to arrays")
			}
			mapDataArray[key] = arrayData
		}
		return mapDataArray, nil
	}
	return mapDataArray, nil
}

type JQTransformer struct {
	// JQTransformer is a struct that contains the required fields for the JQTransformer
	// It has the following fields:
	// 1. jqProgram: Code. It is the JQ program that will be used to transform the input JSON
	// 2. pushable: Pushable. Pushable to send data onto the next stage
	jqProgram *gojq.Code
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
	dataToExtract, err := data.GetData()
	if err != nil {
		return err
	}
	var unmarshalledData any
	err = json.Unmarshal(dataToExtract, &unmarshalledData)
	if err != nil {
		return Server.NewInvalidErrorFromError(err)
	}
	iter := jqt.jqProgram.Run(unmarshalledData)
	outData, err := getDataFromJQIterator(&iter)
	if err != nil {
		return Server.NewInvalidErrorFromError(err)
	}
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancelCause(context.Background())
	for key, value := range outData {
		for _, val := range value {
			wg.Add(1)
			go func() {
				defer wg.Done()
				jsonData, err := json.Marshal(val)
				if err != nil {
					cancel(err)
					return
				}
				appData := Server.NewAppData(jsonData, key)
				err = jqt.pushable.SendTo(appData)
				if err != nil {
					cancel(Server.NewSendErrorFromError(err))
				}
			}()
		}
	}
	wg.Wait()
	if ctx.Err() != nil {
		err = context.Cause(ctx)
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

func (jqt *JQTransformer) Setup(config Server.Config) error {
	// Setup is a method that will set up the JQTransformer
	// It will return an error if the configuration is invalid
	jqtConfig, ok := config.(*JQTransformerConfig)
	if !ok {
		return errors.New("invalid config")
	}
	if jqtConfig.JQQueryStrings == nil {
		return errors.New("JQQueryStrings not set")
	}
	if len(jqtConfig.JQQueryStrings) == 0 {
		return errors.New("JQQueryStrings is empty")
	}
	jqBuiltString := "{ "
	for key, jqString := range jqtConfig.JQQueryStrings {
		checkJqString, err := gojq.Parse(jqString)
		if err != nil {
			return errors.New("The following JQ string for key \"" + key + "\" failed to be parsed correctly:\n" + jqString + "\n" + err.Error())
		}
		_, err = gojq.Compile(checkJqString)
		if err != nil {
			return errors.New("The following JQ string for key \"" + key + "\" failed to be compiled correctly:\n" + jqString + "\n" + err.Error())
		}
		jqBuiltString += key + ": [( " + jqString + " )//empty],\n"
	}
	jqBuiltString += "}"
	jqQuery, err := gojq.Parse(jqBuiltString)
	if err != nil {
		err = errors.New("The following built JQ string failed to be parsed correctly:\n" + jqBuiltString + "\n" + err.Error())
		return err
	}
	jqProgram, err := gojq.Compile(jqQuery)
	if err != nil {
		err = errors.New("The following built JQ string failed to be compiled correctly:\n" + jqBuiltString + "\n" + err.Error())
		return err
	}
	jqt.jqProgram = jqProgram
	return nil
}
