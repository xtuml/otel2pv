package groupandverify

import (
	"errors"
	"fmt"
)

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

// OutgoingData that is used to hold the outgoing data from the GroupAndVerify
// component. It has the following fields:
//
// 1. NodeId: string. It is the identifier of the node.
//
// 2. OrderedChildIds: []string. It is the list of identifiers of the children of the node, in order
// of occurence.
//
// 3. AppJSON: map[string]any. It is the JSON data that is to be sent to the next stage.
type OutgoingData struct {
	NodeId          string
	OrderedChildIds []string
	AppJSON         map[string]any
}

// IncomingData is a struct that is used to hold the incoming data from the previous stage
// in the GroupAndVerify component. It has the following fields:
//
// 1. NodeId: string. It is the identifier of the node.
//
// 2. ParentId: string. It is the identifier of the parent of the node.
//
// 3. ChildIds: []string. It is the list of identifiers of the children of the node.
//
// 4. NodeType: string. It is the type of the node. This can be used to identify if
// bidirectional confirmation is required. Optional.
//
// 5. Timestamp: int. It is the timestamp of the node. Optional.
//
// 6. AppJSON: map[string]any. It is the JSON data that is to be processed and sent on.
type IncomingData struct {
	NodeId    string
	ParentId  string
	ChildIds  []string
	NodeType  string
	Timestamp int
	AppJSON   map[string]any
}

// GroupAndVerifyConfig is a struct that is used to hold the configuration for the GroupAndVerify
// component. It has the following fields:
//
// 1. orderChildrenByTimestamp: bool. It is a boolean that determines if the orderedChildIds array
// should be ordered on the basis of timestamp.
//
// 2. groupApplies: map[string]GroupApply. It is a map that holds the GroupApply structs for each
// to map from one to many in the appJSON.
//
// 3. parentVerifySet: map[string]bool. It is a map that holds the identifiers of the nodes (NodeTypes)
// that do not require bidirectional confirmation.
type GroupAndVerifyConfig struct {
	orderChildrenByTimestamp bool
	groupApplies             map[string]GroupApply
	parentVerifySet          map[string]bool
}

// updateOrderChildrenByTimestamp is a method that is used to update the orderChildrenByTimestamp field
// of the GroupAndVerifyConfig struct from config.
//
// Args:
//
// 1. config: map[string]any. It is a map that holds the raw configuration.
//
// Returns:
//
// 1. error. It returns an error if the configuration is invalid in any way.
func (gavc *GroupAndVerifyConfig) updateOrderChildrenByTimestamp(config map[string]any) error {
	orderChildrenByTimestampRaw, ok := config["orderChildrenByTimestamp"]

	if ok {
		orderChildrenByTimestamp, ok := orderChildrenByTimestampRaw.(bool)
		if !ok {
			return errors.New("orderChildrenByTimestamp is not a boolean")
		}
		gavc.orderChildrenByTimestamp = orderChildrenByTimestamp
	}
	return nil
}

// updateGroupApplies is a method that is used to update the groupApplies field of the GroupAndVerifyConfig
// struct from config.
//
// Args:
//
// 1. config: map[string]any. It is a map that holds the raw configuration.
//
// Returns:
//
// 1. error. It returns an error if the configuration is invalid in any way.
func (gavc *GroupAndVerifyConfig) updateGroupApplies(config map[string]any) error {
	groupAppliesRaw, ok := config["groupApplies"]
	var groupApplies []any
	if ok {
		groupApplies, ok = groupAppliesRaw.([]any)
		if !ok {
			return errors.New("groupApplies is not an array")
		}
	}
	gavc.groupApplies = make(map[string]GroupApply)
	for i, value := range groupApplies {
		groupApplyMap, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("groupApplies[%d] is not a map", i)
		}
		fieldToShare, ok := groupApplyMap["FieldToShare"].(string)
		if !ok {
			return fmt.Errorf("FieldToShare does not exist or is not a string for groupApplies[%d]", i)
		}
		identifyingField, ok := groupApplyMap["IdentifyingField"].(string)
		if !ok {
			return fmt.Errorf("IdentifyingField does not exist or is not a string for groupApplies[%d]", i)
		}
		valueOfIdentifyingField, ok := groupApplyMap["ValueOfIdentifyingField"].(string)
		if !ok {
			return fmt.Errorf("ValueOfIdentifyingField does not exist or is not a string for groupApplies[%d]", i)
		}
		if _, ok := gavc.groupApplies[fieldToShare]; ok {
			return fmt.Errorf("FieldToShare %s already exists in groupApplies", fieldToShare)
		}
		gavc.groupApplies[fieldToShare] = GroupApply{
			FieldToShare:            fieldToShare,
			IdentifyingField:        identifyingField,
			ValueOfIdentifyingField: valueOfIdentifyingField,
		}
	}
	return nil
}

// updateParentVerifySet is a method that is used to update the parentVerifySet field of the GroupAndVerifyConfig
// struct from config.
//
// Args:
//
// 1. config: map[string]any. It is a map that holds the raw configuration.
//
// Returns:
//
// 1. error. It returns an error if the configuration is invalid in any way.
func (gavc *GroupAndVerifyConfig) updateParentVerifySet(config map[string]any) error {
	parentVerifySetRaw, ok := config["parentVerifySet"]
	if ok {
		parentVerifySetAny, ok := parentVerifySetRaw.([]any)
		if !ok {
			return errors.New("parentVerifySet is not an array of strings")
		}
		parentVerifySet := make(map[string]bool)
		for i, value := range parentVerifySetAny {
			valueString, ok := value.(string)
			if !ok {
				return fmt.Errorf("parentVerifySet[%d] is not a string", i)
			}
			parentVerifySet[valueString] = true
		}
		gavc.parentVerifySet = parentVerifySet
	}
	return nil
}

// IngestConfig is a method that is used to set the fields of the GroupAndVerifyConfig struct.
//
// Args:
// 1. config: map[string]any. It is a map that holds the raw configuration.
//
// Returns:
//
// 1. error. It returns an error if the configuration is invalid in any way.
func (gavc *GroupAndVerifyConfig) IngestConfig(config map[string]any) error {
	// IngestConfig is a method that is used to set the fields of the GroupAndVerifyConfig struct
	// It returns an error if the configuration is invalid in any way
	if config == nil {
		return errors.New("config is nil")
	}
	err := gavc.updateOrderChildrenByTimestamp(config)
	if err != nil {
		return err
	}
	err = gavc.updateGroupApplies(config)
	if err != nil {
		return err
	}
	err = gavc.updateParentVerifySet(config)
	if err != nil {
		return err
	}
	return nil
}
