package resp_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/resp"
)

func TestArrayEncode(t *testing.T) {
	tests := []struct {
		testCaseName string
		input        []string
		expected     []byte
		expectError  bool
	}{
		{
			testCaseName: "Empty Array",
			input:        []string{},
			expected:     []byte("*0\r\n"),
			expectError:  false,
		},
		{
			testCaseName: "Array with 1 element",
			input:        []string{"PING"},
			expected:     []byte("*1\r\n$4\r\nPING\r\n"),
			expectError:  false,
		},
		{
			testCaseName: "Array with 2 elements",
			input:        []string{"PING", "PONG"},
			expected:     []byte("*2\r\n$4\r\nPING\r\n$4\r\nPONG\r\n"),
			expectError:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testCaseName, func(t *testing.T) {
			handlers := resp.RESPHandler{}
			res, err := handlers.Array.Encode(tc.input)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !bytes.Equal(res, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, res)
			}
		})
	}

}

func TestDecode(t *testing.T) {
	tests := []struct {
		testCaseName string
		input        []byte
		expected     []string
		expectError  bool
	}{
		{
			testCaseName: "Empty Array",
			input:        []byte("*0\r\n"),
			expected:     []string{},
			expectError:  false,
		},
		{
			testCaseName: "Array with 1 element",
			input:        []byte("*1\r\n$4\r\nPING\r\n"),
			expected:     []string{"PING"},
			expectError:  false,
		},
		{
			testCaseName: "Array with 2 elements",
			input:        []byte("*2\r\n$4\r\nPING\r\n$4\r\nPONG\r\n"),
			expected:     []string{"PING", "PONG"},
			expectError:  false,
		},
		{
			testCaseName: "Invalid Format",
			input:        []byte("Invalid"),
			expected:     nil,
			expectError:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testCaseName, func(t *testing.T) {
			handler := resp.RESPHandler{}
			res, _, err := handler.Array.Decode(tc.input)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
			}

			if !reflect.DeepEqual(res, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, res)
			}
		})
	}
}
