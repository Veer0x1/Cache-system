package handlers

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/types"
	"github.com/codecrafters-io/redis-starter-go/resp"
)

func Info(con net.Conn, serverInfo *types.ServerState) {
	replicationInfo := fmt.Sprintf("role:%s", serverInfo.Role)
	replicationInfo += fmt.Sprintf("\nmaster_replid:%s", serverInfo.MasterReplID)
	replicationInfo += fmt.Sprintf("\nmaster_repl_offset:%d", serverInfo.MasterReplOffset)

	res,err := resp.RESPHandler{}.BulkString.Encode(replicationInfo)
	if err != nil {
		fmt.Println("Error encoding response: ", err)
		return
	}
	con.Write(res)
}