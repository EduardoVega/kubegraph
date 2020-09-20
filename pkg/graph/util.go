package graph

import (
	"encoding/json"
	"strings"

	"k8s.io/klog/v2"
)

// Contains returns true if a string is being contained in a list of strings
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
		klog.V(4).Infof("%v can not be convert to JSON, Error: '%s'", v, err)
		return ""
	}

	return string(b)
}

// GetPrettyString returns a string without dashes
func GetPrettyString(ugly string) string {
	return strings.ReplaceAll(ugly, "-", "")
}
