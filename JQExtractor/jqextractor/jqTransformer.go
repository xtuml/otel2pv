package jqextractor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/itchyny/gojq"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"


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
	ValidatorInfo map[string]validationInfo
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

// validationInfo is a type that is used to define what to use for validation
// It has the following fields:
//
// 1. validate: bool. A boolean that is used to define if validation is required
//
// 2. schema: string. A file path that is used to define the schema
type validationInfo struct {
	validate bool
	schema   string
}

func getValidationSchemaMap(jqQueryStringMapRaw any) (validationInfo, error) {
	// getValidationSchemaMap is a helper function that will get the validation schema map
	// It returns the validation schema map and an error if the map is invalid
	jqQueryStringMap, ok := jqQueryStringMapRaw.(map[string]any)
	if !ok {
		return validationInfo{}, errors.New("invalid JQQueryStrings map in config. Must map a string identifier to a string")
	}
	validationFilePathRaw, ok := jqQueryStringMap["validation"]
	if !ok {
		return validationInfo{validate: false}, nil
	}
	validationFilePath, ok := validationFilePathRaw.(string)
	if !ok {
		return validationInfo{}, errors.New("invalid JQQueryStrings map in config. If field \"validation\" is present it must be a valid file path")
	}
	_, err := os.Stat(validationFilePath)
	if err != nil {
		return validationInfo{}, fmt.Errorf("invalid JQQueryStrings map in config. If field \"validation\" is present it must be a valid file path: %s", err.Error())
	}
	return validationInfo{validate: true, schema: validationFilePath}, nil
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
	validationInfoMap := make(map[string]validationInfo)
	for key, value := range jqQueryStrings {
		jqString, err := getJQStringFromInput(value)
		if err != nil {
			return fmt.Errorf("%s. Key: %s", err.Error(), key)
		}
		jqQueryStringsMap[key] = jqString
		validation, err := getValidationSchemaMap(value)
		if err != nil {
			return fmt.Errorf("%s. Key: %s", err.Error(), key)
		}
		validationInfoMap[key] = validation
	}
	jqt.JQQueryStrings = jqQueryStringsMap
	jqt.ValidatorInfo = validationInfoMap
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
			if err, ok := data.(error); ok {
				return nil, err
			}
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

// JQTransformer is a struct that contains the required fields for the JQTransformer
// It has the following fields:
//
// 1. jqProgram: Code. It is the JQ program that will be used to transform the input JSON
//
// 2. pushable: Pushable. Pushable to send data onto the next stage
//
// 3. validatorMap: map[string]Validator. A map that contains the validators
type JQTransformer struct {
	jqProgram *gojq.Code
	pushable  Server.Pushable
	validatorMap map[string]Server.Validator
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
	if jqt.validatorMap == nil {
		return errors.New("validator map not set")
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
	defer cancel(nil)
	for key, value := range outData {
		for _, val := range value {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if jqt.validatorMap[key] != nil {
					err := jqt.validatorMap[key].Validate(val)
					if err != nil {
						cancel(Server.NewInvalidErrorFromError(err))
						return
					}
				}
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
	if jqt.validatorMap == nil {
		return errors.New("validator map not set")
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
	// Set up the validators
	if jqtConfig.ValidatorInfo == nil {
		return errors.New("ValidatorInfo not set")
	}
	jqt.validatorMap = make(map[string]Server.Validator)
	for key, value := range jqtConfig.ValidatorInfo {
		if !value.validate {
			jqt.validatorMap[key] = Server.DummyValidator{}
		} else {
			schema, err := jsonschema.Compile(value.schema)
			if err != nil {
				return err
			}
			jqt.validatorMap[key] = schema
		}
	}


	return nil
}
