package resp_test

import (
	"reflect"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/resp"
)

func TestEncodeError(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected []byte
	}{
		{"Simple error", "ERR simple error", []byte("-ERR simple error\r\n")},
		{"Empty error", "", []byte("-\r\n")},
		{"Complex error", "ERR complex error with details", []byte("-ERR complex error with details\r\n")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := resp.RESPHandler{}
			result, err := handler.Error.Encode(tc.errMsg)
			if err != nil {
				t.Errorf("EncodeError(%q) returned an error: %v", tc.errMsg, err)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("EncodeError(%q) = %v, want %v", tc.errMsg, result, tc.expected)
			}
		})
	}
}
