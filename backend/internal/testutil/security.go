package testutil

import (
	"fmt"
	"reflect"
	"testing"
)

// AssertNoInternalIDs recursively checks if any field in the given interface
// contains a value that matches the forbidden internal ID.
// This is used to ensure sequential numeric IDs are not leaked in DTOs.
func AssertNoInternalIDs(t *testing.T, data interface{}, forbiddenID uint64) {
	t.Helper()
	v := reflect.ValueOf(data)
	checkValue(t, v, forbiddenID, "")
}

func checkValue(t *testing.T, v reflect.Value, forbiddenID uint64, path string) {
	// Dereference pointers
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			fieldName := field.Name
			if path != "" {
				fieldName = path + "." + fieldName
			}
			checkValue(t, v.Field(i), forbiddenID, fieldName)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			indexPath := fmt.Sprintf("%s[%d]", path, i)
			checkValue(t, v.Index(i), forbiddenID, indexPath)
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			keyPath := fmt.Sprintf("%s{%v}", path, key)
			checkValue(t, v.MapIndex(key), forbiddenID, keyPath)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v.Uint() == forbiddenID {
			t.Errorf("Security Violation: Found internal ID %d at path %s", forbiddenID, path)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Int() == int64(forbiddenID) {
			t.Errorf("Security Violation: Found internal ID %d at path %s", forbiddenID, path)
		}
	}
}
