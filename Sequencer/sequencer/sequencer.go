package sequencer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"reflect"
	"strings"

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

// GroupApply is a struct that is used by config to hold information
// on a field from the appJSON to share with all other grouped nodes
// in the same group identified by a field with a specific value
// in the appJSON. It has the following fields:
//
// 1. FieldToShare: string. It is the field from the appJSON whose value will be shared
// across all grouped nodes.
//
// 2. IdentifyingField: string. It is the field from the appJSON that will be used to
// identify the specific node.
//
// 3. ValueOfIdentifyingField: string. It is the value of the IdentifyingField that will
// be used to identify the specific node.
type GroupApply struct {
	FieldToShare            string
	IdentifyingField        string
	ValueOfIdentifyingField string
}

// childrenByBackwardsLink is a struct to hold the information on whether to use backwards links to set children
// or not.
// It has the following fields:
//
// 1. All: bool. If true, then perform this for all nodes.
//
// 2. NodeTypes: map[string]bool. If All is false, then perform this for the node types in the map.
type childrenByBackwardsLink struct {
	All       bool
	NodeTypes map[string]bool
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
//
// 4. groupApplies: map[string]GroupApply. It is a map that holds the GroupApply structs for each
// to map from one to many in the appJSON.
type SequencerConfig struct {
	outputAppSequenceField      string
	outputAppFieldSequenceIdMap string
	outputAppFieldSequenceType  OutputAppFieldSequenceType
	groupApplies                map[string]GroupApply
	ChildrenByBackwardsLink     childrenByBackwardsLink
}

// updateGroupApplies is a method that is used to update the groupApplies field of the SequencerConfig
// struct from config.
//
// Args:
//
// 1. config: map[string]any. It is a map that holds the raw configuration.
//
// Returns:
//
// 1. error. It returns an error if the configuration is invalid in any way.
func (s *SequencerConfig) updateGroupApplies(config map[string]any) error {
	groupAppliesRaw, ok := config["groupApplies"]
	var groupApplies []any
	if ok {
		groupApplies, ok = groupAppliesRaw.([]any)
		if !ok {
			return errors.New("groupApplies is not an array")
		}
	}
	s.groupApplies = make(map[string]GroupApply)
	for i, value := range groupApplies {
		groupApplyMap, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("groupApplies[%d] is not a map", i)
		}
		fieldToShare, ok := groupApplyMap["FieldToShare"].(string)
		if !ok {
			return fmt.Errorf("FieldToShare does not exist or is not a string for groupApplies[%d]", i)
		}
		if fieldToShare == "" {
			return fmt.Errorf("FieldToShare is empty for groupApplies[%d]", i)
		}
		identifyingField, ok := groupApplyMap["IdentifyingField"].(string)
		if !ok {
			return fmt.Errorf("IdentifyingField does not exist or is not a string for groupApplies[%d]", i)
		}
		if identifyingField == "" {
			return fmt.Errorf("IdentifyingField is empty for groupApplies[%d]", i)
		}
		valueOfIdentifyingField, ok := groupApplyMap["ValueOfIdentifyingField"].(string)
		if !ok {
			return fmt.Errorf("ValueOfIdentifyingField does not exist or is not a string for groupApplies[%d]", i)
		}
		if valueOfIdentifyingField == "" {
			return fmt.Errorf("ValueOfIdentifyingField is empty for groupApplies[%d]", i)
		}
		if _, ok := s.groupApplies[fieldToShare]; ok {
			return fmt.Errorf("FieldToShare %s already exists in groupApplies", fieldToShare)
		}
		s.groupApplies[fieldToShare] = GroupApply{
			FieldToShare:            fieldToShare,
			IdentifyingField:        identifyingField,
			ValueOfIdentifyingField: valueOfIdentifyingField,
		}
	}
	return nil
}

