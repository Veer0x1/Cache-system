package resp

import (
	"bytes"
	"fmt"
	"strconv"
)

type RESPCodec struct{}

func (codec *RESPCodec) Encode(parts []interface{}) ([]byte, error) {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("*%d\r\n", len(parts)))
	for _, part := range parts {
		switch t := part.(type) {
		case string:
			buffer.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(t), t))
		case int:
			buffer.WriteString(fmt.Sprintf(":%d\r\n", t))
		case []byte:
			buffer.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(t), t))
		case nil:
			buffer.WriteString("$-1\r\n")
		case []string:
			buffer.WriteString(fmt.Sprintf("*%d\r\n", len(t)))
			for _, str := range t {
				buffer.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(str), str))
			}
		default:
			return nil, fmt.Errorf("unsupported type: %T", part)
		}
	}
	return buffer.Bytes(), nil
}

func (codec *RESPCodec) Decode(message []byte) ([]interface{}, error) {
	lines := bytes.Split(message, []byte("\r\n"))
	var result []interface{}

	for i := 1; i < len(lines)-2; i++ { // -2 to avoid the last CRLF split parts
		line := lines[i]
		if len(line) == 0 {
			continue
		}
		switch line[0] {
		case '+':
			result = append(result, string(line[1:]))
		case '-':
			result = append(result, fmt.Errorf("%s", line[1:]))
		case ':':
			intVal, err := strconv.Atoi(string(line[1:]))
			if err != nil {
				return nil, fmt.Errorf("invalid integer value: %v", err)
			}
			result = append(result, intVal)
		case '$':
			length, _ := strconv.Atoi(string(line[1:]))
			if length == -1 {
				result = append(result, nil)
				continue
			}
			i++
			result = append(result, lines[i])
		case '*':
			// For simplicity, nested arrays are not handled in this example.
			return nil, fmt.Errorf("nested arrays are not supported")
		default:
			return nil, fmt.Errorf("unknown type: %v", line[0])
		}
	}

	return result, nil
}

func (codec *RESPCodec) OK()([]byte){
	return []byte("+OK\r\n")
}

func (codec *RESPCodec) ErrorResponse(message string) ([]byte){
	return []byte(fmt.Sprintf("-Error: %s\r\n",message))
}
func (codec *RESPCodec) EncodeBulkString(value string) ([]byte) {
    return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(value), value))
}