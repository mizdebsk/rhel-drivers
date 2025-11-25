package dnf

import "testing"

func TestParseNameFromNVRA(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "SimpleCase",
			input:    "example-package-1.2.3-4.el8.src.rpm",
			expected: "example-package",
		},
		{
			name:     "MissingRelease",
			input:    "single-dash-1.0.src.rpm",
			expected: "single",
		},
		{
			name:     "NoDashesAtAll",
			input:    "foobar",
			expected: "",
		},
		{
			name:     "EmptyString",
			input:    "",
			expected: "",
		},
		{
			name:     "ValidMultipleDashes",
			input:    "this-is-a-complex-name-0.0.1-2.fc35.src.rpm",
			expected: "this-is-a-complex-name",
		},
		{
			name:     "LeadingDash",
			input:    "-leading-dash-1.0",
			expected: "-leading",
		},
		{
			name:     "TrailingDash",
			input:    "trailing-dash-1.0-",
			expected: "trailing-dash",
		},
		{
			name:     "TwoTrailingDashes",
			input:    "C--",
			expected: "C",
		},
		{
			name:     "LeadingSpaces",
			input:    "  spaced-package-3.2.1-4.el7.rpm",
			expected: "  spaced-package",
		},
		{
			name:     "VeryShortName",
			input:    "x-1.0-3.fc33.src.rpm",
			expected: "x",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseNameFromNVRA(tt.input)
			if result != tt.expected {
				t.Errorf("parseNameFromNVRA(%q) = %q; expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
