package graph

import (
	"testing"
)

func TestContains(t *testing.T) {

	tests := []struct {
		Element     string
		ElementList []string
		Expected    bool
	}{
		{
			"apple",
			[]string{"orange", "banana", "apple"},
			true,
		},
		{
			"watermelon",
			[]string{"orange", "banana", "apple"},
			false,
		},
	}

	for _, test := range tests {
		r := Contains(test.Element, test.ElementList)

		if r != test.Expected {
			t.Errorf("Returned result was incorrect, got: %t want: %t", r, test.Expected)
		}
	}
}

func TestToJSON(t *testing.T) {

	tests := []struct {
		Data     interface{}
		Expected string
	}{
		{
			[]struct {
				Kind   string
				Name   string
				Labels map[string]string
			}{
				{
					"Service",
					"service-foo",
					map[string]string{
						"app": "foo",
					},
				},
			},
			"[{\"Kind\":\"Service\",\"Name\":\"service-foo\",\"Labels\":{\"app\":\"foo\"}}]",
		},
		{
			make(chan int),
			"",
		},
	}

	for _, test := range tests {
		r := ToJSON(test.Data)

		if r != test.Expected {
			t.Errorf("Returned result was incorrect, got: %s want: %s", r, test.Expected)
		}
	}

}

func TestGetPrettyString(t *testing.T) {
	tests := []struct {
		UglyString string
		Expected   string
	}{
		{
			"-Rl-Uj2-6j-X-n0g",
			"RlUj26jXn0g",
		},
		{
			"me-Pl-va-6V---L0S",
			"mePlva6VL0S",
		},
	}

	for _, test := range tests {
		r := GetPrettyString(test.UglyString)

		if r != test.Expected {
			t.Errorf("Returned result was incorrect, got: %s want: %s", r, test.Expected)
		}
	}
}
