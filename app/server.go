package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/handlers"
	"github.com/codecrafters-io/redis-starter-go/app/types"
	"github.com/codecrafters-io/redis-starter-go/resp"
)

func main() {

	agrs := GetArgs()
	serverState := GetServerState(&agrs)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", serverState.Port))
	if err != nil {
		fmt.Println("Failed to bind the port ",serverState.Port)
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			os.Exit(1)
		}
		fmt.Printf("Accepted connection from %s\n", conn.LocalAddr().String())

		go handleConnection(conn, serverState, false)
	}
}

func handleConnection(conn net.Conn, serverState *types.ServerState, isMasterConnection bool) {
	defer conn.Close()

	for {
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Connection closed by client")
			}
			fmt.Println("Error reading:", err)
			break
		}

		fmt.Printf("Received %d bytes: %q\n", n, buffer[:n])
		handleCommand(buffer[:n], conn, serverState, isMasterConnection)
	}
}

func handleCommand(buffer []byte, conn net.Conn, state *types.ServerState, isMasterCommand bool) {
	respHandler := resp.RESPHandler{}

	arr, next, err := respHandler.Array.Decode(buffer)
	if err != nil {
		fmt.Println("Error decoding RESP array:", err)
		return
	}
	buffer = buffer[:len(buffer)-len(next)]

	fmt.Println("Command received: ", arr)

	switch strings.ToUpper(arr[0]) {
	case "PING":
		handlers.Ping(conn, false)

	case "ECHO":
		handlers.Echo(conn, arr[1])

	case "SET":
		toReply := !isMasterCommand
		handlers.Set(conn, state, toReply, arr[1:]...)
		if state.Role == "master" {
			state.BytesSent += len(buffer)
			streamToReplicas(state.Replicas, buffer)
		}

	case "GET":
		handlers.Get(conn, &state.DB, &state.DBMutex, arr[1])

	case "INFO":
		handlers.Info(conn, state)

	case "REPLCONF":
		handlers.ReplConf(conn, arr[1:], state)

	case "PSYNC":
		handlers.Psync(conn, state.MasterReplID, state.MasterReplOffset)

	default:
		fmt.Println("Unknown command: ", arr[0])
	}

	// If this was a command from master, update the acknowledgment offset
	if isMasterCommand {
		state.AckOffset += len(buffer)
	}

	if len(next) > 0 {
		handleCommand(next, conn, state, isMasterCommand)
	}
}

