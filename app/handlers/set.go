package handlers

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/types"
	"github.com/codecrafters-io/redis-starter-go/resp"
)

func Set(con net.Conn, server *types.ServerState, shouldReply bool, arr ...string) {
	if len(arr) < 2 {
		fmt.Println("Error: SET requires at least 2 arguments, which are the KEY and the VALUE")
		return
	}

	expiry := int64(-1)
	if len(arr) > 2 {
		if arr[2] != "px" {
			fmt.Println("Error: SET only supports px as a third argument")
			return
		}
		n, err := strconv.ParseInt(arr[3], 10, 64)
		if err != nil {
			fmt.Println("Error: EX argument must be an integer")
			return
		}
		expiry = time.Now().UnixMilli() + n
	}

	server.DBMutex.Lock()
	defer server.DBMutex.Unlock()

	key := arr[0]
	value := arr[1]

	if checkIfKeyExists(key, server) {
		fmt.Printf("Key %s already exists for a key-value pair or a stream\n", key)
		return
	}

	server.DB[key] = types.DBItem{Value: value, Expiry: expiry}

	if shouldReply {
		res, err := resp.RESPHandler{}.String.Encode("OK")
		if err != nil {
			fmt.Printf("Error encoding response: %s\n", err)
			return
		}
		con.Write(res)
	}
}