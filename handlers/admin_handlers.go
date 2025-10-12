package handlers

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"minecraft-exchange/models"
	"minecraft-exchange/utils"
)

// 管理员页面处理器
func AdminHandler(w http.ResponseWriter, r *http.Request) {
	// 检查是否已登录
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		// 未登录，重定向到登录页面
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	tmpl, err := template.ParseFiles("templates/admin.html")
	if err != nil {
		http.Error(w, "无法加载模板", http.StatusInternalServerError)
		return
	}

	// 查询所有任务
	tasks, err := models.GetAllTasks()
	if err != nil {
		log.Println("查询任务失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 查询所有任务模板
	taskTemplates, err := models.GetAllTaskTemplates()
	if err != nil {
		log.Println("查询任务模板失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 查询所有兑换记录
	exchangeRecords, err := models.GetAllExchangeRecords()
	if err != nil {
		log.Println("查询兑换记录失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 查询所有物品
	items, err := models.GetAllItems()
	if err != nil {
		log.Println("查询物品失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 准备传递给模板的数据
	data := map[string]interface{}{
		"Tasks":            tasks,
		"TaskTemplates":    taskTemplates,
		"ExchangeRecords":  exchangeRecords,
		"Items":            items,
	}

	// 执行模板渲染
	tmpl.Execute(w, data)
}

// 测试创建任务页面处理器
func TestCreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	// 检查是否已登录
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		// 未登录，重定向到登录页面
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	tmpl, err := template.ParseFiles("templates/create_task.html")
	if err != nil {
		http.Error(w, "无法加载模板", http.StatusInternalServerError)
		return
	}

	// 执行模板渲染
	tmpl.Execute(w, nil)
}

// 登录页面处理器
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// 处理登录请求
		password := r.FormValue("password")
		if password == "admin123" {
			// 登录成功，生成会话token
			sessionToken := utils.GenerateSecureToken(32)

			// 设置Cookie，有效期为1小时
			expiration := time.Now().Add(1 * time.Hour)
			cookie := http.Cookie{
				Name:     "session_token",
				Value:    sessionToken,
				Path:     "/",
				Expires:  expiration,
				HttpOnly: true,
			}
			http.SetCookie(w, &cookie)

			// 重定向到管理员页面
			http.Redirect(w, r, "/admin", http.StatusFound)
			return
		} else {
			// 登录失败，显示错误信息
			tmpl, _ := template.ParseFiles("templates/login.html")
			data := map[string]interface{}{
				"Error": "密码错误",
			}
			tmpl.Execute(w, data)
			return
		}
	}

	// 显示登录页面
	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
		http.Error(w, "无法加载模板", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, nil)
}