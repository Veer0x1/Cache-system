package handlers

import (
	"net"
	"sync"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/types"
	"github.com/codecrafters-io/redis-starter-go/resp"
)

func Get(con net.Conn, db *map[string]types.DBItem, mutex *sync.Mutex, key string) {
	respHandler := resp.RESPHandler{}
	mutex.Lock()
	defer mutex.Unlock()

	value, ok := (*db)[key]

	if !ok {
		res := respHandler.Nil.Encode()
		con.Write(res)
		return
	}

	if value.Expiry == -1 || time.Now().UnixMilli() < value.Expiry {
		res,err := respHandler.BulkString.Encode(value.Value)
		if err != nil {
			res = respHandler.Nil.Encode()
		}
		con.Write(res)
		return
	}

	delete(*db, key)
	con.Write(respHandler.Nil.Encode())
}