package resp

import (
	"bytes"
	"fmt"
)

type array struct{}

func (array) Encode(data []string) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("*%d\r\n", len(data)))

	bulkString := bulkString{}
	for _, str := range data {
		strBytes, err := bulkString.Encode(str)
		if err != nil {
			return nil, err
		}
		buf.Write(strBytes)
	}

	return buf.Bytes(), nil
}

func (array) Decode(b []byte) ([]string, []byte, error) {
	if b[0] != '*' {
		return nil, b, fmt.Errorf("invalid format for array: expected the first byte to be '*', got '%q'", b[0])
	}

	n, b, err := parseLen(b[1:])
	if err != nil {
		return nil, b, fmt.Errorf("invalid format for array: %v", err)
	}

	b, err = parseCRLF(b)
	if err != nil {
		return nil, b, fmt.Errorf("invalid format for array: %v", err)
	}

	arr := make([]string, n)
	bulkString := bulkString{}

	str := ""
	for i := 0; i < n; i++ {
		str, b, err = bulkString.Decode(b)
		if err != nil {
			return nil, b, fmt.Errorf("invalid format for array: %v", err)
		}
		arr[i] = str
	}

	return arr, b, nil
}