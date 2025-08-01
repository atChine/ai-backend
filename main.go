package main

import (
	"ai-backend/config"
	"ai-backend/handler"
	"ai-backend/service"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.Init(); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}

	service.InitTaskService()
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	api := r.Group("/api/v1")
	{
		// 获取所有功能列表
		api.GET("/functions", handler.GetAllFunctions)

		// 调用具体功能（同步）
		api.POST("/ai/call", handler.CallAIFunction)

		// 提交任务（异步）
		api.POST("/ai/task", handler.SubmitTask)

		// 查询任务结果
		api.GET("/ai/task/:taskId", handler.GetTaskResult)

		// 流式调用
		api.POST("/ai/stream", handler.StreamAIFunction)
	}

	port := config.AppConfig.Server.Port
	log.Printf("服务启动成功，监听端口: %s", port)
	log.Fatal(r.Run(":" + port))
}
