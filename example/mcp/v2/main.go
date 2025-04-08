package main

import (
	"context"
	"fmt"
	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/ThinkInAIXYZ/go-mcp/server"
	"github.com/ThinkInAIXYZ/go-mcp/transport"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	s, err := server.NewServer(
		transport.NewStdioServerTransport(),
		server.WithServerInfo(protocol.Implementation{
			Name:    "Server-Name",
			Version: "0.0.1",
		}),
	)
	if err != nil {
		log.Fatalf("Failed to create server:%+v\n", err)
	}

	s.RegisterTool(&protocol.Tool{
		Name:        "current time",
		Description: "Get current time with timezone, Asia/Shanghai is default",
		InputSchema: protocol.InputSchema{
			Type: protocol.Object,
			Properties: map[string]interface{}{
				"timezone": map[string]string{
					"type":        "string",
					"description": "current time timezone",
				},
			},
			Required: []string{"timezone"},
		},
	}, func(req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
		timezone, ok := req.Arguments["timezone"].(string)
		if !ok {
			return nil, errors.New("timezone must be a string")
		}

		loc, err := time.LoadLocation(timezone)
		if err != nil {
			return nil, fmt.Errorf("parse timezone with error: %v", err)
		}
		text := fmt.Sprintf(`current time is %s`, time.Now().In(loc))

		return &protocol.CallToolResult{
			Content: []protocol.Content{
				protocol.TextContent{
					Type: "text",
					Text: text,
				},
			},
		}, nil
	})

	g := new(errgroup.Group)
	g.Go(func() error {
		if err := s.Run(); err != nil {
			return errors.Wrap(err, "Server error:%+v\n")
		}
		return nil
	})
	g.Go(func() error {
		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		log.Println("shutting down server...")
		return s.Shutdown(ctx)
	})
	if err := g.Wait(); err != nil {
		log.Fatalf("%+v", err)
	}
}
