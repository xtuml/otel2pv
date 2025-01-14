package groupandverify

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/SmartDCSITlimited/CDS-OTel-To-PV/Server"
)

// Test GroupAndVerifyConfig
func TestGroupAndVerifyConfig(t *testing.T) {
	t.Run("IngestConfig", func(t *testing.T) {
		t.Run("ImplementsConfig", func(t *testing.T) {
			// test that GroupAndVerifyConfig implements Server.Config
			groupandverifyConfig := GroupAndVerifyConfig{}
			_, ok := interface{}(&groupandverifyConfig).(Server.Config)
			if !ok {
				t.Errorf("GroupAndVerifyConfig does not implement Server.Config")
			}
		})
		t.Run("updateOrderChildrenByTimestamp", func(t *testing.T) {
			// test error case when orderChildrenByTimestamp is not a boolean
			config := map[string]any{
				"orderChildrenByTimestamp": "not a boolean",
			}
			gavc := GroupAndVerifyConfig{}
			err := gavc.updateOrderChildrenByTimestamp(config)
			if err == nil {
				t.Errorf("expected error, got nil")
			}
			// test case when orderChildrenByTimestamp is a true boolean
			config = map[string]any{
				"orderChildrenByTimestamp": true,
			}
			gavc = GroupAndVerifyConfig{}
			err = gavc.updateOrderChildrenByTimestamp(config)
			if err != nil {
				t.Errorf("expected nil, got %v", err)
			}
			if !gavc.orderChildrenByTimestamp {
				t.Errorf("expected true, got false")
			}
			// test case when orderChildrenByTimestamp is not present and should default to false
			config = map[string]any{}
			gavc = GroupAndVerifyConfig{}
			err = gavc.updateOrderChildrenByTimestamp(config)
			if err != nil {
				t.Errorf("expected nil, got %v", err)
			}
			if gavc.orderChildrenByTimestamp {
				t.Errorf("expected false, got true")
			}
		})
		t.Run("updateGroupApplies", func(t *testing.T) {
			// test error case when groupApplies is not an array
			config := map[string]any{
				"groupApplies": "not an array",
			}
			gavc := GroupAndVerifyConfig{}
			err := gavc.updateGroupApplies(config)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if err.Error() != "groupApplies is not an array" {
				t.Errorf("expected groupApplies is not an array, got %v", err)
			}
			// test error case when groupApplies is not an array of maps
			config = map[string]any{
				"groupApplies": []any{"not a map"},
			}
			gavc = GroupAndVerifyConfig{}
			err = gavc.updateGroupApplies(config)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if err.Error() != "groupApplies[0] is not a map" {
				t.Errorf("expected groupApplies[0] is not a map, got %v", err)
			}
			// test error case when FieldToShare does not exist or is not a string
			config = map[string]any{
				"groupApplies": []any{
					map[string]any{},
				},
			}
			gavc = GroupAndVerifyConfig{}
			err = gavc.updateGroupApplies(config)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if err.Error() != "FieldToShare does not exist or is not a string for groupApplies[0]" {
				t.Errorf("expected FieldToShare does not exist or is not a string for groupApplies[0], got %v", err)
			}
			// test error case when IdentifyingField does not exist or is not a string
			config = map[string]any{
				"groupApplies": []any{
					map[string]any{
						"FieldToShare": "field",
					},
				},
			}
			gavc = GroupAndVerifyConfig{}
			err = gavc.updateGroupApplies(config)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if err.Error() != "IdentifyingField does not exist or is not a string for groupApplies[0]" {
				t.Errorf("expected IdentifyingField does not exist or is not a string for groupApplies[0], got %v", err)
			}
			// test error case when ValueOfIdentifyingField does not exist or is not a string
			config = map[string]any{
				"groupApplies": []any{
					map[string]any{
						"FieldToShare":     "field",
						"IdentifyingField": "field",
					},
				},
			}
			gavc = GroupAndVerifyConfig{}
			err = gavc.updateGroupApplies(config)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if err.Error() != "ValueOfIdentifyingField does not exist or is not a string for groupApplies[0]" {
				t.Errorf("expected ValueOfIdentifyingField does not exist or is not a string for groupApplies[0], got %v", err)
			}
			// test error case when FieldToShare already exists in groupApplies
			config = map[string]any{
				"groupApplies": []any{
					map[string]any{
						"FieldToShare":            "field",
						"IdentifyingField":        "field",
						"ValueOfIdentifyingField": "field",
					},
					map[string]any{
						"FieldToShare":            "field",
						"IdentifyingField":        "field",
						"ValueOfIdentifyingField": "field",
					},
				},
			}
			gavc = GroupAndVerifyConfig{}
			err = gavc.updateGroupApplies(config)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if err.Error() != "FieldToShare field already exists in groupApplies" {
				t.Errorf("expected FieldToShare field already exists in groupApplies, got %v", err)
			}
			// test case when groupApplies input is valid
			config = map[string]any{
				"groupApplies": []any{
					map[string]any{
						"FieldToShare":            "field",
						"IdentifyingField":        "field",
						"ValueOfIdentifyingField": "field",
					},
				},
			}
			gavc = GroupAndVerifyConfig{}
			err = gavc.updateGroupApplies(config)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
			if len(gavc.groupApplies) != 1 {
				t.Fatalf("expected 1, got %d", len(gavc.groupApplies))
			}
			if _, ok := gavc.groupApplies["field"]; !ok {
				t.Errorf("expected field, got %v", gavc.groupApplies)
			}
			expected := GroupApply{
				FieldToShare:            "field",
				IdentifyingField:        "field",
				ValueOfIdentifyingField: "field",
			}
			if gavc.groupApplies["field"] != expected {
				t.Errorf("expected %v, got %v", expected, gavc.groupApplies["field"])
			}
			// test default case when groupApplies is not present
			config = map[string]any{}
			gavc = GroupAndVerifyConfig{}
			err = gavc.updateGroupApplies(config)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
			if len(gavc.groupApplies) != 0 {
				t.Fatalf("expected 0, got %d", len(gavc.groupApplies))
			}
		})
		t.Run("updateParentVerifySet", func(t *testing.T) {
			// test error case when parentVerifySet is not an array
			config := map[string]any{
				"parentVerifySet": "not an array",
			}
			gavc := GroupAndVerifyConfig{}
			err := gavc.updateParentVerifySet(config)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if err.Error() != "parentVerifySet is not an array of strings" {
				t.Errorf("expected parentVerifySet is not an array of strings, got %v", err)
			}
			// test error case when parentVerifySet is not an array of strings
			config = map[string]any{
				"parentVerifySet": []any{1},
			}
			gavc = GroupAndVerifyConfig{}
			err = gavc.updateParentVerifySet(config)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if err.Error() != "parentVerifySet[0] is not a string" {
				t.Errorf("expected parentVerifySet[0] is not a string, got %v", err)
			}
			// test case when parentVerifySet input is valid
			config = map[string]any{
				"parentVerifySet": []any{"string"},
			}
			gavc = GroupAndVerifyConfig{}
			err = gavc.updateParentVerifySet(config)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
			if len(gavc.parentVerifySet) != 1 {
				t.Fatalf("expected 1, got %d", len(gavc.parentVerifySet))
			}
			if _, ok := gavc.parentVerifySet["string"]; !ok {
				t.Errorf("expected string to be map key, got %v", gavc.parentVerifySet)
			}
			// test default case when parentVerifySet is not present
			config = map[string]any{}
			gavc = GroupAndVerifyConfig{}
			err = gavc.updateParentVerifySet(config)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
			if len(gavc.parentVerifySet) != 0 {
				t.Fatalf("expected 0, got %d", len(gavc.parentVerifySet))
			}
		})
		t.Run("updateTimeout", func(t *testing.T) {
			// test error case when timeout is not an int
			config := map[string]any{
				"Timeout": "not an int",
			}
			gavc := GroupAndVerifyConfig{}
			err := gavc.updateTimeout(config)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if err.Error() != "Timeout must be a positive integer" {
				t.Errorf("expected Timeout must be a positive integer, got %v", err)
			}
			// test error case when timeout is not a positive integer
			config = map[string]any{
				"Timeout": -1,
			}
			gavc = GroupAndVerifyConfig{}
			err = gavc.updateTimeout(config)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if err.Error() != "Timeout must be a positive integer" {
				t.Errorf("expected Timeout must be a positive integer, got %v", err)
			}
			// test case when timeout is a positive integer
			config = map[string]any{
				"Timeout": 1,
			}
			gavc = GroupAndVerifyConfig{}
			err = gavc.updateTimeout(config)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
			if gavc.Timeout != 1 {
				t.Errorf("expected 1, got %d", gavc.Timeout)
			}
			// test default case when timeout is not present
			config = map[string]any{}
			gavc = GroupAndVerifyConfig{}
			err = gavc.updateTimeout(config)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
			if gavc.Timeout != 2 {
				t.Errorf("expected 2, got %d", gavc.Timeout)
			}
		})
		t.Run("IngestConfig", func(t *testing.T) {
			// test error case when config is nil
			gavc := GroupAndVerifyConfig{}
			err := gavc.IngestConfig(nil)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if err.Error() != "config is nil" {
				t.Errorf("expected config is nil, got %v", err)
			}
			// test error case when updateOrderChildrenByTimestamp returns an error
			config := map[string]any{
				"orderChildrenByTimestamp": "not a boolean",
			}
			gavc = GroupAndVerifyConfig{}
			err = gavc.IngestConfig(config)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			// test error case when updateGroupApplies returns an error
			config = map[string]any{
				"groupApplies": "not an array",
			}
			gavc = GroupAndVerifyConfig{}
			err = gavc.IngestConfig(config)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			// test error case when updateParentVerifySet returns an error
			config = map[string]any{
				"parentVerifySet": "not an array",
			}
			gavc = GroupAndVerifyConfig{}
			err = gavc.IngestConfig(config)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			// test error case when updateTimeout returns an error
			config = map[string]any{
				"Timeout": "not an int",
			}
			gavc = GroupAndVerifyConfig{}
			err = gavc.IngestConfig(config)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			// test case when all update functions return no error
			config = map[string]any{
				"orderChildrenByTimestamp": true,
				"groupApplies": []any{
					map[string]any{
						"FieldToShare":            "field",
						"IdentifyingField":        "field",
						"ValueOfIdentifyingField": "field",
					},
				},
				"parentVerifySet": []any{"string"},
				"Timeout":          1,
			}
			gavc = GroupAndVerifyConfig{}
			err = gavc.IngestConfig(config)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
			if !gavc.orderChildrenByTimestamp {
				t.Errorf("expected true, got false")
			}
			if len(gavc.groupApplies) != 1 {
				t.Fatalf("expected 1, got %d", len(gavc.groupApplies))
			}
			if _, ok := gavc.groupApplies["field"]; !ok {
				t.Errorf("expected field, got %v", gavc.groupApplies)
			}
			expected := GroupApply{
				FieldToShare:            "field",
				IdentifyingField:        "field",
				ValueOfIdentifyingField: "field",
			}
			if gavc.groupApplies["field"] != expected {
				t.Errorf("expected %v, got %v", expected, gavc.groupApplies["field"])
			}
			if len(gavc.parentVerifySet) != 1 {
				t.Fatalf("expected 1, got %d", len(gavc.parentVerifySet))
			}
			if _, ok := gavc.parentVerifySet["string"]; !ok {
				t.Errorf("expected string to be map key, got %v", gavc.parentVerifySet)
			}
			if gavc.Timeout != 1 {
				t.Errorf("expected 1, got %d", gavc.Timeout)
			}
		})
	})
}

