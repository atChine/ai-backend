package service

import (
	"ai-backend/model"
	"errors"
	"time"
)

// 所有支持的功能
var supportedFunctions = []model.Function{
	{
		Name:        "translate_zh_to_en",
		Description: "中文翻译成英文",
	},
	{
		Name:        "translate_en_to_zh",
		Description: "英文翻译成中文",
	},
	{
		Name:        "summarize",
		Description: "文本总结",
	},
}

// GetAllFunctions 获取所有支持的功能列表
func GetAllFunctions() []model.Function {
	return supportedFunctions
}

// 验证功能是否有效
func isValidFunction(function string) bool {
	for _, f := range supportedFunctions {
		if f.Name == function {
			return true
		}
	}
	return false
}

// 同步调用AI功能
func CallAI(function, content string) (string, error) {
	if !isValidFunction(function) {
		return "", errors.New("不支持的功能")
	}
	doubaoService := NewDoubaoService()
	// 调用大模型（伪代码）
	//result, err := callLLM(function, content)
	result, err := doubaoService.Call(function, content)
	if err != nil {
		return "", err
	}

	return result, nil
}

// StreamAI 流式调用AI功能
func StreamAI(function, content string, streamChan chan string, doneChan chan bool, errorChan chan error) {
	defer close(streamChan)
	defer close(doneChan)
	defer close(errorChan)

	if !isValidFunction(function) {
		errorChan <- errors.New("不支持的功能")
		return
	}
	doubaoService := NewDoubaoService()
	//streamChunks, err := streamLLM(function, content) //伪代码
	streamChunks, err := doubaoService.Stream(function, content)
	if err != nil {
		errorChan <- err
		return
	}

	// 将流式结果发送到通道
	for chunk := range streamChunks {
		streamChan <- chunk
	}

	doneChan <- true
}

// 调用大模型（伪代码实现）
func callLLM(function, content string) (string, error) {
	// 模拟API调用延迟
	time.Sleep(1 * time.Second)

	// 根据不同功能返回模拟结果
	switch function {
	case "translate_zh_to_en":
		return "Simulated English translation: " + content, nil
	case "translate_en_to_zh":
		return "模拟中文翻译: " + content, nil
	case "summarize":
		return "Simulated summary of content: " + content[:minInt(len(content), 50)] + "...", nil
	default:
		return "", errors.New("不支持的功能")
	}
}

// 流式调用大模型（伪代码实现）
func streamLLM(function, content string) (chan string, error) {
	ch := make(chan string)

	go func() {
		defer close(ch)

		// 模拟流式返回
		var result string
		switch function {
		case "translate_zh_to_en":
			result = "Streamed English translation: " + content
		case "translate_en_to_zh":
			result = "流式中文翻译: " + content
		case "summarize":
			result = "Streamed summary: " + content
		default:
			return
		}

		// 关键改进：按符文(rune)处理，确保中文字符不被截断
		// 因为中文字符在UTF-8中占多个字节，直接按字节分割会导致乱码
		runes := []rune(result)
		chunkSize := 10 // 每个块大约10个字符（无论中英文）
		for i := 0; i < len(runes); i += chunkSize {
			end := i + chunkSize
			if end > len(runes) {
				end = len(runes)
			}
			// 将rune切片转换回字符串，确保完整字符
			ch <- string(runes[i:end])
			time.Sleep(200 * time.Millisecond)
		}
	}()

	return ch, nil
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
