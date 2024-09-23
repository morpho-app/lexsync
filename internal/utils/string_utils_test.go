package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input      string
		pascalCase bool
		expected   string
	}{
		{"profile_view_basic", true, "ProfileViewBasic"},
		{"profile_view_basic", false, "profileViewBasic"},
		{"allow-incoming", true, "AllowIncoming"},
		{"allow-incoming", false, "allowIncoming"},
		{"#known_followers", true, "KnownFollowers"},
		{"#known_followers", false, "knownFollowers"},
		{"API_response", true, "APIResponse"},
		{"API_response", false, "apiResponse"},
		{"simpleTest", true, "SimpleTest"},
		{"simpleTest", false, "simpleTest"},
		{"multiple#delimiters-test_case", true, "MultipleDelimitersTestCase"},
		{"multiple#delimiters-test_case", false, "multipleDelimitersTestCase"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			output := ToCamelCase(tc.input, tc.pascalCase)
			assert.Equal(t, tc.expected, output, "ToCamelCase(%s) should be %s", tc.input, tc.expected)
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ProfileViewBasic", "profile_view_basic"},
		{"AllowIncoming", "allow_incoming"},
		{"KnownFollowers", "known_followers"},
		{"APIResponse", "api_response"},
		{"SimpleTest", "simple_test"},
		{"MultipleDelimitersTestCase", "multiple_delimiters_test_case"},
		{"", ""},
	}

	for _, test := range tests {
		result := ToSnakeCase(test.input)
		assert.Equal(t, test.expected, result, "ToSnakeCase(%s) should be %s", test.input, test.expected)
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  Hello,   World!  ", "Hello, World!"},
		{"NoExtraSpaces", "NoExtraSpaces"},
		{"   ", ""},
		{"Multiple     Spaces Between Words", "Multiple Spaces Between Words"},
	}

	for _, test := range tests {
		result := SanitizeString(test.input)
		assert.Equal(t, test.expected, result, "SanitizeString(%s) should be %s", test.input, test.expected)
	}
}

func TestReverse(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello, World!", "!dlroW ,olleH"},
		{"", ""},
		{"A", "A"},
		{"Go语言", "言语oG"},
		{"12345", "54321"},
		{"😊👍", "👍😊"},
	}

	for _, test := range tests {
		result := Reverse(test.input)
		assert.Equal(t, test.expected, result, "Reverse(%s) should be %s", test.input, test.expected)
	}
}

func TestContains(t *testing.T) {
	assert.True(t, Contains("The quick brown fox", "fox"))
	assert.False(t, Contains("The quick brown fox", "dog"))
	assert.True(t, Contains("Golang is awesome", "Golang"))
	assert.False(t, Contains("", "anything"))
}

func TestReplaceAll(t *testing.T) {
	input := "Go is awesome. Go is fast."
	expected := "Golang is awesome. Golang is fast."
	result := ReplaceAll(input, "Go", "Golang")
	assert.Equal(t, expected, result, "ReplaceAll should replace all instances of Go with Golang")
}

func TestIsEmpty(t *testing.T) {
	assert.True(t, IsEmpty("   "))
	assert.False(t, IsEmpty("Hello"))
	assert.True(t, IsEmpty(""))
	assert.False(t, IsEmpty("  A  "))
}
