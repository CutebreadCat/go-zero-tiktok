package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"go_zero-tiktok/internal/svc/xerr"

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
		return nil, xerr.Wrap(err, "NewFuuMCPClient.NewStreamableHTTP")
	}
	mcpClient := mcpclient.NewClient(transport)

	_, err = mcpClient.Initialize(ctx, mcpprotocol.InitializeRequest{})
	if err != nil {
		return nil, xerr.Wrap(err, "NewFuuMCPClient.Initialize")
	}

	return &FuuMCP{client: mcpClient}, nil
}

func (f *FuuMCP) ListMCPTools(ctx context.Context) ([]openai.Tool, error) {
	listResp, err := f.client.ListTools(ctx, mcpprotocol.ListToolsRequest{})
	if err != nil {
		return nil, xerr.Wrap(err, "FuuMCP.ListMCPTools")
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
			return nil, xerr.Wrap(err, "FuuMCP.CallMCPTool.Unmarshal")
		}
	} else {
		args = make(map[string]any)
	}

	// 处理本地教务处登录工具
	if tc.Function.Name == "jwch_login" {
		content, err := HandleJwchLogin(ctx, args)
		if err != nil {
			return &openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    fmt.Sprintf("教务处登录失败: %v", err),
				ToolCallID: tc.ID,
			}, nil
		}
		// 解析登录结果，保存到 auth
		var loginResult JwchLogin
		if err := json.Unmarshal([]byte(content), &loginResult); err == nil && auth != nil {
			*auth = loginResult
		}
		return &openai.ChatCompletionMessage{
			Role:       openai.ChatMessageRoleTool,
			Content:    content,
			ToolCallID: tc.ID,
		}, nil
	}

	// 处理远程 MCP 工具
	if auth != nil {
		if auth.JwchId != "" {
			args["user_id"] = auth.JwchId
		}
		if auth.Jwchcookie != "" {
			args["user_cookie"] = auth.Jwchcookie
		}
	}

	finalArgs, err := json.Marshal(args)
	if err != nil {
		return nil, xerr.Wrap(err, "FuuMCP.CallMCPTool.Marshal")
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
		return nil, xerr.Wrap(err, "FuuMCP.CallMCPTool.MarshalJSON")
	}
	msg := openai.ChatCompletionMessage{
		Role:       openai.ChatMessageRoleTool,
		Content:    string(jsonResult),
		ToolCallID: tc.ID,
	}
	return &msg, nil

}
