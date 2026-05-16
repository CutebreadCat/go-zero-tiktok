package ai

import (
	"os"

	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

var (
	aimodel = "mimo-v2.5-pro"
)

type Agent struct {
	openaiClient *openai.Client
	mcpClient    *FuuMCP
	model        string
}

func NewAgent(mcp *FuuMCP) *Agent {
	return &Agent{
		openaiClient: openai.NewClient(os.Getenv("XIAOMI_AI_KEY")),
		mcpClient:    mcp,
		model:        aimodel,
	}
}
func (a *Agent) Run(ctx context.Context, messages []openai.ChatCompletionMessage, jwchid, jwchpassword string) (string, error) {

	auth, err := JwchLoginFunc(ctx, jwchid, jwchpassword)
	tools, err := a.mcpClient.ListMCPTools(ctx)
	if err != nil {
		return "", err
	}

	for i := 0; i < 10; i++ {
		resp, err := a.openaiClient.CreateChatCompletion(
			ctx,
			openai.ChatCompletionRequest{
				Model:    a.model,
				Messages: messages,
				Tools:    tools,
			},
		)
		if err != nil {
			return "", err
		}

		msg := resp.Choices[0].Message
		messages = append(messages, msg)

		if len(msg.ToolCalls) == 0 {
			return msg.Content, nil
		}

		for _, tc := range msg.ToolCalls {
			toolMsg, err := a.mcpClient.CallMCPTool(ctx, tc, &auth)
			if err != nil {
				return "", err
			}

			messages = append(messages, *toolMsg)
		}
	}

	return "", fmt.Errorf("tool loop exceeded max iterations")
}