// MockPushable is a mock implementation of Server.Pushable
type MockPushable struct {
	isSendToError bool
	sentData      *Server.AppData
}

// SendTo is a mock implementation of Server.Pushable.SendTo
func (mp *MockPushable) SendTo(data *Server.AppData) error {
	if mp.isSendToError {
		return errors.New("sendTo error for pushable")
	}
	mp.sentData = data
	return nil
}

// MockChanPushable is a mock implementation of Server.Pushable
type MockChanPushable struct {
	isSendToError bool
	appDataChan   chan *Server.AppData
}

// SendTo is a mock implementation of Server.Pushable.SendTo
func (mcp *MockChanPushable) SendTo(data *Server.AppData) error {
	if mcp.isSendToError {
		return errors.New("sendTo error for pushable")
	}
	mcp.appDataChan <- data
	return nil
}

// MockConfig is a mock implementation of Server.Config
type MockConfig struct{}

// IngestConfig is a mock implementation of Server.Config.IngestConfig
func (mc *MockConfig) IngestConfig(config map[string]any) error {
	return nil
}

// Test GroupAndVerify
func TestGroupAndVerify(t *testing.T) {
	t.Run("ImplementsPipeServer", func(t *testing.T) {
		// test that GroupAndVerify implements Server.PipeServer
		groupandverify := GroupAndVerify{}
		_, ok := interface{}(&groupandverify).(Server.PipeServer)
		if !ok {
			t.Errorf("GroupAndVerify does not implement Server.PipeServer")
		}
	})
	t.Run("AddPushable", func(t *testing.T) {
		// Test valid case
		gavc := GroupAndVerify{}
		pushable := &MockPushable{}
		err := gavc.AddPushable(pushable)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
		if gavc.pushable != pushable {
			t.Errorf("expected %v, got %v", pushable, gavc.pushable)
		}
		// Test error case when pushable is already set
		err = gavc.AddPushable(pushable)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
	t.Run("Setup", func(t *testing.T) {
		// Test error case when config is not GroupAndVerifyConfig
		gav := GroupAndVerify{}
		err := gav.Setup(&MockConfig{})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if err.Error() != "config is not a GroupAndVerifyConfig" {
			t.Errorf("expected config is not a GroupAndVerifyConfig, got %v", err)
		}
		// Test case when config is GroupAndVerifyConfig
		gav = GroupAndVerify{}
		config := &GroupAndVerifyConfig{}
		err = gav.Setup(config)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
		// Test error case when config is already set
		err = gav.Setup(config)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if err.Error() != "config already set" {
			t.Errorf("expected config already set, got %v", err)
		}
	})
	t.Run("Serve", func(t *testing.T) {
		// Test case: config is not set
		gav := GroupAndVerify{}
		err := gav.Serve()
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if err.Error() != "config not set" {
			t.Errorf("expected config not set, got %v", err)
		}
		// Test case: pushable is not set
		gav = GroupAndVerify{config: &GroupAndVerifyConfig{}}
		err = gav.Serve()
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if err.Error() != "pushable not set" {
			t.Errorf("expected pushable not set, got %v", err)
		}
		// Test case: taskChan is not set
		gav = GroupAndVerify{config: &GroupAndVerifyConfig{}, pushable: &MockPushable{}}
		err = gav.Serve()
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if err.Error() != "taskChan not set" {
			t.Errorf("expected taskChan not set, got %v", err)
		}
		// Test case: parentVerifySet is not set
		gav = GroupAndVerify{config: &GroupAndVerifyConfig{}, pushable: &MockPushable{}, taskChan: make(chan *Task)}
		err = gav.Serve()
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if err.Error() != "parentVerifySet not set" {
			t.Errorf("expected parentVerifySet not set, got %v", err)
		}
		// Test case: all required fields are set
		gav = GroupAndVerify{config: &GroupAndVerifyConfig{parentVerifySet: map[string]bool{}}, pushable: &MockPushable{}, taskChan: make(chan *Task, 1)}
		close(gav.taskChan)
		err = gav.Serve()
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})
	t.Run("SendTo", func(t *testing.T) {
		// Test case: taskChan is not set
		gav := GroupAndVerify{}
		err := gav.SendTo(&Server.AppData{})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		// Test case: appData is nil
		gav = GroupAndVerify{taskChan: make(chan *Task, 1)}
		err = gav.SendTo(nil)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if err.Error() != "data is nil" {
			t.Errorf("expected data is nil, got %v", err)
		}
		close(gav.taskChan)
		// Test case: data on appData is not a map
		gav = GroupAndVerify{taskChan: make(chan *Task, 1)}
		appData := Server.NewAppData("not a map", "")
		err = gav.SendTo(appData)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if err.Error() != "data is not a map" {
			t.Errorf("expected data is not a map, got %v", err)
		}
		close(gav.taskChan)
		// Test case: map data does not fit the IncomingData expected format
		gav = GroupAndVerify{taskChan: make(chan *Task, 1)}
		appData = Server.NewAppData(map[string]any{}, "")
		err = gav.SendTo(appData)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if err.Error() != "TreeId does not exist or is not a string" {
			t.Errorf("expected TreeId does not exist or is not a string, got %v", err)
		}
		// Test case: all required fields are set but the taskChan is closed
		gav = GroupAndVerify{taskChan: make(chan *Task, 1)}
		appData = Server.NewAppData(map[string]any{
			"treeId":    "tree",
			"nodeId":    "node",
			"parentId":  "",
			"nodeType":  "",
			"childIds":  []any{},
			"timestamp": 0,
			"appJSON":   map[string]any{},
		}, "")
		close(gav.taskChan)
		err = gav.SendTo(appData)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if err.Error() != "taskChan is closed" {
			t.Errorf("expected taskChan is closed, got %v", err)
		}
		// Test case: all required fields are set
		gav = GroupAndVerify{taskChan: make(chan *Task, 1)}
		appData = Server.NewAppData(map[string]any{
			"treeId":    "tree",
			"nodeId":    "node",
			"parentId":  "",
			"nodeType":  "",
			"childIds":  []any{},
			"timestamp": 0,
			"appJSON":   map[string]any{},
		}, "")
		// no error expected
		go func() {
			task := <-gav.taskChan
			task.errChan <- nil
		}()
		err = gav.SendTo(appData)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
		// error expected
		go func() {
			task := <-gav.taskChan
			task.errChan <- errors.New("task error")
		}()
		err = gav.SendTo(appData)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if err.Error() != "task error" {
			t.Errorf("expected task error, got %v", err)
		}
		close(gav.taskChan)
	})
}

// Test parentStatus struct
func TestParentStatus(t *testing.T) {
	t.Run("UpdateFromChild", func(t *testing.T) {
		// Test case: childId is not in parentStatus.childRefBalance
		ps := newParentStatus()
		childId := "child"
		ps.UpdateFromChild(childId)
		if len(ps.childRefBalance) != 1 {
			t.Errorf("expected 1, got %d", len(ps.childRefBalance))
		}
		if _, ok := ps.childRefBalance[childId]; !ok {
			t.Errorf("expected child, got %v", ps.childRefBalance)
		}
		if ps.childRefBalance[childId].IsVerified() {
			t.Errorf("expected false, got true")
		}
		// Test case: childId is already in parentStatus.childRefBalance (unchanged from default)
		ps = newParentStatus()
		ps.childRefBalance[childId] = &childBalance{}
		ps.UpdateFromChild(childId)
		if len(ps.childRefBalance) != 1 {
			t.Errorf("expected 1, got %d", len(ps.childRefBalance))
		}
		if _, ok := ps.childRefBalance[childId]; !ok {
			t.Errorf("expected child, got %v", ps.childRefBalance)
		}
		if ps.childRefBalance[childId].IsVerified() {
			t.Errorf("expected false, got true")
		}
		// Test case: childId is already in parentStatus.childRefBalance (when the parent has been referenced)
		ps = newParentStatus()
		ps.childRefBalance[childId] = &childBalance{parentRef: true}
		ps.UpdateFromChild(childId)
		if len(ps.childRefBalance) != 1 {
			t.Errorf("expected 0, got %d", len(ps.childRefBalance))
		}
		if _, ok := ps.childRefBalance[childId]; !ok {
			t.Errorf("expected child, got %v", ps.childRefBalance)
		}
		if !ps.childRefBalance[childId].IsVerified() {
			t.Errorf("expected true, got false")
		}
		// Test case: childId is not in parentStatus.childRefBalance and the singleChildNoRef is set to true
		ps = newParentStatus()
		ps.singleChildNoRef = true
		ps.UpdateFromChild(childId)
		if len(ps.childRefBalance) != 1 {
			t.Errorf("expected 1, got %d", len(ps.childRefBalance))
		}
		if _, ok := ps.childRefBalance[childId]; !ok {
			t.Errorf("expected child, got %v", ps.childRefBalance)
		}
		if !ps.childRefBalance[childId].IsVerified() {
			t.Errorf("expected true, got false")
		}
		// Tests case child is already verified
		ps = newParentStatus()
		ps.childRefBalance[childId] = &childBalance{parentRef: true, childRef: true}
		ps.UpdateFromChild(childId)
		if len(ps.childRefBalance) != 1 {
			t.Errorf("expected 1, got %d", len(ps.childRefBalance))
		}
		if _, ok := ps.childRefBalance[childId]; !ok {
			t.Errorf("expected child, got %v", ps.childRefBalance)
		}
		if !ps.childRefBalance[childId].IsVerified() {
			t.Errorf("expected true, got false")
		}
	})
	t.Run("UpdateFromParent", func(t *testing.T) {
		t.Run("isSingleChildNoRef is false", func(t *testing.T) {
			// Test case: childIds is empty
			ps := newParentStatus()
			childIds := []string{}
			err := ps.UpdateFromParent(childIds, false)
			if err != nil {
				t.Errorf("expected nil, got %v", err)
			}
			if len(ps.childRefBalance) != 0 {
				t.Errorf("expected 0, got %d", len(ps.childRefBalance))
			}
			if ps.singleChildNoRef {
				t.Errorf("expected false, got true")
			}
			// Test case: childIds has values and childRefBalance is empty
			ps = newParentStatus()
			childIds = []string{"child"}
			err = ps.UpdateFromParent(childIds, false)
			if err != nil {
				t.Errorf("expected nil, got %v", err)
			}
			if len(ps.childRefBalance) != 1 {
				t.Errorf("expected 1, got %d", len(ps.childRefBalance))
			}
			if _, ok := ps.childRefBalance["child"]; !ok {
				t.Errorf("expected child, got %v", ps.childRefBalance)
			}
			if ps.childRefBalance["child"].IsVerified() {
				t.Errorf("expected false, got true")
			}
			if ps.singleChildNoRef {
				t.Errorf("expected false, got true")
			}
			// Test case: childIds has values and childRefBalance is populated with the same id but from a parent reference
			ps = newParentStatus()
			ps.childRefBalance["child"] = &childBalance{parentRef: true}
			err = ps.UpdateFromParent(childIds, false)
			if err != nil {
				t.Errorf("expected nil, got %v", err)
			}
			if len(ps.childRefBalance) != 1 {
				t.Errorf("expected 1, got %d", len(ps.childRefBalance))
			}
			if _, ok := ps.childRefBalance["child"]; !ok {
				t.Errorf("expected child, got %v", ps.childRefBalance)
			}
			if ps.childRefBalance["child"].IsVerified() {
				t.Errorf("expected false, got true")
			}
			if ps.singleChildNoRef {
				t.Errorf("expected false, got true")
			}
			// Test case: childIds has values and childRefBalance is populated with the same id but from a child reference
			ps = newParentStatus()
			ps.childRefBalance["child"] = &childBalance{childRef: true}
			err = ps.UpdateFromParent(childIds, false)
			if err != nil {
				t.Errorf("expected nil, got %v", err)
			}
			if len(ps.childRefBalance) != 1 {
				t.Errorf("expected 1, got %d", len(ps.childRefBalance))
			}
			if _, ok := ps.childRefBalance["child"]; !ok {
				t.Errorf("expected child, got %v", ps.childRefBalance)
			}
			if !ps.childRefBalance["child"].IsVerified() {
				t.Errorf("expected true, got false")
			}
			if ps.singleChildNoRef {
				t.Errorf("expected false, got true")
			}
		})
		t.Run("isSingleChildNoRef is true", func(t *testing.T) {
			// Test case: childIds has values and childRefBalance is empty
			ps := newParentStatus()
			childIds := []string{"child"}
			err := ps.UpdateFromParent(childIds, true)
			if err == nil {
				t.Errorf("expected error, got nil")
			}
			// Test case: childIds is empty and childRefBalance is empty
			ps = newParentStatus()
			childIds = []string{}
			err = ps.UpdateFromParent(childIds, true)
			if err != nil {
				t.Errorf("expected nil, got %v", err)
			}
			if len(ps.childRefBalance) != 0 {
				t.Errorf("expected 0, got %d", len(ps.childRefBalance))
			}
			if !ps.singleChildNoRef {
				t.Errorf("expected true, got false")
			}
			// Test case: childIds is empty and childRefBalance is populated with references from parent
			ps = newParentStatus()
			ps.childRefBalance["child"] = &childBalance{parentRef: true}
			err = ps.UpdateFromParent(childIds, true)
			if err != nil {
				t.Errorf("expected nil, got %v", err)
			}
			if len(ps.childRefBalance) != 1 {
				t.Errorf("expected 1, got %d", len(ps.childRefBalance))
			}
			if _, ok := ps.childRefBalance["child"]; !ok {
				t.Errorf("expected child, got %v", ps.childRefBalance)
			}
			if ps.childRefBalance["child"].IsVerified() {
				t.Errorf("expected false, got true")
			}
			if !ps.singleChildNoRef {
				t.Errorf("expected true, got false")
			}
			// Test case: childIds is empty and childRefBalance is populated with references from child
			ps = newParentStatus()
			ps.childRefBalance["child"] = &childBalance{childRef: true}
			err = ps.UpdateFromParent(childIds, true)
			if err != nil {
				t.Errorf("expected nil, got %v", err)
			}
			if len(ps.childRefBalance) != 1 {
				t.Errorf("expected 0, got %d", len(ps.childRefBalance))
			}
			if _, ok := ps.childRefBalance["child"]; !ok {
				t.Errorf("expected child, got %v", ps.childRefBalance)
			}
			if !ps.childRefBalance["child"].IsVerified() {
				t.Errorf("expected true, got false")
			}
			if !ps.singleChildNoRef {
				t.Errorf("expected true, got false")
			}
		})
		t.Run("CheckVerified", func(t *testing.T) {
			// Test case: parent has no children but is singleChildNoRef true
			ps := newParentStatus()
			ps.singleChildNoRef = true
			if ps.CheckVerified() {
				t.Errorf("expected false, got true")
			}
			// Test case: parent has no children and is singleChildNoRef false
			ps = newParentStatus()
			if !ps.CheckVerified() {
				t.Errorf("expected true, got false")
			}
			// Test case: parent has children but none are verified
			ps = newParentStatus()
			ps.childRefBalance["child"] = &childBalance{}
			if ps.CheckVerified() {
				t.Errorf("expected false, got true")
			}
			// Test case: parent has children and not all are verified
			ps = newParentStatus()
			ps.childRefBalance["child"] = &childBalance{}
			ps.childRefBalance["child2"] = &childBalance{parentRef: true, childRef: true}
			if ps.CheckVerified() {
				t.Errorf("expected false, got true")
			}
			// Test case: parent has children and all are verified
			ps = newParentStatus()
			ps.childRefBalance["child"] = &childBalance{parentRef: true, childRef: true}
			ps.childRefBalance["child2"] = &childBalance{parentRef: true, childRef: true}
			if !ps.CheckVerified() {
				t.Errorf("expected true, got false")
			}
		})
	})
}

// Test verificationStatusHolder struct
func TestVerificationStatusHolder(t *testing.T) {
	t.Run("UpdateVerificationStatus", func(t *testing.T) {
		t.Run("updateBackwardsLink", func(t *testing.T) {
			// Test case: parentId is not in parentStatusHolder
			vsh := newVerificationStatusHolder(map[string]bool{})
			childId := "child"
			parentId := "parent"
			vsh.updateBackwardsLink(parentId, childId)
			if len(vsh.parentStatusHolder) != 1 {
				t.Errorf("expected 1, got %d", len(vsh.parentStatusHolder))
			}
			if _, ok := vsh.parentStatusHolder[parentId]; !ok {
				t.Errorf("expected parent, got %v", vsh.parentStatusHolder)
			}
			if len(vsh.verifiedNodes) != 0 {
				t.Errorf("expected 0, got %d", len(vsh.verifiedNodes))
			}
			// Test case: parentId is in parentStatusHolder but updating with child does not change the parentStatus to verified
			vsh = newVerificationStatusHolder(map[string]bool{})
			vsh.parentStatusHolder[parentId] = newParentStatus()
			vsh.updateBackwardsLink(parentId, childId)
			if len(vsh.parentStatusHolder) != 1 {
				t.Errorf("expected 1, got %d", len(vsh.parentStatusHolder))
			}
			if _, ok := vsh.parentStatusHolder[parentId]; !ok {
				t.Errorf("expected parent, got %v", vsh.parentStatusHolder)
			}
			if len(vsh.verifiedNodes) != 0 {
				t.Errorf("expected 0, got %d", len(vsh.verifiedNodes))
			}
			// Test case: parentId is in parentStatusHolder and updating with child changes the parentStatus to verified
			vsh = newVerificationStatusHolder(map[string]bool{})
			vsh.parentStatusHolder[parentId] = newParentStatus()
			vsh.parentStatusHolder[parentId].childRefBalance[childId] = &childBalance{parentRef: true}
			vsh.updateBackwardsLink(parentId, childId)
			if len(vsh.parentStatusHolder) != 0 {
				t.Errorf("expected 0, got %d", len(vsh.parentStatusHolder))
			}
			if len(vsh.verifiedNodes) != 1 {
				t.Errorf("expected 1, got %d", len(vsh.verifiedNodes))
			}
			if _, ok := vsh.verifiedNodes[parentId]; !ok {
				t.Errorf("expected parent, got %v", vsh.verifiedNodes)
			}
		})
		t.Run("updateForwardsLinks", func(t *testing.T) {
			// Test case: parentId is not in parentStatusHolder, childIds is empty and nodeType is not in parentVerifySet
			vsh := newVerificationStatusHolder(map[string]bool{})
			parentId := "parent"
			childIds := []string{}
			nodeType := "type"
			err := vsh.updateForwardLinks(parentId, childIds, nodeType)
			if err != nil {
				t.Errorf("expected nil, got %v", err)
			}
			if len(vsh.parentStatusHolder) != 0 {
				t.Errorf("expected 0, got %d", len(vsh.parentStatusHolder))
			}
			if len(vsh.verifiedNodes) != 1 {
				t.Errorf("expected 1, got %d", len(vsh.verifiedNodes))
			}
			if _, ok := vsh.verifiedNodes[parentId]; !ok {
				t.Errorf("expected parent, got %v", vsh.verifiedNodes)
			}
			// Test case: parentId is not in parentStatusHolder, childIds is empty and nodeType is in parentVerifySet
			vsh = newVerificationStatusHolder(map[string]bool{"type": true})
			err = vsh.updateForwardLinks(parentId, childIds, nodeType)
			if err != nil {
				t.Errorf("expected nil, got %v", err)
			}
			if len(vsh.parentStatusHolder) != 1 {
				t.Errorf("expected 1, got %d", len(vsh.parentStatusHolder))
			}
			if _, ok := vsh.parentStatusHolder[parentId]; !ok {
				t.Errorf("expected parent, got %v", vsh.parentStatusHolder)
			}
			if len(vsh.verifiedNodes) != 0 {
				t.Errorf("expected 0, got %d", len(vsh.verifiedNodes))
			}
			// Test case: parentId is in parentStatusHolder with no children, childIds is empty and nodeType is not in parentVerifySet
			vsh = newVerificationStatusHolder(map[string]bool{})
			vsh.parentStatusHolder[parentId] = newParentStatus()
			err = vsh.updateForwardLinks(parentId, childIds, nodeType)
			if err != nil {
				t.Errorf("expected nil, got %v", err)
			}
			if len(vsh.parentStatusHolder) != 0 {
				t.Errorf("expected 0, got %d", len(vsh.parentStatusHolder))
			}
			if len(vsh.verifiedNodes) != 1 {
				t.Errorf("expected 1, got %d", len(vsh.verifiedNodes))
			}
			if _, ok := vsh.verifiedNodes[parentId]; !ok {
				t.Errorf("expected parent, got %v", vsh.verifiedNodes)
			}
			// Test case: parentId is in parentStatusHolder with no children, childIds is empty and nodeType is in parentVerifySet
			vsh = newVerificationStatusHolder(map[string]bool{"type": true})
			vsh.parentStatusHolder[parentId] = newParentStatus()
			err = vsh.updateForwardLinks(parentId, childIds, nodeType)
			if err != nil {
				t.Errorf("expected nil, got %v", err)
			}
			if len(vsh.parentStatusHolder) != 1 {
				t.Errorf("expected 1, got %d", len(vsh.parentStatusHolder))
			}
			if _, ok := vsh.parentStatusHolder[parentId]; !ok {
				t.Errorf("expected parent, got %v", vsh.parentStatusHolder)
			}
			if len(vsh.verifiedNodes) != 0 {
				t.Errorf("expected 0, got %d", len(vsh.verifiedNodes))
			}
			// Test case updateFromParent produces an error
			vsh = newVerificationStatusHolder(map[string]bool{"type": true})
			err = vsh.updateForwardLinks("parent", []string{"child"}, "type")
			if err == nil {
				t.Errorf("expected error, got nil")
			}
			// Tests case when updating with children verifies the parent
			vsh = newVerificationStatusHolder(map[string]bool{})
			vsh.parentStatusHolder["parent"] = newParentStatus()
			vsh.parentStatusHolder["parent"].childRefBalance["child"] = &childBalance{childRef: true}
			err = vsh.updateForwardLinks("parent", []string{"child"}, "type")
			if err != nil {
				t.Errorf("expected nil, got %v", err)
			}
			if len(vsh.parentStatusHolder) != 0 {
				t.Errorf("expected 0, got %d", len(vsh.parentStatusHolder))
			}
			if len(vsh.verifiedNodes) != 1 {
				t.Errorf("expected 1, got %d", len(vsh.verifiedNodes))
			}
			if _, ok := vsh.verifiedNodes["parent"]; !ok {
				t.Errorf("expected parent, got %v", vsh.verifiedNodes)
			}
		})
		t.Run("UpdateVerificationStatus", func(t *testing.T) {
			// Test case: where incoming data is nil
			vsh := newVerificationStatusHolder(map[string]bool{})
			err := vsh.UpdateVerificationStatus(nil)
			if err == nil {
				t.Errorf("expected error, got nil")
			}
			// Test case: where updateForwardLinks throws an error
			vsh = newVerificationStatusHolder(map[string]bool{"type": true})
			incomingData := &IncomingData{
				NodeType: "type",
				ChildIds: []string{"child"},
			}
			err = vsh.UpdateVerificationStatus(incomingData)
			if err == nil {
				t.Errorf("expected error, got nil")
			}
			// Test case: where parentId is "" and there are no child ids
			vsh = newVerificationStatusHolder(map[string]bool{})
			incomingData = &IncomingData{NodeId: "node"}
			err = vsh.UpdateVerificationStatus(incomingData)
			if err != nil {
				t.Errorf("expected nil, got %v", err)
			}
			if len(vsh.parentStatusHolder) != 0 {
				t.Errorf("expected 0, got %d", len(vsh.parentStatusHolder))
			}
			if len(vsh.verifiedNodes) != 1 {
				t.Errorf("expected 1, got %d", len(vsh.verifiedNodes))
			}
			// Test case: backwards link updated with ParentId and NodeId and parentId is not in verified nodes
			// and node will not be verified
			vsh = newVerificationStatusHolder(map[string]bool{"type": true})
			incomingData = &IncomingData{NodeId: "node", ParentId: "parent", NodeType: "type"}
			err = vsh.UpdateVerificationStatus(incomingData)
			if err != nil {
				t.Errorf("expected nil, got %v", err)
			}
			if len(vsh.parentStatusHolder) != 2 {
				t.Errorf("expected 1, got %d", len(vsh.parentStatusHolder))
			}
			if _, ok := vsh.parentStatusHolder["parent"]; !ok {
				t.Errorf("expected parent, got %v", vsh.parentStatusHolder)
			}
			if _, ok := vsh.parentStatusHolder["node"]; !ok {
				t.Errorf("expected node, got %v", vsh.parentStatusHolder)
			}
			if len(vsh.verifiedNodes) != 0 {
				t.Errorf("expected 0, got %d", len(vsh.verifiedNodes))
			}
			parentStatus := vsh.parentStatusHolder["parent"]
			if _, ok := parentStatus.childRefBalance["node"]; !ok {
				t.Errorf("expected node, got %v", parentStatus.childRefBalance)
			}
			// Test case: where parent is already verified
			vsh = newVerificationStatusHolder(map[string]bool{"type": true})
			vsh.verifiedNodes["parent"] = true
			incomingData = &IncomingData{NodeId: "node", ParentId: "parent", NodeType: "type"}
			err = vsh.UpdateVerificationStatus(incomingData)
			if err != nil {
				t.Errorf("expected nil, got %v", err)
			}
			if len(vsh.parentStatusHolder) != 1 {
				t.Errorf("expected 0, got %d", len(vsh.parentStatusHolder))
			}
			if _, ok := vsh.parentStatusHolder["node"]; !ok {
				t.Errorf("expected node, got %v", vsh.parentStatusHolder)
			}
			if len(vsh.verifiedNodes) != 1 {
				t.Errorf("expected 1, got %d", len(vsh.verifiedNodes))
			}
			// Test case: where parent is not verified but node is
			vsh = newVerificationStatusHolder(map[string]bool{})
			vsh.verifiedNodes["node"] = true
			incomingData = &IncomingData{NodeId: "node", ParentId: "parent"}
			err = vsh.UpdateVerificationStatus(incomingData)
			if err != nil {
				t.Errorf("expected nil, got %v", err)
			}
			if len(vsh.parentStatusHolder) != 1 {
				t.Errorf("expected 1, got %d", len(vsh.parentStatusHolder))
			}
			if _, ok := vsh.parentStatusHolder["parent"]; !ok {
				t.Errorf("expected parent, got %v", vsh.parentStatusHolder)
			}
			if len(vsh.verifiedNodes) != 1 {
				t.Errorf("expected 1, got %d", len(vsh.verifiedNodes))
			}
			parentStatus = vsh.parentStatusHolder["parent"]
			if _, ok := parentStatus.childRefBalance["node"]; !ok {
				t.Errorf("expected node, got %v", parentStatus.childRefBalance)
			}
			// Test case: where parent verified and node is verified
			vsh = newVerificationStatusHolder(map[string]bool{})
			vsh.verifiedNodes["node"] = true
			vsh.verifiedNodes["parent"] = true
			incomingData = &IncomingData{NodeId: "node", ParentId: "parent"}
			err = vsh.UpdateVerificationStatus(incomingData)
			if err != nil {
				t.Errorf("expected nil, got %v", err)
			}
			if len(vsh.parentStatusHolder) != 0 {
				t.Errorf("expected 0, got %d", len(vsh.parentStatusHolder))
			}
			if len(vsh.verifiedNodes) != 2 {
				t.Errorf("expected 2, got %d", len(vsh.verifiedNodes))
			}
			if _, ok := vsh.verifiedNodes["node"]; !ok {
				t.Errorf("expected node, got %v", vsh.verifiedNodes)
			}
			if _, ok := vsh.verifiedNodes["parent"]; !ok {
				t.Errorf("expected parent, got %v", vsh.verifiedNodes)
			}
		})
	})
}

// Test updateIncomingDataHolderMap
func TestUpdateIncomingDataHolderMap(t *testing.T) {
	// Test case: where incomingData is nil
	incomingDataHolderMap := map[string]*incomingDataHolder{}
	err := updateIncomingDataHolderMap(incomingDataHolderMap, nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err.Error() != "incomingData is nil" {
		t.Errorf("expected incomingData is nil, got %v", err)
	}
	// Test case: incoming nodeId is not in the map and parentId is not in the map
	incomingData := &IncomingData{NodeId: "node", ParentId: "parent"}
	incomingDataHolderMap = map[string]*incomingDataHolder{}
	err = updateIncomingDataHolderMap(incomingDataHolderMap, incomingData)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if len(incomingDataHolderMap) != 2 {
		t.Errorf("expected 2, got %d", len(incomingDataHolderMap))
	}
	if _, ok := incomingDataHolderMap["node"]; !ok {
		t.Errorf("expected node, got %v", incomingDataHolderMap)
	}
	if _, ok := incomingDataHolderMap["parent"]; !ok {
		t.Errorf("expected parent, got %v", incomingDataHolderMap)
	}
	nodeIncomingDataHolder := incomingDataHolderMap["node"]
	if nodeIncomingDataHolder.incomingData != incomingData {
		t.Errorf("expected %v, got %v", incomingData, nodeIncomingDataHolder.incomingData)
	}
	if len(nodeIncomingDataHolder.backwardsLinks) != 0 {
		t.Errorf("expected empty array, got %v", nodeIncomingDataHolder.backwardsLinks)
	}
	parentIncomingDataHolder := incomingDataHolderMap["parent"]
	if parentIncomingDataHolder.incomingData != nil {
		t.Errorf("expected nil, got %v", parentIncomingDataHolder.incomingData)
	}
	if !reflect.DeepEqual(parentIncomingDataHolder.backwardsLinks, []string{"node"}) {
		t.Errorf("expected [node], got %v", parentIncomingDataHolder.backwardsLinks)
	}
	// Test case: incoming nodeId is in the map but with incomingData
	incomingData = &IncomingData{NodeId: "node"}
	incomingDataHolderMap = map[string]*incomingDataHolder{"node": {}}
	err = updateIncomingDataHolderMap(incomingDataHolderMap, incomingData)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if len(incomingDataHolderMap) != 1 {
		t.Errorf("expected 1, got %d", len(incomingDataHolderMap))
	}
	if _, ok := incomingDataHolderMap["node"]; !ok {
		t.Errorf("expected node, got %v", incomingDataHolderMap)
	}
	nodeIncomingDataHolder = incomingDataHolderMap["node"]
	if nodeIncomingDataHolder.incomingData != incomingData {
		t.Errorf("expected %v, got %v", incomingData, nodeIncomingDataHolder.incomingData)
	}
	// Test case: incoming nodeId is in the map and incomingData is not nil but the new incomingData matches that stored
	incomingData1 := &IncomingData{NodeId: "node"}
	incomingData2 := &IncomingData{NodeId: "node"}
	incomingDataHolderMap = map[string]*incomingDataHolder{"node": {incomingData: incomingData1}}
	err = updateIncomingDataHolderMap(incomingDataHolderMap, incomingData2)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if len(incomingDataHolderMap) != 1 {
		t.Errorf("expected 1, got %d", len(incomingDataHolderMap))
	}
	if _, ok := incomingDataHolderMap["node"]; !ok {
		t.Errorf("expected node, got %v", incomingDataHolderMap)
	}
	nodeIncomingDataHolder = incomingDataHolderMap["node"]
	if nodeIncomingDataHolder.incomingData != incomingData1 {
		t.Errorf("expected %v, got %v", incomingData1, nodeIncomingDataHolder.incomingData)
	}
	// Test case: incoming nodeId is in the map and incomingData is not nil but the new incomingData does not match that stored
	incomingData1 = &IncomingData{NodeId: "node"}
	incomingData2 = &IncomingData{NodeId: "node", ParentId: "parent"}
	incomingDataHolderMap = map[string]*incomingDataHolder{"node": {incomingData: incomingData1}}
	err = updateIncomingDataHolderMap(incomingDataHolderMap, incomingData2)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err.Error() != "incomingData already exists for id: \"node\"" {
		t.Errorf("expected incomingData already exists for id: \"node\", got %v", err)
	}
}

// Test processTasks
func TestProcessTasks(t *testing.T) {
	// Test case: error from updateVerificationStatus
	task := &Task{IncomingData: &IncomingData{NodeId: "node", NodeType: "type", ChildIds: []string{"child"}}}
	taskChan := make(chan *Task, 1)
	taskChan <- task
	close(taskChan)
	_, _, verified, err := processTasks(taskChan, 60, map[string]bool{"type": true})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err.Error() != "node should have no referenced children as it is in the list of node types where children cannot be referenced" {
		t.Errorf("expected node should have no referenced children as it is in the list of node types where children cannot be referenced, got %v", err)
	}
	if verified {
		t.Errorf("expected false, got true")
	}
	// Test case: error from updateIncomingDataHolderMap
	task1 := &Task{IncomingData: &IncomingData{NodeId: "node", NodeType: "type"}}
	task2 := &Task{IncomingData: &IncomingData{TreeId: "tree", NodeId: "node"}}
	taskChan = make(chan *Task, 2)
	taskChan <- task1
	taskChan <- task2
	close(taskChan)
	_, _, verified, err = processTasks(taskChan, 60, map[string]bool{"type": true})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err.Error() != "incomingData already exists for id: \"node\"" {
		t.Errorf("expected incomingData already exists for id: \"node\", got %v", err)
	}
	if verified {
		t.Errorf("expected false, got true")
	}
	// Test case tree is verified
	task = &Task{IncomingData: &IncomingData{NodeId: "node"}}
	taskChan = make(chan *Task, 1)
	taskChan <- task
	close(taskChan)
	nodes, tasks, verified, err := processTasks(taskChan, 60, map[string]bool{})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if len(nodes) != 1 {
		t.Errorf("expected 1, got %d", len(nodes))
	}
	if _, ok := nodes["node"]; !ok {
		t.Errorf("expected node, got %v", nodes)
	}
	if len(tasks) != 1 {
		t.Errorf("expected 1, got %d", len(tasks))
	}
	if tasks[0] != task {
		t.Errorf("expected %v, got %v", task, tasks)
	}
	if tasks[0].IncomingData != nodes["node"].incomingData {
		t.Errorf("expected %v, got %v", nodes["node"].incomingData, tasks[0].IncomingData)
	}
	if !verified {
		t.Errorf("expected true, got false")
	}
	// Test case when there is a timeout
	task = &Task{IncomingData: &IncomingData{NodeId: "node", NodeType: "type"}}
	taskChan = make(chan *Task, 1)
	taskChan <- task
	nodes, tasks, verified, err = processTasks(taskChan, 1, map[string]bool{"type": true})
	close(taskChan)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if len(nodes) != 1 {
		t.Errorf("expected 0, got %d", len(nodes))
	}
	if len(tasks) != 1 {
		t.Errorf("expected 0, got %d", len(tasks))
	}
	if verified {
		t.Errorf("expected false, got true")
	}
	// Test case: there is a timeout and there is a reference to a parent node that has not been seen
	task = &Task{IncomingData: &IncomingData{NodeId: "node", NodeType: "type", ParentId: "parent"}}
	taskChan = make(chan *Task, 1)
	taskChan <- task
	nodes, tasks, verified, err = processTasks(taskChan, 1, map[string]bool{})
	close(taskChan)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if len(nodes) != 1 {
		t.Errorf("expected 0, got %d", len(nodes))
	}
	if len(tasks) != 1 {
		t.Errorf("expected 0, got %d", len(tasks))
	}
	if verified {
		t.Errorf("expected false, got true")
	}
	if node, ok := nodes["node"]; !ok {
		t.Errorf("expected node, got %v", nodes)
	} else {
		if node.incomingData != task.IncomingData {
			t.Errorf("expected %v, got %v", task.IncomingData, node.incomingData)
		}
	}
}

// Test outgoingDataFromIncomingDataHolder
func TestOutgoingDataFromIncomingDataHolder(t *testing.T) {
	// Test case: incomingDataHolder is nil
	outgoingData, err := outgoingDataFromIncomingDataHolder(nil, map[string]bool{})
	if outgoingData != nil {
		t.Fatalf("expected nil, got %v", outgoingData)
	}
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err.Error() != "incomingDataHolder is nil" {
		t.Fatalf("expected incomingDataHolder is nil, got %v", err)
	}
	// Test case: incomingDataHolder has no incomingData
	holder := &incomingDataHolder{}
	outgoingData, err = outgoingDataFromIncomingDataHolder(holder, map[string]bool{})
	if outgoingData != nil {
		t.Fatalf("expected nil, got %v", outgoingData)
	}
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err.Error() != "incomingData is nil" {
		t.Fatalf("expected incomingData is nil, got %v", err)
	}
	// Test case: incomingDataHolder has incomingData and no children and nodeType is not in parentVerifySet
	holder = &incomingDataHolder{incomingData: &IncomingData{NodeId: "node", AppJSON: map[string]interface{}{"key": "value"}}}
	outgoingData, err = outgoingDataFromIncomingDataHolder(holder, map[string]bool{})
	if outgoingData == nil {
		t.Fatalf("expected outgoingData, got nil")
	}
	if err != nil {
		t.Fatalf("expected error, got nil")
	}
	if outgoingData.NodeId != "node" {
		t.Errorf("expected node, got %v", outgoingData.NodeId)
	}
	if !reflect.DeepEqual(outgoingData.AppJSON, map[string]interface{}{"key": "value"}) {
		t.Errorf("expected map[string]interface{}{\"key\": \"value\"}, got %v", outgoingData.AppJSON)
	}
	if len(outgoingData.OrderedChildIds) != 0 {
		t.Errorf("expected 0, got %d", len(outgoingData.OrderedChildIds))
	}
	// Test case: incomingDataHolder has incomingData and children and nodeType is not in parentVerifySet
	holder = &incomingDataHolder{incomingData: &IncomingData{NodeId: "node", AppJSON: map[string]interface{}{"key": "value"}, ChildIds: []string{"child"}}}
	outgoingData, err = outgoingDataFromIncomingDataHolder(holder, map[string]bool{})
	if outgoingData == nil {
		t.Fatalf("expected outgoingData, got nil")
	}
	if err != nil {
		t.Fatalf("expected error, got nil")
	}
	if outgoingData.NodeId != "node" {
		t.Errorf("expected node, got %v", outgoingData.NodeId)
	}
	if !reflect.DeepEqual(outgoingData.AppJSON, map[string]interface{}{"key": "value"}) {
		t.Errorf("expected map[string]interface{}{\"key\": \"value\"}, got %v", outgoingData.AppJSON)
	}
	if len(outgoingData.OrderedChildIds) != 1 {
		t.Errorf("expected 1, got %d", len(outgoingData.OrderedChildIds))
	}
	if outgoingData.OrderedChildIds[0] != "child" {
		t.Errorf("expected child, got %v", outgoingData.OrderedChildIds)
	}
	// Test case: incomingDataHolder has incomingData no children and nodeType is in parentVerifySet and has a single backwards link
	holder = &incomingDataHolder{incomingData: &IncomingData{NodeId: "node", AppJSON: map[string]interface{}{"key": "value"}, NodeType: "type"}, backwardsLinks: []string{"child"}}
	outgoingData, err = outgoingDataFromIncomingDataHolder(holder, map[string]bool{"type": true})
	if outgoingData == nil {
		t.Fatalf("expected outgoingData, got nil")
	}
	if err != nil {
		t.Fatalf("expected error, got nil")
	}
	if outgoingData.NodeId != "node" {
		t.Errorf("expected node, got %v", outgoingData.NodeId)
	}
	if !reflect.DeepEqual(outgoingData.AppJSON, map[string]interface{}{"key": "value"}) {
		t.Errorf("expected map[string]interface{}{\"key\": \"value\"}, got %v", outgoingData.AppJSON)
	}
	if len(outgoingData.OrderedChildIds) != 1 {
		t.Errorf("expected 1, got %d", len(outgoingData.OrderedChildIds))
	}
	if outgoingData.OrderedChildIds[0] != "child" {
		t.Errorf("expected child, got %v", outgoingData.OrderedChildIds)
	}
	// Test case: incomingDataHolder has incomingData no children and nodeType is in parentVerifySet and has multiple backwards links
	holder = &incomingDataHolder{incomingData: &IncomingData{NodeId: "node", AppJSON: map[string]interface{}{"key": "value"}, NodeType: "type"}, backwardsLinks: []string{"child", "child2"}}
	outgoingData, err = outgoingDataFromIncomingDataHolder(holder, map[string]bool{"type": true})
	if outgoingData != nil {
		t.Fatalf("expected nil, got %v", outgoingData)
	}
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err.Error() != "node should have exactly one edge. NodeId: node" {
		t.Fatalf("expected node should have exactly one edge. NodeId: node, got %v", err)
	}
}

// Test treeHandler
func TestTreeHandler(t *testing.T) {
	// Test case where processTasks returns an error
	taskChan := make(chan *Task, 1)
	taskChan <- &Task{IncomingData: &IncomingData{NodeId: "node", NodeType: "type", ChildIds: []string{"child"}}}
	close(taskChan)
	config := &GroupAndVerifyConfig{parentVerifySet: map[string]bool{"type": true}, Timeout: 10}
	_, err := treeHandler(taskChan, config, &MockPushable{})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err.Error() != "node should have no referenced children as it is in the list of node types where children cannot be referenced" {
		t.Errorf("expected node should have no referenced children as it is in the list of node types where children cannot be referenced, got %v", err)
	}
	// Test case where outgoingDataFromIncomingDataHolder returns an error
	taskChan = make(chan *Task, 3)
	taskChan <- &Task{IncomingData: &IncomingData{NodeId: "child1", ParentId: "node"}}
	taskChan <- &Task{IncomingData: &IncomingData{NodeId: "child2", ParentId: "node"}}
	taskChan <- &Task{IncomingData: &IncomingData{NodeId: "node", NodeType: "type"}}
	close(taskChan)
	_, err = treeHandler(taskChan, config, &MockPushable{})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err.Error() != "node should have exactly one edge. NodeId: node" {
		t.Errorf("expected node should have exactly one edge. NodeId: node, got %v", err)
	}
	// Test case where SendTo of pushable returns an error
	taskChan = make(chan *Task, 1)
	taskChan <- &Task{IncomingData: &IncomingData{NodeId: "node"}}
	close(taskChan)
	config = &GroupAndVerifyConfig{parentVerifySet: map[string]bool{}, Timeout: 10}
	_, err = treeHandler(taskChan, config, &MockPushable{isSendToError: true})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err.Error() != "sendTo error for pushable" {
		t.Errorf("expected sendTo error, got %v", err)
	}
	// Test case where everything works as expected
	taskChan = make(chan *Task, 1)
	task := &Task{IncomingData: &IncomingData{NodeId: "node", AppJSON: map[string]interface{}{"key": "value"}}}
	taskChan <- task
	close(taskChan)
	pushable := &MockPushable{}
	tasks, err := treeHandler(taskChan, config, pushable)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if tasks == nil {
		t.Fatalf("expected tasks, got nil")
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1, got %d", len(tasks))
	}
	if tasks[0] != task {
		t.Errorf("expected %v, got %v", task, tasks)
	}
	if pushable.sentData == nil {
		t.Fatalf("expected sentData, got nil")
	}
	gotDataRaw, err := pushable.sentData.GetData()
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	gotData, ok := gotDataRaw.([]*OutgoingData)
	if !ok {
		t.Fatalf("expected OutgoingData, got %v", gotDataRaw)
	}
	if len(gotData) != 1 {
		t.Fatalf("expected 1, got %d", len(gotData))
	}
	if gotData[0].NodeId != "node" {
		t.Errorf("expected node, got %v", gotData[0].NodeId)
	}
	if !reflect.DeepEqual(gotData[0].AppJSON, map[string]interface{}{"key": "value"}) {
		t.Errorf("expected map[string]interface{}{\"key\": \"value\"}, got %v", gotData[0].AppJSON)
	}
	if len(gotData[0].OrderedChildIds) != 0 {
		t.Errorf("expected 0, got %d", len(gotData[0].OrderedChildIds))
	}
}

// Test tasksHandler
func TestTasksHandler(t *testing.T) {
	// Test case where two separate trees both complete without errors
	taskChan := make(chan *Task, 2)
	task1 := &Task{IncomingData: &IncomingData{TreeId: "tree1", NodeId: "node1", AppJSON: map[string]any{"key1": "value1"}}, errChan: make(chan error, 1)}
	task2 := &Task{IncomingData: &IncomingData{TreeId: "tree2", NodeId: "node2", AppJSON: map[string]any{"key2": "value2"}}, errChan: make(chan error, 1)}
	taskChan <- task1
	taskChan <- task2
	close(taskChan)
	pushable := &MockChanPushable{appDataChan: make(chan *Server.AppData, 2)}
	config := &GroupAndVerifyConfig{parentVerifySet: map[string]bool{}, Timeout: 10}
	tasksHandler(taskChan, config, pushable)
	close(pushable.appDataChan)
	seenNodes := map[string]bool{}
	for appData := range pushable.appDataChan {
		if appData == nil {
			t.Fatalf("expected appData, got nil")
		}
		gotDataRaw, err := appData.GetData()
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
		gotData, ok := gotDataRaw.([]*OutgoingData)
		if !ok {
			t.Fatalf("expected OutgoingData, got %v", gotDataRaw)
		}
		if len(gotData) != 1 {
			t.Fatalf("expected 1, got %d", len(gotData))
		}
		gotDataEntry := gotData[0]
		if gotDataEntry.NodeId == "node1" {
			if !reflect.DeepEqual(gotDataEntry.AppJSON, map[string]any{"key1": "value1"}) {
				t.Errorf("expected map[string]any{\"key1\":\"value1\"}, got %v", gotDataEntry.AppJSON)
			}
			if len(gotDataEntry.OrderedChildIds) != 0 {
				t.Errorf("expected 0, got %d", len(gotDataEntry.OrderedChildIds))
			}
			seenNodes["node1"] = true
		} else if gotDataEntry.NodeId == "node2" {
			if !reflect.DeepEqual(gotDataEntry.AppJSON, map[string]any{"key2": "value2"}) {
				t.Errorf("expected map[string]any{\"key2\":\"value2\"}, got %v", gotDataEntry.AppJSON)
			}
			if len(gotDataEntry.OrderedChildIds) != 0 {
				t.Errorf("expected 0, got %d", len(gotDataEntry.OrderedChildIds))
			}
			seenNodes["node2"] = true
		} else {
			t.Fatalf("unexpected NodeId: %s", gotDataEntry.NodeId)
		}
	}
	if !reflect.DeepEqual(seenNodes, map[string]bool{"node1": true, "node2": true}) {
		t.Fatalf("expected map[string]bool{\"node1\": true, \"node2\": true}, got %v", seenNodes)
	}
	err := <-task1.errChan
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	err = <-task2.errChan
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	close(task1.errChan)
	close(task2.errChan)
	// Test case where a tree completes with an error
	taskChan = make(chan *Task, 1)
	task := &Task{IncomingData: &IncomingData{TreeId: "tree", NodeId: "node", AppJSON: map[string]any{"key": "value"}}, errChan: make(chan error, 1)}
	taskChan <- task
	close(taskChan)
	pushable = &MockChanPushable{isSendToError: true, appDataChan: make(chan *Server.AppData, 1)}
	tasksHandler(taskChan, config, pushable)
	close(pushable.appDataChan)
	appData := <-pushable.appDataChan
	if appData != nil {
		t.Fatalf("expected nil, got %v", appData)
	}
	close(task.errChan)
	err = <-task.errChan
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err.Error() != "sendTo error for pushable" {
		t.Fatalf("expected sendTo error for pushable, got %v", err)
	}
}

// MockSourceServer is a mock implementation of the SourceServer interface
type MockSourceServer struct {
	dataToSend []any
	pushable   Server.Pushable
}

// AddPushable is a method that adds a pushable to the SourceServer
func (mss *MockSourceServer) AddPushable(pushable Server.Pushable) error {
	mss.pushable = pushable
	return nil
}

// Serve is a method that serves the SourceServer
func (mss *MockSourceServer) Serve() error {
	errChan := make(chan error, len(mss.dataToSend))
	defer close(errChan)
	for _, data := range mss.dataToSend {
		go func() {
			err := mss.pushable.SendTo(Server.NewAppData(data, ""))
			errChan <- err
		}()
	}
	for range mss.dataToSend {
		err := <-errChan
		fmt.Println(err)
		if err != nil {
			return err
		}
	}
	return nil
}

// Setup is a method that sets up the SourceServer
func (mss *MockSourceServer) Setup(config Server.Config) error {
	return nil
}

// MockSinkServer is a mock implementation of the SinkServer interface
type MockSinkServer struct {
	dataReceived chan (any)
	expectedCount int
	receivedCount int
	taskChan	  chan *Task
}

// SendTo is a method that sends data to the SinkServer
func (mss *MockSinkServer) SendTo(data *Server.AppData) error {
	if data == nil {
		return errors.New("data is nil")
	}
	gotData, err := data.GetData()
	if err != nil {
		return err
	}
	mss.dataReceived <- gotData
	mss.receivedCount++
	if mss.receivedCount == mss.expectedCount {
		close(mss.taskChan)
	}
	return nil
}

// Serve is a method that serves the SinkServer
func (mss *MockSinkServer) Serve() error {
	return nil
}

// Setup is a method that sets up the SinkServer
func (mss *MockSinkServer) Setup(config Server.Config) error {
	return nil
}

// Tests GroupAndVerify integrating with RunApp
func TestGroupAndVerifyRunApp(t *testing.T) {
	dataToSend := []any{}
	for i := 0; i < 5; i++ {
		for j := 0; j < 2; j++ {
			mapToSend := map[string]any{
				"treeId": fmt.Sprintf("%d", i),
				"nodeId": fmt.Sprintf("%d", j),
				"appJSON": map[string]interface{}{
					"jobId": fmt.Sprintf("%d", i),
				},
				"nodeType": "type",
				"timestamp": j,
			}
			if j == 0 {
				mapToSend["parentId"] = fmt.Sprintf("%d", j+1)
				mapToSend["childIds"] = []any{}
			}
			if j == 1 {
				mapToSend["parentId"] = ""
				mapToSend["childIds"] = []any{fmt.Sprintf("%d", j-1)}
			}
			dataToSend = append(dataToSend, mapToSend)
		}
	}
	gav := &GroupAndVerify{
		taskChan: make(chan *Task, 10),
	}
	gavConfig := &GroupAndVerifyConfig{
		parentVerifySet: map[string]bool{},
	}
	mockSourceServer := &MockSourceServer{
		dataToSend: dataToSend,
	}
	chanForData := make(chan (any), 10)
	mockSinkServer := &MockSinkServer{
		dataReceived: chanForData,
		taskChan: gav.taskChan,
		expectedCount: 5,
	}
	producerConfigMap := map[string]func() Server.Config{
		"MockSink": func() Server.Config {
			return &MockConfig{}
		},
	}
	consumerConfigMap := map[string]func() Server.Config{
		"MockSource": func() Server.Config {
			return &MockConfig{}
		},
	}
	producerMap := map[string]func() Server.SinkServer{
		"MockSink": func() Server.SinkServer {
			return mockSinkServer
		},
	}
	consumerMap := map[string]func() Server.SourceServer{
		"MockSource": func() Server.SourceServer {
			return mockSourceServer
		},
	}
	// set config file
	tmpFile, err := os.CreateTemp("", "config.json")
	if err != nil {
		t.Errorf("Error creating temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	data := []byte(
		`{"AppConfig":{},"ProducersSetup":{"ProducerConfigs":[{"Type":"MockSink","ProducerConfig":{}}]},"ConsumersSetup":{"ConsumerConfigs":[{"Type":"MockSource","ConsumerConfig":{}}]}}`,
	)
	err = os.WriteFile(tmpFile.Name(), data, 0644)
	if err != nil {
		t.Errorf("Error writing to temp file: %v", err)
	}
	err = flag.Set("config", tmpFile.Name())
	if err != nil {
		t.Errorf("Error setting flag: %v", err)
	}
	// Run the app
	err = Server.RunApp(
		gav, gavConfig,
		producerConfigMap, consumerConfigMap,
		producerMap, consumerMap,
	)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	seenJobIds := map[string]int{}
	close(chanForData)
	for rawReceivedData := range chanForData {
		jobIds := []string{}
		nodeIds := []string{}
		receivedData, ok := rawReceivedData.([]*OutgoingData)
		if !ok {
			t.Fatalf("expected []*OutgoingData, got %v", rawReceivedData)
		}
		if len(receivedData) != 2 {
			t.Fatalf("expected 2, got %d", len(receivedData))
		}
		for _, receivedOutgoingData := range receivedData {
			nodeId := receivedOutgoingData.NodeId
			nodeIds = append(nodeIds, nodeId)
			orderedChildIds := receivedOutgoingData.OrderedChildIds
			if nodeId == "0" {
				if len(orderedChildIds) != 0 {
					t.Errorf("expected 0, got %d", len(orderedChildIds))
				}
			}
			if nodeId == "1" {
				if len(orderedChildIds) != 1 {
					t.Errorf("expected 1, got %d", len(orderedChildIds))
				}
				if orderedChildIds[0] != "0" {
					t.Errorf("expected 0, got %s", orderedChildIds[0])
				}
			}
			appJSON := receivedOutgoingData.AppJSON
			if len(appJSON) != 1 {
				t.Fatalf("expected 1, got %d", len(appJSON))
			}
			jobId, ok := appJSON["jobId"].(string)
			if !ok {
				t.Fatalf("expected string, got %v", appJSON["jobId"])
			}
			jobIds = append(jobIds, jobId)
		}
		if len(jobIds) != 2 {
			t.Fatalf("expected 2, got %d", len(jobIds))
		}
		if jobIds[0] != jobIds[1] {
			t.Errorf("expected %s, got %s", jobIds[0], jobIds[1])
		}
		seenJobIds[jobIds[0]]++
		if len(nodeIds) != 2 {
			t.Fatalf("expected 2, got %d", len(nodeIds))
		}
		sort.Slice(nodeIds, func(i, j int) bool {
			return nodeIds[i] < nodeIds[j]
		})
		if nodeIds[0] != "0" {
			t.Errorf("expected 0, got %s", nodeIds[0])
		}
		if nodeIds[1] != "1" {
			t.Errorf("expected 1, got %s", nodeIds[1])
		}
	}
	if len(seenJobIds) != 5 {
		t.Fatalf("expected 5, got %d", len(seenJobIds))
	}
	if !reflect.DeepEqual(seenJobIds, map[string]int{"0": 1, "1": 1, "2": 1, "3": 1, "4": 1}) {
		t.Fatalf("expected map[string]int{\"0\": 1, \"1\": 1, \"2\": 1, \"3\": 1, \"4\": 1}, got %v", seenJobIds)
	}
}
