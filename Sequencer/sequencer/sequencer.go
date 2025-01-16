package sequencer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"iter"

	"github.com/SmartDCSITlimited/CDS-OTel-To-PV/Server"
)

// OutputAppFieldSequenceType is a type that represents the type
// that will be used for the sequence field in the output app schema.
// It has the following constants:
//
// 1. Array: OutputAppFieldSequenceType. The sequence field will be an array.
//
// 2. String: OutputAppFieldSequenceType. The sequence field will be a string.
type OutputAppFieldSequenceType string

// OutputAppFieldSequenceType constants
const (
	Array  OutputAppFieldSequenceType = "array"
	String OutputAppFieldSequenceType = "string"
)

// GetOutputAppFieldSequenceType returns the OutputAppFieldSequenceType for the given string.
// Returns an error if the string is not a valid OutputAppFieldSequenceType.
//
// Args:
// 1. s: string. The string to convert to a OutputAppFieldSequenceType.
//
// Returns:
// 1. OutputAppFieldSequenceType. The OutputAppFieldSequenceType for the given string.
// 2. error. An error if the string is not a valid OutputAppFieldSequenceType.
func GetOutputAppFieldSequenceType(s string) (OutputAppFieldSequenceType, error) {
	switch s {
	case "array":
		return Array, nil
	case "string":
		return String, nil
	default:
		return OutputAppFieldSequenceType(""), fmt.Errorf("unknown output app field sequence type: %s", s)
	}
}

// SequencerConfig is a struct that represents the configuration for a Sequencer.
// It has the following fields:
//
// 1. outputAppSequenceField: string. The field in the output app schema that will be used to add
// the sequencing information.
//
// 2. outputAppFieldSequenceIdMap: string. The field in the output app schema that will be used to
// to replace the node id with, if at all. If null then just uses nodeId field.
//
// 3. outputAppFieldSequenceType: OutputAppFieldSequenceType. The type that will be used for the sequence field in the output app schema.
type SequencerConfig struct {
	outputAppSequenceField      string
	outputAppFieldSequenceIdMap string
	outputAppFieldSequenceType  OutputAppFieldSequenceType
}

// IngestConfig ingests the configuration for the SequencerConfig.
// Returns an error if the configuration is invalid.
//
// Args:
// 1. config: map[string]any. The raw configuration for the Sequencer.
//
// Returns:
// 1. error. An error if the configuration is invalid.
func (sc *SequencerConfig) IngestConfig(config map[string]any) error {
	outputAppSequenceField, ok := config["outputAppSequenceField"].(string)
	if !ok {
		return fmt.Errorf("invalid outputAppSequenceField - must be a string and must be set")
	}
	sc.outputAppSequenceField = outputAppSequenceField

	outputAppFieldSequenceIdMap, ok := config["outputAppFieldSequenceIdMap"]
	if ok {
		outputAppFieldSequenceIdMap, ok := outputAppFieldSequenceIdMap.(string)
		if !ok {
			return fmt.Errorf("invalid outputAppFieldSequenceIdMap - must be a string")
		}
		sc.outputAppFieldSequenceIdMap = outputAppFieldSequenceIdMap
	}

	outputAppFieldSequenceType, ok := config["outputAppFieldSequenceType"]
	if !ok {
		sc.outputAppFieldSequenceType = Array
	} else {
		outputAppFieldSequenceType, ok := outputAppFieldSequenceType.(string)
		if !ok {
			return fmt.Errorf("invalid outputAppFieldSequenceType - must be a string")
		}
		ofst, err := GetOutputAppFieldSequenceType(outputAppFieldSequenceType)
		if err != nil {
			return fmt.Errorf("invalid outputAppFieldSequenceType - %s", err)
		}
		sc.outputAppFieldSequenceType = ofst
	}
	return nil
}

// Sequencer struct
// This struct represents a Sequencer.
// It has the following fields:
//
// 1. config: SequencerConfig. The configuration for the Sequencer.
//
// 2. pushable: Server.Pushable. The Pushable interface added to the struct for
// sending on data
type Sequencer struct {
	config   *SequencerConfig
	pushable Server.Pushable
}

// AddPushable is a method that sets the Pushable that will be used to
// send data to the next stage
// Args:
// 1. pushable: Server.Pushable. The pushable to add to the instance
// Returns:
// 1. error. The error if the method fails
func (s *Sequencer) AddPushable(pushable Server.Pushable) error {
	if s.pushable != nil {
		return errors.New("pushable already set")
	}
	s.pushable = pushable
	return nil
}

// Serve is a method that will check everything is
// set up correctly
// Returns:
//
// 1. error. The error if the method fails
func (s *Sequencer) Serve() error {
	if s.pushable == nil {
		return errors.New("pushable not set")
	}
	if s.config == nil {
		return errors.New("config not set")
	}
	return nil
}

// Setup is a method that will setup the instance of
// Sequencer using input Config
//
// Args:
//
// 1. config: Server.Config. The config to set up the instance with
//
// Returns:
//
// 1. error. The error if the method fails
func (s *Sequencer) Setup(config Server.Config) error {
	if s.config != nil {
		return errors.New("configuration already set")
	}
	sequencerConfig, ok := config.(*SequencerConfig)
	if !ok {
		return errors.New("config must be SequencerConfig type")
	}
	s.config = sequencerConfig
	return nil
}

