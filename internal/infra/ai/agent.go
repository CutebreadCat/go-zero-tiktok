package ai

import (
	"context"
	"fmt"
	"go_zero-tiktok/internal/svc/xerr"
	"os"

	"github.com/sashabaranov/go-openai"
)

type Agent struct {
	openaiClient *openai.Client
	mcpClient    *FuuMCP
	model        string
}

func NewAgent(ctx context.Context) (*Agent, error) {
	mcp, err := NewFuuMCPClient(ctx)
	if err != nil {
		return nil, xerr.Wrap(err, "NewAgent.NewFuuMCPClient")
	}
	config := openai.DefaultConfig(os.Getenv(xiaomiAIEnv))
	config.BaseURL = xiaomiAIURL
	return &Agent{
		openaiClient: openai.NewClientWithConfig(config),
		mcpClient:    mcp,
		model:        aiModel,
	}, nil
}
func (a *Agent) Run(ctx context.Context, messages []openai.ChatCompletionMessage, jwchid, jwchpassword string) (string, error) {

	auth, err := JwchLoginFunc(ctx, jwchid, jwchpassword)
	if err != nil {
		return "", xerr.Wrap(err, "Agent.Run.JwchLoginFunc")
	}
	tools, err := a.mcpClient.ListMCPTools(ctx)
	if err != nil {
		return "", xerr.Wrap(err, "Agent.Run.ListMCPTools")
	}

	for i := 0; i < maxToolLoops; i++ {
		resp, err := a.openaiClient.CreateChatCompletion(
			ctx,
			openai.ChatCompletionRequest{
				Model:    a.model,
				Messages: messages,
				Tools:    tools,
			},
		)
		if err != nil {
			return "", xerr.Wrap(err, "Agent.Run.CreateChatCompletion")
		}

		msg := resp.Choices[0].Message
		messages = append(messages, msg)

		if len(msg.ToolCalls) == 0 {
			return msg.Content, nil
		}

		for _, tc := range msg.ToolCalls {
			toolMsg, err := a.mcpClient.CallMCPTool(ctx, tc, &auth)
			if err != nil {
				return "", xerr.Wrap(err, "Agent.Run.CallMCPTool")
			}

			messages = append(messages, *toolMsg)
		}
	}

	return "", fmt.Errorf("tool loop exceeded max iterations")
}
