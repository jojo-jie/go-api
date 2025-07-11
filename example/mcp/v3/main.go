package v3

import (
	"context"
	"fmt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"log"
)

func main() {
	server := mcp.NewServer("greeter-server", "1.0.0", nil)
	server.AddTools(
		mcp.NewServerTool("greet", "Say hi to someone", SayHi),
	)
	log.Println("Greeter server running over stdio...")
	if err := server.Run(context.Background(), mcp.NewStdioTransport()); err != nil {
		log.Fatalf("Server run failed: %+v", err)
	}
}

type HiParams struct {
	Name string `json:"name"`
}

func SayHi(ctx context.Context, session *mcp.ServerSession, c *mcp.CallToolParamsFor[HiParams]) (*mcp.CallToolResultFor[any], error) {
	resultText := fmt.Sprintf("Hi %s", c.Arguments.Name)
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: resultText},
		},
	}, nil
}
