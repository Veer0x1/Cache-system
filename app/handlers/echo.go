package handlers

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/resp"
)

func Echo(con net.Conn, message string) {
	res,err := resp.RESPHandler{}.BulkString.Encode(message)
	if err != nil {
		fmt.Println("Error encoding response: ", err)
		return
	}
	con.Write(res)
}