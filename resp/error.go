package resp

type errorString struct{}

// EncodeError encodes an error message to RESP format.
func (errorString) Encode(errMsg string) ([]byte, error) {
	return []byte("-" + errMsg + "\r\n"), nil
}