package resp

import (
	"errors"
	"fmt"
	"strings"
)

type simpleString struct{}

func (simpleString) Encode(s string) ([]byte, error) {
	if strings.ContainsAny(s, "\r") {
		return nil, errors.New("simple string cannot contain CR")
	}
	if strings.ContainsAny(s, "\n") {
		return nil, errors.New("simple string cannot contain LF")
	}
	if strings.ContainsAny(s, "\r\n") {
		return nil, errors.New("simple string cannot contain CRLF")
	}
	return []byte("+" + s + "\r\n"), nil
}

func (simpleString) Decode(data []byte) (string, []byte, error) {
	if len(data) < 3 {
		return "", data, errors.New("length of data is less than 3, and a valid simple string should atleast contain 3 character +, CR and LF")
	}
	if data[0] != '+' {
		return "", data, errors.New("The first character of the data should be +, but got " + string(data[0]))
	}
	for i := 1; i < len(data); i++ {
		if data[i] == '\r' {
			if i+1 < len(data) && data[i+1] == '\n' {
				return string(data[1:i]), data[i+2:], nil
			}
			return "", data, fmt.Errorf("invalid format for simple string: expected the last two bytes to be \\r\\n, got %q", data[i:])
		}
	}

	return "", data, fmt.Errorf("invalid format for simple string: expected the last two bytes to be \\r\\n, got %q", data[len(data)-1:])
}
