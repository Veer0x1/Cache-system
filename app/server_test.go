package main

import (
	"bufio"
	"net"
	"os"
	"sync"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	go startServer("6379")
	// Give the server time to start
	time.Sleep(time.Second)

	// Run the tests
	code := m.Run()

	os.Exit(code)
}

func TestHandleConnection(t *testing.T) {

	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		t.Errorf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Send a PING command
	_, err = conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
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
		t.Fatalf("Failed to read from server")
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Scanner error: %v", err)
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
		_, err = conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
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

func TestHandleMultipleUser(t *testing.T) {
	userCount := 10
	var wg sync.WaitGroup // synchronize multiple goroutines

	// simulating 10 concurrent user using goroutines
	for i := 0; i < userCount; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			conn, err := net.Dial("tcp", "localhost:6379")
			if err != nil {
				t.Errorf("Failed to connect to server: %v", err)
			}
			defer conn.Close()

			_, err = conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
			if err != nil {
				t.Errorf("Failed to write to server: %v", err)
			}

			scanner := bufio.NewScanner(conn)
			if scanner.Scan() {
				response := scanner.Text()
				if response != "+PONG" {
					t.Errorf("Expected '+PONG', got '%s'", response)
				}
			} else {
				t.Errorf("Failed to read from server")
			}

			if err := scanner.Err(); err != nil {
				t.Errorf("Scanner error: %v", err)
			}
		}()
	}
	wg.Wait()
}

func TestEchoCommand(t *testing.T) {
	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Send ECHO command in RESP format
	_, err = conn.Write([]byte("*2\r\n$4\r\nECHO\r\n$9\r\nraspberry\r\n"))
	if err != nil {
		t.Fatalf("Failed to write ECHO command: %v", err)
	}

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		t.Fatalf("Failed to read response from server: %v", err)
	}

	// Convert the bytes read into a string
	response := string(buffer[:n])
	expectedResponse := "$9\r\nraspberry\r\n"
	if response != expectedResponse {
		t.Errorf("Expected response '%s', got '%s'", expectedResponse, response)
	}
}
