package resp

import (
	"bytes"
	"errors"
	"fmt"
)

type bulkString struct{}

func (bulkString) Encode(s string) ([]byte, error) {
	if s == "" {
		return []byte("$-1\r\n"), nil
	}
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(s), s)), nil
}

func (bulkString) Decode(data []byte) (string, []byte, error) {
	if bytes.Equal(data, []byte("$-1\r\n")) {
		return "",data, nil
	}

	if !bytes.HasPrefix(data, []byte("$")) {
		return "",data, errors.New("invalid format: does not start with '$'")
	}

	n, data, err := parseLen(data[1:])
	if err != nil {
		return "", data, fmt.Errorf("invalid format for bulk string: %v", err)
	}

	data, err = parseCRLF(data)
	if err != nil {
		return "", data, fmt.Errorf("invalid format for bulk string: %v", err)
	}

	if len(data) < n+2 {
		return "", data, fmt.Errorf("invalid format for bulk string: expected length of string to be atleast %d, got %d", n+2, len(data))
	}

	str := string(data[:n])
	data = data[n:]

	data, err = parseCRLF(data)
	if err != nil {
		return "", data, fmt.Errorf("invalid format for bulk string: %v", err)
	}

	return str, data, nil
}
