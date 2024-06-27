package resp_test

import (
	"testing"
	"github.com/codecrafters-io/redis-starter-go/resp"
)

func TestSimpleStringEncode(t *testing.T) {
	tests := []struct {
		testCaseName string
		input        string
		expected     []byte
		expectError   bool
	}{
		{
			testCaseName: "Simple string",
			input:        "Hello",
			expected:     []byte("+Hello\r\n"),
			expectError:   false,
		},
		{
			testCaseName: "Empty string",
			input:        "",
			expected:     []byte("+\r\n"),
			expectError:   false,
		},
		{
			testCaseName: "String with spaces",
			input:        "Hello World",
			expected:     []byte("+Hello World\r\n"),
			expectError:   false,
		},
		{
			testCaseName: "String with CRLF",
			input:        "Hello\r\nWorld",
			expected:     nil,
			expectError:   true,
		},
		{
			testCaseName: "String with CR",
			input:        "Hello\rWorld",
			expected:     nil,
			expectError:   true,
		},
		{
			testCaseName: "String with LF",
			input:        "Hello\nWorld",
			expected:     nil,
			expectError:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testCaseName, func(t *testing.T) {
			handler := resp.RESPHandler{}
			res, err := handler.String.Encode(tc.input)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			} else {
				if err != nil {
					t.Errorf("Expected nil, got %v", err)
				}
			}

			if string(res) != string(tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, res)
			}
		})
	}
}

func TestSimpleStringDecode(t *testing.T) {
	tests := []struct {
		testCaseName string
		input        []byte
		expected     string
		expectError   bool
	}{
		{
			testCaseName: "Simple string",
			input:        []byte("+Hello\r\n"),
			expected:     "Hello",
			expectError:   false,
		},
		{
			testCaseName: "Empty string",
			input:        []byte("+\r\n"),
			expected:     "",
			expectError:   false,
		},
		{
			testCaseName: "String with spaces",
			input:        []byte("+Hello World\r\n"),
			expected:     "Hello World",
			expectError:   false,
		},
		{
			testCaseName: "Invalid string",
			input:        []byte("Hello\r\nWorld"),
			expected:     "",
			expectError:   true,
		},
		{
			testCaseName: "Invalind string with size less than 3",
			input:        []byte("+\r"),
			expected:     "",
			expectError:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testCaseName, func(t *testing.T) {
			handler := resp.RESPHandler{}
			res,_, err := handler.String.Decode(tc.input)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			} else {
				if err != nil {
					t.Errorf("Expected nil, got %v", err)
				}
			}

			if res != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, res)
			}
		})
	}
}