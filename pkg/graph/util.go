package graph

import (
	"encoding/json"
	"strings"

	"k8s.io/klog/v2"
)

// Contains verifies if a string element is being contained in a list of strings
func Contains(element string, elements []string) bool {
	for _, e := range elements {
		if e == element {
			return true
		}
	}

	return false
}

// ToJSON returns a json string
func ToJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		klog.Warningf("could not convert to JSON. Error: %s", err)
		return ""
	}

	return string(b)
}

// GetPrettyString returns a string without dashes
func GetPrettyString(ugly string) string {
	return strings.ReplaceAll(ugly, "-", "")
}