// IncomingData is a struct that represents the required incoming data
// for the Sequencer
// It has the following fields:
//
// 1. NodeId: string. The node id of the incoming data
//
// 2. ParentId: string. The parent id of the incoming data
//
// 3. ChildIds: []string. The ordered child ids of the incoming data (first to last)
//
// 4. NodeType: string. The type of the incoming data
//
// 5. Timestamp: int. The timestamp of the incoming data
//
// 6. AppJSON: map[string]any. The outgoing app JSON of the incoming data
type IncomingData struct {
	NodeId    string
	ParentId  string
	ChildIds  []string
	NodeType  string
	Timestamp int
	AppJSON   map[string]any
}

type XIncomingData IncomingData

type XIncomingDataExceptions struct {
	XIncomingData
	TreeId string `json:"treeId"`
}

// UnmarshalJSON is a method that is used to unmarshal the JSON data into the IncomingData struct raising errors if:
//
// 1. Extra fields are present in the JSON data.
//
// 2. Required fields are missing from the JSON data.
//
// Args:
//
// 1. data: []byte. It is the JSON data that is to be unmarshalled.
//
// Returns:
//
// 1. error. It returns an error if the unmarshalling fails.
func (id *IncomingData) UnmarshalJSON(data []byte) error {
	var xide XIncomingDataExceptions
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&xide); err != nil {
		return errors.New("input JSON does not match the IncomingData format with strictly disallowed unknown fields except \"treeId\"")
	}
	if xide.NodeId == "" {
		return errors.New("nodeId is required")
	}
	if xide.AppJSON == nil {
		return errors.New("appJSON is required")
	}
	*id = IncomingData(xide.XIncomingData)
	return nil
}

// UnmarshalJSON is a method that will unmarshal the JSON data
// into the incomingData struct

// stackIncomingData is a single incomingData when it is used for sequencing
// It has the following fields:
//
// 1. embedded *IncomingData. Pointer to the incoming data
//
// 2. currentChildIdIndex: int. The current child id index
type stackIncomingData struct {
	*IncomingData
	currentChildIdIndex int
}

// nextChildId returns the next child id in the stackIncomingData
// Returns an error if the next child id is not found
//
// Returns:
//
// 1. string. The next child id in the stackIncomingData
//
// 2. error. The error if the next child id is not found
func (sid *stackIncomingData) nextChildId() (string, error) {
	if sid.currentChildIdIndex >= len(sid.ChildIds) {
		return "", errors.New("no more child ids")
	}
	childId := sid.ChildIds[sid.currentChildIdIndex]
	sid.currentChildIdIndex++
	return childId, nil
}

// sequenceWithStack is a method that will sequence the incoming data
// using a stack. It will return the sequenced data as a generator
// Returns an error if the sequencing fails
//
// Args:
//
// 1. rootNode: *IncomingData. The root node of the incoming data
//
// 2. nodeIdToIncomingDataMap: map[string]*IncomingData. The map of node ids to incoming data
//
// Returns:
//
// 1. iter.Seq2[*IncomingData, error]. The sequenced data as a generator
func sequenceWithStack(rootNode *IncomingData, nodeIdToIncomingDataMap map[string]*IncomingData) iter.Seq2[*IncomingData, error] {
	if rootNode == nil {
		return func(yield func(*IncomingData, error) bool) {
			yield(nil, errors.New("root node not set"))
		}
	}
	if nodeIdToIncomingDataMap == nil {
		return func(yield func(*IncomingData, error) bool) {
			yield(nil, errors.New("nodeIdToIncomingDataMap not set"))
		}
	}
	stack := []*stackIncomingData{
		{
			IncomingData:        rootNode,
			currentChildIdIndex: 0,
		},
	}
	return func(yield func(*IncomingData, error) bool) {
		for len(stack) > 0 {
			top := stack[len(stack)-1]
			childId, err := top.nextChildId()
			if err != nil {
				stack = stack[:len(stack)-1]
				if !yield(top.IncomingData, nil) {
					return
				}
				continue
			}
			childNode, ok := nodeIdToIncomingDataMap[childId]
			if !ok {
				yield(nil, fmt.Errorf("child node not found: %s", childId))
				return
			}
			if childNode == nil {
				yield(nil, fmt.Errorf("child node is nil: %s", childId))
				return
			}
			stack = append(stack, &stackIncomingData{
				IncomingData:        childNode,
				currentChildIdIndex: 0,
			})
		}
	}
}

