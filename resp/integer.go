package resp

import (
	"fmt"
	"strconv"
)

type integer struct{}

func (integer) Encode(data int) ([]byte, error) {
	return []byte(":" + strconv.Itoa(data) + "\r\n"), nil
}

func (integer) Decode(b []byte) (int, []byte, error) {
	if b[0] != ':' {
		return 0, b, fmt.Errorf("invalid format for integer: expected the first byte to be ':', got '%q'", b[0])
	}

	n, b, err := parseLen(b[1:])
	if err != nil {
		return 0, b, err
	}

	b, err = parseCRLF(b)
	if err != nil {
		return 0, b, err
	}

	return n, b, nil
}