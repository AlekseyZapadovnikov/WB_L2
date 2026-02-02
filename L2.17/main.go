package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	timeoutFlag := flag.Duration("timeout", 10*time.Second, "Connection timeout (e.g. 5s, 1m)")
	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		fmt.Println("Usage: go-telnet [--timeout <duration>] <host> <port>")
		os.Exit(1)
	}

	host := args[0]
	port := args[1]
	address := net.JoinHostPort(host, port)

	conn, err := net.DialTimeout("tcp", address, *timeoutFlag)
	if err != nil {
		log.Fatalf("Error: could not connect to %s: %v\n", address, err)
	}
	defer conn.Close()

	fmt.Printf("Connected to %s\n", address)

	done := make(chan error, 2)

	// Чтение из STDOUT
	go func() {
		_, err := io.Copy(os.Stdout, conn)
		if err != nil {
			log.Printf("Read error: %v\n", err)
		}
		done <- err
	}()

	// Чтение из stdin
	go func() {
		_, err := io.Copy(conn, os.Stdin)
		if err != nil {
			log.Printf("Write error: %v\n", err)
		}
		done <- err
	}()

	select {
	case <-done:
		fmt.Println("\nConnection closed")
	}
}