package graphql

// TODO - rename
func strPtr(val string) *string {
	if val == "" {
		return nil
	}

	return &val
}

// TODO - rename
func intPtr(val int32) *int32 { return &val }

// TODO - rename
func boolPtr(val bool) *bool { return &val }

// TODO - rename
func int32Ptr(val *int) *int32 {
	if val == nil {
		return nil
	}

	v := int32(*val)
	return &v
}

// TODO - rename
func strDefault(val *string, defaultValue string) string {
	if val != nil {
		return *val
	}
	return defaultValue
}

// TODO - rename
func int32Default(val *int32, defaultValue int) int {
	if val != nil {
		return int(*val)
	}
	return defaultValue
}

// TODO - rename
func boolDefault(val *bool, defaultValue bool) bool {
	return (val != nil && *val) || defaultValue
}
