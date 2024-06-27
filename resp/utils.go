package resp

import (
	"fmt"
	"strconv"
)

// Parses the length of the string from the byte slice and returns the length and the remaining byte slice
func parseLen(b []byte) (int, []byte, error) {
	lenStr := ""
	for i := 0; i < len(b); i++ {
		if b[i] == '\r' {
			break
		}
		lenStr += string(b[i])
	}

	n, err := strconv.Atoi(lenStr)
	if err != nil {
		return 0, nil, fmt.Errorf("cannot parse number from string %s: %v", lenStr, err)
	}

	return n, b[len(lenStr):], nil
}

// Checks if the first two bytes of the byte slice are '\r\n'
// and returns the remaining byte slice
func parseCRLF(b []byte) ([]byte, error) {
	if b[0] != '\r' || b[1] != '\n' {
		return nil, fmt.Errorf("expected the next two bytes to be \\r\\n, got %q", b[0:2])
	}

	return b[2:], nil
}