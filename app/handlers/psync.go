package handlers

import (
	"encoding/hex"
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/resp"
)

var emptyRDBFileHex string = "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2"

func Psync(conn net.Conn, replID string, offset int) {
	respHandler := resp.RESPHandler{}
	// Send the full resync message
	bytes, err := respHandler.String.Encode(fmt.Sprintf("FULLRESYNC %s %d", replID, offset))
	if err != nil {
		fmt.Println("Failed to encode response", err)
		return
	}
	conn.Write(bytes)

	// Send an empty RDB file
	emptyRBDFileBytes, err := hex.DecodeString(emptyRDBFileHex)
	if err != nil {
		fmt.Println("Failed to decode hex", err)
		return
	}

	message := []byte("$")
	message = append(message, []byte(fmt.Sprintf("%d", len(emptyRBDFileBytes)))...)
	message = append(message, []byte("\r\n")...)
	message = append(message, emptyRBDFileBytes...)
	conn.Write(message)
}