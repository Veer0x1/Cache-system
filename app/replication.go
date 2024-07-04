package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/types"
	"github.com/codecrafters-io/redis-starter-go/resp"
)

func sendAndAssertReply(conn net.Conn, messageArr []string, expectedMsg string, respHandler resp.RESPHandler) error {
	respHandler = resp.RESPHandler{}
	bytes, err := respHandler.Array.Encode(messageArr)
	if err != nil {
		return fmt.Errorf("failed to encode message: %s", err)
	}

	conn.Write(bytes)

	resp := make([]byte, 1024)
	n, _ := conn.Read(resp)
	msg, remain, err := respHandler.String.Decode(resp[:n])
	if err != nil {
		return fmt.Errorf("failed to decode response: %s", err)
	}
	if msg != expectedMsg {
		return fmt.Errorf("expected +OK, got %s", string(resp[:n]))
	}
	if len(remain) > 0 {
		return fmt.Errorf("unexpected remaining (buffered) bytes: %q", remain)
	}

	return nil
}

func sendAndGetRBDFile(conn net.Conn, messageArr []string, respHandler resp.RESPHandler, state *types.ServerState) (string, []byte, error) {
	respHandler = resp.RESPHandler{}
	bytes,err := respHandler.Array.Encode(messageArr)
	if err != nil {
		return "", nil, fmt.Errorf("failed to encode message: %s", err)
	}

	conn.Write(bytes)

	// Get the initial PSYNC response
	resp := make([]byte, 1024)
	n, err := conn.Read(resp)
	if err != nil {
		return "", nil, fmt.Errorf("failed to recieve message from master: %s", err)
	}
	psyncResp, rdbBytes, err := respHandler.String.Decode(resp[:n])
	if err != nil {
		return "", nil, fmt.Errorf("failed to decode response: %s", err)
	}

	// Parse the PSYNC response
	responseParts := strings.Split(psyncResp, " ")
	if len(responseParts) != 3 {
		return "", nil, fmt.Errorf("expected 3 parts in PSYNC response, got %d", len(responseParts))
	}
	if responseParts[0] != "FULLRESYNC" {
		return "", nil, fmt.Errorf("expected FULLRESYNC in PSYNC response, got %s", responseParts[0])
	}
	state.MasterReplID = responseParts[1]
	portAsInt, err := strconv.Atoi(responseParts[2])
	if err != nil {
		return "", nil, fmt.Errorf("failed to convert port to int: %s", err)
	}
	state.MasterReplOffset = portAsInt

	if len(rdbBytes) == 0 {
		// Get the RDB file
		rdbBytes = make([]byte, 1024)
		n, err = conn.Read(rdbBytes)
		if err != nil {
			return "", nil, fmt.Errorf("failed to recieve message from master: %s", err)
		}
		rdbBytes = rdbBytes[:n]
	}
	fileContent, remainingBytes, err := parseFileContent(rdbBytes)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse RDB file: %s", err)
	}

	return fileContent, remainingBytes, nil
}

func parseFileContent(dataBytes []byte) (string, []byte, error) {
	if dataBytes[0] != '$' {
		return "", nil, fmt.Errorf("expected $ at the start of the datafile, got %s", string(dataBytes[0]))
	}

	// Parse the length of the data
	lenStr := ""
	for i := 1; i < len(dataBytes); i++ {
		if dataBytes[i] == '\r' {
			break
		}
		lenStr += string(dataBytes[i])
	}
	dataLen, err := strconv.Atoi(lenStr)
	if err != nil {
		return "", nil, fmt.Errorf("failed to convert data length to int: %s", err)
	}

	startIndex := len(lenStr) + 3 // $<len>\r\n
	if len(dataBytes) < startIndex+dataLen {
		return "", nil, fmt.Errorf("expected %d bytes of data, got %d", dataLen, len(dataBytes)-startIndex)
	}
	fileContentBytes := dataBytes[startIndex : startIndex+dataLen]
	remainingBytes := dataBytes[startIndex+dataLen:]

	return string(fileContentBytes), remainingBytes, nil
}

func handshakeWithMaster(server *types.ServerState) {
	masterConn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", server.MasterHost, server.MasterPort))
	if err != nil {
		fmt.Println("Failed to connect to master: ", err)
		return
	}

	respHandler := resp.RESPHandler{}

	// PING
	err = sendAndAssertReply(
		masterConn,
		[]string{"PING"},
		"PONG",
		respHandler,
	)
	if err != nil {
		fmt.Println("Failed to send PING to master: ", err)
		return
	}

	// REPLCONF listening-port <port>
	err = sendAndAssertReply(
		masterConn,
		[]string{"REPLCONF", "listening-port", fmt.Sprintf("%d", server.Port)},
		"OK",
		respHandler,
	)
	if err != nil {
		fmt.Println("Failed to send REPLCONF listening-port to master: ", err)
		return
	}

	// REPLCONF capa psync2
	err = sendAndAssertReply(
		masterConn,
		[]string{"REPLCONF", "capa", "psync2"},
		"OK",
		respHandler,
	)
	if err != nil {
		fmt.Println("Failed to send REPLCONF capa psync2 to master: ", err)
		return
	}

	// PSYNC <replicationid> <offset>
	rdbFile, remainingBytes, err := sendAndGetRBDFile(
		masterConn,
		[]string{"PSYNC", "?", fmt.Sprintf("%d", -1)},
		respHandler,
		server,
	)
	if err != nil {
		fmt.Println("Failed to send PSYNC to master: ", err)
		return
	}
	fmt.Printf("RDB File content: %q\n", rdbFile)

	// Since the handshake was successful, we can now set handle the master connection in a separate goroutine
	go handleConnection(masterConn, server, true)

	// If there are remaining bytes, handle them as a separate command
	if len(remainingBytes) > 0 {
		handleCommand(remainingBytes, masterConn, server, true)
	}
}

func streamToReplicas(replicas []types.Replica, buff []byte) {
	fmt.Printf("Streaming recieved command to %d replicas\n", len(replicas))
	for ind, r := range replicas {
		_, err := r.Conn.Write(buff)
		if err != nil {
			fmt.Printf("Failed to stream to replica %d: %s", ind+1, err.Error())
		}
	}
}
