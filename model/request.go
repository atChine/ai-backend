package model

// FunctionCallRequest AI功能调用请求
type FunctionCallRequest struct {
	Function string `json:"function" binding:"required"` // 功能名称：translate_zh_to_en, translate_en_to_zh, summarize
	Content  string `json:"content" binding:"required"`  // 要处理的内容
}

// TaskRequest 任务提交请求
type TaskRequest struct {
	Function string `json:"function" binding:"required"` // 功能名称
	Content  string `json:"content" binding:"required"`  // 要处理的内容
}
