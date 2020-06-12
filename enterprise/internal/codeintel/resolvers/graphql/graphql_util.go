package resolvers

import (
	"fmt"
)

var severities = map[int]string{
	1: "ERROR",
	2: "WARNING",
	3: "INFORMATION",
	4: "HINT",
}

func toSeverity(val int) (*string, error) {
	severity, ok := severities[val]
	if !ok {
		return nil, fmt.Errorf("unknown diagnostic severity %d", val)
	}

	return &severity, nil
}

// TODO - rename
func strPtr(val string) *string {
	if val == "" {
		return nil
	}

	return &val
}

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
