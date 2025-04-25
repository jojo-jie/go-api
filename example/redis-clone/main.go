package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"redis-clone/store"
	"strings"
	"time"
)

type Server struct {
	cache *store.Cache
}

/*
NewServer initializes a new instance of the Server.
It creates a Cache instance with AOF persistence enabled, reading from the specified AOF file.
It also replays the AOF file to restore the cache state and starts a background cleaning process for the cache.
*/
func NewServer(aofFilename string) (*Server, error) {
	persist, err := store.NewAOF(aofFilename)
	if err != nil {
		return nil, err
	}

	cache := store.NewCache(5, persist)

	if err := replayAOF(aofFilename, cache); err != nil {
		return nil, err
	}

	go cache.StartCleaningServer()

	return &Server{
		cache: cache,
	}, nil
}

func replayAOF(aofFilename string, cache *store.Cache) error {
	file, err := os.Open(aofFilename)
	if err != nil {
		return fmt.Errorf("could not open AOF file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "SET") {
			parts := strings.Fields(line)
			if len(parts) == 3 {
				cache.Set(parts[1], parts[2], time.Hour, true)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading AOF file: %v", err)
	}

	return nil
}

func (s *Server) HandleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		// Read the incoming message from the client
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Println(err)
			return
		}

		message = strings.TrimSpace(message)

		tokens := strings.Split(message, " ")

		if len(tokens) < 2 {
			conn.Write([]byte("Invalid command\n"))
			continue
		}

		cmd := strings.ToUpper(tokens[0])

		switch cmd {
		case "SET":

			if len(tokens) < 3 {
				conn.Write([]byte("Usage: SET key value\n"))
				continue
			}

			s.cache.Set(tokens[1], tokens[2], time.Hour, false)
			conn.Write([]byte("OK\n"))

		case "GET":
			key := tokens[1]
			if item, exists := s.cache.Get(key); !exists {
				conn.Write([]byte("nil\n"))
			} else {
				conn.Write([]byte(item + "\n"))
			}

		default:
			conn.Write([]byte("Unknown Command\n"))
		}
	}
}

func (s *Server) Start(address string) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}

	defer listener.Close()

	fmt.Println("Cache server running on port", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go s.HandleConnection(conn)
	}
}

func main() {
	aofFilename := "cache.aof"

	server, err := NewServer(aofFilename)
	if err != nil {
		log.Fatalf("Error initializing server: %v", err)
	}

	server.Start(":6379")
}
