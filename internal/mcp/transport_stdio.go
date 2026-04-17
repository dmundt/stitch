package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type registeredTool struct {
	Name        string
	Description string
	InputSchema any
}

func (a *app) serveSTDIO() error {
	server, err := a.newSDKServer()
	if err != nil {
		return err
	}
	return server.Run(context.Background(), &sdkmcp.StdioTransport{})
}

func (a *app) newSDKServer() (*sdkmcp.Server, error) {
	server := sdkmcp.NewServer(&sdkmcp.Implementation{
		Name:    "stitch-mcp",
		Version: "0.1.0",
	}, nil)

	tools, err := registeredTools()
	if err != nil {
		return nil, err
	}

	for _, tool := range tools {
		t := tool
		server.AddTool(&sdkmcp.Tool{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.InputSchema,
		}, func(ctx context.Context, req *sdkmcp.CallToolRequest) (*sdkmcp.CallToolResult, error) {
			args := map[string]any{}
			if len(req.Params.Arguments) > 0 {
				if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
					return &sdkmcp.CallToolResult{
						IsError: true,
						Content: []sdkmcp.Content{&sdkmcp.TextContent{Text: fmt.Sprintf("invalid arguments: %v", err)}},
					}, nil
				}
			}

			result, err := a.safeDispatch(t.Name, args)
			if err != nil {
				return &sdkmcp.CallToolResult{
					IsError: true,
					Content: []sdkmcp.Content{&sdkmcp.TextContent{Text: err.Error()}},
				}, nil
			}

			encoded, _ := json.MarshalIndent(result, "", "  ")
			return &sdkmcp.CallToolResult{
				IsError:           false,
				StructuredContent: result,
				Content: []sdkmcp.Content{
					&sdkmcp.TextContent{Text: string(encoded)},
				},
			}, nil
		})
	}

	return server, nil
}

func registeredTools() ([]registeredTool, error) {
	raw := mcpTools()
	out := make([]registeredTool, 0, len(raw))
	for _, tool := range raw {
		name, _ := tool["name"].(string)
		if name == "" {
			return nil, errors.New("tool definition is missing name")
		}
		description, _ := tool["description"].(string)
		out = append(out, registeredTool{
			Name:        name,
			Description: description,
			InputSchema: tool["inputSchema"],
		})
	}
	return out, nil
}
