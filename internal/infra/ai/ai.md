# 社交模块 AI 聊天 + MCP ToolCall 完整方案（Go 实现）

本文给出一个完整、可落地的架构设计与代码方案，实现以下目标：

1. 在社交聊天中引入 AI 用户（群聊或私聊中自动参与）
2. AI 可根据消息自动决定是否回复
3. AI 可调用 Tool（OpenAI Tool Calling）
4. 服务端接入 MCP（Model Context Protocol）
5. MCP Tool 通过 `mcp-go` 动态注册到 OpenAI Tool Call
6. 某些 MCP Tool 需要用户教务系统账号密码
7. 服务端通过 `jwch` 获取 `student_id + cookie`
8. 将这些凭证作为 MCP Tool 参数注入
9. 实现完整 Agent Loop
10. 提供可直接使用的 Go 代码结构

---

# 一、整体架构

```text
用户发送消息
      ↓
Chat Service
      ↓
AI Trigger 判断是否触发 AI
      ↓
Agent Runner
      ↓
OpenAI Chat Completion
      ↓
模型请求 tool_calls
      ↓
Tool Router
      ↓
MCP Client
      ↓
UU MCP Server
      ↓
返回工具结果
      ↓
继续 Agent Loop
      ↓
生成最终回复
      ↓
Message Service
      ↓
将 AI 回复作为普通消息发送到聊天室
```

---

# 二、推荐参考项目

## 1. urlBifrost Gateway[https://github.com/maximhq/bifrost](https://github.com/maximhq/bifrost)

适合作为 AI Gateway 参考。

## 2. urlmark3labs/mcp-go[https://github.com/mark3labs/mcp-go](https://github.com/mark3labs/mcp-go)

Go MCP 官方实现。

## 3. urlgo-openai[https://github.com/sashabaranov/go-openai](https://github.com/sashabaranov/go-openai)

Go OpenAI SDK。

## 4. urlClaude Code Open Source Parts[https://github.com/anthropics/claude-code-sdk-go](https://github.com/anthropics/claude-code-sdk-go)

可参考 Agent Loop。

---

# 三、核心模块设计

```text
/internal
  /ai
    agent.go
    runner.go
    trigger.go
    memory.go
    prompt.go

  /mcp
    manager.go
    converter.go
    executor.go

  /jwch
    auth.go

  /chat
    service.go

  /models
    message.go
    tool.go
    user_context.go
```

---

# 四、数据结构

## Chat Message

```go
type ChatMessage struct {
    ID        string
    RoomID    string
    SenderID  string
    Content   string
    CreatedAt time.Time
}
```

## AI User Context

```go
type AIUserContext struct {
    UserID      string
    JWUsername  string
    JWPassword  string
    StudentID   string
    Cookie      string
}
```

---

# 五、AI 触发策略

## 简单方案（推荐先实现）

以下条件触发：

* @AI
* 包含关键词：`帮我`, `查询`, `课表`, `成绩`
* 每收到一条消息都触发

## 智能方案

先调用轻量模型判断：

```json
{
  "should_reply": true,
  "reason": "用户在询问成绩"
}
```

---

# 六、MCP Manager

## manager.go

```go
package mcp

import (
    "context"

    mcpclient "github.com/mark3labs/mcp-go/client"
    mcptransport "github.com/mark3labs/mcp-go/client/transport"
)

type Manager struct {
    client *mcpclient.Client
}

func NewManager(serverURL string) (*Manager, error) {
    transport := mcptransport.NewSSE(serverURL)

    client := mcpclient.NewClient(transport)

    if err := client.Start(context.Background()); err != nil {
        return nil, err
    }

    return &Manager{client: client}, nil
}
```

---

# 七、获取 MCP Tools

```go
func (m *Manager) ListTools(ctx context.Context) ([]mcpprotocol.Tool, error) {
    resp, err := m.client.ListTools(ctx)
    if err != nil {
        return nil, err
    }
    return resp.Tools, nil
}
```

---

# 八、MCP Tool → OpenAI Tool 转换

```go
func ConvertMCPTool(tool mcpprotocol.Tool) openai.Tool {
    return openai.Tool{
        Type: openai.ToolTypeFunction,
        Function: &openai.FunctionDefinition{
            Name:        tool.Name,
            Description: tool.Description,
            Parameters:  tool.InputSchema,
        },
    }
}
```

