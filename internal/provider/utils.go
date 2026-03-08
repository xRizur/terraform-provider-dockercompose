package provider

// getStr safely extracts a string value from a map.
func getStr(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// getBool safely extracts a bool value from a map.
func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok && v != nil {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// getBoolPtr returns a *bool. Returns nil for false/unset (omitted from YAML).
func getBoolPtr(m map[string]interface{}, key string) *bool {
	if v, ok := m[key]; ok && v != nil {
		if b, ok := v.(bool); ok && b {
			return &b
		}
	}
	return nil
}

// getIntPtr returns a *int. Returns nil for 0/unset (omitted from YAML).
func getIntPtr(m map[string]interface{}, key string) *int {
	if v, ok := m[key]; ok && v != nil {
		if i, ok := v.(int); ok && i != 0 {
			return &i
		}
	}
	return nil
}

// getStrList safely extracts a string slice. Returns nil if empty (omitted from YAML).
func getStrList(m map[string]interface{}, key string) []string {
	if v, ok := m[key]; ok && v != nil {
		if raw, ok := v.([]interface{}); ok && len(raw) > 0 {
			result := make([]string, 0, len(raw))
			for _, item := range raw {
				if s, ok := item.(string); ok && s != "" {
					result = append(result, s)
				}
			}
			if len(result) > 0 {
				return result
			}
		}
	}
	return nil
}

// getStrMap safely extracts a string map. Returns nil if empty (omitted from YAML).
func getStrMap(m map[string]interface{}, key string) map[string]string {
	if v, ok := m[key]; ok && v != nil {
		if raw, ok := v.(map[string]interface{}); ok && len(raw) > 0 {
			result := make(map[string]string, len(raw))
			for k, val := range raw {
				if s, ok := val.(string); ok {
					result[k] = s
				}
			}
			if len(result) > 0 {
				return result
			}
		}
	}
	return nil
}
