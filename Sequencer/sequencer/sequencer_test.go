package sequencer

import (
	"errors"
	"reflect"
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
	})
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
			err = sequencer.SendTo(Server.NewAppData(
				1, "",
			))
			if err == nil {
				t.Fatalf("Expected error from SendTo, got nil")
			}
			if err.Error() != "data must be an array of maps" {
				t.Errorf("Expected error message 'data must be an array of maps', got %v", err.Error())
			}
			// check error case where there are no root nodes
			sequencer.config = &SequencerConfig{}
			err = sequencer.SendTo(Server.NewAppData(
				[]map[string]any{}, "",
			))
			if err == nil {
				t.Fatalf("Expected error from SendTo, got nil")
			}
			if err.Error() != "no root nodes" {
				t.Errorf("Expected error message 'no root nodes', got %v", err.Error())
			}
			// check error case where sequenceWithStack returns an error
			sequencer.config = &SequencerConfig{}
			err = sequencer.SendTo(Server.NewAppData(
				[]map[string]any{
					{
						"nodeId":          "1",
						"orderedChildIds": []any{"2"},
						"appJSON":         map[string]any{},
					},
				},
				"",
			))
			if err == nil {
				t.Fatalf("Expected error from SendTo, got nil")
			}
			if err.Error() != "child node not found: 2" {
				t.Errorf("Expected error message 'child node not found: 2', got %v", err.Error())
			}
			// check error case where getPrevIdData returns an error
			sequencer.config = &SequencerConfig{}
			appData := Server.NewAppData(
				[]map[string]any{
					{
						"nodeId":          "1",
						"orderedChildIds": []any{"2"},
						"appJSON":         map[string]any{},
					},
					{
						"nodeId":          "2",
						"orderedChildIds": []any{},
						"appJSON":         map[string]any{},
					},
				},
				"",
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
				[]map[string]any{
					{
						"nodeId":          "1",
						"orderedChildIds": []any{"2"},
						"appJSON":         map[string]any{},
					},
					{
						"nodeId":          "2",
						"orderedChildIds": []any{},
						"appJSON":         map[string]any{},
					},
				},
				"",
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
			expectedData := []map[string]any{
				{},
				{"SeqField": "2"},
			}
			if !reflect.DeepEqual(gotData, expectedData) {
				t.Errorf("Expected appData to be %v, got %v", expectedData, gotData)
			}
			// check valid case with multiple root nodes
			pushableChan = make(chan (*Server.AppData), 1)
			sequencer = &Sequencer{
				config: &SequencerConfig{
					outputAppFieldSequenceType: String,
					outputAppSequenceField:     "SeqField",
				},
				pushable: &MockPushable{
					incomingData: pushableChan,
				},
			}
			appData = Server.NewAppData(
				[]map[string]any{
					{
						"nodeId":          "1",
						"orderedChildIds": []any{"2"},
						"appJSON":         map[string]any{},
					},
					{
						"nodeId":          "2",
						"orderedChildIds": []any{},
						"appJSON":         map[string]any{"field": "2"},
					},
					{
						"nodeId":          "3",
						"orderedChildIds": []any{"4"},
						"appJSON":         map[string]any{},
					},
					{
						"nodeId":          "4",
						"orderedChildIds": []any{},
						"appJSON":         map[string]any{"field": "4"},
					},
				},
				"",
			)
			err = sequencer.SendTo(appData)
			if err != nil {
				t.Errorf("Expected no error from SendTo, got %v", err)
			}
			close(pushableChan)
			receivedAppData = <-pushableChan
			if receivedAppData == nil {
				t.Fatalf("Expected appData to be sent")
			}
			gotData, err = receivedAppData.GetData()
			if err != nil {
				t.Fatalf("Expected no error from GetData, got %v", err)
			}
			gotDataArray, ok := gotData.([]map[string]any)
			if !ok {
				t.Fatalf("Expected gotData to be an array of maps")
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
		})
	})
}

