package main

import (
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/types"
)

func GetServerState(args *Args) *types.ServerState {
	state := types.ServerState{
		DB:   map[string]types.DBItem{},
		Port: args.port,

		Role:             "master",
		MasterReplID:     "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb",
		MasterReplOffset: 0,

		// DBDir:      args.dir,
		// DBFilename: args.dbfilename,
	}

	if args.replicaof != "" {
		state.Role = "slave"
		state.MasterHost = strings.Split(args.replicaof, " ")[0]
		state.MasterPort = strings.Split(args.replicaof, " ")[1]
		state.MasterReplID = "?"
		state.MasterReplOffset = -1
		handshakeWithMaster(&state)
	}

	// if args.dir != "" && args.dbfilename != "" {
	// 	file.InitialiseDB(&state, args.dbfilename, args.dir)
	// }

	// Initialise the streams map
	state.Streams = map[string][]types.StreamEntry{}

	return &state
}
