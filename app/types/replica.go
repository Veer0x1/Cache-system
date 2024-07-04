package types

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/resp"
)

type Replica struct {
	Conn              net.Conn
	BytesAcknowledged int
}

func (r *Replica) GetAcknowlegment() error {
	respHandler := resp.RESPHandler{}

	// Send the GETACK command to the replica
	messageBytes,err := respHandler.Array.Encode([]string{"REPLCONF", "GETACK", "*"})
	if err != nil {
		return fmt.Errorf("failed to encode GETACK command: %v", err)
	}
	_, err = r.Conn.Write(messageBytes)
	if err != nil {
		return fmt.Errorf("failed to write GETACK command to replica connection: %v", err)
	}

	return nil
}