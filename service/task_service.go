package service

import (
	"ai-backend/model"
	"ai-backend/util"
	"errors"
	"sync"
	"time"
)

// 任务状态
const (
	TaskStatusPending    = "pending"
	TaskStatusProcessing = "processing"
	TaskStatusCompleted  = "completed"
	TaskStatusFailed     = "failed"
)

// 任务存储
var (
	taskStore = make(map[string]*model.Task)
	mutex     sync.RWMutex
)

// InitTaskService 初始化任务服务
func InitTaskService() {
	// 启动定时清理过期任务的协程
	go cleanupExpiredTasks()
}

// SubmitTask 提交任务
func SubmitTask(function, content string) (string, error) {
	// 验证功能是否存在
	if !isValidFunction(function) {
		return "", errors.New("无效的功能名称")
	}

	// 创建任务
	taskID := util.GenerateUUID()
	task := &model.Task{
		TaskID:    taskID,
		Status:    TaskStatusPending,
		Function:  function,
		Content:   content,
		CreatedAt: time.Now(),
	}

	// 保存任务
	mutex.Lock()
	taskStore[taskID] = task
	mutex.Unlock()

	// 异步处理任务
	go processTask(taskID)

	return taskID, nil
}

// GetTaskResult 获取任务结果
func GetTaskResult(taskID string) (*model.Task, error) {
	mutex.RLock()
	defer mutex.RUnlock()

	task, exists := taskStore[taskID]
	if !exists {
		return nil, nil
	}

	return task, nil
}

// 处理任务
func processTask(taskID string) {
	mutex.RLock()
	task, exists := taskStore[taskID]
	mutex.RUnlock()

	if !exists {
		return
	}

	// 更新任务状态为处理中
	mutex.Lock()
	task.Status = TaskStatusProcessing
	task.UpdatedAt = time.Now()
	mutex.Unlock()

	// 调用AI服务处理任务
	doubaoService := NewDoubaoService()
	result, err := doubaoService.Call(task.Function, task.Content)

	// 更新任务结果
	mutex.Lock()
	defer mutex.Unlock()

	task.UpdatedAt = time.Now()
	if err != nil {
		task.Status = TaskStatusFailed
		task.Result = err.Error()
	} else {
		task.Status = TaskStatusCompleted
		task.Result = result
	}
}

// 清理过期任务（比如10分钟前的任务）
func cleanupExpiredTasks() {
	ticker := time.NewTicker(10 * time.Minute) // 每10分钟检查一次
	defer ticker.Stop()

	for range ticker.C {
		expirationTime := time.Now().Add(-10 * time.Minute)
		mutex.Lock()

		for taskID, task := range taskStore {
			if task.CreatedAt.Before(expirationTime) {
				delete(taskStore, taskID)
			}
		}

		mutex.Unlock()
	}
}
