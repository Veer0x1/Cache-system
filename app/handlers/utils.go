package handlers

import (
	"github.com/codecrafters-io/redis-starter-go/app/types"
)

func checkIfKeyExists(key string, server *types.ServerState) bool {
	_, okString := server.DB[key]
	_, okStream := server.Streams[key]

	return okString || okStream
}