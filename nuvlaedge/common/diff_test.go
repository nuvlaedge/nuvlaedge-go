package common

import (
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

	b.FieldB = "test"
	diff, del = GetStructDiff(a, b)

}
