package timetool

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"time"
)

func New() server.ServerTool {
	tool := mcp.NewTool("current time",
		mcp.WithDescription("Get current time with timezone, Asia/Shanghai is default"),
		mcp.WithString("timezone",
			mcp.Required(),
			mcp.Description("current time timezone"),
		),
	)
	return server.ServerTool{
		Tool:    tool,
		Handler: currentTimeHandler,
	}
}

func currentTimeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	timezone, ok := request.Params.Arguments["timezone"].(string)
	if !ok {
		return mcp.NewToolResultText("timezone must be a string"), nil
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("parse timezone with error: %v", err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf(`current time is %s`, time.Now().In(loc))), nil
}
