package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

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
	//PING command
	// conn.Write([]byte("+PONG\r\n"))
	// conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.ToUpper(line) == "PING" {
			_,err := conn.Write([]byte("+PONG\r\n"))
			if err != nil {
				fmt.Println("Error sending PONG response: ",err.Error())
				return
			}
		}
	}

	if err:= scanner.Err(); err!=nil {
		fmt.Println("Error reading from connections: ",err.Error())
	}
}

func main() {
	startServer()
}
