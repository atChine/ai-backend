package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"os"
	"strings"
)

type LLMService interface {
	Call(function, content string) (string, error)
	Stream(function, content string) (chan string, error)
}

// 实现服务接口
type DoubaoService struct{}

func NewDoubaoService() LLMService {
	return &DoubaoService{}
}

func (s *DoubaoService) Call(function, content string) (string, error) {
	return CallLLM(function, content)
}

func (s *DoubaoService) Stream(function, content string) (chan string, error) {
	return StreamLLM(function, content)
}

// 初始化豆包客户端（全局只初始化一次）
var doubaoClient *arkruntime.Client
var modelID string

func init() {
	// 从环境变量获取API密钥初始化客户端
	apiKey := os.Getenv("ARK_API_KEY")
	if apiKey == "" {
		panic("请设置ARK_API_KEY环境变量")
	}
	modelID = "doubao-1-5-pro-32k-250115"
	doubaoClient = arkruntime.NewClientWithApiKey(apiKey)
}

// CallLLM 调用豆包SDK实现同步AI功能
func CallLLM(function, content string) (string, error) {
	if doubaoClient == nil {
		return "", errors.New("豆包客户端未初始化")
	}
	prompt := getPromptByFunction(function, content)
	if prompt == "" {
		return "", errors.New("不支持的功能")
	}

	ctx := context.Background()
	req := model.ChatCompletionRequest{
		// 将推理接入点 <Model>替换为 Model ID
		Model: modelID,
		Messages: []*model.ChatCompletionMessage{
			{
				Role: model.ChatMessageRoleUser,
				Content: &model.ChatCompletionMessageContent{
					StringValue: volcengine.String(prompt),
				},
			},
		},
	}
	resp, err := doubaoClient.CreateChatCompletion(ctx, req)
	if err != nil {
		fmt.Printf("standard chat error: %v\n", err)
		return "", fmt.Errorf("豆包API调用失败: %v", err)
	}
	fmt.Println(*resp.Choices[0].Message.Content.StringValue)

	return *resp.Choices[0].Message.Content.StringValue, nil
}

// StreamLLM 调用豆包SDK实现流式AI功能
func StreamLLM(function, content string) (chan string, error) {
	if doubaoClient == nil {
		return nil, errors.New("豆包客户端未初始化")
	}

	// 根据不同功能构造提示词
	prompt := getPromptByFunction(function, content)
	if prompt == "" {
		return nil, errors.New("不支持的功能")
	}

	// 创建上下文
	ctx := context.Background()

	// 使用与CallLLM相同的ChatCompletionRequest结构体
	req := model.ChatCompletionRequest{
		Model: modelID, // 替换为实际的豆包模型ID
		Messages: []*model.ChatCompletionMessage{
			{
				Role: model.ChatMessageRoleUser,
				Content: &model.ChatCompletionMessageContent{
					StringValue: volcengine.String(prompt),
				},
			},
		},
		Stream: true,
	}

	// 调用豆包流式API
	stream, err := doubaoClient.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("豆包流式API调用失败: %v", err)
	}

	ch := make(chan string)

	go func() {
		defer close(ch)
		defer stream.Close()

		// 读取流式响应
		for {
			resp, err := stream.Recv()
			if err != nil {
				if errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "EOF") {
					// 正常结束
					return
				}
				// 将错误发送到通道
				ch <- fmt.Sprintf("[ERROR] %v", err)
				return
			}

			// 解析流式响应内容
			if len(resp.Choices) > 0 {
				// 检查Delta字段（部分响应可能包含在Delta中）
				if &resp.Choices[0].Delta != nil &&
					&resp.Choices[0].Delta.Content != nil &&
					resp.Choices[0].Delta.Content != "" {
					// 获取增量内容
					delta := resp.Choices[0].Delta.Content
					if delta != "" {
						ch <- delta
					}
				}
			}
		}
	}()

	return ch, nil
}

// 根据功能类型生成对应的提示词
func getPromptByFunction(function, content string) string {
	switch function {
	case "translate_zh_to_en":
		return fmt.Sprintf("请将以下中文翻译成英文，保持原意准确，只回答我翻译的内容：\n%s", content)
	case "translate_en_to_zh":
		return fmt.Sprintf("请将以下英文翻译成中文，保持原意准确，只回答我翻译的内容：\n%s", content)
	case "summarize":
		return fmt.Sprintf("请简要总结以下内容的核心观点，控制在原长度的10%%以内：\n%s", content)
	default:
		return ""
	}
}
