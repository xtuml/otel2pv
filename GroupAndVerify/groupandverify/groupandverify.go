package groupandverify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"sync"
	"time"

	"github.com/SmartDCSITlimited/CDS-OTel-To-PV/Server"
)

var groupAndVerifyLogger slog.Logger

func init() {
	groupAndVerifyLogger = *Server.Logger.With(
		slog.Group(
			"logger",
			slog.String("logger_name", "GroupAndVerify"),
		),
	)
}
// IncomingData is a struct that is used to hold the incoming data from the previous stage
// in the GroupAndVerify component. It has the following fields:
//
// 1. TreeId: string. It is the identifier of the tree.
//
// 2. NodeId: string. It is the identifier of the node.
//
// 3. ParentId: string. It is the identifier of the parent of the node.
//
// 4. ChildIds: []string. It is the list of identifiers of the children of the node.
//
// 5. NodeType: string. It is the type of the node. This can be used to identify if
// bidirectional confirmation is required. Optional.
//
// 6. Timestamp: int. It is the timestamp of the node. Optional.
//
// 7. AppJSON: json.RawMessage. It is the JSON data that is to be sent on.
type IncomingData struct {
	TreeId    string          `json:"treeId"`
	NodeId    string          `json:"nodeId"`
	ParentId  string          `json:"parentId"`
	ChildIds  []string        `json:"childIds"`
	NodeType  string          `json:"nodeType"`
	Timestamp int             `json:"timestamp"`
	AppJSON   json.RawMessage `json:"appJSON"`
}

type XIncomingData IncomingData

type XIncomingDataExceptions struct {
	XIncomingData
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
		return err
	}
	if xide.TreeId == "" {
		return errors.New("TreeId is required")
	}
	if xide.NodeId == "" {
		return errors.New("NodeId is required")
	}
	if xide.AppJSON == nil {
		return errors.New("AppJSON is required")
	}
	*id = IncomingData(xide.XIncomingData)
	return nil
}

// Task is a struct that is used to hold the task that is to be performed by the GroupAndVerify
// component. It has the following fields:
//
// 1. IncomingData: IncomingData. It is the incoming data that is to be processed.
//
// 2. errChan: chan error. It is the channel that is used to receive errors from the Serve routine.
type Task struct {
	IncomingData *IncomingData
	errChan      chan error
}

// GroupAndVerifyConfig is a struct that is used to hold the configuration for the GroupAndVerify
// component. It has the following fields:
//
// 1. parentVerifySet: map[string]*int. It is a map that holds the identifiers of the nodes (NodeTypes)
// that do not require bidirectional confirmation and their expected number of children.
//
// 2. Timeout: int. It is the time out for the processing of the tree.
//
// 3. MaxTrees: int. It is the maximum number of trees that can be processed at a time. If 0, then there is no limit.
type GroupAndVerifyConfig struct {
	parentVerifySet map[string]*int
	Timeout         int
	MaxTrees        int
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
		parentVerifySet := make(map[string]*int)
		for i, value := range parentVerifySetAny {
			valueString, ok := value.(map[string]any)
			if !ok {
				return fmt.Errorf("parentVerifySet[%d] is not an object", i)
			}
			nodeType, ok := valueString["nodeType"].(string)
			if !ok {
				return fmt.Errorf("nodeType in parentVerifySet[%d] is not present or is not a string", i)
			}
			expectedChildrenRaw, ok := valueString["expectedChildren"].(float64)
			if !ok {
				return fmt.Errorf("expectedChildren in parentVerifySet[%d] is not present or is not a number", i)
			}
			expectedChildren := int(expectedChildrenRaw)
			if expectedChildren <= 0 {
				return fmt.Errorf("expectedChildren in parentVerifySet[%d] must be a positive integer greater than zero", i)
			}
			if _, ok := parentVerifySet[nodeType]; ok {
				return fmt.Errorf("nodeType in parentVerifySet[%d] is a duplicate", i)
			}
			parentVerifySet[nodeType] = &expectedChildren
		}
		gavc.parentVerifySet = parentVerifySet
	} else {
		gavc.parentVerifySet = make(map[string]*int)
	}
	return nil
}

