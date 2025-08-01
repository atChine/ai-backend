package model

import (
	"time"
)

// BaseResponse 基础响应结构
type BaseResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Function 功能信息
type Function struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// FunctionResult 功能调用结果
type FunctionResult struct {
	Result string `json:"result"`
}

// Task 任务信息
type Task struct {
	TaskID    string    `json:"task_id"`
	Status    string    `json:"status"` // pending, processing, completed, failed
	Function  string    `json:"function"`
	Content   string    `json:"content,omitempty"`
	Result    string    `json:"result,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// StreamData 流式返回数据
type StreamData struct {
	Chunk  string `json:"chunk"`
	Finish bool   `json:"finish"`
}
