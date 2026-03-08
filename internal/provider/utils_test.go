package provider

import (
	"testing"
)

// ============================================================
// Unit Tests for utils.go helper functions
// ============================================================

func TestGetStr(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		key      string
		expected string
	}{
		{"existing string", map[string]interface{}{"key": "value"}, "key", "value"},
		{"missing key", map[string]interface{}{}, "key", ""},
		{"nil value", map[string]interface{}{"key": nil}, "key", ""},
		{"non-string value", map[string]interface{}{"key": 123}, "key", ""},
		{"empty string", map[string]interface{}{"key": ""}, "key", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStr(tt.data, tt.key)
			if result != tt.expected {
				t.Errorf("getStr(%v, %q) = %q, want %q", tt.data, tt.key, result, tt.expected)
			}
		})
	}
}

func TestGetBool(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		key      string
		expected bool
	}{
		{"true", map[string]interface{}{"key": true}, "key", true},
		{"false", map[string]interface{}{"key": false}, "key", false},
		{"missing", map[string]interface{}{}, "key", false},
		{"nil", map[string]interface{}{"key": nil}, "key", false},
		{"non-bool", map[string]interface{}{"key": "true"}, "key", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBool(tt.data, tt.key)
			if result != tt.expected {
				t.Errorf("getBool(%v, %q) = %v, want %v", tt.data, tt.key, result, tt.expected)
			}
		})
	}
}

func TestGetBoolPtr(t *testing.T) {
	tests := []struct {
		name        string
		data        map[string]interface{}
		key         string
		expectNil   bool
		expectValue bool
	}{
		{"true returns pointer", map[string]interface{}{"key": true}, "key", false, true},
		{"false returns nil", map[string]interface{}{"key": false}, "key", true, false},
		{"missing returns nil", map[string]interface{}{}, "key", true, false},
		{"nil returns nil", map[string]interface{}{"key": nil}, "key", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBoolPtr(tt.data, tt.key)
			if tt.expectNil {
				if result != nil {
					t.Errorf("getBoolPtr(%v, %q) = %v, want nil", tt.data, tt.key, *result)
				}
			} else {
				if result == nil {
					t.Errorf("getBoolPtr(%v, %q) = nil, want %v", tt.data, tt.key, tt.expectValue)
				} else if *result != tt.expectValue {
					t.Errorf("getBoolPtr(%v, %q) = %v, want %v", tt.data, tt.key, *result, tt.expectValue)
				}
			}
		})
	}
}

func TestGetIntPtr(t *testing.T) {
	tests := []struct {
		name        string
		data        map[string]interface{}
		key         string
		expectNil   bool
		expectValue int
	}{
		{"positive int", map[string]interface{}{"key": 3}, "key", false, 3},
		{"zero returns nil", map[string]interface{}{"key": 0}, "key", true, 0},
		{"missing returns nil", map[string]interface{}{}, "key", true, 0},
		{"nil returns nil", map[string]interface{}{"key": nil}, "key", true, 0},
		{"non-int returns nil", map[string]interface{}{"key": "3"}, "key", true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getIntPtr(tt.data, tt.key)
			if tt.expectNil {
				if result != nil {
					t.Errorf("getIntPtr(%v, %q) = %v, want nil", tt.data, tt.key, *result)
				}
			} else {
				if result == nil {
					t.Errorf("getIntPtr(%v, %q) = nil, want %v", tt.data, tt.key, tt.expectValue)
				} else if *result != tt.expectValue {
					t.Errorf("getIntPtr(%v, %q) = %v, want %v", tt.data, tt.key, *result, tt.expectValue)
				}
			}
		})
	}
}

func TestGetStrList(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		key      string
		expected []string
	}{
		{
			"normal list",
			map[string]interface{}{"key": []interface{}{"a", "b", "c"}},
			"key",
			[]string{"a", "b", "c"},
		},
		{
			"empty list returns nil",
			map[string]interface{}{"key": []interface{}{}},
			"key",
			nil,
		},
		{
			"missing key returns nil",
			map[string]interface{}{},
			"key",
			nil,
		},
		{
			"nil value returns nil",
			map[string]interface{}{"key": nil},
			"key",
			nil,
		},
		{
			"filters empty strings",
			map[string]interface{}{"key": []interface{}{"a", "", "c"}},
			"key",
			[]string{"a", "c"},
		},
		{
			"all empty strings returns nil",
			map[string]interface{}{"key": []interface{}{"", ""}},
			"key",
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStrList(tt.data, tt.key)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("getStrList() = %v, want nil", result)
				}
				return
			}
			if len(result) != len(tt.expected) {
				t.Errorf("getStrList() len = %d, want %d", len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("getStrList()[%d] = %q, want %q", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestGetStrMap(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		key      string
		expected map[string]string
	}{
		{
			"normal map",
			map[string]interface{}{"key": map[string]interface{}{"a": "1", "b": "2"}},
			"key",
			map[string]string{"a": "1", "b": "2"},
		},
		{
			"empty map returns nil",
			map[string]interface{}{"key": map[string]interface{}{}},
			"key",
			nil,
		},
		{
			"missing key returns nil",
			map[string]interface{}{},
			"key",
			nil,
		},
		{
			"nil value returns nil",
			map[string]interface{}{"key": nil},
			"key",
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStrMap(tt.data, tt.key)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("getStrMap() = %v, want nil", result)
				}
				return
			}
			if len(result) != len(tt.expected) {
				t.Errorf("getStrMap() len = %d, want %d", len(result), len(tt.expected))
				return
			}
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("getStrMap()[%q] = %q, want %q", k, result[k], v)
				}
			}
		})
	}
}
