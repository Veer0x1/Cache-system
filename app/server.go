package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/resp"
)

var (
	serverRole       = "master"
	port             = flag.Int("port", 6379, "server port")
	replicaOf        = flag.String("replicaof", "", "host and port of master server")
	masterReplID     = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
	masterReplOffset = 0
	store            = make(map[string]StoredData)
	masterPort       = "6379"
	masterHost       = "localhost"
	// storeMutex       = sync.RWMutex{} // Mutex to protect the store map
)

type StoredData struct {
	Data     string
	ExpireAt int64
}

func startServer(port int) {
	fmt.Println("I was here")
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
					infoFields := []resp.KeyValuePair{
						{Key: "role", Value: serverRole},
						{Key: "master_replid", Value: masterReplID},
						{Key: "master_repl_offset", Value: masterReplOffset},
					}
					conn.Write(codec.EncodeMultipleBulkStrings(infoFields))
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

func connectToMasterAndReplicate() {
	masterAddress := fmt.Sprintf("%s:%s", masterHost, masterPort)
	masterConn, err := net.Dial("tcp", masterAddress)
	if err != nil {
		fmt.Println("Failed to connect to the master:", err)
	}
	defer masterConn.Close()

	pingCommand := "*1\r\n$4\r\nPING\r\n"
	_, err = masterConn.Write([]byte(pingCommand))
	if err != nil {
		fmt.Println("Failed to send PING to the master:", err)
		return
	}

	// Read response from the master
	buffer := make([]byte, 1024)
	n, err := masterConn.Read(buffer)
	if err != nil {
		fmt.Println("Failed to read PING response from the master:", err)
		return
	}

	fmt.Printf("Received response from master: %s\n", string(buffer[:n]))
}

func main() {
	flag.Parse()
	if *replicaOf != "" {
		serverRole = "slave"
		formattedReplicaOf := strings.Replace(*replicaOf, " ", ":", 1)
		var err error
		masterHost, masterPort, err = net.SplitHostPort(formattedReplicaOf)
		if err != nil {
			log.Fatalf("Failed to parse master address '%s' : %v", *replicaOf, err)
		}

		go connectToMasterAndReplicate()
	}
	startServer(*port)
}
