package handler

import (
	"ai-backend/model"
	"ai-backend/service"
	"ai-backend/util"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetAllFunctions 获取所有功能列表
func GetAllFunctions(c *gin.Context) {
	functions := service.GetAllFunctions()
	c.JSON(http.StatusOK, model.BaseResponse{
		Code:    0,
		Message: "success",
		Data:    functions,
	})
}

// CallAIFunction 调用AI功能（同步）
func CallAIFunction(c *gin.Context) {
	var req model.FunctionCallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.BaseResponse{
			Code:    1,
			Message: "无效的请求参数: " + err.Error(),
		})
		return
	}

	result, err := service.CallAI(req.Function, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.BaseResponse{
			Code:    2,
			Message: "调用AI失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.BaseResponse{
		Code:    0,
		Message: "success",
		Data:    model.FunctionResult{Result: result},
	})
}

// SubmitTask 提交异步任务
func SubmitTask(c *gin.Context) {
	var req model.TaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.BaseResponse{
			Code:    1,
			Message: "无效的请求参数: " + err.Error(),
		})
		return
	}

	taskID, err := service.SubmitTask(req.Function, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.BaseResponse{
			Code:    2,
			Message: "提交任务失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.BaseResponse{
		Code:    0,
		Message: "任务已提交",
		Data:    gin.H{"task_id": taskID},
	})
}

// GetTaskResult 查询任务结果
func GetTaskResult(c *gin.Context) {
	taskID := c.Param("taskId")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, model.BaseResponse{
			Code:    1,
			Message: "taskId不能为空",
		})
		return
	}

	task, err := service.GetTaskResult(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.BaseResponse{
			Code:    2,
			Message: "查询任务失败: " + err.Error(),
		})
		return
	}

	if task == nil {
		c.JSON(http.StatusNotFound, model.BaseResponse{
			Code:    3,
			Message: "任务不存在",
		})
		return
	}

	c.JSON(http.StatusOK, model.BaseResponse{
		Code:    0,
		Message: "success",
		Data:    task,
	})
}

// StreamAIFunction 流式调用AI功能
// StreamAIFunction 流式调用AI功能，符合前端交互格式
func StreamAIFunction(c *gin.Context) {
	var req model.FunctionCallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.BaseResponse{
			Code:    1,
			Message: "无效的请求参数: " + err.Error(),
		})
		return
	}

	// 设置流式响应头，确保UTF-8编码避免乱码
	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Content-Type-Options", "nosniff")

	// 创建通道接收流式结果
	streamChan := make(chan string)
	doneChan := make(chan bool)
	errorChan := make(chan error)

	// 异步调用AI服务
	go service.StreamAI(req.Function, req.Content, streamChan, doneChan, errorChan)

	// 向客户端发送流式数据
	for {
		select {
		case chunk := <-streamChan:
			jsonChunk, err := json.Marshal(map[string]string{
				"type": "text",
				"msg":  chunk,
			})
			if err != nil {
				// 处理JSON序列化错误
				errorJSON := fmt.Sprintf(`{"type":"error","msg":"序列化错误: %s"}`, escapeJSON(err.Error()))
				c.Writer.WriteString("data: " + errorJSON + "\n\n")
				c.Writer.Flush()
				return
			}

			// 使用标准JSON序列化替代手动字符串拼接，避免编码问题
			if _, err := c.Writer.WriteString("data: " + string(jsonChunk) + "\n\n"); err != nil {
				return
			}
			c.Writer.Flush()
		case <-doneChan:
			// 发送完成信号和元数据
			completionMsg, _ := json.Marshal(map[string]string{"type": "text"})
			c.Writer.WriteString("data: " + string(completionMsg) + "\n\n")

			metaMsg, _ := json.Marshal(map[string]interface{}{
				"type":      "meta",
				"messageId": util.GenerateUUID(),
				"index":     1,
				"tokenUsageInfo": map[string]int{
					"promptTokens":     0,
					"completionTokens": 0,
					"totalTokens":      0,
				},
			})
			c.Writer.WriteString("data: " + string(metaMsg) + "\n\n")
			c.Writer.WriteString("data: [DONE]\n\n")
			c.Writer.Flush()
			return
		case err := <-errorChan:
			// 发送错误信息
			errorJSON := fmt.Sprintf(`{"type":"error","msg":"%s"}`, escapeJSON(err.Error()))
			c.Writer.WriteString("data: " + errorJSON + "\n\n")
			c.Writer.WriteString("data: [DONE]\n\n")
			c.Writer.Flush()
			return
		case <-c.Request.Context().Done():
			// 客户端断开连接
			return
		}
	}
}

// escapeJSON 转义JSON特殊字符，防止格式错误
func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\b", "\\b")
	s = strings.ReplaceAll(s, "\f", "\\f")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}
