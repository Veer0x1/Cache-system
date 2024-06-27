package resp_test

import (
	"fmt"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/resp"
)

func TestBulkStringEncode(t *testing.T) {
	tests := []struct {
		testCaseName string
		input        string
		expected     []byte
		expectError  bool
	}{
		{
			testCaseName: "Empty string",
			input:        "",
			expected:     []byte("$-1\r\n"),
			expectError:  false,
		},
		{
			testCaseName: "Non-empty string",
			input:        "Hello",
			expected:     []byte("$5\r\nHello\r\n"),
			expectError:  false,
		},
		{
			testCaseName: "String with special characters",
			input:        "!@#$%^&*()",
			expected:     []byte("$10\r\n!@#$%^&*()\r\n"),
			expectError:  false,
		},
		{
			testCaseName: "Long string",
			input:        fmt.Sprintf("%01024d", 1), // Generates a string of 1024 '1's
			expected:     []byte("$1024\r\n" + fmt.Sprintf("%01024d", 1) + "\r\n"),
			expectError:  false,
		},
		{
			testCaseName: "Unicode characters",
			input:        "こんにちは",                    // "Hello" in Japanese
			expected:     []byte("$15\r\nこんにちは\r\n"), // Length is byte length, not character count
			expectError:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testCaseName, func(t *testing.T) {
			handler := resp.RESPHandler{}
			res, err := handler.BulkString.Encode(tc.input)

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

func TestBulkStringDecode(t *testing.T) {
	tests := []struct {
		testCaseName string
		input        []byte
		expected     string
		expectError  bool
	}{
		{
			testCaseName: "Null bulk string",
			input:        []byte("$-1\r\n"),
			expected:     "",
			expectError:  false,
		},
		{
			testCaseName: "Valid bulk string",
			input:        []byte("$5\r\nHello\r\n"),
			expected:     "Hello",
			expectError:  false,
		},
		{
			testCaseName: "Invalid format - no dollar sign",
			input:        []byte("5\r\nHello\r\n"),
			expected:     "",
			expectError:  true,
		},
		{
			testCaseName: "Invalid format - no CRLF",
			input:        []byte("$5Hello"),
			expected:     "",
			expectError:  true,
		},
		{
			testCaseName: "Invalid length - not a number",
			input:        []byte("$x\r\nHello\r\n"),
			expected:     "",
			expectError:  true,
		},
		{
			testCaseName: "Content length does not match specified length",
			input:        []byte("$5\r\nHell\r\n"),
			expected:     "",
			expectError:  true,
		},
		{
			testCaseName: "Unicode characters",
			input:        []byte("$15\r\nこんにちは\r\n"),
			expected:     "こんにちは",
			expectError:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testCaseName, func(t *testing.T) {
			handler := resp.RESPHandler{}
			res, _, err := handler.BulkString.Decode(tc.input)

			if tc.expectError {
				if err == nil {
					t.Errorf("Test '%s' failed: expected error, got nil", tc.testCaseName)
				}
			} else {
				if err != nil {
					t.Errorf("Test '%s' failed: expected no error, got %v", tc.testCaseName, err)
				} else if res != tc.expected {
					t.Errorf("Test '%s' failed: expected %s, got %s", tc.testCaseName, tc.expected, res)
				}
			}
		})
	}
}
