package main

import (
	"bufio"
	"fmt"
	"os"
	"redis-clone/store"
	"strings"
	"time"
)

type Server struct {
	cache *store.Cache // Cache used to store key-value pairs
}

/*
NewServer initializes a new instance of the Server.
It creates a Cache instance with AOF persistence enabled, reading from the specified AOF file.
It also replays the AOF file to restore the cache state and starts a background cleaning process for the cache.
*/
func NewServer(aofFilename string) (*Server, error) {
	// Initialize AOF persistence
	persist, err := store.NewAOF(aofFilename)
	if err != nil {
		return nil, err
	}

	// Create the cache with a maximum size of 5 items and enable AOF persistence
	cache := store.NewCache(5, persist)

	// Replay the AOF file to restore the cache state
	if err := replayAOF(aofFilename, cache); err != nil {
		return nil, err
	}

	// Start the cache cleaning server in the background
	go cache.StartCleaningServer()

	return &Server{
		cache: cache, // Initialize the server with the cache
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