// TestStackIncomingData tests the stackIncomingData struct
func TestStackIncomingData(t *testing.T) {
	t.Run("nextChildId", func(t *testing.T) {
		sid := &stackIncomingData{}
		// check error case when currentChildIdIndex is greater than or equal to len(incomingData)
		sid.incomingData = &incomingData{
			orderedChildIds: []string{},
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
		sid.incomingData = &incomingData{
			orderedChildIds: []string{"1", "2"},
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
		rootNode := &incomingData{}
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
		nodeIdToIncomingDataMap := map[string]*incomingData{}
		rootNode := &incomingData{
			orderedChildIds: []string{"1"},
		}
		sequence := sequenceWithStack(rootNode, nodeIdToIncomingDataMap)
		counter := 0
		for node, err := range sequence {
			if node != nil {
				t.Errorf("Expected node to be nil, got %v", node)
			}
			if err == nil {
				t.Fatalf("Expected error from sequence, got nil")
			}
			if err.Error() != "child node not found: 1" {
				t.Errorf("Expected error message 'child node not found: 1', got %v", err.Error())
			}
			counter++
		}
		if counter != 1 {
			t.Fatalf("Expected counter to be 1, got %v", counter)
		}
	})
	t.Run("childNodeIsNil", func(t *testing.T) {
		// check error case when child node is nil
		nodeIdToIncomingDataMap := map[string]*incomingData{
			"1": nil,
		}
		rootNode := &incomingData{
			orderedChildIds: []string{"1"},
		}
		sequence := sequenceWithStack(rootNode, nodeIdToIncomingDataMap)
		counter := 0
		for node, err := range sequence {
			if node != nil {
				t.Errorf("Expected node to be nil, got %v", node)
			}
			if err == nil {
				t.Fatalf("Expected error from sequence, got nil")
			}
			if err.Error() != "child node is nil: 1" {
				t.Errorf("Expected error message 'child node is nil: 1', got %v", err.Error())
			}
			counter++
		}
		if counter != 1 {
			t.Fatalf("Expected counter to be 1, got %v", counter)
		}
	})
	t.Run("valid", func(t *testing.T) {
		// check valid case
		nodeIdToIncomingDataMap := map[string]*incomingData{
			"1": {
				nodeId:          "1",
				orderedChildIds: []string{"2", "3"},
			},
			"2": {
				nodeId:          "2",
				orderedChildIds: []string{"4", "5"},
			},
			"3": {
				nodeId:          "3",
				orderedChildIds: []string{"6", "7"},
			},
			"4": {nodeId: "4"},
			"5": {nodeId: "5"},
			"6": {nodeId: "6"},
			"7": {nodeId: "7"},
		}
		rootNode := nodeIdToIncomingDataMap["1"]
		sequence := sequenceWithStack(rootNode, nodeIdToIncomingDataMap)
		counter := 0
		expectedSequence := []string{"4", "5", "2", "6", "7", "3", "1"}
		for node, err := range sequence {
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

// Tests convertRawDataToIncomingData
func TestConvertRawDataMapToIncomingData(t *testing.T) {
	t.Run("invalidData", func(t *testing.T) {
		// check error case when nodeId is not set
		data := map[string]any{}
		_, err := convertRawDataMapToIncomingData(data)
		if err == nil {
			t.Fatalf("Expected error from convertRawDataToIncomingData, got nil")
		}
		if err.Error() != "nodeId must be set and must be a string" {
			t.Errorf("Expected error message 'nodeId must be set and must be a string', got %v", err.Error())
		}
		// check error case when nodeId is not a string
		data = map[string]any{
			"nodeId": 1,
		}
		_, err = convertRawDataMapToIncomingData(data)
		if err == nil {
			t.Fatalf("Expected error from convertRawDataToIncomingData, got nil")
		}
		if err.Error() != "nodeId must be set and must be a string" {
			t.Errorf("Expected error message 'nodeId must be set and must be a string', got %v", err.Error())
		}
		// check error case when orderedChildIds is not set
		data = map[string]any{
			"nodeId": "1",
		}
		_, err = convertRawDataMapToIncomingData(data)
		if err == nil {
			t.Fatalf("Expected error from convertRawDataToIncomingData, got nil")
		}
		if err.Error() != "orderedChildIds must be set and must be an array of strings" {
			t.Errorf("Expected error message 'orderedChildIds must be set and must be an array of strings', got %v", err.Error())
		}
		// check error case when orderedChildIds is not an array
		data = map[string]any{
			"nodeId":          "1",
			"orderedChildIds": 1,
		}
		_, err = convertRawDataMapToIncomingData(data)
		if err == nil {
			t.Fatalf("Expected error from convertRawDataToIncomingData, got nil")
		}
		if err.Error() != "orderedChildIds must be set and must be an array of strings" {
			t.Errorf("Expected error message 'orderedChildIds must be set and must be an array of strings', got %v", err.Error())
		}
		// check error case when orderedChildIds is not an array of strings
		data = map[string]any{
			"nodeId":          "1",
			"orderedChildIds": []any{1},
		}
		_, err = convertRawDataMapToIncomingData(data)
		if err == nil {
			t.Fatalf("Expected error from convertRawDataToIncomingData, got nil")
		}
		if err.Error() != "orderedChildIds must be set and must be an array of strings" {
			t.Errorf("Expected error message 'orderedChildIds must be set and must be an array of strings', got %v", err.Error())
		}
		// check error case where appJSON is not set
		data = map[string]any{
			"nodeId":          "1",
			"orderedChildIds": []any{"2"},
		}
		_, err = convertRawDataMapToIncomingData(data)
		if err == nil {
			t.Fatalf("Expected error from convertRawDataToIncomingData, got nil")
		}
		if err.Error() != "appJSON must be set and must be a map" {
			t.Errorf("Expected error message 'appJSON must be set and must be a map', got %v", err.Error())
		}
		// check error case where appJSON is not a map
		data = map[string]any{
			"nodeId":          "1",
			"orderedChildIds": []any{"2"},
			"appJSON":         1,
		}
		_, err = convertRawDataMapToIncomingData(data)
		if err == nil {
			t.Fatalf("Expected error from convertRawDataToIncomingData, got nil")
		}
		if err.Error() != "appJSON must be set and must be a map" {
			t.Errorf("Expected error message 'appJSON must be set and must be a map', got %v", err.Error())
		}
	})
	t.Run("valid", func(t *testing.T) {
		// check valid case
		data := map[string]any{
			"nodeId":          "1",
			"orderedChildIds": []any{"2", "3"},
			"appJSON": map[string]any{
				"key": "value",
			},
		}
		incomingData, err := convertRawDataMapToIncomingData(data)
		if err != nil {
			t.Errorf("Expected no error from convertRawDataToIncomingData, got %v", err)
		}
		if incomingData.nodeId != "1" {
			t.Errorf("Expected nodeId to be '1', got %v", incomingData.nodeId)
		}
		if !reflect.DeepEqual(incomingData.orderedChildIds, []string{"2", "3"}) {
			t.Errorf("Expected orderedChildIds to be ['2', '3'], got %v", incomingData.orderedChildIds)
		}
		if !reflect.DeepEqual(incomingData.appJSON, map[string]any{"key": "value"}) {
			t.Errorf("Expected appJSON to be {'key': 'value'}, got %v", incomingData.appJSON)
		}
	})
}

// Tests convertToIncomingDataMapAndRootNodes
func TestConvertToIncomingDataMapAndRootNodes(t *testing.T) {
	t.Run("invalidData", func(t *testing.T) {
		// check error case when data is not an array
		_, _, err := convertToIncomingDataMapAndRootNodes(1)
		if err == nil {
			t.Fatalf("Expected error from convertAppDataToIncomingDataMapAndRootNodes, got nil")
		}
		if err.Error() != "data must be an array of maps" {
			t.Errorf("Expected error message 'data must be an array of maps', got %v", err.Error())
		}
		// check error case when data is not an array of maps
		_, _, err = convertToIncomingDataMapAndRootNodes([]any{1})
		if err == nil {
			t.Fatalf("Expected error from convertAppDataToIncomingDataMapAndRootNodes, got nil")
		}
		if err.Error() != "data must be an array of maps" {
			t.Errorf("Expected error message 'data must be an array of maps', got %v", err.Error())
		}
		// check error case where convertRawDataMapToIncomingData returns an error
		_, _, err = convertToIncomingDataMapAndRootNodes([]map[string]any{
			{},
		})
		if err == nil {
			t.Fatalf("Expected error from convertAppDataToIncomingDataMapAndRootNodes, got nil")
		}
		if err.Error() != "nodeId must be set and must be a string" {
			t.Errorf("Expected error message 'nodeId must be set and must be a string', got %v", err.Error())
		}
	})
	t.Run("valid", func(t *testing.T) {
		// check valid case
		nodes := []map[string]any{
			{
				"nodeId":          "1",
				"orderedChildIds": []any{"2", "3"},
				"appJSON":         map[string]any{},
			},
			{
				"nodeId":          "2",
				"orderedChildIds": []any{},
				"appJSON":         map[string]any{},
			},
			{
				"nodeId":          "3",
				"orderedChildIds": []any{},
				"appJSON":         map[string]any{},
			},
		}
		incomingDataMap, rootNodes, err := convertToIncomingDataMapAndRootNodes(nodes)
		if err != nil {
			t.Errorf("Expected no error from convertAppDataToIncomingDataMapAndRootNodes, got %v", err)
		}
		if len(incomingDataMap) != 3 {
			t.Errorf("Expected incomingDataMap to have 3 elements, got %v", len(incomingDataMap))
		}
		if len(rootNodes) != 1 {
			t.Errorf("Expected rootNodes to have 1 element, got %v", len(rootNodes))
		}
		for _, inputMap := range nodes {
			incomingData, ok := incomingDataMap[inputMap["nodeId"].(string)]
			if !ok {
				t.Errorf("Expected incomingDataMap to have key %v", inputMap["nodeId"])
			}
			if incomingData.nodeId != inputMap["nodeId"] {
				t.Errorf("Expected nodeId to be %v, got %v", inputMap["nodeId"], incomingData.nodeId)
			}
			expectedOrderedChildIds := []string{}
			for _, childId := range inputMap["orderedChildIds"].([]any) {
				expectedOrderedChildIds = append(expectedOrderedChildIds, childId.(string))
			}
			if !reflect.DeepEqual(incomingData.orderedChildIds, expectedOrderedChildIds) {
				t.Errorf("Expected orderedChildIds to be %v, got %v", inputMap["orderedChildIds"], incomingData.orderedChildIds)
			}
			if !reflect.DeepEqual(incomingData.appJSON, inputMap["appJSON"]) {
				t.Errorf("Expected appJSON to be %v, got %v", inputMap["appJSON"], incomingData.appJSON)
			}
		}
		rootNode, ok := rootNodes["1"]
		if !ok {
			t.Errorf("Expected rootNodes to have key '1'")
		}
		if rootNode != incomingDataMap["1"] {
			t.Errorf("Expected rootNode to be %v, got %v", incomingDataMap["1"], rootNode)
		}
	})
}

// Test getPrevIdFromPrevIncomingData
func TestGetPrevIdFromPrevIncomingData(t *testing.T) {
	t.Run("invalidData", func(t *testing.T) {
		// check error case outputAppFieldSequenceIdMap is not found
		// in prevIncomingData appJSON
		prevIncomingData := &incomingData{
			appJSON: map[string]any{},
		}
		_, err := getPrevIdFromPrevIncomingData(prevIncomingData, "test")
		if err == nil {
			t.Fatalf("Expected error from getPrevIdFromPrevIncomingData, got nil")
		}
		if err.Error() != "outputAppFieldSequenceIdMap must be a string and must exist in the input JSON" {
			t.Errorf("Expected error message 'outputAppFieldSequenceIdMap must be a string and must exist in the input JSON', got %v", err.Error())
		}
		// check error case outputAppFieldSequenceIdMap is not a string
		prevIncomingData = &incomingData{
			appJSON: map[string]any{
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
		prevIncomingData := &incomingData{
			nodeId: "1",
			appJSON: map[string]any{
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
		prevIncomingData := &incomingData{}
		config := &SequencerConfig{outputAppFieldSequenceIdMap: "test"}
		_, err := getPrevIdData(prevIncomingData, config)
		if err == nil {
			t.Fatalf("Expected error from getPrevIdData, got nil")
		}
		if err.Error() != "outputAppFieldSequenceIdMap must be a string and must exist in the input JSON" {
			t.Errorf("Expected error message 'outputAppFieldSequenceIdMap must be a string and must exist in the input JSON', got %v", err.Error())
		}
		// check when there the config field outputAppFieldSequenceType is not correct
		prevIncomingData = &incomingData{nodeId: "1"}
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
		prevIncomingData := &incomingData{nodeId: "1"}
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
		prevIncomingData = &incomingData{nodeId: "1"}
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