// updateTimeout is a method that is used to update the Timeout field of the GroupAndVerifyConfig
// struct from config.
//
// Args:
//
// 1. config: map[string]any. It is a map that holds the raw configuration.
//
// Returns:
//
// 1. error. It returns an error if the configuration is invalid in any way.
func (gavc *GroupAndVerifyConfig) updateTimeout(config map[string]any) error {
	timeoutRaw, ok := config["Timeout"]
	if ok {
		timeout, ok := timeoutRaw.(int)
		if !ok {
			if timeoutFloat, ok := timeoutRaw.(float64); !ok {
				return errors.New("Timeout must be a positive integer")
			} else {
				timeout = int(timeoutFloat)
			}
		}
		if timeout < 0 {
			return errors.New("Timeout must be a positive integer")
		}
		gavc.Timeout = timeout
	} else {
		gavc.Timeout = 2
	}
	return nil
}

// updateMaxTrees is a method that is used to update the MaxTrees field of the GroupAndVerifyConfig
// struct from config.
//
// Args:
//
// 1. config: map[string]any. It is a map that holds the raw configuration.
//
// Returns:
//
// 1. error. It returns an error if the configuration is invalid in any way.
func (gavc *GroupAndVerifyConfig) updateMaxTrees(config map[string]any) error {
	maxTreesRaw, ok := config["maxTrees"]
	if ok {
		maxTrees, ok := maxTreesRaw.(int)
		if !ok {
			return errors.New("maxTrees must be a positive integer")
		}
		if maxTrees < 0 {
			return errors.New("maxTrees must be a positive integer")
		}
		gavc.MaxTrees = maxTrees
	} else {
		gavc.MaxTrees = 0
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
	err := gavc.updateParentVerifySet(config)
	if err != nil {
		return err
	}
	err = gavc.updateTimeout(config)
	if err != nil {
		return err
	}
	err = gavc.updateMaxTrees(config)
	if err != nil {
		return err
	}
	return nil
}

// GroupAndVerify is a struct that is used to hold the GroupAndVerify component. It has the following fields:
//
// It has the following fields:
//
// 1. config: GroupAndVerifyConfig. It is the configuration for the GroupAndVerify component.
//
// 2. pushable: Server.Pushable. It is the pushable that is used to send data to the next stage.
//
// 3. taskChan: chan *Task. It is the channel that holds the incoming tasks.
type GroupAndVerify struct {
	config   *GroupAndVerifyConfig
	pushable Server.Pushable
	taskChan chan *Task
}

// AddPushable is a method that is used to set the pushable field of the GroupAndVerify struct.
//
// Args:
//
// 1. pushable: Server.Pushable. It is the pushable that is used to send data to the next stage.
//
// Returns:
//
// 1. error. It returns an error if the pushable field is already set.
func (gav *GroupAndVerify) AddPushable(pushable Server.Pushable) error {
	// AddPushable is a method that is used to set the pushable field of the GroupAndVerify struct
	// It returns an error if the pushable field is already set
	if gav.pushable != nil {
		return errors.New("pushable is already set")
	}
	gav.pushable = pushable
	return nil
}

// Setup is a method that is used to set the configuration for the GroupAndVerify component.
//
// Args:
//
// 1. config: Server.Config. It is the configuration for the GroupAndVerify component.
//
// Returns:
//
// 1. error. It returns an error if the configuration is invalid in any way.
func (gav *GroupAndVerify) Setup(config Server.Config) error {
	if gav.config != nil {
		return errors.New("config already set")
	}
	groupAndVerifyConfig, ok := config.(*GroupAndVerifyConfig)
	if !ok {
		return errors.New("config is not a GroupAndVerifyConfig")
	}
	gav.config = groupAndVerifyConfig
	if gav.taskChan == nil {
		gav.taskChan = make(chan *Task)
	}
	return nil
}

// Serve is a method that is used to handle incoming data and process it.
//
// Returns:
//
// 1. error. It returns an error if the processing of data fails.
func (gav *GroupAndVerify) Serve() error {
	if gav.config == nil {
		return errors.New("config not set")
	}
	if gav.pushable == nil {
		return errors.New("pushable not set")
	}
	if gav.taskChan == nil {
		return errors.New("taskChan not set")
	}
	if gav.config.parentVerifySet == nil {
		return errors.New("parentVerifySet not set")
	}
	tasksHandler(gav.taskChan, gav.config, gav.pushable)
	return nil
}

// tasksHandler is a method that is used to handle the incoming tasks and process them, sending the output to the pushable
//
// Args:
//
// 1. taskChan: chan *Task. It is the channel that holds the incoming tasks.
//
// 2. config: *GroupAndVerifyConfig. It is the configuration for the GroupAndVerify component.
//
// 3. pushable: Server.Pushable. It is the pushable that is used to send data to the next stage.
func tasksHandler(taskChan chan *Task, config *GroupAndVerifyConfig, pushable Server.Pushable) {
	treeToTaskChannelMap := make(map[string]chan *Task)
	treeCompletionChannel := make(chan string)
	wg := sync.WaitGroup{}
	// make sure we wait for all the goroutines to finish
	defer wg.Wait()
WORK:
	for {
		select {
		case task, ok := <-taskChan:
			if !ok {
				break WORK
			}
			incomingData := task.IncomingData
			treeId := incomingData.TreeId
			if _, ok := treeToTaskChannelMap[treeId]; !ok {
				if config.MaxTrees != 0 && len(treeToTaskChannelMap) >= config.MaxTrees {
					task.errChan <- Server.NewFullError(fmt.Sprintf("max trees reached. MaxTrees: %d", config.MaxTrees))
					continue
				}
				treeToTaskChannelMap[treeId] = make(chan *Task)
				// handle the tree in a separate goroutine and send the errors to all tasks gathered
				wg.Add(1)
				go func(treeChan chan *Task, treeId string) {
					groupAndVerifyLogger.Debug("tree started", slog.Group("details", slog.String("treeId", treeId)))
					defer wg.Done()
					tasks, err := treeHandler(treeChan, config, pushable)
					// defer to send errors to all tasks and remove the tree from the map
					for _, task := range tasks {
						task.errChan <- err
					}
					treeCompletionChannel <- treeId
					for task := range treeChan {
						taskChan <- task
					}
					groupAndVerifyLogger.Debug("tree completed", slog.Group("details", slog.String("treeId", treeId)))
				}(treeToTaskChannelMap[treeId], treeId)
			}
			treeChannel := treeToTaskChannelMap[treeId]
			treeChannel <- task
		case treeId := <-treeCompletionChannel:
			// check if there have been any tasks received in between the time the tree was completed
			if treeChannel, ok := treeToTaskChannelMap[treeId]; ok {
				close(treeChannel)
				delete(treeToTaskChannelMap, treeId)
			}
		}
	}
	// wait for all trees to be completed and close the channels
	if len(treeToTaskChannelMap) == 0 {
		close(treeCompletionChannel)
		return
	}
LOOPBREAK:
	for {
		select {
		case treeId := <-treeCompletionChannel:
			if treeChannel, ok := treeToTaskChannelMap[treeId]; ok {
				close(treeChannel)
				delete(treeToTaskChannelMap, treeId)
			}
		default:
			if len(treeToTaskChannelMap) == 0 {
				break LOOPBREAK
			}
		}
	}
	close(treeCompletionChannel)
}

// treeHandler is a function that is used to handle the incoming data for a tree.
//
// Args:
//
// 1. taskChan: chan *Task. It is the channel that holds the incoming tasks.
//
// 2. config: *GroupAndVerifyConfig. It is the configuration for the GroupAndVerify component.
//
// 3. pushable: Server.Pushable. It is the pushable that is used to send data to the next stage.
//
// Returns:
//
// 1. []*Task. It returns the tasks that were processed.
//
// 2. error. It returns an error if the processing of the tree fails.
func treeHandler(taskChan chan *Task, config *GroupAndVerifyConfig, pushable Server.Pushable) ([]*Task, error) {
	verifiedNodes, tasks, verified, err := processTasks(taskChan, config.Timeout, config.parentVerifySet)
	if err != nil {
		return tasks, err
	}
	if len(tasks) == 0 {
		return tasks, fmt.Errorf("attempted to handle tree with no tasks received. Consider increasing the timeout in config")
	}
	if !verified {
		groupAndVerifyLogger.Warn("tree not verified", slog.Group("details", slog.String("treeId", tasks[0].IncomingData.TreeId)))
	}
	var outgoingData []*IncomingData
	for _, node := range verifiedNodes {
		outgoingData = append(outgoingData, node.incomingData)
		outgoingData = append(outgoingData, node.Duplicates...)
	}
	// create AppData and send to pushable
	jsonData, err := json.Marshal(outgoingData)
	if err != nil {
		return tasks, err
	}
	appData := Server.NewAppData(jsonData, "")
	return tasks, pushable.SendTo(appData)
}

// childBalance is a struct that is used to hold the verification status of the child node.
//
// It has the following fields:
//
// 1. parentRef: bool. It is a boolean that is used to determine if the child node was from a forwards link.
//
// 2. childRef: bool. It is a boolean that is used to determine if the child node was from a backwards link.
type childBalance struct {
	parentRef bool
	childRef  bool
}

// ParentRef is a method that is used to set the parentRef field of the childBalance struct.
func (cb *childBalance) ParentRef() {
	cb.parentRef = true
}

// ChildRef is a method that is used to set the childRef field of the childBalance struct.
func (cb *childBalance) ChildRef() {
	cb.childRef = true
}

// IsVerified is a method that is used to check if the child node has been verified.
//
// Returns:
//
// 1. bool. It returns true if the child node has been verified.
func (cb *childBalance) IsVerified() bool {
	return cb.parentRef && cb.childRef
}

// parentStatus is a struct that is used to hold the verification status of the parent node.
//
// It has the following fields:
//
// 1. childRefBalance: map[string]int. It is the map that holds the reference balance of the children.
// The reference balance is a number that can be added to or subtracted from to determine if the child
// has been verified and can be removed from the map i.e. the value for a child is zero as it has been
// seen twice.
//
// 2. expectedChildren: *int. It is a pointer to the expected number of children for the parent node.
// This is used for the nodes that do not require bidirectional confirmation.
type parentStatus struct {
	childRefBalance  map[string]*childBalance
	expectedChildren *int	
}

// newParentStatus is a function that is used to create a new parentStatus struct.
//
// Returns:
//
// 1. *parentStatus. It returns a new parentStatus struct.
func newParentStatus() *parentStatus {
	return &parentStatus{
		childRefBalance: make(map[string]*childBalance),
	}
}

// updateFromChild is a method that is used to update the verification status of the parent node.
//
// Args:
//
// 1.childId: string. The identifier of the child node.
func (ps *parentStatus) UpdateFromChild(childId string) {
	if _, ok := ps.childRefBalance[childId]; !ok {
		ps.childRefBalance[childId] = &childBalance{}
	}
	ps.childRefBalance[childId].ChildRef()
	if ps.expectedChildren != nil {
		ps.childRefBalance[childId].ParentRef()
	}
}

// updateFromParent is a method that is used to update the verification status of the child node.
//
// Args:
//
// 1. childId: string. The identifier of the child node.
//
// 2. expectedChildren: *int. A pointer to the expected number of children for the parent node.
//
// Returns:
//
// 1. error. It returns an error if there is a problem with the data.
func (ps *parentStatus) UpdateFromParent(childIds []string, expectedChildren *int) error {
	if expectedChildren != nil {
		if len(childIds) != 0 {
			return errors.New("node should have no referenced children as it is in the list of node types where children cannot be referenced")
		}
		ps.expectedChildren = expectedChildren
		for childId := range ps.childRefBalance {
			ps.childRefBalance[childId].ParentRef()
		}
	} else {
		for _, childId := range childIds {
			if _, ok := ps.childRefBalance[childId]; !ok {
				ps.childRefBalance[childId] = &childBalance{}
			}
			ps.childRefBalance[childId].ParentRef()
		}
	}
	return nil
}

// CheckVerified is a method that is used to check if the parent node has been verified.
//
// Returns:
//
// 1. bool. It returns true if the parent node has been verified.
func (ps *parentStatus) CheckVerified() bool {
	if ps.expectedChildren != nil {
		if len(ps.childRefBalance) != *ps.expectedChildren {
			return false
		} else {
			return true
		}
	}
	for _, child := range ps.childRefBalance {
		if !child.IsVerified() {
			return false
		}
	}
	return true
}

// verificationStatusHolder is a struct that is used to hold the verification status.
//
// It has the following fields:
//
// 1. parentStatusHolder: map[string]*parentStatus. It is the map that holds the verification status of the parent nodes.
//
// 2. parentVerifySet: map[string]bool. It is the map that holds the identifiers of the nodes (NodeTypes)
// that do not require bidirectional confirmation and their expected number of children.
//
// 3. verifiedNodes: map[string]bool. It is the map that holds the identifiers of the nodes that have already been verified.
// This is used to prevent the same node from being verified multiple times if duplicates appear in the data.
type verificationStatusHolder struct {
	parentStatusHolder map[string]*parentStatus
	parentVerifySet    map[string]*int
	verifiedNodes      map[string]bool
}

// newVerificationStatusHolder is a function that is used to create a new verificationStatusHolder struct.
//
// Args:
//
// 1. parentVerifySet: map[string]*int. It is the map that holds the identifiers of the nodes (NodeTypes)
// that do not require bidirectional confirmation and their expected number of children.
// Returns:
//
// 1. *verificationStatusHolder. It returns a new verificationStatusHolder struct.
func newVerificationStatusHolder(parentVerifySet map[string]*int) *verificationStatusHolder {
	return &verificationStatusHolder{
		parentStatusHolder: make(map[string]*parentStatus),
		parentVerifySet:    parentVerifySet,
		verifiedNodes:      make(map[string]bool),
	}
}

// updateBackwardsLink is a method that is used to update the backwards link.
//
// Args:
//
// 1. parentId: string. It is the identifier of the parent node.
//
// 2. childId: string. It is the identifier of the child node.
func (vsh *verificationStatusHolder) updateBackwardsLink(parentId string, childId string) {
	// create the parentStatus for the node if it does not exist
	if _, ok := vsh.parentStatusHolder[parentId]; !ok {
		vsh.parentStatusHolder[parentId] = newParentStatus()
	}
	nodeStatus := vsh.parentStatusHolder[parentId]
	nodeStatus.UpdateFromChild(childId)
	if nodeStatus.CheckVerified() {
		delete(vsh.parentStatusHolder, parentId)
		vsh.verifiedNodes[parentId] = true
	}
}

// updateForwardLinks is a method that is used to update the forward links.
//
// Args:
//
// 1. nodeId: string. It is the identifier of the node.
//
// 2. childIds: []string. It is the list of identifiers of the children of the node.
//
// 3. nodeType: string. It is the type of the node. This can be used to identify if
// bidirectional confirmation is required.
//
// Returns:
//
// 1. error. It returns an error if there is a problem with the data.
func (vsh *verificationStatusHolder) updateForwardLinks(nodeId string, childIds []string, nodeType string) error {
	if _, ok := vsh.parentStatusHolder[nodeId]; !ok {
		// create the parentStatus for the node if it does not exist
		vsh.parentStatusHolder[nodeId] = newParentStatus()
	}
	nodeStatus := vsh.parentStatusHolder[nodeId]
	expectedChildren := vsh.parentVerifySet[nodeType]
	err := nodeStatus.UpdateFromParent(childIds, expectedChildren)
	if err != nil {
		return err
	}
	if nodeStatus.CheckVerified() {
		delete(vsh.parentStatusHolder, nodeId)
		vsh.verifiedNodes[nodeId] = true
	}
	return nil
}

// updateVerificationStatus is a method that is used to update the verification status of the nodes and edges.
//
// Args:
//
// 1. incomingData: *IncomingData. It is the incoming data that is to be processed.
//
// Returns:
//
// 1. error. It returns an error if there is a problem with that data
func (vsh *verificationStatusHolder) UpdateVerificationStatus(incomingData *IncomingData) error {
	// updateVerificationStatus is a method that is used to update the verification status of the nodes and edges
	// It returns an error if there is a problem with that data
	if incomingData == nil {
		return errors.New("incomingData is nil")
	}
	nodeId := incomingData.NodeId
	parentId := incomingData.ParentId
	nodeType := incomingData.NodeType
	childIds := incomingData.ChildIds
	// update backwards link
	if parentId != "" {
		if _, ok := vsh.verifiedNodes[parentId]; !ok {
			vsh.updateBackwardsLink(parentId, nodeId)
		}
	}
	// check if node has already been verified
	if _, ok := vsh.verifiedNodes[nodeId]; ok {
		return nil
	}
	err := vsh.updateForwardLinks(nodeId, childIds, nodeType)
	if err != nil {
		return err
	}
	return nil
}

// checkVerificationStatus is a method that is used to check if the verification status
//
// Returns:
//
// 1. bool. It returns true if the verification status is empty.
func (vsh *verificationStatusHolder) CheckVerificationStatus() bool {
	return len(vsh.parentStatusHolder) == 0
}

// incomingDataHolder is a struct that is used to hold the incoming data.
//
// It has the following fields:
//
// 1. incomingData: *IncomingData. It is the incoming data that is to be processed.
//
// 2. backwardsLinks: []string. It is the list of identifiers of the parent nodes.
type incomingDataHolder struct {
	incomingData      *IncomingData
	Duplicates        []*IncomingData
}

// AddDuplicate is a method that is used to add a duplicate to the incomingDataHolder.
//
// Args:
//
// 1. duplicate: *IncomingData. It is the duplicate that is to be added.
func (idh *incomingDataHolder) AddDuplicate(duplicate *IncomingData) {
	idh.Duplicates = append(idh.Duplicates, duplicate)
	if !reflect.DeepEqual(idh.incomingData, duplicate) {
		groupAndVerifyLogger.Warn("incomingData already exists and is not equal", slog.Group("details", slog.String("nodeId", idh.incomingData.NodeId), slog.String("treeId", idh.incomingData.TreeId)))
	}
}

// updateIncomingDataHolderMap is a method that is used to update the incomingDataHolder map.
//
// Args:
//
// 1. incomingDataHolderMap: map[string]*incomingDataHolder. It is the map that holds the incoming data.
//
// 2. incomingData: *IncomingData. It is the incoming data that is to be processed.
//
// Returns:
//
// 1. error. It returns an error if there is a problem with the data.
func updateIncomingDataHolderMap(incomingDataHolderMap map[string]*incomingDataHolder, incomingData *IncomingData) error {
	if incomingData == nil {
		return errors.New("incomingData is nil")
	}
	if _, ok := incomingDataHolderMap[incomingData.NodeId]; !ok {
		incomingDataHolderMap[incomingData.NodeId] = &incomingDataHolder{
			incomingData: incomingData,
		}
	} else {
		if incomingDataHolderMap[incomingData.NodeId].incomingData != nil {
			incomingDataHolderMap[incomingData.NodeId].AddDuplicate(incomingData)
			return nil
		} else {
			incomingDataHolderMap[incomingData.NodeId].incomingData = incomingData
		}
	}
	if incomingData.ParentId != "" {
		if _, ok := incomingDataHolderMap[incomingData.ParentId]; !ok {
			incomingDataHolderMap[incomingData.ParentId] = &incomingDataHolder{}
		}
	}
	return nil
}

// processTree is a method that is used to process the tree with the given treeId.
//
// Args:
//
// 1. incomingDataChannel: chan *IncomingData. It is the channel that holds the incoming data.
//
// 2. timeOut: int. It is the time out for the processing of the tree.
// Returns:
//
// 1. error. It returns an error if the processing of the tree fails.
func processTasks(
	taskChan chan *Task, timeOut int, parentVerifySet map[string]*int,
) (map[string]*incomingDataHolder, []*Task, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeOut)*time.Second)
	verificationStatusHolder := newVerificationStatusHolder(parentVerifySet)
	nodes := make(map[string]*incomingDataHolder)
	var tasks []*Task
	treeVerified := false
	defer cancel()
