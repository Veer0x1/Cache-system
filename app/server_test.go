package main

import (
	"bufio"
	"net"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
    go startServer()
    // Give the server time to start
    time.Sleep(time.Second)

    // Run the tests
    code := m.Run()

    os.Exit(code)
}

func TestHandleConnection(t *testing.T){

	conn, err := net.Dial("tcp","localhost:6379")
	if err!= nil{
		t.Fatalf("Failed to connect to server: %v",err)
	}
	defer conn.Close()

	// Send a PING command
	_,err = conn.Write([]byte("PING\r\n"))
	if err!=nil{
		t.Fatalf("Failed to write to server: %v",err)
	}

	// Read the response
	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		response := scanner.Text()
		if response != "+PONG"{
			t.Errorf("Expected '+PONG', got '%s'",response)
		}
	}else {
		t.Fatalf("Failed to read from server")
	}

	if err := scanner.Err(); err !=nil {
		t.Fatalf("Scanner error: %v",err)
	}
}

func TestHandleMultiplePings(t *testing.T) {

	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Number of PING commands to send
	pingCount := 3

	for i := 0; i < pingCount; i++ {
		// Send a PING command
		_, err = conn.Write([]byte("PING\r\n"))
		if err != nil {
			t.Fatalf("Failed to write to server: %v", err)
		}

		// Read the response
		scanner := bufio.NewScanner(conn)
		if scanner.Scan() {
			response := scanner.Text()
			if response != "+PONG" {
				t.Errorf("Expected '+PONG', got '%s'", response)
			}
		} else {
			t.Fatal("Failed to read from server")
		}

		if err := scanner.Err(); err != nil {
			t.Fatalf("Scanner error: %v", err)
		}
	}
}