package handlers

import (
	"fmt"
	"net"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/types"
	"github.com/codecrafters-io/redis-starter-go/resp"
)

func ReplConf(conn net.Conn, args []string, serverState *types.ServerState) {
	// If the command is REPLCONF listening-port, add the connection to the list of replica connections
	if len(args) >= 2 && args[0] == "listening-port" {
		serverState.Replicas = append(serverState.Replicas, types.Replica{
			Conn:              conn,
			BytesAcknowledged: 0,
		})
		sendOk(conn)
		return
	}

	// If the command is REPLCONF capa, send OK
	if len(args) >= 2 && args[0] == "capa" {
		sendOk(conn)
		return
	}

	// If the command is REPLCONF GETACK *, send an ACK back to the master
	if len(args) >= 2 && args[0] == "GETACK" && args[1] == "*" {
		sendAck(conn, serverState.AckOffset)
		return
	}

	// If the command is REPLCONF ACK <bytes>, update the acknowledgment offset
	if len(args) >= 2 && args[0] == "ACK" {
		bytesOffset, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Printf("Error converting bytes offset to integer: %v\n", err)
			return
		}

		for ind, replica := range serverState.Replicas {
			if replica.Conn == conn {
				serverState.Replicas[ind].BytesAcknowledged = bytesOffset
				fmt.Printf("Bytes acknowledged by replica (%s) updated: %d\n", replica.Conn.RemoteAddr().String(), bytesOffset)
				return
			}
		}

		fmt.Printf("Replica connection not found to update bytes: %s\n", conn.RemoteAddr().String())
		return
	}

	fmt.Printf("Unknown REPLCONF command: %s\n", args)
}

func sendAck(conn net.Conn, bytesOffset int) {
	respHandler := resp.RESPHandler{}
	bytes, err := respHandler.Array.Encode([]string{"REPLCONF", "ACK", fmt.Sprintf("%d", bytesOffset)})
	if err != nil {
		fmt.Println("Failed to encode ACK response: ", err)
		return
	}
	_, err = conn.Write(bytes)
	if err != nil {
		fmt.Println("Failed to write response ACK response to master: ", err)
	}
}

func sendOk(conn net.Conn) {
	respHandler := resp.RESPHandler{}
	bytes, err := respHandler.String.Encode("OK")
	if err != nil {
		fmt.Println("Failed to encode OK response: ", err)
		return
	}
	_, err = conn.Write(bytes)
	if err != nil {
		fmt.Println("Failed to write OK response to master: ", err)
	}
}
