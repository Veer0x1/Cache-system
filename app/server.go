package main

import (
	"encoding/base64"
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
	port             = flag.String("port", "6379", "server port")
	replicaOf        = flag.String("replicaof", "", "host and port of master server")
	masterReplID     = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
	masterReplOffset = 0
	store            = make(map[string]StoredData)
	masterPort       = "6379"
	masterHost       = "localhost"
)

type StoredData struct {
	Data     string
	ExpireAt int64
}


func startServer(port string,replicaManager *ReplicaManager) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
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

		go handleConnection(&conn,replicaManager)
	}
}

func handleConnection(conn *net.Conn,replicaManager *ReplicaManager) {
	defer (*conn).Close()

	// conn.SetReadDeadline(time.Now().Add(5 * time.Minute))

	for {
		buffer := make([]byte, 1024)
		n, err := (*conn).Read(buffer)
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
				(*conn).Write([]byte("+PONG\r\n"))
			case "ECHO":
				if len(parts) > 1 {
					message := strings.Join(parts[1:], " ")
					response := fmt.Sprintf("$%d\r\n%s\r\n", len(message), message)

					(*conn).Write([]byte(response))
				} else {
					(*conn).Write([]byte("-Error: ECHO command requires an argument\r\n"))
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
					(*conn).Write([]byte("+OK\r\n"))

					respCommand := fmt.Sprintf("*%d\r\n", len(parts))
					for _, part := range parts {
						respCommand += fmt.Sprintf("$%d\r\n%s\r\n", len(part), part)
					}

					replicaManager.SendUpdateToReplicas(respCommand)
				} else {
					(*conn).Write([]byte("-Error: SET command requires at least a key and a value\r\n"))
				}
			case "GET":
				if len(parts) == 2 {
					key := parts[1]
					valueWithExpiry, exists := store[key]
					if exists {
						// Check if the key has expired.
						if valueWithExpiry.ExpireAt != 0 && time.Now().UnixNano()/1e6 > valueWithExpiry.ExpireAt {
							delete(store, key)            // Remove expired key.
							(*conn).Write([]byte("$-1\r\n")) // Return null bulk string.
						} else {
							(*conn).Write(codec.EncodeBulkString(valueWithExpiry.Data))
						}
					} else {
						(*conn).Write([]byte("$-1\r\n"))
					}
				} else {
					(*conn).Write(codec.ErrorResponse("GET command require a key"))
				}
			case "INFO":
				if len(parts) == 2 && parts[1] == "replication" {
					infoFields := []resp.KeyValuePair{
						{Key: "role", Value: serverRole},
						{Key: "master_replid", Value: masterReplID},
						{Key: "master_repl_offset", Value: masterReplOffset},
					}
					(*conn).Write(codec.EncodeMultipleBulkStrings(infoFields))
				} else {
					// Optionally handle other sections or provide a generic response.
					(*conn).Write(codec.ErrorResponse("Unsupported INFO section"))
				}
			case "REPLCONF":
				if len(parts) < 2 {
					(*conn).Write(codec.ErrorResponse("Not enough arguments for REPLCONF"))
					break
				}
				switch parts[1] {
				case "listening-port":
					
					
			
					(*conn).Write(codec.OK())
				case "capa":
					if len(parts) > 2 && parts[2] == "psync2" {
						(*conn).Write(codec.OK())
					} else {
						(*conn).Write(codec.ErrorResponse("Unsupported REPLCONF capa option"))
					}
				default:
					(*conn).Write(codec.ErrorResponse("Unsupported REPLCONF option"))
				}
			case "PSYNC":
				replicaPort := parts[2]
				remoteIP := (*conn).RemoteAddr().String()
				replicaManager.AddReplica(remoteIP, replicaPort, conn)

				response := fmt.Sprintf("+FULLRESYNC %s 0\r\n", masterReplID)
				_, err := (*conn).Write([]byte(response))
				if err != nil {
					log.Fatalf("Failed to send FULLRESYNC response: %v", err)
				}

				// Decode the base64-encoded empty RDB file
				rdbBase64 := "UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog=="
				rdbBytes, err := base64.StdEncoding.DecodeString(rdbBase64)
				if err != nil {
					log.Fatalf("Failed to decode RDB file: %v", err)
				}

				// Send the encoded empty RDB file
				rdbLength := len(rdbBytes)
				rdbHeader := fmt.Sprintf("$%d\r\n", rdbLength)
				_, err =(*conn).Write([]byte(rdbHeader))
				if err != nil {
					log.Fatalf("Failed to send RDB file header: %v", err)
				}
				_, err = (*conn).Write(rdbBytes)
				if err != nil {
					log.Fatalf("Failed to send RDB file contents: %v", err)
				}
			default:
				(*conn).Write([]byte("-Error: Unknown command\r\n"))
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

	replConfListeningPortCmd := fmt.Sprintf("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$4\r\n%s\r\n", *port)
	_, err = masterConn.Write([]byte(replConfListeningPortCmd))
	if err != nil {
		log.Fatalf("Failed to send REPLCONF listening-port command: %v", err)
	}

	_, err = masterConn.Read(buffer)
	if err != nil || !strings.Contains(string(buffer), "+OK\r\n") {
		log.Fatalf("Failed to receive OK response for REPLCONF listening-port command")
	}

	replConfCapaCmd := "*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n"
	_, err = masterConn.Write([]byte(replConfCapaCmd))
	if err != nil {
		log.Fatalf("Failed to send REPLCONF capa psync2 command: %v", err)
	}

	_, err = masterConn.Read(buffer)
	if err != nil || !strings.Contains(string(buffer), "+OK\r\n") {
		log.Fatalf("Failed to receive OK response for REPLCONF listening-port command")
	}

	pysncCommand := "*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n"
	_, err = masterConn.Write([]byte(pysncCommand))
	if err != nil || !strings.Contains(string(buffer), "+OK\r\n") {
		log.Fatalf("Failed to receive OK response for PSYNC command")
	}

	fmt.Printf("Received response from master: %s\n", string(buffer[:n]))
}

func main() {
	flag.Parse()

	replicaManager := NewReplicaManager()
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
	startServer(*port,replicaManager)
}
