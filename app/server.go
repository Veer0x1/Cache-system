package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/resp"
)

var port = flag.Int("port", 6379, "Port number for redis server")
var serverRole = "master"

type StoredData struct {
	Data     string
	ExpireAt int64
}

var store = make(map[string]StoredData) // In-memory key value

func startServer(port int) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
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
				if len(parts) >= 3 {
					key := parts[1]
					expiryIndex := len(parts) // Assume no optional arguments by default

					// Check for "PX" argument and adjust expiryIndex accordingly.
					for i, part := range parts {
						if strings.ToUpper(part) == "PX" && i+1 < len(parts) {
							expiryIndex = i
							break
						}
					}

					// Extract the value correctly by joining parts up to the expiryIndex.
					value := strings.Join(parts[2:expiryIndex], " ")
					expiry := int64(0) // No expiry by default

					// Parse expiry time if "PX" argument is present.
					if expiryIndex != len(parts) {
						expiryMillis, err := strconv.ParseInt(parts[expiryIndex+1], 10, 64)
						if err == nil {
							expiry = time.Now().UnixNano()/1e6 + expiryMillis // Current time in ms + expiry duration
						}
					}

					// Store the key-value with optional expiry.
					store[key] = StoredData{Data: value, ExpireAt: expiry}
					conn.Write([]byte("+OK\r\n"))
				} else {
					conn.Write([]byte("-Error: SET command requires at least a key and a value\r\n"))
				}
			case "GET":
				if len(parts) == 2 {
					key := parts[1]
					valueWithExpiry, exists := store[key]
					if exists {
						// Check if the key has expired.
						if valueWithExpiry.ExpireAt != 0 && time.Now().UnixNano()/1e6 > valueWithExpiry.ExpireAt {
							delete(store, key)            // Remove expired key.
							conn.Write([]byte("$-1\r\n")) // Return null bulk string.
						} else {
							conn.Write(codec.EncodeBulkString(valueWithExpiry.Data))
						}
					} else {
						conn.Write([]byte("$-1\r\n"))
					}
				} else {
					conn.Write(codec.ErrorResponse("GET command require a key"))
				}
			case "INFO":
				if len(parts) == 2 && parts[1] == "replication" {
					// Since only the role key is needed for this stage, construct the response.
					response := fmt.Sprintf("role:%s\r\n",serverRole) // Ensure to include \r\n for proper formatting.

					// Encode the response as a Bulk string.
					// Assuming codec.EncodeBulkString properly encodes a string as a Redis Bulk string.
					encodedResponse := codec.EncodeBulkString(response)

					// Send the encoded response back to the client.
					conn.Write([]byte(encodedResponse))
				} else {
					// Optionally handle other sections or provide a generic response.
					conn.Write(codec.ErrorResponse("Unsupported INFO section"))
				}
			default:
				conn.Write([]byte("-Error: Unknown command\r\n"))
			}
		}
	}
}

func main() {
	var replicaOf = flag.String("replicaof","","host and port of master server")
	flag.Parse()
	if *replicaOf != "" {
		serverRole = "slave"
	}
	startServer(*port)
}
