package main

import (
	"context"
	"fmt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"log"
	"net/http"
	"time"
)

/*
printf '%s\n' \
'{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"clientInfo":{"name":"test-cli","version":"0.1"},"protocolVersion":"2025-03-26"}}' \
'{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}' \
'{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"greet","arguments":{"name":"Go MCP Enthusiast"}}}' \
| go run main.go
*/
func main() {
	greeterServer := mcp.NewServer(&mcp.Implementation{Name: "greeter-service", Version: "v1.0.0"}, nil)
	mcp.AddTool(greeterServer, &mcp.Tool{Name: "greet", Description: "say hi"}, SayHi)

	mathServer := mcp.NewServer(&mcp.Implementation{Name: "math-service", Title: "1.0", Version: "v1.0.0"}, nil)
	mcp.AddTool(mathServer, &mcp.Tool{Name: "add", Description: "Add two integers"}, Add)

	timezoneServer := mcp.NewServer(&mcp.Implementation{Name: "timezone-service", Title: "1.0", Version: "v1.0.0"}, nil)
	mcp.AddTool(timezoneServer, &mcp.Tool{Name: "timezone", Description: "Get current time with timezone, Asia/Shanghai is default"}, Timezone)

	handler := mcp.NewStreamableHTTPHandler(func(request *http.Request) *mcp.Server {
		log.Printf("Routing request for URL: %s\n", request.URL.Path)
		switch request.URL.Path {
		case "/greeter":
			return greeterServer
		case "/math":
			return mathServer
		case "/timezone":
			return timezoneServer
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
	resultText := fmt.Sprintf("Hi %s nice", c.Arguments.Name)
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: resultText},
		},
	}, nil
}

type AddParams struct{ A, B int }

func Add(ctx context.Context, session *mcp.ServerSession, c *mcp.CallToolParamsFor[AddParams]) (*mcp.CallToolResultFor[any], error) {
	result := c.Arguments.A + c.Arguments.B
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("The sum is: %d", result)}},
	}, nil
}

type TimezoneParams struct {
	Timezone string `json:"timezone"`
}

func Timezone(ctx context.Context, session *mcp.ServerSession, c *mcp.CallToolParamsFor[TimezoneParams]) (*mcp.CallToolResultFor[any], error) {
	loc, err := time.LoadLocation(c.Arguments.Timezone)
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("current time is %s", time.Now().In(loc))}},
	}, err
}
