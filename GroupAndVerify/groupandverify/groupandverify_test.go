package groupandverify

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"reflect"
	"sort"
	"sync"
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
		t.Run("updateMaxTrees", func(t *testing.T) {
			// test error case when maxTrees is not an int
			config := map[string]any{
				"maxTrees": "not an int",
			}
			gavc := GroupAndVerifyConfig{}
			err := gavc.updateMaxTrees(config)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if err.Error() != "maxTrees must be a positive integer" {
				t.Errorf("expected maxTrees must be a positive integer, got %v", err)
			}
			// test error case when maxTrees is not a positive integer
			config = map[string]any{
				"maxTrees": -1,
			}
			gavc = GroupAndVerifyConfig{}
			err = gavc.updateMaxTrees(config)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if err.Error() != "maxTrees must be a positive integer" {
				t.Errorf("expected maxTrees must be a positive integer, got %v", err)
			}
			// test case when maxTrees is a positive integer
			config = map[string]any{
				"maxTrees": 1,
			}
			gavc = GroupAndVerifyConfig{}
			err = gavc.updateMaxTrees(config)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
			if gavc.MaxTrees != 1 {
				t.Errorf("expected 1, got %d", gavc.MaxTrees)
			}
			// test default case when maxTrees is not present
			config = map[string]any{}
			gavc = GroupAndVerifyConfig{}
			err = gavc.updateMaxTrees(config)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
			if gavc.MaxTrees != 0 {
				t.Errorf("expected 0, got %d", gavc.MaxTrees)
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
			// test error case when updateParentVerifySet returns an error
			config := map[string]any{
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
				"parentVerifySet": []any{"string"},
				"Timeout":         1,
			}
			gavc = GroupAndVerifyConfig{}
			err = gavc.IngestConfig(config)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
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
		// Test case: map data does not fit the IncomingData expected format
		gav = GroupAndVerify{taskChan: make(chan *Task, 1)}
		appData := Server.NewAppData([]byte(`{}`), "")
		err = gav.SendTo(appData)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if err.Error() != "TreeId is required" {
			t.Errorf("expected TreeId is required, got %v", err)
		}
		// Test case: all required fields are set but the taskChan is closed
		gav = GroupAndVerify{taskChan: make(chan *Task, 1)}
		appData = Server.NewAppData(
			[]byte(`{"treeId":"tree","nodeId":"node","parentId":"","nodeType":"","childIds":[],"timestamp":0,"appJSON":{}}`),
			"",
		)
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
		appData = Server.NewAppData(
			[]byte(`{"treeId":"tree","nodeId":"node","parentId":"","nodeType":"","childIds":[],"timestamp":0,"appJSON":{}}`),
			"",
		)
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
	parentIncomingDataHolder := incomingDataHolderMap["parent"]
	if parentIncomingDataHolder.incomingData != nil {
		t.Errorf("expected nil, got %v", parentIncomingDataHolder.incomingData)
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
	if err != nil {
		t.Fatalf("expected error, got nil")
	}
	if len(incomingDataHolderMap) != 1 {
		t.Errorf("expected 1, got %d", len(incomingDataHolderMap))
	}
	if _, ok := incomingDataHolderMap["node"]; !ok {
		t.Errorf("expected node, got %v", incomingDataHolderMap)
	}
	nodeIncomingDataHolder = incomingDataHolderMap["node"]
	if len(nodeIncomingDataHolder.Duplicates) != 1 {
		t.Errorf("expected 1, got %d", len(nodeIncomingDataHolder.Duplicates))
	}
	if nodeIncomingDataHolder.Duplicates[0] != incomingData2 {
		t.Errorf("expected %v, got %v", incomingData2, nodeIncomingDataHolder.Duplicates)
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
	// Test case: duplicate node id without verification
	task1 := &Task{IncomingData: &IncomingData{NodeId: "node", NodeType: "type"}}
	task2 := &Task{IncomingData: &IncomingData{TreeId: "tree", NodeId: "node"}}
	taskChan = make(chan *Task, 2)
	taskChan <- task1
	taskChan <- task2
	close(taskChan)
	nodes, tasks, verified, err := processTasks(taskChan, 60, map[string]bool{"type": true})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if verified {
		t.Errorf("expected false, got true")
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2, got %d", len(tasks))
	}
	if tasks[0] != task1 {
		t.Errorf("expected %v, got %v", task1, tasks[0])
	}
	if tasks[1] != task2 {
		t.Errorf("expected %v, got %v", task2, tasks[1])
	}
	if len(nodes) != 1 {
		t.Errorf("expected 1, got %d", len(nodes))
	}
	if _, ok := nodes["node"]; !ok {
		t.Errorf("expected node, got %v", nodes)
	}
	if nodes["node"].incomingData != task1.IncomingData {
		t.Errorf("expected %v, got %v", task1.IncomingData, nodes["node"].incomingData)
	}
	if len(nodes["node"].Duplicates) != 1 {
		t.Errorf("expected 1, got %d", len(nodes["node"].Duplicates))
	}
	if nodes["node"].Duplicates[0] != task2.IncomingData {
		t.Errorf("expected %v, got %v", task2.IncomingData, nodes["node"].Duplicates)
	}
	// Test case: duplicate node id with verification after duplicate received
	task1 = &Task{IncomingData: &IncomingData{NodeId: "node", ChildIds: []string{"child"}}}
	task2 = &Task{IncomingData: &IncomingData{NodeId: "node", ChildIds: []string{"child"}}}
	task3 := &Task{IncomingData: &IncomingData{NodeId: "child", ParentId: "node"}}
	taskChan = make(chan *Task, 3)
	taskChan <- task1
	taskChan <- task2
	taskChan <- task3
	close(taskChan)
	nodes, tasks, verified, err = processTasks(taskChan, 60, map[string]bool{})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if !verified {
		t.Errorf("expected true, got false")
	}
	if len(tasks) != 3 {
		t.Errorf("expected 3, got %d", len(tasks))
	}
	if tasks[0] != task1 {
		t.Errorf("expected %v, got %v", task1, tasks[0])
	}
	if tasks[1] != task2 {
		t.Errorf("expected %v, got %v", task2, tasks[1])
	}
	if tasks[2] != task3 {
		t.Errorf("expected %v, got %v", task3, tasks[2])
	}
	if len(nodes) != 2 {
		t.Errorf("expected 2, got %d", len(nodes))
	}
	if _, ok := nodes["node"]; !ok {
		t.Errorf("expected node, got %v", nodes)
	}
	if nodes["node"].incomingData != task1.IncomingData {
		t.Errorf("expected %v, got %v", task1.IncomingData, nodes["node"].incomingData)
	}
	if len(nodes["node"].Duplicates) != 1 {
		t.Errorf("expected 1, got %d", len(nodes["node"].Duplicates))
	}
	if nodes["node"].Duplicates[0] != task2.IncomingData {
		t.Errorf("expected %v, got %v", task2.IncomingData, nodes["node"].Duplicates)
	}
	if _, ok := nodes["child"]; !ok {
		t.Errorf("expected child, got %v", nodes)
	}
	if nodes["child"].incomingData != task3.IncomingData {
		t.Errorf("expected %v, got %v", task3.IncomingData, nodes["child"].incomingData)
	}
	// Test case: duplicate node id with verification before duplicate received
	task1 = &Task{IncomingData: &IncomingData{NodeId: "node", ChildIds: []string{"child"}}}
	task2 = &Task{IncomingData: &IncomingData{NodeId: "child", ParentId: "node"}}
	task3 = &Task{IncomingData: &IncomingData{NodeId: "node", ChildIds: []string{"child"}}}
	taskChan = make(chan *Task, 3)
	taskChan <- task1
	taskChan <- task2
	taskChan <- task3
	close(taskChan)
	nodes, tasks, verified, err = processTasks(taskChan, 2, map[string]bool{})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if !verified {
		t.Errorf("expected true, got false")
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2, got %d", len(tasks))
	}
	if tasks[0] != task1 {
		t.Errorf("expected %v, got %v", task1, tasks[0])
	}
	if tasks[1] != task2 {
		t.Errorf("expected %v, got %v", task2, tasks[1])
	}
	if len(nodes) != 2 {
		t.Errorf("expected 2, got %d", len(nodes))
	}
	if _, ok := nodes["node"]; !ok {
		t.Errorf("expected node, got %v", nodes)
	}
	if nodes["node"].incomingData != task1.IncomingData {
		t.Errorf("expected %v, got %v", task1.IncomingData, nodes["node"].incomingData)
	}
	if len(nodes["node"].Duplicates) != 0 {
		t.Errorf("expected 0, got %d", len(nodes["node"].Duplicates))
	}
	if _, ok := nodes["child"]; !ok {
		t.Errorf("expected child, got %v", nodes)
	}
	if nodes["child"].incomingData != task2.IncomingData {
		t.Errorf("expected %v, got %v", task2.IncomingData, nodes["child"].incomingData)
	}
	select {
		case leftOverTask := <-taskChan:
			if leftOverTask != task3 {
				t.Fatalf("expected %v, got %v", task3, leftOverTask)
			}
		default:
			t.Fatalf("expected task, got none")
	}
	// Test case tree is verified
	task = &Task{IncomingData: &IncomingData{NodeId: "node"}}
	taskChan = make(chan *Task, 1)
	taskChan <- task
	close(taskChan)
	nodes, tasks, verified, err = processTasks(taskChan, 60, map[string]bool{})
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
	task := &Task{IncomingData: &IncomingData{TreeId: "tree", NodeId: "node", AppJSON: json.RawMessage(`{"key": "value"}`)}}
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
	var outgoingData []*IncomingData
	err = json.Unmarshal(gotDataRaw, &outgoingData)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if len(outgoingData) != 1 {
		t.Fatalf("expected 1, got %d", len(outgoingData))
	}
	if outgoingData[0].NodeId != "node" {
		t.Errorf("expected node, got %v", outgoingData[0].NodeId)
	}
	if !reflect.DeepEqual(outgoingData[0].AppJSON, json.RawMessage(`{"key":"value"}`)) {
		t.Errorf("expected %v, got %v", json.RawMessage(`{"key":"value"}`), outgoingData[0].AppJSON)
	}
	if len(outgoingData[0].ChildIds) != 0 {
		t.Errorf("expected 0, got %d", len(outgoingData[0].ChildIds))
	}
	// Tests case where a duplicate node is received and there is verification
	taskChan = make(chan *Task, 3)
	task1 := &Task{IncomingData: &IncomingData{TreeId: "tree", NodeId: "node", NodeType: "type", ChildIds: []string{"child"}, AppJSON: json.RawMessage(`{}`)}}
	task2 := &Task{IncomingData: &IncomingData{TreeId: "tree", NodeId: "node", NodeType: "type", ChildIds: []string{"child"}, AppJSON: json.RawMessage(`{}`)}}
	task3 := &Task{IncomingData: &IncomingData{TreeId: "tree", NodeId: "child", ParentId: "node", AppJSON: json.RawMessage(`{}`)}}
	taskChan <- task1
	taskChan <- task2
	taskChan <- task3
	close(taskChan)
	config = &GroupAndVerifyConfig{parentVerifySet: map[string]bool{}, Timeout: 10}
	pushable = &MockPushable{}
	tasks, err = treeHandler(taskChan, config, pushable)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if tasks == nil {
		t.Fatalf("expected tasks, got nil")
	}
	if len(tasks) != 3 {
		t.Fatalf("expected 3, got %d", len(tasks))
	}
	if tasks[0] != task1 {
		t.Errorf("expected %v, got %v", task1, tasks[0])
	}
	if tasks[1] != task2 {
		t.Errorf("expected %v, got %v", task2, tasks[1])
	}
	if tasks[2] != task3 {
		t.Errorf("expected %v, got %v", task3, tasks[2])
	}
	if pushable.sentData == nil {
		t.Fatalf("expected sentData, got nil")
	}
	gotDataRaw, err = pushable.sentData.GetData()
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	err = json.Unmarshal(gotDataRaw, &outgoingData)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if len(outgoingData) != 3 {
		t.Fatalf("expected 3, got %d", len(outgoingData))
	}
	seenNodes := map[string]int{}
	for _, outgoingDataEntry := range outgoingData {
		if outgoingDataEntry.NodeId == "node" {
			if !reflect.DeepEqual(outgoingDataEntry.AppJSON, json.RawMessage(`{}`)) {
				t.Errorf("expected %v, got %v", json.RawMessage(`{}`), outgoingDataEntry.AppJSON)
			}
			if len(outgoingDataEntry.ChildIds) != 1 {
				t.Errorf("expected 1, got %d", len(outgoingDataEntry.ChildIds))
			}
			if outgoingDataEntry.ChildIds[0] != "child" {
				t.Errorf("expected child, got %v", outgoingDataEntry.ChildIds[0])
			}
			seenNodes["node"]++
		} else if outgoingDataEntry.NodeId == "child" {
			if !reflect.DeepEqual(outgoingDataEntry.AppJSON, json.RawMessage(`{}`)) {
				t.Errorf("expected %v, got %v", json.RawMessage(`{}`), outgoingDataEntry.AppJSON)
			}
			if len(outgoingDataEntry.ChildIds) != 0 {
				t.Errorf("expected 0, got %d", len(outgoingDataEntry.ChildIds))
			}
			seenNodes["child"]++
		} else {
			t.Fatalf("unexpected NodeId: %s", outgoingDataEntry.NodeId)
		}
	}
	if !reflect.DeepEqual(seenNodes, map[string]int{"node": 2, "child": 1}) {
		t.Fatalf("expected map[string]int{\"node\": 2, \"child\": 1}, got %v", seenNodes)
	}
}

// Test tasksHandler
func TestTasksHandler(t *testing.T) {
	// Test case where two separate trees both complete without errors
	taskChan := make(chan *Task, 2)
	task1 := &Task{IncomingData: &IncomingData{TreeId: "tree1", NodeId: "node1", AppJSON: json.RawMessage(`{"key1":"value1"}`)}, errChan: make(chan error, 1)}
	task2 := &Task{IncomingData: &IncomingData{TreeId: "tree2", NodeId: "node2", AppJSON: json.RawMessage(`{"key2":"value2"}`)}, errChan: make(chan error, 1)}
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
		var outgoingData []*IncomingData
		err = json.Unmarshal(gotDataRaw, &outgoingData)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
		if len(outgoingData) != 1 {
			t.Fatalf("expected 1, got %d", len(outgoingData))
		}
		outgoingDataEntry := outgoingData[0]
		if outgoingDataEntry.NodeId == "node1" {
			if !reflect.DeepEqual(outgoingDataEntry.AppJSON, json.RawMessage(`{"key1":"value1"}`)) {
				t.Errorf("expected %v, got %v", json.RawMessage(`{"key1":"value1"}`), outgoingDataEntry.AppJSON)
			}
			if len(outgoingDataEntry.ChildIds) != 0 {
				t.Errorf("expected 0, got %d", len(outgoingDataEntry.ChildIds))
			}
			seenNodes["node1"] = true
		} else if outgoingDataEntry.NodeId == "node2" {
			if !reflect.DeepEqual(outgoingDataEntry.AppJSON, json.RawMessage(`{"key2":"value2"}`)) {
				t.Errorf("expected %v, got %v", json.RawMessage(`{"key2":"value2"}`), outgoingDataEntry.AppJSON)
			}
			if len(outgoingDataEntry.ChildIds) != 0 {
				t.Errorf("expected 0, got %d", len(outgoingDataEntry.ChildIds))
			}
			seenNodes["node2"] = true
		} else {
			t.Fatalf("unexpected NodeId: %s", outgoingDataEntry.NodeId)
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
	task := &Task{IncomingData: &IncomingData{TreeId: "tree", NodeId: "node", AppJSON: json.RawMessage(`{"key":"value"}`)}, errChan: make(chan error, 1)}
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
	t.Run("maxTrees", func(t *testing.T) {
		// Test case where maxTrees is reached
		taskChan = make(chan *Task, 2)
		task1 := &Task{IncomingData: &IncomingData{TreeId: "tree1", NodeId: "node1", ChildIds: []string{"nodeNotPresent"}, AppJSON: json.RawMessage(`{"key1":"value1"}`)}, errChan: make(chan error, 1)}
		task2 := &Task{IncomingData: &IncomingData{TreeId: "tree2", NodeId: "node2", AppJSON: json.RawMessage(`{"key2":"value2"}`)}, errChan: make(chan error, 1)}
		taskChan <- task1
		taskChan <- task2
		close(taskChan)
		pushable = &MockChanPushable{appDataChan: make(chan *Server.AppData, 2)}
		config = &GroupAndVerifyConfig{parentVerifySet: map[string]bool{}, Timeout: 2, MaxTrees: 1}
		tasksHandler(taskChan, config, pushable)
		close(pushable.appDataChan)
		seenNodes = map[string]bool{}
		for appData := range pushable.appDataChan {
			if appData == nil {
				t.Fatalf("expected appData, got nil")
			}
			gotDataRaw, err := appData.GetData()
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
			var outgoingData []*IncomingData
			err = json.Unmarshal(gotDataRaw, &outgoingData)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
			if len(outgoingData) != 1 {
				t.Fatalf("expected 1, got %d", len(outgoingData))
			}
			outgoingDataEntry := outgoingData[0]
			if outgoingDataEntry.NodeId == "node1" {
				if !reflect.DeepEqual(outgoingDataEntry.AppJSON, json.RawMessage(`{"key1":"value1"}`)) {
					t.Errorf("expected %v, got %v", json.RawMessage(`{"key1":"value1"}`), outgoingDataEntry.AppJSON)
				}
				if !reflect.DeepEqual(outgoingDataEntry.ChildIds, []string{"nodeNotPresent"}) {
					t.Errorf("expected [nodeNotPresent], got %v", outgoingDataEntry.ChildIds)
				}
				seenNodes["node1"] = true

			}
		}
		if !reflect.DeepEqual(seenNodes, map[string]bool{"node1": true}) {
			t.Fatalf("expected map[string]bool{\"node1\": true}, got %v", seenNodes)
		}
		err = <-task1.errChan
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
		err = <-task2.errChan
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if err.Error() != "max trees reached. MaxTrees: 1" {
			t.Fatalf("expected max trees reached. MaxTrees: 1, got %v", err)
		}
		close(task1.errChan)
		close(task2.errChan)
		if _, ok := err.(*Server.FullError); !ok {
			t.Fatalf("expected FullError, got %T", err)
		}
	})
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
	wg := sync.WaitGroup{}
	for _, data := range mss.dataToSend {
		wg.Add(1)
		go func() {
			defer wg.Done()
			jsonBytes, err := json.Marshal(data)
			if err != nil {
				errChan <- err
				return
			}
			err = mss.pushable.SendTo(Server.NewAppData(jsonBytes, ""))
			errChan <- err
		}()
	}
	wg.Wait()
	for range mss.dataToSend {
		err := <-errChan
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
	dataReceived  chan ([]byte)
	expectedCount int
	receivedCount int
	taskChan      chan *Task
	mu 		  sync.Mutex
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
	mss.mu.Lock()
	mss.receivedCount++
	if mss.receivedCount == mss.expectedCount {
		close(mss.taskChan)
	}
	mss.mu.Unlock()
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
				"nodeType":  "type",
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
	chanForData := make(chan ([]byte), 10)
	mockSinkServer := &MockSinkServer{
		dataReceived:  chanForData,
		taskChan:      gav.taskChan,
		expectedCount: 5,
		mu: sync.Mutex{},
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
		var receivedData []*IncomingData
		err := json.Unmarshal(rawReceivedData, &receivedData)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
		if len(receivedData) != 2 {
			t.Fatalf("expected 2, got %d", len(receivedData))
		}
		for _, receivedOutgoingData := range receivedData {
			nodeId := receivedOutgoingData.NodeId
			nodeIds = append(nodeIds, nodeId)
			orderedChildIds := receivedOutgoingData.ChildIds
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
			appJSONRaw := receivedOutgoingData.AppJSON
			var appJSON map[string]interface{}
			err = json.Unmarshal(appJSONRaw, &appJSON)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
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

// Benchmark RunApp with 100 trees each with 10 nodes i.e. 1000 nodes
func BenchmarkGroupAndVerifyRunApp(b *testing.B) {
	dataToSend := []any{}
	numTrees := 100
	numNodes := 10
	for j := 0; j < numNodes; j++ {
		for i := 0; i < numTrees; i++ {
			mapToSend := map[string]any{
				"treeId": fmt.Sprintf("%d", i),
				"nodeId": fmt.Sprintf("%d", j),
				"appJSON": map[string]interface{}{
					"jobId": fmt.Sprintf("%d", i),
				},
				"nodeType":  "type",
				"timestamp": j,
			}
			if j == 0 {
				mapToSend["parentId"] = fmt.Sprintf("%d", j+1)
				mapToSend["childIds"] = []any{}
			} else if j == numNodes-1 {
				mapToSend["parentId"] = ""
				mapToSend["childIds"] = []any{fmt.Sprintf("%d", j-1)}
			} else {
				mapToSend["parentId"] = fmt.Sprintf("%d", j+1)
				mapToSend["childIds"] = []any{fmt.Sprintf("%d", j-1)}
			}
			dataToSend = append(dataToSend, mapToSend)
		}
	}
	// set config file
	tmpFile, err := os.CreateTemp("", "config.json")
	if err != nil {
		b.Errorf("Error creating temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	data := []byte(
		`{"AppConfig":{"Timeout":10},"ProducersSetup":{"ProducerConfigs":[{"Type":"MockSink","ProducerConfig":{}}]},"ConsumersSetup":{"ConsumerConfigs":[{"Type":"MockSource","ConsumerConfig":{}}]}}`,
	)
	err = os.WriteFile(tmpFile.Name(), data, 0644)
	if err != nil {
		b.Errorf("Error writing to temp file: %v", err)
	}
	err = flag.Set("config", tmpFile.Name())
	if err != nil {
		b.Errorf("Error setting flag: %v", err)
	}
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		gav := &GroupAndVerify{
			taskChan: make(chan *Task, numTrees*numNodes),
		}
		gavConfig := &GroupAndVerifyConfig{
			parentVerifySet: map[string]bool{},
		}
		mockSourceServer := &MockSourceServer{
			dataToSend: dataToSend,
		}
		chanForData := make(chan ([]byte), numTrees*numNodes)
		mockSinkServer := &MockSinkServer{
			dataReceived:  chanForData,
			taskChan:      gav.taskChan,
			expectedCount: numTrees,
			mu: sync.Mutex{},
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
		// Run the app
		b.StartTimer()
		err = Server.RunApp(
			gav, gavConfig,
			producerConfigMap, consumerConfigMap,
			producerMap, consumerMap,
		)
		if err != nil {
			b.Fatalf("Error: %v", err)
		}
		close(chanForData)
	}
}
