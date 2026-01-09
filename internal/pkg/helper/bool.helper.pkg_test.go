package helper

import (
	"testing"
)

func TestGetMapBoolValue(t *testing.T) {
	t.Run("valid true string", func(t *testing.T) {
		h := map[string]interface{}{"flag": "true"}
		result := GetMapBoolValue(h, "flag")
		if result == nil || *result != true {
			t.Errorf("expected true, got %v", result)
		}
	})

	t.Run("valid false string", func(t *testing.T) {
		h := map[string]interface{}{"flag": "false"}
		result := GetMapBoolValue(h, "flag")
		if result == nil || *result != false {
			t.Errorf("expected false, got %v", result)
		}
	})

	t.Run("case insensitive true", func(t *testing.T) {
		h := map[string]interface{}{"flag": "TrUe"}
		result := GetMapBoolValue(h, "flag")
		if result == nil || *result != true {
			t.Errorf("expected true, got %v", result)
		}
	})

	t.Run("invalid bool string", func(t *testing.T) {
		h := map[string]interface{}{"flag": "notabool"}
		result := GetMapBoolValue(h, "flag")
		if result == nil || *result != false {
			t.Errorf("expected false for invalid bool, got %v", result)
		}
	})

	t.Run("missing key", func(t *testing.T) {
		h := map[string]interface{}{"other": "true"}
		result := GetMapBoolValue(h, "flag")
		if result == nil || *result != false {
			t.Errorf("expected false for missing key, got %v", result)
		}
	})

	t.Run("non-string value", func(t *testing.T) {
		h := map[string]interface{}{"flag": 123}
		result := GetMapBoolValue(h, "flag")
		if result == nil || *result != false {
			t.Errorf("expected false for non-string value, got %v", result)
		}
	})
}

func TestContains(t *testing.T) {
	t.Run("string slice contains element", func(t *testing.T) {
		arr := []string{"apple", "banana", "cherry"}
		result := Contains(arr, "banana")
		if !result {
			t.Errorf("expected true, got %v", result)
		}
	})

	t.Run("string slice does not contain element", func(t *testing.T) {
		arr := []string{"apple", "banana", "cherry"}
		result := Contains(arr, "orange")
		if result {
			t.Errorf("expected false, got %v", result)
		}
	})

	t.Run("int slice contains element", func(t *testing.T) {
		arr := []int{1, 2, 3, 4, 5}
		result := Contains(arr, 3)
		if !result {
			t.Errorf("expected true, got %v", result)
		}
	})

	t.Run("int slice does not contain element", func(t *testing.T) {
		arr := []int{1, 2, 3, 4, 5}
		result := Contains(arr, 6)
		if result {
			t.Errorf("expected false, got %v", result)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		arr := []string{}
		result := Contains(arr, "test")
		if result {
			t.Errorf("expected false for empty slice, got %v", result)
		}
	})

	t.Run("bool slice contains element", func(t *testing.T) {
		arr := []bool{true, false, true}
		result := Contains(arr, false)
		if !result {
			t.Errorf("expected true, got %v", result)
		}
	})

	t.Run("float slice contains element", func(t *testing.T) {
		arr := []float64{1.1, 2.2, 3.3}
		result := Contains(arr, 2.2)
		if !result {
			t.Errorf("expected true, got %v", result)
		}
	})

	t.Run("single element slice", func(t *testing.T) {
		arr := []string{"only"}
		result := Contains(arr, "only")
		if !result {
			t.Errorf("expected true, got %v", result)
		}
	})
}
