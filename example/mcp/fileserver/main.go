package main

import (
	"context"
	"github.com/pkg/errors"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	prefixFileURI = "file:///"
	prefixListURI = "mcp://fs/list"
	mimeTextPlain = "text/plain"
)

func main() {
	server := mcp.NewServer(
		&mcp.Implementation{Name: "filesystem-server", Version: "1.0.1"},
		nil,
	)

	root, err := os.Getwd()
	if err != nil {
		log.Fatalf("cannot get working directory: %v", err)
	}
	log.Printf("File server serving from: %s", root)

	// 注册资源
	server.AddResource(
		&mcp.Resource{
			URI:         prefixListURI,
			Name:        "list_files",
			Description: "List all non-directory files in the current directory.",
		},
		listDir(root),
	)

	server.AddResourceTemplate(
		&mcp.ResourceTemplate{
			Name:        "read_file",
			URITemplate: "file:///{+filename}",
			Description: "Read a specific file from the directory.",
		},
		readFile(root),
	)

	log.Println("File system server running over stdio...")
	if err := server.Run(context.Background(), mcp.NewStdioTransport()); err != nil {
		log.Fatalf("server run failed: %v", err)
	}
}

// readFile 返回一个 ResourceHandler，用于读取指定文件
func readFile(base string) mcp.ResourceHandler {
	return func(ctx context.Context, _ *mcp.ServerSession, p *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
		rel := strings.TrimPrefix(p.URI, prefixFileURI)
		rel = filepath.FromSlash(rel)
		abs, err := secureJoin(base, rel)
		if err != nil {
			return nil, mcp.ResourceNotFoundError(p.URI)
		}

		data, err := os.ReadFile(abs)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, mcp.ResourceNotFoundError(p.URI)
			}
			return nil, errors.Wrap(err, "read file failed: %v")
		}

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: p.URI, MIMEType: mimeTextPlain, Text: string(data)},
			},
		}, nil
	}
}

// listDir 返回一个 ResourceHandler，用于列出 base 目录下的文件
func listDir(base string) mcp.ResourceHandler {
	return func(ctx context.Context, _ *mcp.ServerSession, p *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
		entries, err := os.ReadDir(base)
		if err != nil {
			return nil, errors.Wrap(err, "cannot read directory: %v")
		}

		var files []string
		for _, e := range entries {
			if !e.IsDir() {
				files = append(files, e.Name())
			}
		}

		content := "(The directory is empty or contains no files)"
		if len(files) > 0 {
			content = strings.Join(files, "\n")
		}

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: p.URI, MIMEType: mimeTextPlain, Text: content},
			},
		}, nil
	}
}

// secureJoin 防止目录穿越，确保最终路径在 base 之内
func secureJoin(base, rel string) (string, error) {
	abs := filepath.Join(base, rel)
	relToBase, err := filepath.Rel(base, abs)
	if err != nil || strings.HasPrefix(relToBase, "..") {
		return "", os.ErrNotExist
	}
	return abs, nil
}