---

# 九、JWCH 登录

## auth.go

使用的库：urlwest2-online/jwch[https://github.com/west2-online/jwch](https://github.com/west2-online/jwch)

```go
package jwch

import (
    "context"

    jwch "github.com/west2-online/jwch"
)

func LoginJWCH(ctx context.Context, username, password string) (studentID, cookie string, err error) {
    // 以下 API 名称可能会随着 jwch 版本略有变化，请以实际版本为准。

    client := jwch.NewClient()

    // 登录教务系统
    if err = client.Login(ctx, username, password); err != nil {
        return "", "", err
    }

    // 获取学号
    profile, err := client.GetProfile(ctx)
    if err != nil {
        return "", "", err
    }

    // 导出 Cookie 字符串
    cookie, err = client.ExportCookieString()
    if err != nil {
        return "", "", err
    }

    return profile.StudentID, cookie, nil
}
```

> 注意：不同版本的 `west2-online/jwch` 在方法命名上可能有差异，例如 `NewClient`、`Login`、`GetProfile`、`ExportCookieString`。整体思路不变：登录 → 获取用户信息 → 导出 Cookie。

---

# 十、Tool 参数注入

某些 MCP Tool 需要：

* student_id
* cookie

### 自动注入

```go
func InjectAuth(args map[string]any, ctx *AIUserContext) {
    args["student_id"] = ctx.StudentID
    args["cookie"] = ctx.Cookie
}
```

---

# 十一、Tool Executor

```go
func (m *Manager) CallTool(
    ctx context.Context,
    name string,
    args map[string]any,
) (string, error) {
    resp, err := m.client.CallTool(ctx, name, args)
    if err != nil {
        return "", err
    }

    if len(resp.Content) == 0 {
        return "", nil
    }

    return resp.Content[0].Text, nil
}
```

---

# 十二、OpenAI Client

```go
client := openai.NewClient(apiKey)
```

---

# 十三、Agent Runner

```go
type Runner struct {
    openaiClient *openai.Client
    mcpManager   *mcp.Manager
}
```

---

# 十四、完整 Agent Loop

```go
func (r *Runner) Run(
    ctx context.Context,
    messages []openai.ChatCompletionMessage,
    userCtx *AIUserContext,
) (string, error) {

    tools, err := r.mcpManager.ListTools(ctx)
    if err != nil {
        return "", err
    }

    var openaiTools []openai.Tool
    for _, t := range tools {
        openaiTools = append(openaiTools, ConvertMCPTool(t))
    }

    for i := 0; i < 10; i++ {
        resp, err := r.openaiClient.CreateChatCompletion(
            ctx,
            openai.ChatCompletionRequest{
                Model:    openai.GPT4oMini,
                Messages: messages,
                Tools:    openaiTools,
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
            var args map[string]any
            _ = json.Unmarshal([]byte(tc.Function.Arguments), &args)

            InjectAuth(args, userCtx)

            result, err := r.mcpManager.CallTool(
                ctx,
                tc.Function.Name,
                args,
            )
            if err != nil {
                result = "tool error: " + err.Error()
            }

            messages = append(messages,
                openai.ChatCompletionMessage{
                    Role:       openai.ChatMessageRoleTool,
                    ToolCallID: tc.ID,
                    Content:    result,
                },
            )
        }
    }

    return "达到最大循环次数", nil
}
```

---

# 十五、系统 Prompt

```text
你是校园社交平台中的 AI 助手。

能力：
1. 参与群聊。
2. 回答校园问题。
3. 调用工具查询课表、成绩、考试安排。
4. 若工具需要 student_id 和 cookie，系统会自动注入。
5. 回复要自然，像一个真实用户。
```

---

# 十六、聊天接入

```go
func (s *ChatService) OnMessage(msg ChatMessage) {
    if !s.aiTrigger.ShouldReply(msg) {
        return
    }

    go func() {
        reply, err := s.agent.Run(
            context.Background(),
            BuildConversation(msg.RoomID),
            LoadUserContext(msg.SenderID),
        )
        if err != nil {
            return
        }

        s.SendMessage(ChatMessage{
            RoomID:   msg.RoomID,
            SenderID: "ai-assistant",
            Content:  reply,
        })
    }()
}
```

---

# 十七、消息上下文构建

建议取最近 20 条消息。

```go
func BuildConversation(roomID string) []openai.ChatCompletionMessage
```

---

# 十八、AI 作为普通用户

数据库中插入一个特殊用户：

```text
id: ai-assistant
nickname: 校园助手
avatar: ...
```

AI 回复后直接调用普通消息发送接口。

---

# 十九、Tool 权限控制

建议建立配置：

```go
type ToolPolicy struct {
    RequireAuth bool
    AllowedRoles []string
}
```

例如：

| Tool         | RequireAuth |
| ------------ | ----------- |
| get_schedule | true        |
| get_grades   | true        |
| search_news  | false       |

---

# 二十、JWCH 凭证缓存

建议 Redis：

```text
ai:jwch:{user_id}
TTL = 2h
```

字段：

* student_id
* cookie

---

# 二十一、Tool 调用日志

```go
type ToolCallLog struct {
    UserID    string
    ToolName  string
    Args      string
    Result    string
    Duration  int64
    Success   bool
}
```

---

# 二十二、建议的回复触发模式

## 模式 1：@AI

## 模式 2：概率回复

## 模式 3：工具需求回复

## 模式 4：静默观察

---

# 二十三、推荐实现策略

## Phase 1

* OpenAI Tool Calling
* MCP 接入
* JWCH 登录
* AI 自动回复

## Phase 2

* 智能 Trigger
* Memory
* Tool 权限

## Phase 3

* 多 Agent
* RAG
* 长期记忆

---

# 二十四、生产级增强

## 超时

```go
ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
defer cancel()
```

## 并发控制

每个 room 限制一个 AI 任务。

## 熔断

连续失败时暂停 AI。

## Token 控制

上下文裁剪。

---

# 二十五、完整时序图

```text
User → ChatService
ChatService → AI Trigger
AI Trigger → Agent Runner
Runner → OpenAI
OpenAI → tool_calls
Runner → MCP Manager
MCP Manager → UU MCP
UU MCP → Tool Result
Runner → OpenAI
OpenAI → Final Reply
Runner → ChatService
ChatService → Room
```

---

# 二十六、为什么推荐 MCP + Tool Calling

优势：

* 工具动态发现
* Schema 自动转换
* 无需手写每个 Tool
* 与 OpenAI 原生兼容
* 可复用现有 UU MCP Tool

---

# 二十七、最关键的三个模块

1. MCP Tool Discovery
2. Agent Loop
3. Auth Injection

---

# 二十八、最终建议架构

```text
Chat Service
  └── AI Gateway
        ├── Trigger
        ├── Prompt Builder
        ├── Agent Runner
        ├── Tool Router
        ├── MCP Manager
        ├── Auth Injector
        └── Memory
```

---

# 二十九、依赖安装

```bash
go get github.com/sashabaranov/go-openai
go get github.com/mark3labs/mcp-go
go get github.com/redis/go-redis/v9
```

---

# 三十、完整流程总结

1. 用户发送消息
2. Trigger 判断是否需要 AI 回复
3. 构建 Prompt + 最近聊天记录
4. 获取 MCP Tool 列表
5. 转换为 OpenAI Tools
6. 调用模型
7. 如果出现 Tool Calls
8. 自动注入 student_id + cookie
9. 调用 MCP Tool
10. 将结果追加到上下文
11. 继续 Loop
12. 得到最终回复
13. 以 AI 用户身份发送消息

---

# 三十一、推荐实施方式

最推荐的方案：

* OpenAI Tool Calling
* `mcp-go`
* `go-openai`
* `jwch`
* Redis 缓存
* Bifrost 风格 AI Gateway

这是目前最稳健、扩展性最强的架构。

---

# 三十二、后续可提供内容

如果你需要，我还可以继续提供：

1. 完整可运行 Demo（含 main.go）
2. Gin/Fiber HTTP API
3. Redis 缓存实现
4. Trigger 智能判断
5. Docker Compose
6. 单元测试
7. 数据库表结构
8. WebSocket 聊天集成
9. 多 MCP Server 支持
10. Bifrost 风格 AI Gateway 完整实现
