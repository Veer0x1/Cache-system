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

type KeyValuePair struct {
	Key   string
	Value interface{}
}

func (codec *RESPCodec) EncodeMultipleBulkStrings(pairs []KeyValuePair) ([]byte) {
	var infoString string
    for _, pair := range pairs {
        // Convert the value to a string, regardless of its actual type
        var valueStr string
        switch v := pair.Value.(type) {
        case string:
            valueStr = v
        case int, int64:
            valueStr = fmt.Sprintf("%d", v)
        default:
            // Handle other types as needed, or skip unsupported types
            continue
        }
        // Append the key-value pair to the info string, separated by a newline
        infoString += pair.Key + ":" + valueStr + "\n"
    }

    // Encode the entire info string as a bulk string
    return codec.EncodeBulkString(infoString)
}

func (codec *RESPCodec) EncodeCommand(command string, args []string) []byte {
	var respParts []string

	// Add the command and its arguments to the parts slice
	respParts = append(respParts, command)
	respParts = append(respParts, args...)

	// Start building the RESP command with the array prefix and the number of elements
	respString := fmt.Sprintf("*%d\r\n", len(respParts))

	// Encode each part of the command as a bulk string
	for _, part := range respParts {
		respString += fmt.Sprintf("$%d\r\n%s\r\n", len(part), part)
	}

	return []byte(respString)
}