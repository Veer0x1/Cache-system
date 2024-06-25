package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	//PING command
	conn.Write([]byte("+PONG\r\n"))
	// if len(os.Args) > 1 {
	// 	fmt.Println("Command: ", os.Args[1])
	// }
	// if os.Args[1]=="PING"{
	//   	_,err = conn.Write([]byte("+PONG\r\n"))
	// 	if err != nil {
	// 		fmt.Println("Error writing to connection: ", err.Error())
	// 		os.Exit(1)
	// 	}
	// }
}
