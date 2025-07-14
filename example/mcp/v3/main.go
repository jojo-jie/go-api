package main

import (
	"context"
	"fmt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"log"
	"net/http"
)

/*
printf '%s\n' \
'{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"clientInfo":{"name":"test-cli","version":"0.1"},"protocolVersion":"2025-03-26"}}' \
'{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}' \
'{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"greet","arguments":{"name":"Go MCP Enthusiast"}}}' \
| go run main.go
*/
func main() {
	greeterServer := mcp.NewServer(&mcp.Implementation{Name: "greeter-server", Version: "v1.0.0"}, nil)
	mcp.AddTool(greeterServer, &mcp.Tool{Name: "greet", Description: "say hi"}, SayHi)

	handler := mcp.NewStreamableHTTPHandler(func(request *http.Request) *mcp.Server {
		log.Printf("Routing request for URL: %s\n", request.URL.Path)
		switch request.URL.Path {
		case "/greeter":
			return greeterServer
		default:
			return nil
		}
	}, nil)
	addr := ":8080"
	log.Printf("Multi-service MCP server Listening on %s\n", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalln(err)
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
