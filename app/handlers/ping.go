package handlers

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/resp"
)

func Ping(con net.Conn, isMasterCommand bool) {
	if isMasterCommand {
		// If the PING command is from the master, we should not reply
		// The master is only sending PING to check if the replica is alive
		return
	}

	res, err := resp.RESPHandler{}.String.Encode("PONG")
	if err != nil {
		fmt.Printf("Error encoding response: %s\n", err)
		return
	}

	con.Write(res)
}
