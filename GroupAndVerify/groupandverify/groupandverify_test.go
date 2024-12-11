package groupandverify

import (
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
		})
	})
}