// updateChildrenByBackwardsLink is a method that is used to update the ChildrenByBackwardsLink field of the SequencerConfig
// struct from config.
//
// Args:
//
// 1. config: map[string]any. It is a map that holds the raw configuration.
//
// Returns:
//
// 1. error. It returns an error if the configuration is invalid in any way.
func (s *SequencerConfig) updateChildrenByBackwardsLink(config map[string]any) error {
	childrenByBackwardsLinkRaw, ok := config["childrenByBackwardsLink"]
	if !ok {
		s.ChildrenByBackwardsLink = childrenByBackwardsLink{
			NodeTypes: map[string]bool{},
		}
		return nil
	}
	childrenByBackwardsLinkMap, ok := childrenByBackwardsLinkRaw.(map[string]any)
	if !ok {
		return errors.New("childrenByBackwardsLink is not a map")
	}
	var all bool
	allRaw, ok := childrenByBackwardsLinkMap["all"]
	if ok {
		all, ok = allRaw.(bool)
		if !ok {
			return errors.New("field \"all\" in childrenByBackwardsLink is not a boolean")
		}
	}
	nodeTypesMap := make(map[string]bool)
	nodeTypesRaw, ok := childrenByBackwardsLinkMap["nodeTypes"]
	if ok {
		nodeTypes, ok := nodeTypesRaw.([]any)
		if !ok {
			return errors.New("nodeTypes in childrenByBackwardsLink is not an array of strings")
		}
		for _, nodeType := range nodeTypes {
			nodeTypeString, ok := nodeType.(string)
			if !ok {
				return errors.New("nodeTypes in childrenByBackwardsLink is not an array of strings")
			}

			nodeTypesMap[nodeTypeString] = true
		}
	}
	s.ChildrenByBackwardsLink = childrenByBackwardsLink{
		All:       all,
		NodeTypes: nodeTypesMap,
	}
	return nil
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
	err := sc.updateGroupApplies(config)
	if err != nil {
		return err
	}
	err = sc.updateChildrenByBackwardsLink(config)
	if err != nil {
		return err
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
	NodeId    string         `json:"nodeId"`
	ParentId  string         `json:"parentId"`
	ChildIds  []string       `json:"childIds"`
	NodeType  string         `json:"nodeType"`
	Timestamp int            `json:"timestamp"`
	AppJSON   map[string]any `json:"appJSON"`
}

// incomingDataEquality is a method that is used to check if two incoming data are equal
//
// Args:
//
// 1. id1: *IncomingData. The first incoming data
//
// 2. id2: *IncomingData. The second incoming data
//
// Returns:
//
// 1. bool. It returns true if the incoming data are equal, false otherwise
func incomingDataEquality(id1 *IncomingData, id2 *IncomingData) bool {
	if id1.NodeId != id2.NodeId {
		return false
	}
	if id1.ParentId != id2.ParentId {
		return false
	}
	if len(id1.ChildIds) != len(id2.ChildIds) {
		return false
	}
	for i, childId := range id1.ChildIds {
		if childId != id2.ChildIds[i] {
			return false
		}
	}
	if id1.NodeType != id2.NodeType {
		return false
	}
	if id1.Timestamp != id2.Timestamp {
		return false
	}
	if !reflect.DeepEqual(id1.AppJSON, id2.AppJSON) {
		return false
	}
	return true
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
		return errors.New(strings.Replace(err.Error(), "XIncomingDataExceptions", "IncomingData", 1))
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

// incomingDataWithDuplicates is a struct that is used to hold the incoming data and any duplicates found
// in the incoming data.
// It has the following fields:
//
// 1. incomingData: *IncomingData. The incoming data.
//
// 2. duplicates: []*IncomingData. The duplicates found in the incoming data.
type incomingDataWithDuplicates struct {
	*IncomingData
	Duplicates   []*IncomingData
}

// stackIncomingData is a single incomingData when it is used for sequencing
// It has the following fields:
//
// 1. embedded *IncomingData. Pointer to the incoming data
//
// 2. currentChildIdIndex: int. The current child id index
type stackIncomingData struct {
	*incomingDataWithDuplicates
	currentChildIdIndex int
	IsDummy             bool
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
func sequenceWithStack(rootNode *incomingDataWithDuplicates, nodeIdToIncomingDataMap map[string]*incomingDataWithDuplicates) iter.Seq2[*stackIncomingData, error] {
	if rootNode == nil {
		return func(yield func(*stackIncomingData, error) bool) {
			yield(nil, errors.New("root node not set"))
		}
	}
	if nodeIdToIncomingDataMap == nil {
		return func(yield func(*stackIncomingData, error) bool) {
			yield(nil, errors.New("nodeIdToIncomingDataMap not set"))
		}
	}
	stack := []*stackIncomingData{
		{
			incomingDataWithDuplicates:        rootNode,
			currentChildIdIndex: 0,
		},
	}
	return func(yield func(*stackIncomingData, error) bool) {
		for len(stack) > 0 {
			top := stack[len(stack)-1]
			childId, err := top.nextChildId()
			if err != nil {
				stack = stack[:len(stack)-1]
				if !yield(top, nil) {
					return
				}
				continue
			}
			childNode, ok := nodeIdToIncomingDataMap[childId]
			if !ok || childNode == nil {
				stack = append(stack, &stackIncomingData{
					incomingDataWithDuplicates: &incomingDataWithDuplicates{
						IncomingData: &IncomingData{
							NodeId: childId,
						},
					},
					IsDummy: true,
				})
				continue
			}
			stack = append(stack, &stackIncomingData{
				incomingDataWithDuplicates:        childNode,
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
func convertToIncomingDataMapAndRootNodes(rawDataArray []json.RawMessage, childrenBybackwardsLink *childrenByBackwardsLink) (map[string]*incomingDataWithDuplicates, map[string]*incomingDataWithDuplicates, bool, error) {
	nodeIdToIncomingDataMap := make(map[string]*incomingDataWithDuplicates)
	nodeIdToNoForwardRefMap := make(map[string]*incomingDataWithDuplicates)
	nodeIdToForwardRefMap := make(map[string]bool)
	backwardsLinks := make(map[string][]string)
	nodeTypeToNodeMap := make(map[string][]*incomingDataWithDuplicates)
	hasUnequalDuplicates := false
	for _, rawData := range rawDataArray {
		incomingData := &IncomingData{}
		err := json.Unmarshal(rawData, incomingData)
		if err != nil {
			return nil, nil, hasUnequalDuplicates, Server.NewInvalidErrorFromError(err)
		}
		if firstNode, ok := nodeIdToIncomingDataMap[incomingData.NodeId]; ok {
			if !incomingDataEquality(firstNode.IncomingData, incomingData) {
				hasUnequalDuplicates = true
			}
			firstNode.Duplicates = append(firstNode.Duplicates, incomingData)
			continue
		}
		incomingDataDuplicate := &incomingDataWithDuplicates{
			IncomingData: incomingData,
			Duplicates:   []*IncomingData{},
		}
		if incomingData.ParentId != "" {
			_, ok := backwardsLinks[incomingData.ParentId]
			if !ok {
				backwardsLinks[incomingData.ParentId] = []string{}
			}
			backwardsLinks[incomingData.ParentId] = append(backwardsLinks[incomingData.ParentId], incomingData.NodeId)
		}
		if _, ok := childrenBybackwardsLink.NodeTypes[incomingData.NodeType]; ok && !childrenBybackwardsLink.All {
			if !ok {
				nodeTypeToNodeMap[incomingData.NodeType] = []*incomingDataWithDuplicates{}
			}
			nodeTypeToNodeMap[incomingData.NodeType] = append(nodeTypeToNodeMap[incomingData.NodeType], incomingDataDuplicate)
		}
		nodeIdToIncomingDataMap[incomingData.NodeId] = incomingDataDuplicate
		_, ok := nodeIdToForwardRefMap[incomingData.NodeId]
		if !ok {
			nodeIdToNoForwardRefMap[incomingData.NodeId] = incomingDataDuplicate
		}
		for _, childId := range incomingData.ChildIds {
			nodeIdToForwardRefMap[childId] = true
			_, ok := nodeIdToNoForwardRefMap[childId]
			if ok {
				delete(nodeIdToNoForwardRefMap, childId)
			}
		}
	}
	if childrenBybackwardsLink.All {
		for parentId, childIds := range backwardsLinks {
			parentNode, ok := nodeIdToIncomingDataMap[parentId]
			if !ok {
				for _, childId := range childIds {
					nodeIdToNoForwardRefMap[childId] = nodeIdToIncomingDataMap[childId]
				}
				continue
			}
			parentNode.ChildIds = childIds
			for _, childId := range childIds {
				if _, ok := nodeIdToIncomingDataMap[childId]; ok {
					delete(nodeIdToNoForwardRefMap, childId)
				}
			}
			err := orderChildrenByTimestamp(parentNode, nodeIdToIncomingDataMap)
			if err != nil {
				return nil, nil, hasUnequalDuplicates, Server.NewInvalidErrorFromError(err)
			}
		}
	} else if len(childrenBybackwardsLink.NodeTypes) != 0 {
		for _, nodes := range nodeTypeToNodeMap {
			for _, node := range nodes {
				childIds, ok := backwardsLinks[node.NodeId]
				if !ok {
					node.ChildIds = []string{}
				} else {
					node.ChildIds = childIds
				}
				for _, childId := range node.ChildIds {
					if _, ok := nodeIdToIncomingDataMap[childId]; ok {
						delete(nodeIdToNoForwardRefMap, childId)
					}
				}
				err := orderChildrenByTimestamp(node, nodeIdToIncomingDataMap)
				if err != nil {
					return nil, nil, hasUnequalDuplicates ,Server.NewInvalidErrorFromError(err)
				}
			}
		}
	}
	return nodeIdToIncomingDataMap, nodeIdToNoForwardRefMap, hasUnequalDuplicates, nil
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
		return Server.NewInvalidError("incoming data is not an array")
	}
	nodeIdToIncomingDataMap, rootNodes, hasUnequalDuplicates, err := convertToIncomingDataMapAndRootNodes(rawDataArray, &s.config.ChildrenByBackwardsLink)
	if err != nil {
		return err
	}
	if len(rootNodes) == 0 {
		return Server.NewInvalidError("no root nodes")
	}
	appJSONArray := []map[string]any{}
	groupAppliesMap := make(map[string]string)
	for _, rootIncomingData := range rootNodes {
		var prevIncomingData *IncomingData
		for stackIncomingData, err := range sequenceWithStack(rootIncomingData, nodeIdToIncomingDataMap) {
			if stackIncomingData.IsDummy {
				slog.Warn("child node not present in data", "details", fmt.Sprintf("childId=%s", stackIncomingData.NodeId))
				prevIncomingData = nil
				continue
			}
			if err != nil {
				return err
			}
			appJSON := stackIncomingData.AppJSON
			var prevID any
			if prevIncomingData != nil && !hasUnequalDuplicates {
				prevID, err = getPrevIdData(prevIncomingData, s.config)
				if err != nil {
					return err
				}
				appJSON[s.config.outputAppSequenceField] = prevID
			}
			// GroupApplies get data
			for fieldToShare, groupApply := range s.config.groupApplies {
				if _, ok := groupAppliesMap[fieldToShare]; !ok {
					fieldValue, err := getGroupApplyValueFromAppJSON(appJSON, groupApply)
					if err != nil {
						continue
					}
					groupAppliesMap[fieldToShare] = fieldValue
				}
			}
			appJSONArray = append(appJSONArray, appJSON)
			// update duplicate nodes
			for _, duplicate := range stackIncomingData.Duplicates {
				appJSON := duplicate.AppJSON
				if !incomingDataEquality(duplicate, stackIncomingData.IncomingData) {
					slog.Warn("duplicate node not equal to original node", "details", fmt.Sprintf("nodeId=%s", duplicate.NodeId))
				}
				if prevIncomingData != nil && !hasUnequalDuplicates {
					appJSON[s.config.outputAppSequenceField] = prevID
				}
				appJSONArray = append(appJSONArray, appJSON)
			}

			prevIncomingData = stackIncomingData.IncomingData
		}
	}
	// GroupApplies set data
	for fieldToShare, fieldValue := range groupAppliesMap {
		for _, appJSON := range appJSONArray {
			appJSON[fieldToShare] = fieldValue
		}
	}
	jsonData, err := json.Marshal(appJSONArray)
	if err != nil {
		return err
	}
	appData := Server.NewAppData(jsonData, "")
	err = s.pushable.SendTo(appData)
	if err != nil {
		return Server.NewSendErrorFromError(err)
	}
	return nil
}

// getGroupAppliesValueFromAppJSON is a helper function that will get the value of the fieldToShare,
// if it exists, from the appJSON. It returns an error if the value is not found or it is not a string.
//
// Args:
//
// 1. appJSON: map[string]any. The appJSON to get the value from.
//
// 2. groupApply: GroupApply. The GroupApply struct that holds the information on the field to get.
//
// Returns:
//
// 1. string. The value of the fieldToShare.
//
// 2. error. The error if the value is not found or it is not a string.
func getGroupApplyValueFromAppJSON(appJSON map[string]any, groupApply GroupApply) (string, error) {
	identifyingValue, ok := appJSON[groupApply.IdentifyingField].(string)
	if !ok {
		return "", errors.New("groupApplies identifying field not found in appJSON or identifying value not a string")
	}
	if identifyingValue != groupApply.ValueOfIdentifyingField {
		return "", errors.New("groupApplies identifying field value does not match")
	}
	fieldValue, ok := appJSON[groupApply.FieldToShare].(string)
	if !ok {
		return "", errors.New("groupApplies field to share value not found or not a string")
	}
	return fieldValue, nil
}

// orderChildrenByTimestamp is a helper function that will order the children of the incoming data
// by their timestamp. It returns an error if the ordering fails.
//
// Args:
//
// 1. incomingData: *IncomingData. The incoming data to order the children of.
//
// 2. nodeIdToIncomingDataMap: map[string]*IncomingData. The map of node ids to incoming data.
//
// Returns:
//
// 1. error. The error if the ordering fails.
func orderChildrenByTimestamp(incomingData *incomingDataWithDuplicates, nodeIdToIncomingDataMap map[string]*incomingDataWithDuplicates) error {
	if len(incomingData.ChildIds) == 0 {
		return nil
	}
	childIdToTimestampMap := make(map[string]int)
	for _, childId := range incomingData.ChildIds {
		childNode, ok := nodeIdToIncomingDataMap[childId]
		if !ok || childNode == nil {
			return fmt.Errorf("child node %s not found but attempting to order children by timestamp", childId)
		}
		if childNode.Timestamp == 0 {
			return fmt.Errorf("child node %s has no timestamp but attempting to order children by timestamp", childId)
		}
		childIdToTimestampMap[childId] = childNode.Timestamp
	}
	orderedChildIds := make([]string, len(incomingData.ChildIds))
	copy(orderedChildIds, incomingData.ChildIds)
	for i := 0; i < len(orderedChildIds); i++ {
		for j := i + 1; j < len(orderedChildIds); j++ {
			if childIdToTimestampMap[orderedChildIds[i]] > childIdToTimestampMap[orderedChildIds[j]] {
				orderedChildIds[i], orderedChildIds[j] = orderedChildIds[j], orderedChildIds[i]
			}
		}
	}
	incomingData.ChildIds = orderedChildIds
	return nil
}
