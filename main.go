package main

import (
	"log"
	"net/http"

	"minecraft-exchange/handlers"
	"minecraft-exchange/models"
	"minecraft-exchange/utils"
)

func main() {
	// 初始化数据库
	models.InitDB()

	// 启动日常任务自动刷新机制
	utils.StartDailyTaskRefresh()

	// 设置静态文件服务
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// 设置路由
	http.HandleFunc("/", handlers.IndexHandler)
	http.HandleFunc("/tasks", handlers.TasksHandler)
	http.HandleFunc("/claim_task", handlers.ClaimTaskHandler)
	http.HandleFunc("/complete_task", handlers.CompleteTaskHandler)
	http.HandleFunc("/verify_task", handlers.VerifyTaskHandler)
	http.HandleFunc("/shop", handlers.ShopHandler)
	http.HandleFunc("/exchange", handlers.ExchangeHandler)
	http.HandleFunc("/exchange_reward", handlers.ExchangeRewardHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/admin", handlers.AdminHandler)
	http.HandleFunc("/test_create_task", handlers.TestCreateTaskHandler)
	http.HandleFunc("/delete_task", handlers.DeleteTaskHandler)
	http.HandleFunc("/delete_task_template", handlers.DeleteTaskTemplateHandler)
	http.HandleFunc("/create_task", handlers.CreateTaskHandler)
	http.HandleFunc("/create_item", handlers.CreateItemHandler)
	http.HandleFunc("/delete_item", handlers.DeleteItemHandler)

	// 启动HTTP服务器
	log.Println("服务器启动在 http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
