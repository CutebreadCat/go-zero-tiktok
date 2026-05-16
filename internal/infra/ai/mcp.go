package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	mcpclient "github.com/mark3labs/mcp-go/client"
	mcptransport "github.com/mark3labs/mcp-go/client/transport"
	mcpprotocol "github.com/mark3labs/mcp-go/mcp"
	"github.com/sashabaranov/go-openai"
)

type FuuMCP struct {
	client *mcpclient.Client
}

func NewFuuMCPClient(ctx context.Context) (*FuuMCP, error) {
	transport, err := mcptransport.NewStreamableHTTP("https://fzuhelper.west2.online/mcp")
	if err != nil {
		log.Printf("Failed to create MCP transport: %v", err)
		return nil, err
	}
	mcpClient := mcpclient.NewClient(transport)

	_, err = mcpClient.Initialize(ctx, mcpprotocol.InitializeRequest{})
	if err != nil {
		log.Printf("Failed to initialize MCP client: %v", err)
		return nil, err
	}

	return &FuuMCP{client: mcpClient}, nil
}

func (f *FuuMCP) ListMCPTools(ctx context.Context) ([]openai.Tool, error) {
	listResp, err := f.client.ListTools(ctx, mcpprotocol.ListToolsRequest{})
	if err != nil {
		log.Printf("Failed to list MCP tools: %v", err)
		return nil, err
	}
	tools := make([]openai.Tool, len(listResp.Tools))
	for i, tool := range listResp.Tools {
		tools[i] = openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.InputSchema,
			},
		}
	}
	return tools, nil
}
func (f *FuuMCP) CallMCPTool(ctx context.Context, tc openai.ToolCall, auth *JwchLogin) (*openai.ChatCompletionMessage, error) {
	var args map[string]any

	if tc.Function.Arguments != "" {
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			return nil, err
		}
	} else {
		args = make(map[string]any)
	}

	// 2. 自动注入 student_id 和 cookie
	if auth != nil {
		if auth.JwchId != "" {
			args["user_id"] = auth.JwchId
		}
		if auth.Jwchcookie != "" {
			args["user_cookie"] = auth.Jwchcookie
		}
	}

	// 3. 重新编码为 JSON
	finalArgs, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}
	result, err := f.client.CallTool(ctx, mcpprotocol.CallToolRequest{
		Params: mcpprotocol.CallToolParams{
			Name:      tc.Function.Name,
			Arguments: json.RawMessage(finalArgs),
		},
	})
	if err != nil {
		result = &mcpprotocol.CallToolResult{
			Content: []mcpprotocol.Content{
				mcpprotocol.TextContent{
					Text: fmt.Sprintf("Failed to call MCP tool: %v", err),
				},
			},
			IsError: true,
		}

	}
	jsonResult, err := result.MarshalJSON()
	if err != nil {
		log.Printf("Failed to marshal MCP tool result: %v", err)
		return nil, err
	}
	msg := openai.ChatCompletionMessage{
		Role:       openai.ChatMessageRoleTool,
		Content:    string(jsonResult),
		ToolCallID: tc.ID,
	}
	return &msg, nil

}
