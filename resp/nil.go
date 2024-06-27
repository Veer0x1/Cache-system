package resp

type nilString struct{}

func (nilString) Encode() []byte {
	return []byte("$-1\r\n")
}