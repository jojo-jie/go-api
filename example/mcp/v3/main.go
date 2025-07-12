package main

import (
	"context"
	"fmt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"log"
)

/*
printf '%s\n' \
'{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"clientInfo":{"name":"test-cli","version":"0.1"},"protocolVersion":"2025-03-26"}}' \
'{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}' \
'{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"greet","arguments":{"name":"Go MCP Enthusiast"}}}' \
| go run main.go
*/
func main() {
	server := mcp.NewServer(&mcp.Implementation{Name: "greeter", Version: "v1.0.0"}, nil)
	mcp.AddTool(server, &mcp.Tool{Name: "greet", Description: "say hi"}, SayHi)

	log.Println("Greeter server running over stdio...")
	if err := server.Run(context.Background(), mcp.NewStdioTransport()); err != nil {
		log.Fatalf("Server run failed: %+v", err)
	}
}

type HiParams struct {
	Name string `json:"name"`
}

func SayHi(ctx context.Context, session *mcp.ServerSession, c *mcp.CallToolParamsFor[HiParams]) (*mcp.CallToolResultFor[any], error) {
	resultText := fmt.Sprintf("Hi %s Name==%s", c.Arguments.Name, c.Name)
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: resultText},
		},
	}, nil
}
