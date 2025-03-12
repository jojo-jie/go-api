package main

import (
	"fmt"
	"github.com/mark3labs/mcp-go/server"
	"mcp/timetool"
)

func main() {
	s := server.NewMCPServer(
		"Demo 🚀",
		"1.0.0",
	)

	s.SetTools(
		timetool.New(),
	)
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
