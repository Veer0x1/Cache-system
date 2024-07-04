package main

import "flag"

type Args struct {
	port      int
	replicaof string
}

func GetArgs() Args {
	port := flag.Int("port", 6379, "server port")
	replicaof := flag.String("replicaof", "", "host and port of master server")
	flag.Parse()
	return Args{
		port:      *port,
		replicaof: *replicaof,
	}
}