// convertToIncomingDataMapAndRootNodes
// converts the AppData to a map of node id to incoming data
// and a map of node id to incoming data with no forward references
// Returns an error if the conversion fails
//
// Args:
//
// 1. data: *Server.AppData. The data to convert
//
// Returns:
//
// 1. map[string]*IncomingData. The converted data mapping node id to incoming data
//
// 2. map[string]*IncomingData. The converted data mapping node id to incoming data with no forward references i.e. root nodes
//
// 3. error. The error if the conversion fails
func convertToIncomingDataMapAndRootNodes(rawDataArray []json.RawMessage) (map[string]*IncomingData, map[string]*IncomingData, error) {
	nodeIdToIncomingDataMap := make(map[string]*IncomingData)
	nodeIdToNoForwardRefMap := make(map[string]*IncomingData)
	nodeIdToForwardRefMap := make(map[string]bool)
	for _, rawData := range rawDataArray {
		incomingData := &IncomingData{}
		err := json.Unmarshal(rawData, incomingData)
		if err != nil {
			return nil, nil, err
		}
		nodeIdToIncomingDataMap[incomingData.NodeId] = incomingData
		_, ok := nodeIdToForwardRefMap[incomingData.NodeId]
		if !ok {
			nodeIdToNoForwardRefMap[incomingData.NodeId] = incomingData
		}
		for _, childId := range incomingData.ChildIds {
			nodeIdToForwardRefMap[childId] = true
			_, ok := nodeIdToNoForwardRefMap[childId]
			if ok {
				delete(nodeIdToNoForwardRefMap, childId)
			}
		}

	}
	return nodeIdToIncomingDataMap, nodeIdToNoForwardRefMap, nil
}

// getPrevIdFromPrevIncomingData
// Returns an error if the previous id is not found
//
// Args:
//
// 1. prevIncomingData: *IncomingData. The previous incoming data
//
// 2. outputAppFieldSequenceIdMap: string. The field in the output app schema that will be used to
// to replace the node id with, if at all. If "" then just uses nodeId field.
//
// Returns:
//
// 1. string. The previous id
//
// 2. error. The error if the previous id is not found
func getPrevIdFromPrevIncomingData(prevIncomingData *IncomingData, outputAppFieldSequenceIdMap string) (string, error) {
	if outputAppFieldSequenceIdMap != "" {
		prevIDUnTyped, ok := prevIncomingData.AppJSON[outputAppFieldSequenceIdMap]
		if !ok {
			return "", errors.New(
				"outputAppFieldSequenceIdMap must be a string and must exist in the input JSON",
			)
		}
		prevID, ok := prevIDUnTyped.(string)
		if !ok {
			return "", errors.New(
				"outputAppFieldSequenceIdMap must be a string and must exist in the input JSON",
			)
		}
		return prevID, nil
	}
	return prevIncomingData.NodeId, nil
}

// getPrevIdData
// Returns an error if the process fails
//
// Args:
//
// 1. prevIncomingData: *IncomingData. The previous incoming data
//
// 2. config: *SequencerConfig. The configuration for the Sequencer
//
// Returns:
//
// 1. any. The previous id (string or []string)
//
// 2. error. The error if the process fails
func getPrevIdData(prevIncomingData *IncomingData, config *SequencerConfig) (any, error) {
	prevID, err := getPrevIdFromPrevIncomingData(prevIncomingData, config.outputAppFieldSequenceIdMap)
	if err != nil {
		return nil, err
	}
	switch config.outputAppFieldSequenceType {
	case Array:
		return []string{prevID}, nil
	case String:
		return prevID, nil
	default:
		return nil, errors.New("unknown outputAppFieldSequenceType")
	}
}

// SendTo is a method that will handle incoming data
// It will sequence the data and send it to the next stage
// Returns an error if the passing of data fails
//
// Args:
//
// 1. data: *Server.AppData. The data to send to the next stage
//
// Returns:
//
// 1. error. The error if the method fails
func (s *Sequencer) SendTo(data *Server.AppData) error {
	if s.pushable == nil {
		return errors.New("pushable not set")
	}
	if s.config == nil {
		return errors.New("config not set")
	}
	if data == nil {
		return errors.New("data not sent")
	}
	rawData, err := data.GetData()
	if err != nil {
		return err
	}
	var rawDataArray []json.RawMessage
	err = json.Unmarshal(rawData, &rawDataArray)
	if err != nil {
		return errors.New("incoming data is not an array")
	}
	nodeIdToIncomingDataMap, rootNodes, err := convertToIncomingDataMapAndRootNodes(rawDataArray)
	if err != nil {
		return err
	}
	if len(rootNodes) == 0 {
		return errors.New("no root nodes")
	}
	appJSONArray := []map[string]any{}
	for _, rootIncomingData := range rootNodes {
		var prevIncomingData *IncomingData
		for incomingData, err := range sequenceWithStack(rootIncomingData, nodeIdToIncomingDataMap) {
			if err != nil {
				return err
			}
			appJSON := incomingData.AppJSON
			if prevIncomingData != nil {
				prevID, err := getPrevIdData(prevIncomingData, s.config)
				if err != nil {
					return err
				}
				appJSON[s.config.outputAppSequenceField] = prevID
			}
			appJSONArray = append(appJSONArray, appJSON)
			prevIncomingData = incomingData
		}
	}
	jsonData, err := json.Marshal(appJSONArray)
	if err != nil {
		return err
	}
	appData := Server.NewAppData(jsonData, "")
	return s.pushable.SendTo(appData)
}
