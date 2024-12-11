package groupandverify

// GroupApply is a struct that is used by config to hold information
// on a field from the appJSON to share with all other grouped nodes
// in the same group identified by a field with a specific value
// in the appJSON.
//
// It has the following fields:
// 1. FieldToShare: string. It is the field from the appJSON whose value will be shared
// across all grouped nodes.
//
// 2. IdentifyingField: string. It is the field from the appJSON that will be used to
// identify the specific node.
//
// 3. ValueOfIdentifyingField: string. It is the value of the IdentifyingField that will
// be used to identify the specific node.
type GroupApply struct {
	FieldToShare           string
	IdentifyingField       string
	ValueOfIdentifyingField string
}

// OutgoingData that is used to hold the outgoing data from the GroupAndVerify
// component. It has the following fields:
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
	AppJSON  map[string]any
}
