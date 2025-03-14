package sequencer

import (
	"encoding/json"
	"errors"
	"flag"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/SmartDCSITlimited/CDS-OTel-To-PV/Server"
)

// Test OutputAppFieldSequenceType
func TestOutputAppFieldSequenceType(t *testing.T) {
	tests := []struct {
		name string
		want OutputAppFieldSequenceType
	}{
		{
			name: "array",
			want: Array,
		},
		{
			name: "string",
			want: String,
		},
	}
	incorrect := OutputAppFieldSequenceType("incorrect")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := OutputAppFieldSequenceType(tt.name); got != tt.want {
				t.Errorf("OutputAppFieldSequenceType() = %v, want %v", got, tt.want)
			}
			// Test incorrect value
			if incorrect == tt.want {
				t.Errorf("OutputAppFieldSequenceType() = %v, want %v", incorrect, tt.want)
			}
		})
	}
}

// Test GetOutputAppFieldSequenceType
func TestGetOutputAppFieldSequenceType(t *testing.T) {
	tests := []struct {
		name    string
		want    OutputAppFieldSequenceType
		wantErr bool
	}{
		{
			name:    "array",
			want:    Array,
			wantErr: false,
		},
		{
			name:    "string",
			want:    String,
			wantErr: false,
		},
		{
			name:    "incorrect",
			want:    OutputAppFieldSequenceType(""),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOutputAppFieldSequenceType(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOutputAppFieldSequenceType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetOutputAppFieldSequenceType() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Tests SequencerConfig
func TestSequencerConfig(t *testing.T) {
	t.Run("ImplementsConfig", func(t *testing.T) {
		sc := &SequencerConfig{}
		if _, ok := interface{}(sc).(Server.Config); !ok {
			t.Errorf("SequencerConfig does not implement Config")
		}
	})
	t.Run("IngestConfig", func(t *testing.T) {
		tests := []struct {
			name    string
			config  map[string]any
			wantErr bool
		}{
			{
				name: "valid",
				config: map[string]any{
					"outputAppSequenceField":      "SeqField",
					"outputAppFieldSequenceIdMap": "SeqIdMap",
					"outputAppFieldSequenceType":  "string",
				},
				wantErr: false,
			},
			{
				name: "validDefaultFields",
				config: map[string]any{
					"outputAppSequenceField": "SeqField",
				},
				wantErr: false,
			},
			{
				name:    "invalidOutputAppSequenceFieldNotSet",
				config:  map[string]any{},
				wantErr: true,
			},
			{
				name: "invalidOutputAppSequenceFieldNotString",
				config: map[string]any{
					"outputAppSequenceField": 1,
				},
				wantErr: true,
			},
			{
				name: "invalidOutputAppFieldSequenceIdMapNotString",
				config: map[string]any{
					"outputAppSequenceField":      "SeqField",
					"outputAppFieldSequenceIdMap": 1,
				},
				wantErr: true,
			},
			{
				name: "invalidOutputAppFieldSequenceTypeNotString",
				config: map[string]any{
					"outputAppSequenceField":     "SeqField",
					"outputAppFieldSequenceType": 1,
				},
				wantErr: true,
			},
			{
				name: "invalidOutputAppFieldSequenceTypeUnrecognisedType",
				config: map[string]any{
					"outputAppSequenceField":     "SeqField",
					"outputAppFieldSequenceType": "incorrect",
				},
				wantErr: true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				sc := &SequencerConfig{}
				if err := sc.IngestConfig(tt.config); (err != nil) != tt.wantErr {
					t.Errorf("SequencerConfig.IngestConfig() error = %v, wantErr %v", err, tt.wantErr)
				}
				if tt.name == "valid" {
					if sc.outputAppSequenceField != "SeqField" {
						t.Errorf("SequencerConfig.IngestConfig() outputAppSequenceField = %v, want %v", sc.outputAppSequenceField, "SeqField")
					}
					if sc.outputAppFieldSequenceIdMap != "SeqIdMap" {
						t.Errorf("SequencerConfig.IngestConfig() outputAppFieldSequenceIdMap = %v, want %v", sc.outputAppFieldSequenceIdMap, "SeqIdMap")
					}
					if sc.outputAppFieldSequenceType != String {
						t.Errorf("SequencerConfig.IngestConfig() outputAppFieldSequenceType = %v, want %v", sc.outputAppFieldSequenceType, String)
					}
				}
				if tt.name == "validDefaultFields" {
					if sc.outputAppSequenceField != "SeqField" {
						t.Errorf("SequencerConfig.IngestConfig() outputAppSequenceField = %v, want %v", sc.outputAppSequenceField, "SeqField")
					}
					if sc.outputAppFieldSequenceIdMap != "" {
						t.Errorf("SequencerConfig.IngestConfig() outputAppFieldSequenceIdMap = %v, want %v", sc.outputAppFieldSequenceIdMap, "")
					}
					if sc.outputAppFieldSequenceType != Array {
						t.Errorf("SequencerConfig.IngestConfig() outputAppFieldSequenceType = %v, want %v", sc.outputAppFieldSequenceType, Array)
					}
				}
			})
		}
		t.Run("updateGroupApplies", func(t *testing.T) {
			// test error case when groupApplies is not an array
			config := map[string]any{
				"groupApplies": "not an array",
			}
			sequencer := SequencerConfig{}
			err := sequencer.updateGroupApplies(config)
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
			sequencer = SequencerConfig{}
			err = sequencer.updateGroupApplies(config)
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
			sequencer = SequencerConfig{}
			err = sequencer.updateGroupApplies(config)
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
			sequencer = SequencerConfig{}
			err = sequencer.updateGroupApplies(config)
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
			sequencer = SequencerConfig{}
			err = sequencer.updateGroupApplies(config)
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
			sequencer = SequencerConfig{}
			err = sequencer.updateGroupApplies(config)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
			if len(sequencer.groupApplies) != 1 {
				t.Fatalf("expected 2, got %d", len(sequencer.groupApplies))
			}
			if _, ok := sequencer.groupApplies["field"]; !ok {
				t.Errorf("expected field, got %v", sequencer.groupApplies)
			}
			expected := GroupApply{
				FieldToShare:            "field",
				IdentifyingField:        "field",
				ValueOfIdentifyingField: "field",
			}
			if len(sequencer.groupApplies["field"]) != 2 {
				t.Fatalf("expected 2, got %d", len(sequencer.groupApplies["field"]))
			}
			if sequencer.groupApplies["field"][0] != expected {
				t.Errorf("expected %v, got %v", expected, sequencer.groupApplies["field"])
			}
			if sequencer.groupApplies["field"][1] != expected {
				t.Errorf("expected %v, got %v", expected, sequencer.groupApplies["field"])
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
			sequencer = SequencerConfig{}
			err = sequencer.updateGroupApplies(config)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
			if len(sequencer.groupApplies) != 1 {
				t.Fatalf("expected 1, got %d", len(sequencer.groupApplies))
			}
			if _, ok := sequencer.groupApplies["field"]; !ok {
				t.Errorf("expected field, got %v", sequencer.groupApplies)
			}
			expected = GroupApply{
				FieldToShare:            "field",
				IdentifyingField:        "field",
				ValueOfIdentifyingField: "field",
			}
			if sequencer.groupApplies["field"][0] != expected {
				t.Errorf("expected %v, got %v", expected, sequencer.groupApplies["field"])
			}
			// test default case when groupApplies is not present
			config = map[string]any{}
			sequencer = SequencerConfig{}
			err = sequencer.updateGroupApplies(config)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
			if len(sequencer.groupApplies) != 0 {
				t.Fatalf("expected 0, got %d", len(sequencer.groupApplies))
			}
		})
		t.Run("updateChildrenByBackwardsLink", func(t *testing.T) {
			// test error case when childrenByBackwardsLink is not a map
			config := map[string]any{
				"childrenByBackwardsLink": "not map",
			}
			sequencer := SequencerConfig{}
			err := sequencer.updateChildrenByBackwardsLink(config)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if err.Error() != "childrenByBackwardsLink is not a map" {
				t.Errorf("expected childrenByBackwardsLink is not a map, got %v", err)
			}
			// test error when the map contains the field "all" but it is not a boolean
			config = map[string]any{
				"childrenByBackwardsLink": map[string]any{
					"all": "not boolean",
				},
			}
			sequencer = SequencerConfig{}
			err = sequencer.updateChildrenByBackwardsLink(config)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if err.Error() != "field \"all\" in childrenByBackwardsLink is not a boolean" {
				t.Errorf("expected field \"all\" in childrenByBackwardsLink is not a boolean, got %v", err)
			}
			// test error when nodeTypes field exists but it is not an array of strings
			config = map[string]any{
				"childrenByBackwardsLink": map[string]any{
					"nodeTypes": "not array",
				},
			}
			sequencer = SequencerConfig{}
			err = sequencer.updateChildrenByBackwardsLink(config)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if err.Error() != "nodeTypes in childrenByBackwardsLink is not an array of strings" {
				t.Errorf("expected nodeTypes in childrenByBackwardsLink is not an array of strings, got %v", err)
			}
			// test default case when childrenByBackwardsLink is not present
			config = map[string]any{}
			sequencer = SequencerConfig{}
			err = sequencer.updateChildrenByBackwardsLink(config)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
			if sequencer.ChildrenByBackwardsLink.All {
				t.Fatalf("expected false, got %v", sequencer.ChildrenByBackwardsLink.All)
			}
			if len(sequencer.ChildrenByBackwardsLink.NodeTypes) != 0 {
				t.Fatalf("expected 0, got %d", len(sequencer.ChildrenByBackwardsLink.NodeTypes))
			}
			// test defaults when childrenByBackwardsLink is present
			config = map[string]any{
				"childrenByBackwardsLink": map[string]any{},
			}
			sequencer = SequencerConfig{}
			err = sequencer.updateChildrenByBackwardsLink(config)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
			if sequencer.ChildrenByBackwardsLink.All {
				t.Fatalf("expected false, got %v", sequencer.ChildrenByBackwardsLink.All)
			}
			if len(sequencer.ChildrenByBackwardsLink.NodeTypes) != 0 {
				t.Fatalf("expected 0, got %d", len(sequencer.ChildrenByBackwardsLink.NodeTypes))
			}
			// test case when seeting "all" and "nodeTypes"
			config = map[string]any{
				"childrenByBackwardsLink": map[string]any{
					"all":       true,
					"nodeTypes": []any{"type1", "type2"},
				},
			}
			sequencer = SequencerConfig{}
			err = sequencer.updateChildrenByBackwardsLink(config)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
			if !sequencer.ChildrenByBackwardsLink.All {
				t.Fatalf("expected true, got %v", sequencer.ChildrenByBackwardsLink.All)
			}
			if !reflect.DeepEqual(sequencer.ChildrenByBackwardsLink.NodeTypes, map[string]bool{
				"type1": true,
				"type2": true,
			}) {
				t.Fatalf("expected map[string]bool{\"type1\": true, \"type2\": true}, got %v", sequencer.ChildrenByBackwardsLink.NodeTypes)
			}
		})
	})
}

// Tests for IncomingData
func TestIncomingData(t *testing.T) {
	t.Run("UnmarshalJSON", func(t *testing.T) {
		tests := []struct {
			name       string
			data       []byte
			wantErr    bool
			errMessage string
			expected   *IncomingData
		}{
			{
				name:       "validAllFieldsPresent",
				data:       []byte(`{"nodeId":"1","parentId":"3","childIds":["2"], "nodeType":"type", "timestamp":1, "appJSON":{}}`),
				wantErr:    false,
				errMessage: "",
				expected: &IncomingData{
					NodeId:    "1",
					ParentId:  "3",
					ChildIds:  []string{"2"},
					NodeType:  "type",
					Timestamp: 1,
					AppJSON:   map[string]any{},
				},
			},
			{
				name:       "validAllOptionalFieldsNotPresent",
				data:       []byte(`{"nodeId":"1","appJSON":{}}`),
				wantErr:    false,
				errMessage: "",
				expected: &IncomingData{
					NodeId:    "1",
					ParentId:  "",
					ChildIds:  []string{},
					NodeType:  "",
					Timestamp: 0,
					AppJSON:   map[string]any{},
				},
			},
			{
				name:       "validAllOptionalFieldsNull",
				data:       []byte(`{"nodeId":"1","parentId":null,"childIds":null,"nodeType":null,"timestamp":null,"appJSON":{}}`),
				wantErr:    false,
				errMessage: "",
				expected: &IncomingData{
					NodeId:    "1",
					ParentId:  "",
					ChildIds:  []string{},
					NodeType:  "",
					Timestamp: 0,
					AppJSON:   map[string]any{},
				},
			},
			{
				name:       "validTreeIdPresent",
				data:       []byte(`{"nodeId":"1", "appJSON":{},"treeId":"tree"}`),
				wantErr:    false,
				errMessage: "",
				expected: &IncomingData{
					NodeId:    "1",
					ParentId:  "",
					ChildIds:  []string{},
					NodeType:  "",
					Timestamp: 0,
					AppJSON:   map[string]any{},
				},
			},
			{
				name:       "invalidNodeIdNotString",
				data:       []byte(`{"nodeId":1,"appJSON":{}}`),
				wantErr:    true,
				errMessage: "json: cannot unmarshal number into Go struct field IncomingData.nodeId of type string",
				expected:   nil,
			},
			{
				name:       "invalidParentIdNotString",
				data:       []byte(`{"nodeId":"1","parentId":1,"appJSON":{}}`),
				wantErr:    true,
				errMessage: "json: cannot unmarshal number into Go struct field IncomingData.parentId of type string",
				expected:   nil,
			},
			{
				name:       "invalidChildIdsNotArray",
				data:       []byte(`{"nodeId":"1","childIds":1,"appJSON":{}}`),
				wantErr:    true,
				errMessage: "json: cannot unmarshal number into Go struct field IncomingData.childIds of type []string",
				expected:   nil,
			},
			{
				name:       "invalidChildIdsNotString",
				data:       []byte(`{"nodeId":"1","childIds":[1],"appJSON":{}}`),
				wantErr:    true,
				errMessage: "json: cannot unmarshal number into Go struct field IncomingData.childIds of type string",
				expected:   nil,
			},
			{
				name:       "invalidNodeTypeNotString",
				data:       []byte(`{"nodeId":"1","nodeType":1,"appJSON":{}}`),
				wantErr:    true,
				errMessage: "json: cannot unmarshal number into Go struct field IncomingData.nodeType of type string",
				expected:   nil,
			},
			{
				name:       "invalidTimestampNotNumber",
				data:       []byte(`{"nodeId":"1","timestamp":"1","appJSON":{}}`),
				wantErr:    true,
				errMessage: "json: cannot unmarshal string into Go struct field IncomingData.timestamp of type int",
				expected:   nil,
			},
			{
				name:       "invalidAppJSONNotMap",
				data:       []byte(`{"nodeId":"1","appJSON":1}`),
				wantErr:    true,
				errMessage: "json: cannot unmarshal number into Go struct field IncomingData.appJSON of type map[string]interface {}",
				expected:   nil,
			},
			{
				name:       "invalidUnknownField",
				data:       []byte(`{"nodeId":"1","unknown":1,"appJSON":{}}`),
				wantErr:    true,
				errMessage: "json: unknown field \"unknown\"",
				expected:   nil,
			},
			{
				name:       "invalidNodeIdNotPresent",
				data:       []byte(`{"appJSON":{}}`),
				wantErr:    true,
				errMessage: "nodeId is required",
				expected:   nil,
			},
			{
				name:       "invalidAppJSONNotPresent",
				data:       []byte(`{"nodeId":"1"}`),
				wantErr:    true,
				errMessage: "appJSON is required",
				expected:   nil,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				id := &IncomingData{}
				err := json.Unmarshal(tt.data, id)
				if err == nil {
					if tt.wantErr {
						t.Fatalf("IncomingData.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
					}
					if tt.expected.NodeId != id.NodeId {
						t.Fatalf("IncomingData.UnmarshalJSON() = %v, want %v", id, tt.expected)
					}
					if tt.expected.ParentId != id.ParentId {
						t.Fatalf("IncomingData.UnmarshalJSON() = %v, want %v", id, tt.expected)
					}
					if tt.expected.NodeType != id.NodeType {
						t.Fatalf("IncomingData.UnmarshalJSON() = %v, want %v", id, tt.expected)
					}
					if tt.expected.Timestamp != id.Timestamp {
						t.Fatalf("IncomingData.UnmarshalJSON() = %v, want %v", id, tt.expected)
					}
					if !reflect.DeepEqual(tt.expected.AppJSON, id.AppJSON) {
						t.Fatalf("IncomingData.UnmarshalJSON() = %v, want %v", id, tt.expected)
					}
					if len(tt.expected.ChildIds) == 0 {
						if len(id.ChildIds) != 0 {
							t.Fatalf("IncomingData.UnmarshalJSON() = %v, want %v", id, tt.expected)
						}
					} else {
						if !reflect.DeepEqual(tt.expected.ChildIds, id.ChildIds) {
							t.Fatalf("IncomingData.UnmarshalJSON() = %v, want %v", id, tt.expected)
						}
					}
				} else {
					if !tt.wantErr {
						t.Fatalf("IncomingData.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
					}
					if err.Error() != tt.errMessage {
						t.Fatalf("IncomingData.UnmarshalJSON() error = %v, wantErr %v", err, tt.errMessage)
					}
				}
			})
		}
	})
}

// Test for incomingDataEquality
func TestIncomingDataEquality(t *testing.T) {
	tests := []struct {
		name     string
		incoming *IncomingData
		other    *IncomingData
		want     bool
	}{
		{
			name: "equal",
			incoming: &IncomingData{
				NodeId:    "1",
				ParentId:  "2",
				ChildIds:  []string{"3"},
				NodeType:  "type",
				Timestamp: 1,
				AppJSON: map[string]any{
					"key": []any{},
				},
			},
			other: &IncomingData{
				NodeId:    "1",
				ParentId:  "2",
				ChildIds:  []string{"3"},
				NodeType:  "type",
				Timestamp: 1,
				AppJSON: map[string]any{
					"key": []any{},
				},
			},
			want: true,
		},
		{
			name: "notEqual",
			incoming: &IncomingData{
				NodeId:    "1",
				ParentId:  "2",
				ChildIds:  []string{"3"},
				NodeType:  "type",
				Timestamp: 1,
				AppJSON:   map[string]any{},
			},
			other: &IncomingData{
				NodeId:    "2",
				ParentId:  "2",
				ChildIds:  []string{"3"},
				NodeType:  "type",
				Timestamp: 1,
				AppJSON:   map[string]any{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := incomingDataEquality(tt.incoming, tt.other); got != tt.want {
				t.Errorf("incomingDataEquality() = %v, want %v", got, tt.want)
			}
		})
	}
}

// MockPushable is a mock implementation of the Pushable interface
type MockPushable struct {
	isSendToError bool
	incomingData  chan (*Server.AppData)
}

func (p *MockPushable) SendTo(data *Server.AppData) error {
	if p.isSendToError {
		return errors.New("test error")
	}
	p.incomingData <- data
	return nil
}

// NotSequencerConfig is a struct that is not SequencerConfig struct
type NotSequencerConfig struct{}

func (s *NotSequencerConfig) IngestConfig(config map[string]any) error {
	return nil
}

// TestSequencer tests the Sequencer struct and its methods
func TestSequencer(t *testing.T) {
	t.Run("ImplmentsPipeServer", func(t *testing.T) {
		s := &Sequencer{}
		if _, ok := interface{}(s).(Server.PipeServer); !ok {
			t.Errorf("Sequencer does not implement PipeServer")
		}
	})
	t.Run("AddPushable", func(t *testing.T) {
		pushable := &MockPushable{}
		sequencer := &Sequencer{}
		err := sequencer.AddPushable(pushable)
		if err != nil {
			t.Errorf("Expected no error from AddPushable, got %v", err)
		}
		if sequencer.pushable != pushable {
			t.Errorf("Expected pushable to be %v, got %v", pushable, sequencer.pushable)
		}
		// Test when pushable is set
		err = sequencer.AddPushable(pushable)
		if err == nil {
			t.Errorf("Expected error from AddPushable, got nil")
		}
	})
	t.Run("Serve", func(t *testing.T) {
		// Test error case when pushable is not set
		sequencer := &Sequencer{}
		err := sequencer.Serve()
		if err == nil {
			t.Errorf("Expected error from Serve, got nil")
		}
		if err.Error() != "pushable not set" {
			t.Errorf("Expected error message 'pushable not set', got %v", err.Error())
		}
		// Test error case when config is not set
		sequencer.pushable = &MockPushable{}
		err = sequencer.Serve()
		if err == nil {
			t.Errorf("Expected error from Serve, got nil")
		}
		if err.Error() != "config not set" {
			t.Errorf("Expected error message 'config not set', got %v", err.Error())
		}
		// Test valid case
		sequencer.config = &SequencerConfig{}
		err = sequencer.Serve()
		if err != nil {
			t.Errorf("Expected no error from Serve, got %v", err)
		}
	})
	t.Run("Setup", func(t *testing.T) {
		// Check case where config is not SequencerConfig
		sequencer := &Sequencer{}
		err := sequencer.Setup(&NotSequencerConfig{})
		if err == nil {
			t.Fatalf("Expected error from Setup, got nil")
		}
		if err.Error() != "config must be SequencerConfig type" {
			t.Errorf("Expected error message 'config must be SequencerConfig type', got %v", err.Error())
		}
		// check case where config is SequencerConfig
		sequencerConfig := &SequencerConfig{}
		err = sequencer.Setup(sequencerConfig)
		if err != nil {
			t.Errorf("Expected no error from Setup, got %v", err)
		}
		if sequencer.config != sequencerConfig {
			t.Errorf("Expected config to be %v, got %v", sequencerConfig, sequencer.config)
		}
	})
	t.Run("SendTo", func(t *testing.T) {
		t.Run("errorCases", func(t *testing.T) {
			// check error case when pushable is not set
			sequencer := &Sequencer{}
			err := sequencer.SendTo(&Server.AppData{})
			if err == nil {
				t.Fatalf("Expected error from SendTo, got nil")
			}
			if err.Error() != "pushable not set" {
				t.Errorf("Expected error message 'pushable not set', got %v", err.Error())
			}
			// check error case when config is not set
			sequencer.pushable = &MockPushable{}
			err = sequencer.SendTo(&Server.AppData{})
			if err == nil {
				t.Fatalf("Expected error from SendTo, got nil")
			}
			if err.Error() != "config not set" {
				t.Errorf("Expected error message 'config not set', got %v", err.Error())
			}
			// check error case when appData is nil
			sequencer.config = &SequencerConfig{}
			err = sequencer.SendTo(nil)
			if err == nil {
				t.Fatalf("Expected error from SendTo, got nil")
			}
			if err.Error() != "data not sent" {
				t.Errorf("Expected error message 'data not sent', got %v", err.Error())
			}
			// check error in convertToIncomingDataMapAndRootNodes
			sequencer.config = &SequencerConfig{}
			err = sequencer.SendTo(Server.NewAppData([]byte(`[1]`), ""))
			if err == nil {
				t.Fatalf("Expected error from SendTo, got nil")
			}
			if err.Error() != "json: cannot unmarshal number into Go value of type sequencer.IncomingData" {
				t.Errorf("Expected error message 'json: cannot unmarshal number into Go value of type sequencer.IncomingData', got %v", err.Error())
			}
			// check error case where there are no root nodes
			sequencer.config = &SequencerConfig{}
			err = sequencer.SendTo(Server.NewAppData(
				[]byte(`[]`), "",
			))
			if err == nil {
				t.Fatalf("Expected error from SendTo, got nil")
			}
			if err.Error() != "no data present in incoming data array" {
				t.Errorf("Expected error message 'no data present in incoming data array', got %v", err.Error())
			}
			// check error case where getPrevIdData returns an error
			sequencer.config = &SequencerConfig{}
			appData := Server.NewAppData(
				[]byte(`[{"nodeId":"1","childIds":["2"],"appJSON":{}},{"nodeId":"2","childIds":[],"appJSON":{}}]`), "",
			)
			err = sequencer.SendTo(appData)
			if err == nil {
				t.Fatalf("Expected error from SendTo, got nil")
			}
			if err.Error() != "unknown outputAppFieldSequenceType" {
				t.Errorf("Expected error message 'unknown outputAppFieldSequenceType', got %v", err.Error())
			}
			// check error case when pushable.SendTo returns an error
			sequencer.config = &SequencerConfig{
				outputAppFieldSequenceType: String,
			}
			sequencer.pushable = &MockPushable{isSendToError: true}
			err = sequencer.SendTo(appData)
			if err == nil {
				t.Fatalf("Expected error from SendTo, got nil")
			}
			if err.Error() != "test error" {
				t.Errorf("Expected error message 'test error', got %v", err.Error())
			}
		})
		t.Run("valid", func(t *testing.T) {
			// check valid case with one root node
			t.Run("oneRootNode", func(t *testing.T) {
				tt := []struct {
					name     string
					input    []byte
					expected []byte
				}{
					{
						name:     "NoDuplicates",
						input:    []byte(`[{"nodeId":"1","childIds":["2"],"appJSON":{}},{"nodeId":"2","childIds":[],"appJSON":{}}]`),
						expected: []byte(`[{},{"SeqField":"2"}]`),
					},
					{
						name:     "EqualDuplicates",
						input:    []byte(`[{"nodeId":"1","childIds":["2"],"appJSON":{}},{"nodeId":"2","childIds":[],"appJSON":{}},{"nodeId":"1","childIds":["2"],"appJSON":{}}]`),
						expected: []byte(`[{},{"SeqField":"2"},{"SeqField":"2"}]`),
					},
					{
						name:     "UnequalDuplicates",
						input:    []byte(`[{"nodeId":"1","childIds":["2"],"appJSON":{}},{"nodeId":"2","childIds":[],"appJSON":{}},{"nodeId":"1","childIds":["3"],"appJSON":{}}]`),
						expected: []byte(`[{},{},{}]`),
					},
					{
						name:     "selfReference",
						input:    []byte(`[{"nodeId":"1","childIds":["1", "2"],"appJSON":{}},{"nodeId":"2","childIds":[],"appJSON":{}}]`),
						expected: []byte(`[{},{}]`),
					},
				}
				for _, tc := range tt {
					t.Run(tc.name, func(t *testing.T) {
						pushableChan := make(chan (*Server.AppData), 1)
						sequencer := &Sequencer{
							config: &SequencerConfig{
								outputAppFieldSequenceType: String,
								outputAppSequenceField:     "SeqField",
							},
							pushable: &MockPushable{
								incomingData: pushableChan,
							},
						}
						appData := Server.NewAppData(
							tc.input, "",
						)
						err := sequencer.SendTo(appData)
						if err != nil {
							t.Errorf("Expected no error from SendTo, got %v", err)
						}
						close(pushableChan)
						receivedAppData := <-pushableChan
						if receivedAppData == nil {
							t.Fatalf("Expected appData to be sent")
						}
						gotData, err := receivedAppData.GetData()
						if err != nil {
							t.Fatalf("Expected no error from GetData, got %v", err)
						}
						if !reflect.DeepEqual(gotData, tc.expected) {
							t.Errorf("Expected appData to be %v, got %v", tc.expected, gotData)
						}
					})
				}
			})

			t.Run("multipleRootNodes", func(t *testing.T) {
				// check valid case with multiple root nodes
				pushableChan := make(chan (*Server.AppData), 1)
				sequencer := &Sequencer{
					config: &SequencerConfig{
						outputAppFieldSequenceType: String,
						outputAppSequenceField:     "SeqField",
						groupApplies: map[string][]GroupApply{
							"share": {
								{
									FieldToShare:            "share",
									IdentifyingField:        "field",
									ValueOfIdentifyingField: "2",
								},
							},
						},
					},
					pushable: &MockPushable{
						incomingData: pushableChan,
					},
				}
				jsonBytes, err := json.Marshal(
					[]any{
						map[string]any{
							"nodeId":   "1",
							"childIds": []any{"2"},
							"appJSON":  map[string]any{},
						},
						map[string]any{
							"nodeId":   "2",
							"childIds": []any{},
							"appJSON":  map[string]any{"field": "2", "share": "was shared"},
						},
						map[string]any{
							"nodeId":   "3",
							"childIds": []any{"4"},
							"appJSON":  map[string]any{},
						},
						map[string]any{
							"nodeId":   "4",
							"childIds": []any{},
							"appJSON":  map[string]any{"field": "4"},
						},
					},
				)
				if err != nil {
					t.Fatalf("Expected no error from Marshal, got %v", err)
				}
				appData := Server.NewAppData(
					jsonBytes,
					"",
				)
				err = sequencer.SendTo(appData)
				if err != nil {
					t.Errorf("Expected no error from SendTo, got %v", err)
				}
				close(pushableChan)
				receivedAppData := <-pushableChan
				if receivedAppData == nil {
					t.Fatalf("Expected appData to be sent")
				}
				gotData, err := receivedAppData.GetData()
				if err != nil {
					t.Fatalf("Expected no error from GetData, got %v", err)
				}
				var gotDataArray []map[string]any
				err = json.Unmarshal(gotData, &gotDataArray)
				if err != nil {
					t.Fatalf("Expected no error from Unmarshal, got %v", err)
				}
				expectedSeqFields := map[string]bool{
					"2": true,
					"4": true,
				}
				for i := 0; i < len(gotDataArray); i += 2 {
					if gotDataArray[i+1]["SeqField"] != gotDataArray[i]["field"] {
						t.Errorf("Expected SeqField to be %v, got %v", gotDataArray[i]["field"], gotDataArray[i+1]["SeqField"])
					}
					seqFieldValue, ok := gotDataArray[i+1]["SeqField"].(string)
					if !ok {
						t.Fatalf("Expected SeqField to be a string")
					}
					if !expectedSeqFields[seqFieldValue] {
						t.Errorf("Expected SeqField to be in %v", expectedSeqFields)
					}
					delete(expectedSeqFields, seqFieldValue)
				}
				if len(expectedSeqFields) != 0 {
					t.Errorf("Expected all SeqFields to be used")
				}
				// make sure the shared field is shared
				for _, record := range gotDataArray {
					shared, ok := record["share"].(string)
					if !ok {
						t.Fatalf("Expected share to exist and be a string")
					}
					if shared != "was shared" {
						t.Errorf("Expected share to be 'was shared', got %v", shared)
					}
				}
			})
			t.Run("missingChild", func(t *testing.T) {
				// check valid case with missing child
				pushableChan := make(chan (*Server.AppData), 1)
				sequencer := &Sequencer{
					config: &SequencerConfig{
						outputAppFieldSequenceType: String,
						outputAppSequenceField:     "SeqField",
					},
					pushable: &MockPushable{
						incomingData: pushableChan,
					},
				}
				appData := Server.NewAppData(
					[]byte(`[{"nodeId":"1","childIds":["2", "3", "4"],"appJSON":{}},{"nodeId":"2","appJSON":{}},{"nodeId":"4","appJSON":{}}]`), "",
				)
				err := sequencer.SendTo(appData)
				if err != nil {
					t.Errorf("Expected no error from SendTo, got %v", err)
				}
				close(pushableChan)
				receivedAppData := <-pushableChan
				if receivedAppData == nil {
					t.Fatalf("Expected appData to be sent")
				}
				gotData, err := receivedAppData.GetData()
				if err != nil {
					t.Fatalf("Expected no error from GetData, got %v", err)
				}
				expectedData := []byte(`[{},{},{"SeqField":"4"}]`)
				if !reflect.DeepEqual(gotData, expectedData) {
					t.Errorf("Expected appData to be %v, got %v", expectedData, gotData)
				}
			})
		})
	})
}

// TestStackIncomingData tests the stackIncomingData struct
func TestStackIncomingData(t *testing.T) {
	t.Run("nextChildId", func(t *testing.T) {
		sid := &stackIncomingData{
			incomingDataWithDuplicates: &incomingDataWithDuplicates{},
		}
		// check error case when currentChildIdIndex is greater than or equal to len(IncomingData)
		sid.IncomingData = &IncomingData{
			ChildIds: []string{},
		}
		sid.currentChildIdIndex = 0
		_, err := sid.nextChildId()
		if err == nil {
			t.Fatalf("Expected error from nextChildId, got nil")
		}
		if err.Error() != "no more child ids" {
			t.Errorf("Expected error message 'no more child ids', got %v", err.Error())
		}
		// check valid case
		sid.IncomingData = &IncomingData{
			ChildIds: []string{"1", "2"},
		}
		childId, err := sid.nextChildId()
		if err != nil {
			t.Errorf("Expected no error from nextChildId, got %v", err)
		}
		if childId != "1" {
			t.Errorf("Expected childId to be '1', got %v", childId)
		}
		if sid.currentChildIdIndex != 1 {
			t.Errorf("Expected currentChildIdIndex to be 1, got %v", sid.currentChildIdIndex)
		}
		// check valid case when called again
		childId, err = sid.nextChildId()
		if err != nil {
			t.Errorf("Expected no error from nextChildId, got %v", err)
		}
		if childId != "2" {
			t.Errorf("Expected childId to be '2', got %v", childId)
		}
		if sid.currentChildIdIndex != 2 {
			t.Errorf("Expected currentChildIdIndex to be 2, got %v", sid.currentChildIdIndex)
		}
	})
}

// TestSequenceWithStack tests the sequenceWithStack function
func TestSequenceWithStack(t *testing.T) {
	t.Run("rootNodeNotSet", func(t *testing.T) {
		// check error case when rootNode is nil
		counter := 0
		sequence := sequenceWithStack(nil, nil)
		for node, err := range sequence {
			if node != nil {
				t.Errorf("Expected node to be nil, got %v", node)
			}
			if err == nil {
				t.Fatalf("Expected error from sequence, got nil")
			}
			if err.Error() != "root node not set" {
				t.Errorf("Expected error message 'root node not set', got %v", err.Error())
			}
			counter++
		}
		if counter != 1 {
			t.Fatalf("Expected counter to be 1, got %v", counter)
		}
	})
	t.Run("nodeIdToIncomingDataMapNotSet", func(t *testing.T) {
		// check error case when nodeIdToIncomingDataMap is nil
		counter := 0
		rootNode := &incomingDataWithDuplicates{}
		sequence := sequenceWithStack(rootNode, nil)
		for node, err := range sequence {
			if node != nil {
				t.Errorf("Expected node to be nil, got %v", node)
			}
			if err == nil {
				t.Fatalf("Expected error from sequence, got nil")
			}
			if err.Error() != "nodeIdToIncomingDataMap not set" {
				t.Errorf("Expected error message 'nodeIdToIncomingDataMap not set', got %v", err.Error())
			}
			counter++
		}
		if counter != 1 {
			t.Fatalf("Expected counter to be 1, got %v", counter)
		}
	})
	t.Run("childNodeNotFound", func(t *testing.T) {
		// check error case when child node is not found
		nodeIdToIncomingDataMap := map[string]*incomingDataWithDuplicates{}
		rootNode := &incomingDataWithDuplicates{
			IncomingData: &IncomingData{
				ChildIds: []string{"1"},
			},
		}
		sequence := sequenceWithStack(rootNode, nodeIdToIncomingDataMap)
		counter := 0
		for node, err := range sequence {
			if err != nil {
				t.Errorf("Expected no error from sequence, got %v", err)
			}
			if node == nil {
				t.Errorf("Expected node to be non-nil")
			}
			if counter == 0 {
				if !node.IsDummy {
					t.Errorf("Expected node to be a dummy node")
				}
				if node.NodeId != "1" {
					t.Errorf("Expected node.NodeId to be '1', got %v", node.NodeId)
				}
			} else {
				if node.incomingDataWithDuplicates != rootNode {
					t.Errorf("Expected node.incomingDataWithDuplicates to be rootNode")
				}
			}
			counter++
		}
		if counter != 2 {
			t.Fatalf("Expected counter to be 1, got %v", counter)
		}
	})
	t.Run("childNodeIsNil", func(t *testing.T) {
		// check case when child node is nil
		nodeIdToIncomingDataMap := map[string]*incomingDataWithDuplicates{
			"1": nil,
		}
		rootNode := &incomingDataWithDuplicates{
			IncomingData: &IncomingData{
				ChildIds: []string{"1"},
			},
		}
		sequence := sequenceWithStack(rootNode, nodeIdToIncomingDataMap)
		counter := 0
		for node, err := range sequence {
			if err != nil {
				t.Errorf("Expected no error from sequence, got %v", err)
			}
			if node == nil {
				t.Errorf("Expected node to be non-nil")
			}
			if counter == 0 {
				if !node.IsDummy {
					t.Errorf("Expected node to be a dummy node")
				}
				if node.NodeId != "1" {
					t.Errorf("Expected node.NodeId to be '1', got %v", node.NodeId)
				}
			} else {
				if node.incomingDataWithDuplicates != rootNode {
					t.Errorf("Expected node.incomingDataWithDuplicates to be rootNode")
				}
			}
			counter++
		}
		if counter != 2 {
			t.Fatalf("Expected counter to be 1, got %v", counter)
		}
	})
	t.Run("valid", func(t *testing.T) {
		// check valid case
		nodeIdToIncomingDataMap := map[string]*incomingDataWithDuplicates{
			"1": {
				IncomingData: &IncomingData{
					NodeId:   "1",
					ChildIds: []string{"2", "3"},
				},
			},
			"2": {
				IncomingData: &IncomingData{
					NodeId:   "2",
					ChildIds: []string{"4", "5"},
				},
			},
			"3": {
				IncomingData: &IncomingData{
					NodeId:   "3",
					ChildIds: []string{"6", "7"},
				},
			},
			"4": {IncomingData: &IncomingData{NodeId: "4"}},
			"5": {IncomingData: &IncomingData{NodeId: "5"}},
			"6": {IncomingData: &IncomingData{NodeId: "6"}},
			"7": {IncomingData: &IncomingData{NodeId: "7"}},
		}
		rootNode := nodeIdToIncomingDataMap["1"]
		sequence := sequenceWithStack(rootNode, nodeIdToIncomingDataMap)
		counter := 0
		expectedSequence := []string{"4", "5", "2", "6", "7", "3", "1"}
		for member, err := range sequence {
			node := member.incomingDataWithDuplicates
			if node != nodeIdToIncomingDataMap[expectedSequence[counter]] {
				t.Errorf("Expected node to be %v, got %v", nodeIdToIncomingDataMap[expectedSequence[counter]], node)
			}
			if err != nil {
				t.Errorf("Expected no error from sequence, got %v", err)
			}
			counter++
		}
		if counter != len(expectedSequence) {
			t.Fatalf("Expected counter to be %v, got %v", len(expectedSequence), counter)
		}
	})
}

// Tests convertToIncomingDataMapAndRootNodes
func TestConvertToIncomingDataMapAndRootNodes(t *testing.T) {
	t.Run("invalidData", func(t *testing.T) {
		// check error case where unmarshalling returns an error
		_, _, _, err := convertToIncomingDataMapAndRootNodes([]json.RawMessage{json.RawMessage(`1`)}, &childrenByBackwardsLink{})
		if err == nil {
			t.Fatalf("Expected error from convertAppDataToIncomingDataMapAndRootNodes, got nil")
		}
		if err.Error() != "json: cannot unmarshal number into Go value of type sequencer.IncomingData" {
			t.Errorf("Expected error message 'json: cannot unmarshal number into Go value of type sequencer.IncomingData', got %v", err.Error())
		}
	})
	t.Run("valid", func(t *testing.T) {
		tt := []struct {
			name             string
			isUnSequenceable bool
		}{
			{
				name:             "noDuplicates",
				isUnSequenceable: false,
			},
			{
				name:             "equalDuplicates",
				isUnSequenceable: false,
			},
			{
				name:             "unequalDuplicates",
				isUnSequenceable: true,
			},
			{
				name:             "selfReferenceParentId",
				isUnSequenceable: true,
			},
			{
				name:             "selfReferenceChildId",
				isUnSequenceable: true,
			},
		}
		for _, test := range tt {
			// check valid case
			nodes := []any{
				map[string]any{
					"nodeId":   "1",
					"childIds": []any{"2", "3"},
					"appJSON":  map[string]any{},
				},
				map[string]any{
					"nodeId":   "2",
					"childIds": []any{},
					"appJSON":  map[string]any{},
				},
				map[string]any{
					"nodeId":   "3",
					"childIds": []any{},
					"appJSON":  map[string]any{},
				},
			}
			switch test.name {
			case "equalDuplicates":
				nodes = append(nodes, map[string]any{
					"nodeId":   "1",
					"childIds": []any{"2", "3"},
					"appJSON":  map[string]any{},
				})
			case "unequalDuplicates":
				nodes = append(nodes, map[string]any{
					"nodeId":   "1",
					"childIds": []any{"2", "3", "4"},
					"appJSON":  map[string]any{},
				})
			case "selfReferenceParentId":
				nodes[0].(map[string]any)["parentId"] = "1"
			case "selfReferenceChildId":
				nodes[0].(map[string]any)["childIds"] = []any{"1", "2", "3"}
			}
			jsonBytes, err := json.Marshal(nodes)
			if err != nil {
				t.Fatalf("Expected no error from Marshal, got %v", err)
			}
			var rawDataArray []json.RawMessage
			err = json.Unmarshal(jsonBytes, &rawDataArray)
			if err != nil {
				t.Fatalf("Expected no error from Unmarshal, got %v", err)
			}
			incomingDataMap, rootNodes, isUnSequenceable, err := convertToIncomingDataMapAndRootNodes(rawDataArray, &childrenByBackwardsLink{})
			if err != nil {
				t.Errorf("Expected no error from convertAppDataToIncomingDataMapAndRootNodes, got %v", err)
			}
			if len(incomingDataMap) != 3 {
				t.Errorf("Expected incomingDataMap to have 3 elements, got %v", len(incomingDataMap))
			}
			if test.name == "selfReferenceChildId" {
				if len(rootNodes) != 0 {
					t.Errorf("Expected rootNodes to have 0 elements, got %v", len(rootNodes))
				}
			} else {
				if len(rootNodes) != 1 {
					t.Errorf("Expected rootNodes to have 1 element, got %v", len(rootNodes))
				}
			}
			if isUnSequenceable != test.isUnSequenceable {
				t.Errorf("Expected isUnSequenceable to be %v, got %v", test.isUnSequenceable, isUnSequenceable)
			}
			for i, inputMapAny := range nodes {
				if i == 3 && test.name == "unequalDuplicates" {
					continue
				}
				inputMap := inputMapAny.(map[string]any)
				incomingData, ok := incomingDataMap[inputMap["nodeId"].(string)]
				if !ok {
					t.Errorf("Expected incomingDataMap to have key %v", inputMap["nodeId"])
				}
				if incomingData.NodeId != inputMap["nodeId"] {
					t.Errorf("Expected nodeId to be %v, got %v", inputMap["nodeId"], incomingData.NodeId)
				}
				expectedOrderedChildIds := []string{}
				for _, childId := range inputMap["childIds"].([]any) {
					expectedOrderedChildIds = append(expectedOrderedChildIds, childId.(string))
				}
				if !reflect.DeepEqual(incomingData.ChildIds, expectedOrderedChildIds) {
					t.Errorf("Expected childIds to be %v, got %v", inputMap["childIds"], incomingData.ChildIds)
				}
				if !reflect.DeepEqual(incomingData.AppJSON, inputMap["appJSON"]) {
					t.Errorf("Expected appJSON to be %v, got %v", inputMap["appJSON"], incomingData.AppJSON)
				}
			}
			if test.name != "selfReferenceChildId" {
				rootNode, ok := rootNodes["1"]
				if !ok {
					t.Errorf("Expected rootNodes to have key '1'")
				}
				if rootNode != incomingDataMap["1"] {
					t.Errorf("Expected rootNode to be %v, got %v", incomingDataMap["1"], rootNode)
				}
			}
		}
	})
	t.Run("childrenByBackwardsLink", func(t *testing.T) {
		tests := []struct {
			name                    string
			childrenByBackwardsLink *childrenByBackwardsLink
		}{
			{
				name: "allSetToTrue",
				childrenByBackwardsLink: &childrenByBackwardsLink{
					All: true,
				},
			},
			{
				name: "allSetToFalseNodeTypesGiven",
				childrenByBackwardsLink: &childrenByBackwardsLink{
					All: false,
					NodeTypes: map[string]bool{
						"1": true,
					},
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// valid case
				nodes := []any{
					map[string]any{
						"nodeId":    "1",
						"nodeType":  "1",
						"timestamp": 4,
						"appJSON":   map[string]any{},
					},
					map[string]any{
						"nodeId":    "2",
						"parentId":  "1",
						"timestamp": 2,
						"appJSON":   map[string]any{},
					},
					map[string]any{
						"nodeId":    "3",
						"parentId":  "1",
						"timestamp": 3,
						"appJSON":   map[string]any{},
					},
				}
				for _, testCase := range []string{"noChildren", "children"} {
					if testCase == "children" {
						nodes[0].(map[string]any)["childIds"] = []any{"2", "3"}
					}

					jsonBytes, err := json.Marshal(nodes)
					if err != nil {
						t.Fatalf("Expected no error from Marshal, got %v", err)
					}
					var rawDataArray []json.RawMessage
					err = json.Unmarshal(jsonBytes, &rawDataArray)
					if err != nil {
						t.Fatalf("Expected no error from Unmarshal, got %v", err)
					}
					incomingDataMap, rootNodes, hasUnequalDuplicates, err := convertToIncomingDataMapAndRootNodes(rawDataArray, tt.childrenByBackwardsLink)
					if err != nil {
						t.Fatalf("Expected no error from Marshal, got %v", err)
					}
					if err != nil {
						t.Errorf("Expected no error from convertAppDataToIncomingDataMapAndRootNodes, got %v", err)
					}
					if len(incomingDataMap) != 3 {
						t.Errorf("Expected incomingDataMap to have 3 elements, got %v", len(incomingDataMap))
					}
					if len(rootNodes) != 1 {
						t.Errorf("Expected rootNodes to have 1 element, got %v", len(rootNodes))
					}
					if hasUnequalDuplicates {
						t.Errorf("Expected hasUnequalDuplicates to be false")
					}
					if _, ok := rootNodes["1"]; !ok {
						t.Errorf("Expected rootNodes to have key '1'")
					}
					expectedRootNodeIncomingData := &IncomingData{
						NodeId:    "1",
						Timestamp: 4,
						NodeType:  "1",
						AppJSON:   map[string]any{},
						ChildIds:  []string{"2", "3"},
					}
					expectedRootNode := &incomingDataWithDuplicates{
						IncomingData: expectedRootNodeIncomingData,
						Duplicates:   []*IncomingData{},
					}
					if !reflect.DeepEqual(rootNodes["1"], expectedRootNode) {
						t.Errorf("Expected root node to be %v, got %v", expectedRootNode, rootNodes["1"])
					}
					if rootNode, ok := incomingDataMap["1"]; !ok {
						t.Errorf("Expected incomingDataMap to have key '1'")
					} else {
						if rootNode != rootNodes["1"] {
							t.Errorf("Expected rootNode to be rootNodes['1']")
						}
					}
					for i := 2; i < 4; i++ {
						nodeId := strconv.Itoa(i)
						if node, ok := incomingDataMap[nodeId]; !ok {
							t.Errorf("Expected incomingDataMap to have key '2'")
						} else {
							if node.NodeId != nodeId || node.Timestamp != i || len(node.AppJSON) != 0 || node.ParentId != "1" || len(node.ChildIds) != 0 {
								t.Errorf("Expected node to be %v, got %v", map[string]any{
									"nodeId":    nodeId,
									"timestamp": i,
									"appJSON":   map[string]any{},
									"parentId":  "1",
									"childIds":  []string{},
								}, node)
							}
						}
					}
				}
				// timestamp missing in one child node
				delete(nodes[1].(map[string]any), "timestamp")
				jsonBytes, err := json.Marshal(nodes)
				if err != nil {
					t.Fatalf("Expected no error from Marshal, got %v", err)
				}
				var rawDataArray2 []json.RawMessage
				err = json.Unmarshal(jsonBytes, &rawDataArray2)
				if err != nil {
					t.Fatalf("Expected no error from Unmarshal, got %v", err)
				}
				_, _, _, err = convertToIncomingDataMapAndRootNodes(rawDataArray2, tt.childrenByBackwardsLink)
				if err == nil {
					t.Fatalf("Expected error from convertAppDataToIncomingDataMapAndRootNodes, got nil")
				}
				if err.Error() != "child node 2 has no timestamp but attempting to order children by timestamp" {
					t.Errorf("Expected error message 'child node 2 has no timestamp but attempting to order children by timestamp', got %v", err.Error())
				}
			})
		}
	})
}

// Test getPrevIdFromPrevIncomingData
func TestGetPrevIdFromPrevIncomingData(t *testing.T) {
	t.Run("invalidData", func(t *testing.T) {
		// check error case outputAppFieldSequenceIdMap is not found
		// in prevIncomingData appJSON
		prevIncomingData := &IncomingData{
			AppJSON: map[string]any{},
		}
		_, err := getPrevIdFromPrevIncomingData(prevIncomingData, "test")
		if err == nil {
			t.Fatalf("Expected error from getPrevIdFromPrevIncomingData, got nil")
		}
		if err.Error() != "outputAppFieldSequenceIdMap must be a string and must exist in the input JSON" {
			t.Errorf("Expected error message 'outputAppFieldSequenceIdMap must be a string and must exist in the input JSON', got %v", err.Error())
		}
		// check error case outputAppFieldSequenceIdMap is not a string
		prevIncomingData = &IncomingData{
			AppJSON: map[string]any{
				"test": 1,
			},
		}
		_, err = getPrevIdFromPrevIncomingData(prevIncomingData, "test")
		if err == nil {
			t.Fatalf("Expected error from getPrevIdFromPrevIncomingData, got nil")
		}
		if err.Error() != "outputAppFieldSequenceIdMap must be a string and must exist in the input JSON" {
			t.Errorf("Expected error message 'outputAppFieldSequenceIdMap must be a string and must exist in the input JSON', got %v", err.Error())
		}
	})
	t.Run("valid", func(t *testing.T) {
		// check case where outputAppFieldSequenceIdMap is not set
		prevIncomingData := &IncomingData{
			NodeId: "1",
			AppJSON: map[string]any{
				"key": "value",
			},
		}
		prevId, err := getPrevIdFromPrevIncomingData(prevIncomingData, "")
		if err != nil {
			t.Errorf("Expected no error from getPrevIdFromPrevIncomingData, got %v", err)
		}
		if prevId != "1" {
			t.Errorf("Expected prevId to be '1', got %v", prevId)
		}
		// check case where outputAppFieldSequenceIdMap is set
		prevId, err = getPrevIdFromPrevIncomingData(prevIncomingData, "key")
		if err != nil {
			t.Errorf("Expected no error from getPrevIdFromPrevIncomingData, got %v", err)
		}
		if prevId != "value" {
			t.Errorf("Expected prevId to be 'value', got %v", prevId)
		}
	})
}

// Test getPrevIdData
func TestGetPrevIdData(t *testing.T) {
	t.Run("errorCases", func(t *testing.T) {
		// check when there is an error in getPrevIdFromPrevIncomingData
		prevIncomingData := &IncomingData{}
		config := &SequencerConfig{outputAppFieldSequenceIdMap: "test"}
		_, err := getPrevIdData(prevIncomingData, config)
		if err == nil {
			t.Fatalf("Expected error from getPrevIdData, got nil")
		}
		if err.Error() != "outputAppFieldSequenceIdMap must be a string and must exist in the input JSON" {
			t.Errorf("Expected error message 'outputAppFieldSequenceIdMap must be a string and must exist in the input JSON', got %v", err.Error())
		}
		// check when there the config field outputAppFieldSequenceType is not correct
		prevIncomingData = &IncomingData{NodeId: "1"}
		config = &SequencerConfig{}
		_, err = getPrevIdData(prevIncomingData, config)
		if err == nil {
			t.Fatalf("Expected error from getPrevIdData, got nil")
		}
		if err.Error() != "unknown outputAppFieldSequenceType" {
			t.Errorf("Expected error message 'unknown outputAppFieldSequenceType', got %v", err.Error())
		}
	})
	t.Run("valid", func(t *testing.T) {
		// check valid array case
		prevIncomingData := &IncomingData{NodeId: "1"}
		config := &SequencerConfig{outputAppFieldSequenceType: Array}
		prevIdData, err := getPrevIdData(prevIncomingData, config)
		if err != nil {
			t.Errorf("Expected no error from getPrevIdData, got %v", err)
		}
		prevIdDataArray, ok := prevIdData.([]string)
		if !ok {
			t.Fatalf("Expected prevIdData to be an array of strings")
		}
		if !reflect.DeepEqual(prevIdDataArray, []string{"1"}) {
			t.Errorf("Expected prevIdData to be ['1'], got %v", prevIdDataArray)
		}
		// check valid string case
		prevIncomingData = &IncomingData{NodeId: "1"}
		config = &SequencerConfig{outputAppFieldSequenceType: String}
		prevIdData, err = getPrevIdData(prevIncomingData, config)
		if err != nil {
			t.Errorf("Expected no error from getPrevIdData, got %v", err)
		}
		if prevIdData != "1" {
			t.Errorf("Expected prevIdData to be '1', got %v", prevIdData)
		}
	})
}

// MockConfig is a mock implementation of the Config interface
type MockConfig struct{}

func (mc *MockConfig) IngestConfig(config map[string]any) error {
	return nil
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
	for _, data := range mss.dataToSend {
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return err
		}
		err = mss.pushable.SendTo(Server.NewAppData(jsonBytes, ""))
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
	dataReceived chan ([]byte)
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

// Test Sequencer integrating with RunApp
func TestSeqeuncerRunApp(t *testing.T) {
	// Setup
	dataArray := []any{}
	for i := 0; i < 10; i++ {
		dataToAppend := map[string]any{
			"nodeId":  strconv.Itoa(i),
			"appJSON": map[string]any{"key": strconv.Itoa(i)},
		}
		if i != 9 {
			dataToAppend["childIds"] = []any{strconv.Itoa(i + 1)}
		} else {
			dataToAppend["childIds"] = []any{}
		}
		dataArray = append(dataArray, dataToAppend)
	}
	dataToSend := []any{
		dataArray,
	}
	mockSourceServer := &MockSourceServer{
		dataToSend: dataToSend,
	}
	chanForData := make(chan ([]byte), 10)
	mockSinkServer := &MockSinkServer{
		dataReceived: chanForData,
	}
	sequencer := &Sequencer{}
	sequencerConfig := &SequencerConfig{}
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
		`{"AppConfig":{"outputAppSequenceField":"seqField", "outputAppFieldSequenceType":"string"},"ProducersSetup":{"ProducerConfigs":[{"Type":"MockSink","ProducerConfig":{}}]},"ConsumersSetup":{"ConsumerConfigs":[{"Type":"MockSource","ConsumerConfig":{}}]}}`,
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
		sequencer, sequencerConfig,
		producerConfigMap, consumerConfigMap,
		producerMap, consumerMap,
	)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	close(chanForData)
	// Check if the data was transformed correctly
	counter := 0
	for data := range mockSinkServer.dataReceived {
		var dataAsArray []map[string]any
		err = json.Unmarshal(data, &dataAsArray)
		if err != nil {
			t.Fatalf("Expected no error from Unmarshal, got %v", err)
		}
		if len(dataAsArray) != 10 {
			t.Fatalf("Expected data to have 10 elements, got %v", len(dataAsArray))
		}
		for i, appJSON := range dataAsArray {
			if i == 0 {
				if !reflect.DeepEqual(appJSON, map[string]any{
					"key": "9",
				}) {
					t.Errorf("Expected appJSON to be %v, got %v", map[string]any{
						"key": "9",
					}, appJSON)
				}
			} else {
				if !reflect.DeepEqual(appJSON, map[string]any{
					"seqField": strconv.Itoa(10 - i),
					"key":      strconv.Itoa(9 - i),
				}) {
					t.Errorf("Expected appJSON to be %v, got %v", map[string]any{
						"seqField": strconv.Itoa(10 - i),
						"key":      strconv.Itoa(9 - i),
					}, appJSON)
				}
			}
		}
		counter++
	}
	if counter != 1 {
		t.Errorf("Expected 1 data points, got %d", counter)
	}
}

// Test getGroupAppliesValueOfListFromAppJSON
func TestGetGroupAppliesValueOfListFromAppJSON(t *testing.T) {
	// Test case: groupApplies is empty
	groupApplies := []GroupApply{}
	appJSON := map[string]any{}
	_, err := getGroupAppliesValueOfListFromAppJSON(appJSON, groupApplies)
	if err == nil {
		t.Fatalf("Expected error from getGroupAppliesValueOfListFromAppJSON, got nil")
	}
	if err.Error() != "groupApplies field to share value not found or not a string" {
		t.Errorf("Expected error message 'groupApplies field to share value not found or not a string', got %v", err.Error())
	}
	// Test case: groupApplies is not empty but there is not matching value in appJSON
	groupApplies = []GroupApply{
		{
			FieldToShare: "field",
		},
	}
	appJSON = map[string]any{}
	_, err = getGroupAppliesValueOfListFromAppJSON(appJSON, groupApplies)
	if err == nil {
		t.Fatalf("Expected error from getGroupAppliesValueOfListFromAppJSON, got nil")
	}
	if err.Error() != "groupApplies field to share value not found or not a string" {
		t.Errorf("Expected error message 'groupApplies field to share value not found or not a string', got %v", err.Error())
	}
	// Test case: groupApplies is not empty and there is a matching value in appJSON for a single groupApply
	groupApplies = []GroupApply{
		{
			FieldToShare: "field",
			IdentifyingField: "key",
			ValueOfIdentifyingField: "value",
		},
	}
	appJSON = map[string]any{
		"key":   "value",
		"field": "fieldValue",
	}
	value, err := getGroupAppliesValueOfListFromAppJSON(appJSON, groupApplies)
	if err != nil {
		t.Fatalf("Expected no error from getGroupAppliesValueOfListFromAppJSON, got %v", err)
	}
	if value != "fieldValue" {
		t.Errorf("Expected value to be 'fieldValue', got %v", value)
	}
	// Test case: groupApplies is not empty and there is a matching value in appJSON for multiple groupApplies
	// (the first matching value should be used)
	groupApplies = []GroupApply{
		{
			FieldToShare: "field",
			IdentifyingField: "key",
			ValueOfIdentifyingField: "value",
		},
		{
			FieldToShare: "otherField",
			IdentifyingField: "key",
			ValueOfIdentifyingField: "value",
		},
	}
	appJSON = map[string]any{
		"key":        "value",
		"field":      "fieldValue",
		"otherField": "otherFieldValue",
	}
	value, err = getGroupAppliesValueOfListFromAppJSON(appJSON, groupApplies)
	if err != nil {
		t.Fatalf("Expected no error from getGroupAppliesValueOfListFromAppJSON, got %v", err)
	}
	if value != "fieldValue" {
		t.Errorf("Expected value to be 'fieldValue', got %v", value)
	}
	// Test case: groupApplies is not empty and there is a matching value in appJSON for the second groupApply
	groupApplies = []GroupApply{
		{
			FieldToShare: "field",
			IdentifyingField: "key",
			ValueOfIdentifyingField: "value",
		},
		{
			FieldToShare: "otherField",
			IdentifyingField: "key",
			ValueOfIdentifyingField: "value",
		},
	}
	appJSON = map[string]any{
		"key":        "value",
		"otherField": "otherFieldValue",
	}
	value, err = getGroupAppliesValueOfListFromAppJSON(appJSON, groupApplies)
	if err != nil {
		t.Fatalf("Expected no error from getGroupAppliesValueOfListFromAppJSON, got %v", err)
	}
	if value != "otherFieldValue" {
		t.Errorf("Expected value to be 'otherFieldValue', got %v", value)
	}
	// Test case: groupApplies is not empty and the field to share is the same for multiple groupApplies
	groupApplies = []GroupApply{
		{
			FieldToShare: "field",
			IdentifyingField: "key",
			ValueOfIdentifyingField: "value",
		},
		{
			FieldToShare: "field",
			IdentifyingField: "key2",
			ValueOfIdentifyingField: "value2",
		},
	}
	appJSON = map[string]any{
		"key":   "value",
		"field": "fieldValue",
	}
	value, err = getGroupAppliesValueOfListFromAppJSON(appJSON, groupApplies)
	if err != nil {
		t.Fatalf("Expected no error from getGroupAppliesValueOfListFromAppJSON, got %v", err)
	}
	if value != "fieldValue" {
		t.Errorf("Expected value to be 'fieldValue', got %v", value)
	}
}



// Test getGroupApplyValueFromAppJSON
func TestGetGroupApplyValueFromAppJSON(t *testing.T) {
	// Test case: groupApply.IdentifyingField not in appJSON
	groupApply := GroupApply{
		IdentifyingField: "key",
	}
	appJSON := map[string]any{}
	_, err := getGroupApplyValueFromAppJSON(appJSON, groupApply)
	if err == nil {
		t.Fatalf("Expected error from getGroupApplyValueFromAppJSON, got nil")
	}
	if err.Error() != "groupApplies identifying field not found in appJSON or identifying value not a string" {
		t.Errorf("Expected error message 'groupApplies identifying field not found in appJSON or identifying value not a string', got %v", err.Error())
	}
	// Test case: groupApply.IdentifyingField in appJSON but value is not a string
	groupApply = GroupApply{
		IdentifyingField: "key",
	}
	appJSON = map[string]any{
		"key": 1,
	}
	_, err = getGroupApplyValueFromAppJSON(appJSON, groupApply)
	if err == nil {
		t.Fatalf("Expected error from getGroupApplyValueFromAppJSON, got nil")
	}
	if err.Error() != "groupApplies identifying field not found in appJSON or identifying value not a string" {
		t.Errorf("Expected error message 'groupApplies identifying field not found in appJSON or identifying value not a string', got %v", err.Error())
	}
	// Test case: groupApply.IdentifyingField in appJSON and value is a string but does not match groupApply.ValueOfIdentifyingField
	groupApply = GroupApply{
		IdentifyingField:        "key",
		ValueOfIdentifyingField: "value",
	}
	appJSON = map[string]any{
		"key": "notValue",
	}
	_, err = getGroupApplyValueFromAppJSON(appJSON, groupApply)
	if err == nil {
		t.Fatalf("Expected error from getGroupApplyValueFromAppJSON, got nil")
	}
	if err.Error() != "groupApplies identifying field value does not match" {
		t.Errorf("Expected error message 'groupApplies identifying field value does not match', got %v", err.Error())
	}
	// Test case: groupApply.IdentifyingField in appJSON and value is a string and matches groupApply.ValueOfIdentifyingField but groupApply.FieldToShare is not in appJSON
	groupApply = GroupApply{
		IdentifyingField:        "key",
		ValueOfIdentifyingField: "value",
		FieldToShare:            "field",
	}
	appJSON = map[string]any{
		"key": "value",
	}
	_, err = getGroupApplyValueFromAppJSON(appJSON, groupApply)
	if err == nil {
		t.Fatalf("Expected error from getGroupApplyValueFromAppJSON, got nil")
	}
	if err.Error() != "groupApplies field to share value not found or not a string" {
		t.Errorf("Expected error message 'groupApplies field to share value not found or not a string', got %v", err.Error())
	}
	// Test case: groupApply.IdentifyingField in appJSON and value is a string and matches groupApply.ValueOfIdentifyingField and groupApply.FieldToShare is in appJSON but not a string
	groupApply = GroupApply{
		IdentifyingField:        "key",
		ValueOfIdentifyingField: "value",
		FieldToShare:            "field",
	}
	appJSON = map[string]any{
		"key":   "value",
		"field": 1,
	}
	_, err = getGroupApplyValueFromAppJSON(appJSON, groupApply)
	if err == nil {
		t.Fatalf("Expected error from getGroupApplyValueFromAppJSON, got nil")
	}
	if err.Error() != "groupApplies field to share value not found or not a string" {
		t.Errorf("Expected error message 'groupApplies field to share value not found or not a string', got %v", err.Error())
	}
	// Test case: everything works fine
	groupApply = GroupApply{
		IdentifyingField:        "key",
		ValueOfIdentifyingField: "value",
		FieldToShare:            "field",
	}
	appJSON = map[string]any{
		"key":   "value",
		"field": "fieldValue",
	}
	value, err := getGroupApplyValueFromAppJSON(appJSON, groupApply)
	if err != nil {
		t.Fatalf("Expected no error from getGroupApplyValueFromAppJSON, got %v", err)
	}
	if value != "fieldValue" {
		t.Errorf("Expected value to be 'fieldValue', got %v", value)
	}
}

// Test orderChildrenByTimestamp
func TestOrderChildrenByTimestamp(t *testing.T) {
	// Test case: len of children is 0
	incomingData := &incomingDataWithDuplicates{
		IncomingData: &IncomingData{},
	}
	err := orderChildrenByTimestamp(incomingData, map[string]*incomingDataWithDuplicates{})
	if err != nil {
		t.Fatalf("Expected no error from orderChildrenByTimestamp, got %v", err)
	}
	if len(incomingData.ChildIds) != 0 {
		t.Errorf("Expected incomingData.ChildIds to be empty, got %v", incomingData.ChildIds)
	}
	// Test case: childId is not found in nodeIdToIncomingDataMap
	incomingData = &incomingDataWithDuplicates{
		IncomingData: &IncomingData{
			ChildIds: []string{"1"},
		},
	}
	err = orderChildrenByTimestamp(incomingData, map[string]*incomingDataWithDuplicates{})
	if err == nil {
		t.Fatalf("Expected error from orderChildrenByTimestamp, got nil")
	}
	if err.Error() != "child node 1 not found but attempting to order children by timestamp" {
		t.Errorf("Expected error message 'child node 1 not found but attempting to order children by timestamp', got %v", err.Error())
	}
	// Test case: child Timestamp is not set
	incomingData = &incomingDataWithDuplicates{
		IncomingData: &IncomingData{
			ChildIds: []string{"1"},
		},
	}
	err = orderChildrenByTimestamp(incomingData, map[string]*incomingDataWithDuplicates{
		"1": {
			IncomingData: &IncomingData{
				NodeId: "1",
			},
		},
	})
	if err == nil {
		t.Fatalf("Expected error from orderChildrenByTimestamp, got nil")
	}
	if err.Error() != "child node 1 has no timestamp but attempting to order children by timestamp" {
		t.Errorf("Expected error message 'child node 1 has no timestamp but attempting to order children by timestamp', got %v", err.Error())
	}
	// Test case: valid case
	incomingData = &incomingDataWithDuplicates{
		IncomingData: &IncomingData{
			ChildIds: []string{"1", "2", "3", "4"},
		},
	}
	nodeIdToIncomingDataMap := map[string]*incomingDataWithDuplicates{
		"1": {IncomingData: &IncomingData{Timestamp: 3}},
		"2": {IncomingData: &IncomingData{Timestamp: 1}},
		"3": {IncomingData: &IncomingData{Timestamp: 4}},
		"4": {IncomingData: &IncomingData{Timestamp: 2}},
	}
	err = orderChildrenByTimestamp(incomingData, nodeIdToIncomingDataMap)
	if err != nil {
		t.Fatalf("Expected no error from orderChildrenByTimestamp, got %v", err)
	}
	if !reflect.DeepEqual(incomingData.ChildIds, []string{"2", "4", "1", "3"}) {
		t.Errorf("Expected incomingData.ChildIds to be ['2', '4', '1', '3'], got %v", incomingData.ChildIds)
	}
}
