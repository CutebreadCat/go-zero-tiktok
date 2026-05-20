package ai

import (
	"context"
	"fmt"
	"go_zero-tiktok/internal/svc/xerr"
	"os"

	"github.com/sashabaranov/go-openai"
)

type Limiter interface {
	Allow(key string) (bool, error)
}

type Breaker interface {
	Do(name string, req func() error) error
}

type Agent struct {
	openaiClient *openai.Client
	mcpClient    *FuuMCP
	model        string
	limiter      Limiter
	breaker      Breaker
}

func NewAgent(ctx context.Context, l Limiter, b Breaker) (*Agent, error) {
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
		limiter:      l,
		breaker:      b,
	}, nil
}
func (a *Agent) Run(ctx context.Context, userId string, messages []openai.ChatCompletionMessage) (string, error) {
	if allowed, err := a.limiter.Allow(userId); err != nil {
		return "", xerr.Wrap(err, "Agent.Run.Limiter.Allow")
	} else if !allowed {
		return "", xerr.New(429, aiRateLimitMessage)
	}

	var result string
	var resultErr error

	err := a.breaker.Do(aiBreakerName, func() error {
		tools, err := a.mcpClient.ListMCPTools(ctx)
		if err != nil {
			return xerr.Wrap(err, "Agent.Run.ListMCPTools")
		}

		tools = append(tools, JwchLoginToolDef)

		var auth JwchLogin

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
				return xerr.Wrap(err, "Agent.Run.CreateChatCompletion")
			}

			msg := resp.Choices[0].Message
			messages = append(messages, msg)

			if len(msg.ToolCalls) == 0 {
				result = msg.Content
				return nil
			}

			for _, tc := range msg.ToolCalls {
				toolMsg, err := a.mcpClient.CallMCPTool(ctx, tc, &auth)
				if err != nil {
					return xerr.Wrap(err, "Agent.Run.CallMCPTool")
				}
				messages = append(messages, *toolMsg)
			}
		}

		return fmt.Errorf("tool loop exceeded max iterations")
	})

	if err != nil {
		return "", err
	}
	return result, resultErr
}
