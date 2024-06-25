package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/resp"
)

var store = make(map[string]string) // In-memory key value

func startServer() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind the port 6379")
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(5 * time.Minute))

	for {
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading:", err)
			}
			break
		}

		commandLine := string(buffer[:n])
		lines := strings.Split(commandLine, "\r\n")

		if len(lines) > 0 && lines[0][0] == '*' {
			numElements, _ := strconv.Atoi(lines[0][1:])
			parts := make([]string, 0, numElements)

			for i := 1; i < len(lines) && len(parts) < numElements; {
				i++
				if i < len(lines) {
					parts = append(parts, lines[i])
					i++
				}
			}

			codec := resp.RESPCodec{}

			switch strings.ToUpper(parts[0]) {
			case "PING":
				conn.Write([]byte("+PONG\r\n"))
			case "ECHO":
				if len(parts) > 1 {
					message := strings.Join(parts[1:], " ")
					response := fmt.Sprintf("$%d\r\n%s\r\n", len(message), message)

					conn.Write([]byte(response))
				} else {
					conn.Write([]byte("-Error: ECHO command requires an argument\r\n"))
				}
			case "SET":
				if len(parts)>=3 {
					key := parts[1]
					value := strings.Join(parts[2:]," ")
					store[key] = value
					conn.Write(codec.OK())
				}else{
					conn.Write(codec.ErrorResponse("SET command require a key and a value"))
				}
			case "GET":
				if len(parts) == 2 {
					key := parts[1]
					value,exists := store[key]
					if exists {
						conn.Write(codec.EncodeBulkString(value))
					}else {
						conn.Write([]byte("$-1\r\n"))
					}
				} else {
					conn.Write(codec.ErrorResponse("GET command require a key"))
				}
			default:
				conn.Write([]byte("-Error: Unknown command\r\n"))
			}
		}
	}
}

func main() {
	startServer()
}
