package common

import (
	"reflect"
	"strings"
)

// GetMapDiff returns the difference between two maps
func GetMapDiff(oldMap, newMap map[string]interface{}) (map[string]interface{}, []string) {
	diff := make(map[string]interface{})
	remove := make([]string, 0)

	for k, v := range oldMap {
		if _, ok := newMap[k]; !ok {
			remove = append(remove, k)
		}

		if newMap[k] != v {
			diff[k] = newMap[k]
		}
	}

	return diff, remove
}

// GetStructDiff returns the differences between two struct values of the same type.
// It returns a map with field names and their values in the second struct for differing fields,
// and a slice of field names that are zero-valued in the second struct but not in the first.
func GetStructDiff(oldStruct, newStruct interface{}) (map[string]interface{}, []string) {
	diff := make(map[string]interface{})
	remove := make([]string, 0)

	oldVal := reflect.ValueOf(oldStruct)
	newVal := reflect.ValueOf(newStruct)

	for i := 0; i < oldVal.NumField(); i++ {
		oldField := oldVal.Field(i)
		newField := newVal.Field(i)

		fieldType := oldVal.Type().Field(i)
		jsonTag := fieldType.Tag.Get("json")
		if strings.Contains(jsonTag, ",") {
			jsonTagParts := strings.Split(jsonTag, ",")
			jsonTag = jsonTagParts[0]
		}

		if jsonTag == "" {
			jsonTag = fieldType.Name
		}

		// Check if the field is zero-valued in the new struct but not in the old struct
		if !oldField.IsZero() && newField.IsZero() {
			remove = append(remove, jsonTag)
			continue
		}

		// Check if the field values are different
		if !reflect.DeepEqual(oldField.Interface(), newField.Interface()) {
			diff[jsonTag] = newField.Interface()
		}
	}

	return diff, remove
}
