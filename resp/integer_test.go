package resp_test

import (
	"reflect"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/resp"
)

func TestIntegerEncode(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected []byte
	}{
		{"Positive integer", 123, []byte(":123\r\n")},
		{"Negative integer", -456, []byte(":-456\r\n")},
		{"Zero", 0, []byte(":0\r\n")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := resp.RESPHandler{}
			res, err := handler.Integer.Encode(tc.input)
			if err != nil {
				t.Errorf("Encode(%d) returned an error: %v", tc.input, err)
			}
			if !reflect.DeepEqual(res, tc.expected) {
				t.Errorf("Encode(%d) = %v, want %v", tc.input, res, tc.expected)
			}
		})
	}
}

func TestIntegerDecode(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		expected    int
		expectError bool
	}{
		{"Positive integer", []byte(":123\r\n"), 123, false},
		{"Negative integer", []byte(":-456\r\n"), -456, false},
		{"Zero", []byte(":0\r\n"), 0, false},
		{"Invalid format", []byte("123\r\n"), 0, true},
		{"Non-integer", []byte(":abc\r\n"), 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := resp.RESPHandler{}
			res, _, err := handler.Integer.Decode(tc.input)
			if tc.expectError {
				if err == nil {
					t.Errorf("Decode(%v) expected an error, got nil", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("Decode(%v) returned an error: %v", tc.input, err)
				}
				if res != tc.expected {
					t.Errorf("Decode(%v) = %d, want %d", tc.input, res, tc.expected)
				}
			}
		})
	}
}