OUTER:
	for {
		select {
		case task, ok := <-taskChan:
			if !ok {
				break OUTER
			}
			tasks = append(tasks, task)
			incomingData := task.IncomingData
			err := updateIncomingDataHolderMap(nodes, incomingData)
			if err != nil {
				return nil, nil, false, err
			}
			err = verificationStatusHolder.UpdateVerificationStatus(incomingData)
			if err != nil {
				return nil, nil, false, err
			}
			treeVerified = verificationStatusHolder.CheckVerificationStatus()
			if treeVerified {
				break OUTER
			}
		case <-ctx.Done():
			break OUTER
		}
	}
	// remove nodes that have no incomingData that are a reference to a parent that is not present
	for nodeId, node := range nodes {
		if node.incomingData == nil {
			delete(nodes, nodeId)
		}
	}
	return nodes, tasks, treeVerified, nil
}

// SendTo is a method that is used to handle incoming data.
//
// Args:
//
// 1. data: *Server.AppData. It is the incoming data that is to be processed.
//
// Returns:
//
// 1. error. It returns an error if the processing of data fails.
func (gav *GroupAndVerify) SendTo(data *Server.AppData) (err error) {
	if gav.taskChan == nil {
		return errors.New("taskChan not set")
	}
	if data == nil {
		return errors.New("data is nil")
	}
	gotData, err := data.GetData()
	if err != nil {
		return err
	}
	var incomingData *IncomingData
	err = json.Unmarshal(gotData, &incomingData)
	if err != nil {
		return Server.NewInvalidErrorFromError(err)
	}
	groupAndVerifyLogger.Debug("message received", slog.Group("details", slog.String("nodeId", incomingData.NodeId), slog.String("treeId", incomingData.TreeId)))
	task := &Task{
		IncomingData: incomingData,
		errChan:      make(chan error),
	}
	defer close(task.errChan)
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("taskChan is closed")
		}
	}()
	gav.taskChan <- task
	groupAndVerifyLogger.Debug("task sent for processing", slog.Group("details", slog.String("nodeId", incomingData.NodeId), slog.String("treeId", incomingData.TreeId)))
	err = <-task.errChan
	groupAndVerifyLogger.Debug("task processed", slog.Group("details", slog.String("nodeId", incomingData.NodeId), slog.String("treeId", incomingData.TreeId)))
	return err
}
