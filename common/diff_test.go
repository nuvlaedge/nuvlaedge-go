package common

import (
	"reflect"
	"testing"
)

func TestGetStructDiff(t *testing.T) {
	type TestStruct struct {
		FieldA int    `json:"field-a,omitempty"`
		FieldB string `json:"field-b,omitempty"`
	}

	a := TestStruct{FieldA: 1, FieldB: "test"}
	b := TestStruct{FieldA: 1, FieldB: "test"}

	diff, del := GetStructDiff(a, b)
	if len(diff) != 0 {
		t.Errorf("Expected no differences, got %v", diff)
	}
	if len(del) != 0 {
		t.Errorf("Expected no deletions, got %v", del)
	}

	b.FieldA = 2
	diff, del = GetStructDiff(a, b)
	if _, ok := diff["field-a"]; !ok {
		t.Errorf("Expected field name to be json tag, got %v", diff)
	}

	if len(diff) != 1 || diff["field-a"] != 2 {
		t.Errorf("Expected difference in FieldA, got %v", diff)
	}

	if len(del) != 0 {
		t.Errorf("Expected no deletions, got %v", del)
	}

	b.FieldB = ""
	diff, del = GetStructDiff(a, b)

	if _, ok := diff["field-b"]; ok {
		t.Errorf("New struct has an empty field %v", diff)
	}
	if len(del) != 1 || del[0] != "field-b" {
		t.Errorf("Expected deletion in FieldB, got %v", del)
	}

}

func TestGetMapDiff(t *testing.T) {
	oldMap := map[string]interface{}{
		"key1": "value1",
		"key2": 2,
		"key3": true,
	}

	newMapSame := map[string]interface{}{
		"key1": "value1",
		"key2": 2,
		"key3": true,
	}

	newMapDiff := map[string]interface{}{
		"key1": "value1",
		"key2": 3,     // Changed value
		"key4": "new", // New key
	}

	newMapMissing := map[string]interface{}{
		"key1": "value1",
	}

	// Test case: Maps are the same
	diff, remove := GetMapDiff(oldMap, newMapSame)
	if len(diff) != 0 || len(remove) != 0 {
		t.Errorf("Expected no differences or removals, got diff: %v, remove: %v", diff, remove)
	}

	// Test case: Maps have differences
	diff, remove = GetMapDiff(oldMap, newMapDiff)
	if !reflect.DeepEqual(diff, map[string]interface{}{"key2": 3, "key4": "new"}) || len(remove) != 1 || remove[0] != "key3" {
		t.Errorf("Expected differences and removals, got diff: %v, remove: %v", diff, remove)
	}

	// Test case: New map is missing keys
	diff, remove = GetMapDiff(oldMap, newMapMissing)
	if len(diff) != 0 || len(remove) != 2 || !contains(remove, "key2") || !contains(remove, "key3") {
		t.Errorf("Expected removals for missing keys, got diff: %v, remove: %v", diff, remove)
	}
}

// Helper function to check if slice contains a string
func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}
