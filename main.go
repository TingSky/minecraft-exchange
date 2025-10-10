package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// 任务结构体
type Task struct {
	ID          int
	Title       string
	Description string
	Difficulty  string // easy, medium, hard
	Type        string // daily, limited
	Reward      int
	ExpiryTime  string
	Status      string // available, claimed, completed, verified
	PlayerID    *int
	RepeatDays  string // 用于存储日常任务的重复周期，格式为逗号分隔的星期几，如"1,2,3,4,5"
	TemplateID  *int   // 关联的任务模板ID
	CreatedAt   time.Time // 创建时间，用于显示即将开始的任务的开始时间
}

// 任务模板结构体
type TaskTemplate struct {
	ID          int
	Title       string
	Description string
	Difficulty  string // easy, medium, hard
	Type        string // daily, limited
	Reward      int
	RepeatDays  string // 用于存储日常任务的重复周期，格式为逗号分隔的星期几，如"1,2,3,4,5"
	CreatedAt   string
	UpdatedAt   string
}

// 物品结构体
type Item struct {
	ID          int
	Name        string
	Description string
	Cost        int
	Stock       int
	ExpiryTime  string
}

// 兑换记录结构体
type ExchangeRecord struct {
	ID        int
	PlayerID  int
	ItemID    int
	ItemName  string
	Cost      int
	Timestamp string
}

var db *sql.DB

func main() {
	// 初始化数据库
	initDB()

	// 启动日常任务自动刷新机制
	go startDailyTaskRefresh()

	// 设置静态文件服务
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// 创建任务处理器
	http.HandleFunc("/create-task", createTaskHandler)
	// 测试创建任务页面处理器
	http.HandleFunc("/test-create-task", testCreateTaskHandler)
	
	// 路由设置
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/tasks", tasksHandler)
	http.HandleFunc("/shop", shopHandler)
	http.HandleFunc("/exchange", exchangeHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/admin", adminHandler)
	http.HandleFunc("/delete-task", deleteTaskHandler)
	http.HandleFunc("/delete-task-template", deleteTaskTemplateHandler)
	http.HandleFunc("/claim-task", claimTaskHandler)
	http.HandleFunc("/complete-task", completeTaskHandler)
	http.HandleFunc("/verify-task", verifyTaskHandler)

	// 启动服务器
	fmt.Println("服务器已启动，访问 http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// 初始化数据库
func initDB() {
	var err error
	// 检查数据库文件是否存在
	_, err = os.Stat("minecraft_exchange.db")
	if os.IsNotExist(err) {
		// 创建数据库文件
		file, err := os.Create("minecraft_exchange.db")
		if err != nil {
			log.Fatal("无法创建数据库文件:", err)
		}
		file.Close()
	}

	// 从环境变量获取数据库路径，默认为./minecraft_exchange.db
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./minecraft_exchange.db"
	}
	
	// 连接数据库
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("无法连接到数据库:", err)
	}

	// 创建表
	createTables()

	// 初始化一些示例数据
	initSampleData()
}

// 创建数据库表
func createTables() {
	tables := []string{
		// 玩家表
		`CREATE TABLE IF NOT EXISTS players (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			emeralds INTEGER DEFAULT 0
		);`,
		// 任务表
		`CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			difficulty TEXT NOT NULL,
			type TEXT NOT NULL,
			reward INTEGER NOT NULL,
			expiry_time TEXT,
			status TEXT DEFAULT 'available',
			player_id INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (player_id) REFERENCES players(id)
		);`,
		// 物品表
		`CREATE TABLE IF NOT EXISTS items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			cost INTEGER NOT NULL,
			stock INTEGER NOT NULL,
			expiry_time TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		// 兑换记录表
		`CREATE TABLE IF NOT EXISTS exchange_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			player_id INTEGER NOT NULL,
			item_id INTEGER NOT NULL,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (player_id) REFERENCES players(id),
			FOREIGN KEY (item_id) REFERENCES items(id)
		);`,
	}

	for _, table := range tables {
		_, err := db.Exec(table)
		if err != nil {
			log.Fatal("无法创建表:", err)
		}
	}
}

// 初始化示例数据
func initSampleData() {
	// 检查是否已有玩家数据
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM players").Scan(&count)
	if err != nil {
		log.Fatal("查询玩家数据失败:", err)
	}

	if count == 0 {
		// 插入玩家数据
		_, err = db.Exec("INSERT INTO players (name, emeralds) VALUES (?, ?)", "张屹程", 10)
		if err != nil {
			log.Fatal("插入玩家数据失败:", err)
		}

		// 插入任务数据
	tasks := []struct {
			title       string
			description string
			difficulty  string
			taskType    string
			reward      int
			expiryTime  string
		}{{
			"完成数学作业",
			"完成今天的数学作业并检查正确",
			"easy",
			"daily",
			5,
			time.Now().Add(24 * time.Hour).Format("2006-01-02 15:04:05"),
		}, {
			"阅读30分钟",
			"阅读喜欢的书籍30分钟",
			"easy",
			"daily",
			5,
			time.Now().Add(24 * time.Hour).Format("2006-01-02 15:04:05"),
		}, {
			"帮忙做家务",
			"帮助家长打扫房间或洗碗",
			"medium",
			"daily",
			10,
			time.Now().Add(24 * time.Hour).Format("2006-01-02 15:04:05"),
		}, {
			"写一篇短文",
			"写一篇关于你的周末的短文，至少5句话",
			"hard",
			"limited",
			15,
			time.Now().Add(7 * 24 * time.Hour).Format("2006-01-02 15:04:05"),
		}}

		for _, task := range tasks {
			_, err = db.Exec(
				"INSERT INTO tasks (title, description, difficulty, type, reward, expiry_time) VALUES (?, ?, ?, ?, ?, ?)",
				task.title, task.description, task.difficulty, task.taskType, task.reward, task.expiryTime,
			)
			if err != nil {
				log.Fatal("插入任务数据失败:", err)
			}
		}

		// 插入物品数据
		items := []struct {
			name        string
			description string
			cost        int
			stock       int
			expiryTime  string
		}{{
			"小玩具",
			"一个有趣的小玩具",
			10,
			10,
			time.Now().Add(30 * 24 * time.Hour).Format("2006-01-02 15:04:05"),
		}, {
			"漫画书",
			"一本好看的漫画书",
			20,
			5,
			time.Now().Add(30 * 24 * time.Hour).Format("2006-01-02 15:04:05"),
		}, {
			"游戏时间",
			"额外30分钟游戏时间",
			15,
			20,
			time.Now().Add(30 * 24 * time.Hour).Format("2006-01-02 15:04:05"),
		}, {
			"外出游玩",
			"周末去公园玩耍",
			50,
			3,
			time.Now().Add(30 * 24 * time.Hour).Format("2006-01-02 15:04:05"),
		}}

		for _, item := range items {
			_, err = db.Exec(
				"INSERT INTO items (name, description, cost, stock, expiry_time) VALUES (?, ?, ?, ?, ?)",
				item.name, item.description, item.cost, item.stock, item.expiryTime,
			)
			if err != nil {
				log.Fatal("插入物品数据失败:", err)
			}
		}
	}
}

// 获取第一个玩家ID
func getFirstPlayerID() (int, error) {
	var playerID int
	err := db.QueryRow("SELECT id FROM players ORDER BY id LIMIT 1").Scan(&playerID)
	if err != nil {
		return 0, err
	}
	return playerID, nil
}

// 首页处理器
func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "无法加载模板", http.StatusInternalServerError)
		return
	}

	// 获取玩家信息
	var playerName string
	var emeralds int
	playerID, err := getFirstPlayerID()
	if err != nil {
		log.Println("获取玩家ID失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}
	err = db.QueryRow("SELECT name, emeralds FROM players WHERE id = ?", playerID).Scan(&playerName, &emeralds)
	if err != nil {
		log.Println("查询玩家信息失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 传递数据到模板
	data := map[string]interface{}{
		"PlayerName": playerName,
		"Emeralds":   emeralds,
	}

	tmpl.Execute(w, data)
}

// 任务页面处理器
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/tasks.html")
	if err != nil {
		http.Error(w, "无法加载模板", http.StatusInternalServerError)
		return
	}

	// 获取可用任务
	rows, err := db.Query("SELECT id, title, description, difficulty, type, reward, expiry_time, created_at FROM tasks WHERE status = 'available' ORDER BY created_at DESC")
	if err != nil {
		log.Println("查询任务失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Difficulty, &task.Type, &task.Reward, &task.ExpiryTime, &task.CreatedAt)
		if err != nil {
			log.Println("扫描任务数据失败:", err)
			continue
		}
		tasks = append(tasks, task)
	}

	// 获取玩家已领取任务
	claimedRows, err := db.Query("SELECT id, title, description, difficulty, type, reward, expiry_time, created_at FROM tasks WHERE status IN ('claimed', 'completed') AND player_id = 1 ORDER BY updated_at DESC")
	if err != nil {
		log.Println("查询已领取任务失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}
	defer claimedRows.Close()

	var claimedTasks []Task
	for claimedRows.Next() {
		var task Task
		err := claimedRows.Scan(&task.ID, &task.Title, &task.Description, &task.Difficulty, &task.Type, &task.Reward, &task.ExpiryTime, &task.CreatedAt)
		if err != nil {
			log.Println("扫描已领取任务数据失败:", err)
			continue
		}
		claimedTasks = append(claimedTasks, task)
	}

	// 获取玩家信息
	var playerName string
	var emeralds int
	playerID, err := getFirstPlayerID()
	if err != nil {
		log.Println("获取玩家ID失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}
	err = db.QueryRow("SELECT name, emeralds FROM players WHERE id = ?", playerID).Scan(&playerName, &emeralds)
	if err != nil {
		log.Println("查询玩家信息失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 获取即将开始的任务（未来24小时内开始，但现在还不能领取）
	upcomingRows, err := db.Query(`
		SELECT id, title, description, difficulty, type, reward, expiry_time, created_at 
		FROM tasks 
		WHERE status = 'available' 
		AND datetime(created_at) > datetime('now') 
		ORDER BY created_at ASC
	`)
	var upcomingTasks []Task
	if err != nil {
		log.Println("查询即将开始任务失败:", err)
	} else {
		defer upcomingRows.Close()

		for upcomingRows.Next() {
			var task Task
			err := upcomingRows.Scan(&task.ID, &task.Title, &task.Description, &task.Difficulty, &task.Type, &task.Reward, &task.ExpiryTime, &task.CreatedAt)
			if err != nil {
				log.Println("扫描即将开始任务数据失败:", err)
				continue
			}
			upcomingTasks = append(upcomingTasks, task)
		}

		// 从可领取任务中移除即将开始的任务
		availableTasks := []Task{}
		for _, task := range tasks {
			isUpcoming := false
			for _, upcoming := range upcomingTasks {
				if task.ID == upcoming.ID {
					isUpcoming = true
					break
				}
			}
			if !isUpcoming {
				availableTasks = append(availableTasks, task)
			}
		}
		tasks = availableTasks
	}

	// 准备传递给模板的数据
	data := map[string]interface{}{
		"PlayerName":   playerName,
		"Emeralds":     emeralds,
		"Tasks":        tasks,
		"UpcomingTasks": upcomingTasks,
		"ClaimedTasks": claimedTasks,
	}

	// 执行模板渲染
	tmpl.Execute(w, data)
}

// 商店页面处理器
func shopHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/shop.html")
	if err != nil {
		http.Error(w, "无法加载模板", http.StatusInternalServerError)
		return
	}

	// 获取物品列表
	rows, err := db.Query("SELECT id, name, description, cost, stock, expiry_time FROM items WHERE stock > 0 ORDER BY created_at DESC")
	if err != nil {
		log.Println("查询物品失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Cost, &item.Stock, &item.ExpiryTime)
		if err != nil {
			log.Println("扫描物品数据失败:", err)
			continue
		}
		items = append(items, item)
	}

	// 获取玩家信息
	var playerName string
	var emeralds int
	playerID, err := getFirstPlayerID()
	if err != nil {
		log.Println("获取玩家ID失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}
	err = db.QueryRow("SELECT name, emeralds FROM players WHERE id = ?", playerID).Scan(&playerName, &emeralds)
	if err != nil {
		log.Println("查询玩家信息失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 传递数据到模板
	data := map[string]interface{}{
		"PlayerName": playerName,
		"Emeralds":   emeralds,
		"Items":      items,
	}

	tmpl.Execute(w, data)
}

// 兑换处理器
func exchangeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	// 获取物品ID
	itemID := r.FormValue("item_id")
	if itemID == "" {
		http.Error(w, "物品ID不能为空", http.StatusBadRequest)
		return
	}

	// 事务处理兑换
	tx, err := db.Begin()
	if err != nil {
		log.Println("开始事务失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// 获取物品信息
	var itemName string
	var cost int
	var stock int
	err = tx.QueryRow("SELECT name, cost, stock FROM items WHERE id = ?", itemID).Scan(&itemName, &cost, &stock)
	if err != nil {
		log.Println("查询物品信息失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 检查库存
	if stock <= 0 {
		http.Error(w, "物品库存不足", http.StatusBadRequest)
		return
	}

	// 获取玩家绿宝石数量
	var emeralds int
	playerID, err := getFirstPlayerID()
	if err != nil {
		log.Println("获取玩家ID失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}
	err = tx.QueryRow("SELECT emeralds FROM players WHERE id = ?", playerID).Scan(&emeralds)
	if err != nil {
		log.Println("查询玩家信息失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 检查绿宝石是否足够
	if emeralds < cost {
		http.Error(w, "绿宝石不足", http.StatusBadRequest)
		return
	}

	// 扣减玩家绿宝石
	_, err = tx.Exec("UPDATE players SET emeralds = emeralds - ? WHERE id = ?", cost, playerID)
	if err != nil {
		log.Println("扣减绿宝石失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 扣减物品库存
	_, err = tx.Exec("UPDATE items SET stock = stock - ? WHERE id = ?", 1, itemID)
	if err != nil {
		log.Println("扣减库存失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 记录兑换记录
	_, err = tx.Exec("INSERT INTO exchange_records (player_id, item_id) VALUES (?, ?)", 1, itemID)
	if err != nil {
		log.Println("记录兑换失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		log.Println("提交事务失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 重定向到商店页面
	http.Redirect(w, r, "/shop", http.StatusFound)
}

// 生成安全的随机字符串用于Cookie值
func generateSecureToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// 登录页面处理器
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// 获取输入的密码
		password := r.FormValue("password")
		
		// 验证密码（写死为190830zyc）
		if password == "190830zyc" {
			// 生成会话令牌
		sessionToken := generateSecureToken(32)
			
			// 设置Cookie，有效期为7天
			cookie := http.Cookie{
				Name:     "session_token",
				Value:    sessionToken,
				Path:     "/",
				Expires:  time.Now().Add(7 * 24 * time.Hour),
				HttpOnly: true,
				Secure:   false, // 在生产环境中应设为true（使用HTTPS）
			}
			http.SetCookie(w, &cookie)
			
			// 登录成功后重定向到管理页面
			http.Redirect(w, r, "/admin", http.StatusFound)
			return
		} else {
			// 密码错误，显示错误信息
			data := map[string]interface{}{
				"Error": "密码错误，请重新输入",
				"Year":  time.Now().Year(),
			}
			tmpl, _ := template.ParseFiles("templates/login.html")
			tmpl.Execute(w, data)
			return
		}
	}
	
	// GET请求，显示登录页面
	data := map[string]interface{}{
		"Year": time.Now().Year(),
	}
	tmpl, _ := template.ParseFiles("templates/login.html")
	tmpl.Execute(w, data)
}

// 管理员页面处理器
// 测试创建任务页面处理器
func testCreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	// 检查是否已登录
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		// 未登录，重定向到登录页面
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	
	tmpl, err := template.ParseFiles("templates/test-create-task.html")
	if err != nil {
		http.Error(w, "无法加载模板", http.StatusInternalServerError)
		return
	}
	
	// 传递数据到模板
	data := map[string]interface{}{
		"Year": time.Now().Year(),
	}
	
	tmpl.Execute(w, data)
}

// 删除任务处理器
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	// 检查是否已登录
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		// 未登录，重定向到登录页面
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	
	// 确保是POST请求
	if r.Method != "POST" {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}
	
	// 获取任务ID
	taskID := r.FormValue("task_id")
	if taskID == "" {
		http.Error(w, "任务ID不能为空", http.StatusBadRequest)
		return
	}
	
	// 从数据库中删除任务
	_, err = db.Exec("DELETE FROM tasks WHERE id = ?", taskID)
	if err != nil {
		log.Println("删除任务失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}
	
	// 删除成功后重定向回管理员页面
	http.Redirect(w, r, "/admin", http.StatusFound)
}

// 删除任务模板处理器
func deleteTaskTemplateHandler(w http.ResponseWriter, r *http.Request) {
	// 检查是否已登录
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		// 未登录，重定向到登录页面
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	
	// 确保是POST请求
	if r.Method != "POST" {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}
	
	// 获取任务模板ID
	templateID := r.FormValue("template_id")
	if templateID == "" {
		http.Error(w, "任务模板ID不能为空", http.StatusBadRequest)
		return
	}
	
	// 从数据库中删除任务模板
	_, err = db.Exec("DELETE FROM task_templates WHERE id = ?", templateID)
	if err != nil {
		log.Println("删除任务模板失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}
	
	// 删除成功后重定向回管理员页面
	http.Redirect(w, r, "/admin", http.StatusFound)
}

// 领取任务处理器
func claimTaskHandler(w http.ResponseWriter, r *http.Request) {
	// 确保是POST请求
	if r.Method != "POST" {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}
	
	// 获取任务ID
	taskID := r.FormValue("task_id")
	if taskID == "" {
		http.Error(w, "任务ID不能为空", http.StatusBadRequest)
		return
	}
	
	// 事务处理领取任务
	tx, err := db.Begin()
	if err != nil {
		log.Println("开始事务失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()
	
	// 检查任务状态是否为available
	var status string
	err = tx.QueryRow("SELECT status FROM tasks WHERE id = ?", taskID).Scan(&status)
	if err != nil {
		log.Println("查询任务状态失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}
	
	if status != "available" {
		http.Error(w, "该任务已被领取", http.StatusBadRequest)
		return
	}
	
	// 更新任务状态为claimed并关联玩家ID
	_, err = tx.Exec("UPDATE tasks SET status = 'claimed', player_id = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?", taskID)
	if err != nil {
		log.Println("更新任务状态失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}
	
	// 提交事务
	err = tx.Commit()
	if err != nil {
		log.Println("提交事务失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}
	
	// 领取成功后重定向回任务页面
	http.Redirect(w, r, "/tasks", http.StatusFound)
}

// 提交完成任务处理器
func completeTaskHandler(w http.ResponseWriter, r *http.Request) {
	// 确保是POST请求
	if r.Method != "POST" {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}
	
	// 获取任务ID
	taskID := r.FormValue("task_id")
	if taskID == "" {
		http.Error(w, "任务ID不能为空", http.StatusBadRequest)
		return
	}
	
	// 事务处理提交完成任务
	tx, err := db.Begin()
	if err != nil {
		log.Println("开始事务失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()
	
	// 检查任务状态是否为claimed且属于当前玩家
	var status string
	var playerID int
	err = tx.QueryRow("SELECT status, player_id FROM tasks WHERE id = ?", taskID).Scan(&status, &playerID)
	if err != nil {
		log.Println("查询任务状态失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}
	
	if status != "claimed" || playerID != 1 {
		http.Error(w, "你不能提交此任务", http.StatusBadRequest)
		return
	}
	
	// 更新任务状态为completed
	_, err = tx.Exec("UPDATE tasks SET status = 'completed', updated_at = CURRENT_TIMESTAMP WHERE id = ?", taskID)
	if err != nil {
		log.Println("更新任务状态失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}
	
	// 提交事务
	err = tx.Commit()
	if err != nil {
		log.Println("提交事务失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}
	
	// 提交成功后重定向回任务页面
	http.Redirect(w, r, "/tasks", http.StatusFound)
}

// 验证任务完成并发放奖励处理器
func verifyTaskHandler(w http.ResponseWriter, r *http.Request) {
	// 检查是否已登录
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		// 未登录，重定向到登录页面
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	
	// 确保是POST请求
	if r.Method != "POST" {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}
	
	// 获取任务ID
	taskID := r.FormValue("task_id")
	if taskID == "" {
		http.Error(w, "任务ID不能为空", http.StatusBadRequest)
		return
	}
	
	// 事务处理验证任务并发放奖励
	tx, err := db.Begin()
	if err != nil {
		log.Println("开始事务失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()
	
	// 获取任务信息
	var taskReward int
	var taskStatus string
	var playerID int
	err = tx.QueryRow("SELECT reward, status, player_id FROM tasks WHERE id = ?", taskID).Scan(&taskReward, &taskStatus, &playerID)
	if err != nil {
		log.Println("查询任务信息失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}
	
	// 检查任务状态是否为completed
	if taskStatus != "completed" {
		http.Error(w, "该任务未完成，无法验证", http.StatusBadRequest)
		return
	}
	
	// 增加玩家绿宝石数量
	_, err = tx.Exec("UPDATE players SET emeralds = emeralds + ? WHERE id = ?", taskReward, playerID)
	if err != nil {
		log.Println("增加绿宝石失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}
	
	// 更新任务状态为verified
	_, err = tx.Exec("UPDATE tasks SET status = 'verified', updated_at = CURRENT_TIMESTAMP WHERE id = ?", taskID)
	if err != nil {
		log.Println("更新任务状态失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}
	
	// 提交事务
	err = tx.Commit()
	if err != nil {
		log.Println("提交事务失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}
	
	// 验证成功后重定向回管理员页面
	http.Redirect(w, r, "/admin", http.StatusFound)
}

// 创建任务模板处理器
func createTaskHandler(w http.ResponseWriter, r *http.Request) {
	// 检查是否已登录
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		// 未登录，重定向到登录页面
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// 确保是POST请求
	if r.Method != "POST" {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	// 获取表单数据
	title := r.FormValue("title")
	description := r.FormValue("description")
	difficulty := r.FormValue("difficulty")
	taskType := r.FormValue("type")
	rewardStr := r.FormValue("reward")
	expiryTime := r.FormValue("expiry_time")
	// 获取日常任务的重复周期
	err = r.ParseForm()
	if err != nil {
		log.Println("解析表单失败:", err)
	}
	repeatDaysValues := r.Form["repeat_days"]
	var repeatDays string
	if len(repeatDaysValues) > 0 {
		repeatDays = strings.Join(repeatDaysValues, ",")
	}

	// 验证必要字段
	if title == "" || difficulty == "" || taskType == "" || rewardStr == "" {
		http.Error(w, "请填写所有必要字段", http.StatusBadRequest)
		return
	}

	// 日常任务需要设置重复周期
	if taskType == "daily" && repeatDays == "" {
		http.Error(w, "日常任务必须设置重复周期", http.StatusBadRequest)
		return
	}

	// 转换奖励值为整数
	var reward int
	_, err = fmt.Sscanf(rewardStr, "%d", &reward)
	if err != nil || reward <= 0 {
		http.Error(w, "奖励必须是正整数", http.StatusBadRequest)
		return
	}

	// 插入任务模板到数据库
	result, err := db.Exec(
		"INSERT INTO task_templates (title, description, difficulty, type, reward, repeat_days) VALUES (?, ?, ?, ?, ?, ?)",
		title, description, difficulty, taskType, reward, repeatDays,
	)
	if err != nil {
		log.Println("创建任务模板失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 获取刚插入的模板ID
	templateID, err := result.LastInsertId()
	if err != nil {
		log.Println("获取模板ID失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 如果是限时任务，立即创建一个任务实例
	if taskType == "limited" {
		// 为限时任务创建一个任务实例，包含created_at字段
		_, err = db.Exec(
			"INSERT INTO tasks (title, description, difficulty, type, reward, expiry_time, status, template_id, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
			title, description, difficulty, taskType, reward, expiryTime, "available", templateID, time.Now().Format("2006-01-02 15:04:05"),
		)
		if err != nil {
			log.Println("创建限时任务实例失败:", err)
			http.Error(w, "服务器错误", http.StatusInternalServerError)
			return
		}
	} else if taskType == "daily" {
		// 对于日常任务，立即为明天创建任务实例，使其显示在即将开始的任务列表中
		today := time.Now()
		tomorrow := today.Add(24 * time.Hour)
		tomorrowDate := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, tomorrow.Location())
		expiryTime := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 23, 59, 59, 0, tomorrow.Location())
		expiryTimeStr := expiryTime.Format("2006-01-02 15:04:05")
		createdAtStr := tomorrowDate.Format("2006-01-02 15:04:05")

		// 创建明天的任务实例，设置created_at为明天的开始时间
		_, err = db.Exec(
			"INSERT INTO tasks (title, description, difficulty, type, reward, expiry_time, status, repeat_days, template_id, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			title, description, difficulty, "daily", reward, expiryTimeStr, "available", repeatDays, templateID, createdAtStr,
		)
		if err != nil {
			log.Println("创建日常任务实例失败:", err)
			http.Error(w, "服务器错误", http.StatusInternalServerError)
			return
		}
	}

	// 创建成功后重定向回管理员页面
	http.Redirect(w, r, "/admin", http.StatusFound)
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
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

	// 获取所有任务实例，使用COALESCE函数将NULL的repeat_days和template_id转换为默认值
	rows, err := db.Query("SELECT id, title, description, difficulty, type, reward, expiry_time, status, player_id, COALESCE(repeat_days, '') as repeat_days, COALESCE(template_id, 0) as template_id FROM tasks ORDER BY created_at DESC")
	if err != nil {
		log.Println("查询任务失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		var templateID int
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Difficulty, &task.Type, &task.Reward, &task.ExpiryTime, &task.Status, &task.PlayerID, &task.RepeatDays, &templateID)
		if err != nil {
			log.Println("扫描任务数据失败:", err)
			continue
		}
		// 如果template_id不为0，则设置Task的TemplateID字段
		if templateID > 0 {
			task.TemplateID = &templateID
		}
		tasks = append(tasks, task)
	}

	// 获取所有任务模板
	templateRows, err := db.Query("SELECT id, title, description, difficulty, type, reward, COALESCE(repeat_days, '') as repeat_days FROM task_templates ORDER BY created_at DESC")
	if err != nil {
		log.Println("查询任务模板失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}
	defer templateRows.Close()

	var taskTemplates []TaskTemplate
	for templateRows.Next() {
		var template TaskTemplate
		err := templateRows.Scan(&template.ID, &template.Title, &template.Description, &template.Difficulty, &template.Type, &template.Reward, &template.RepeatDays)
		if err != nil {
			log.Println("扫描任务模板数据失败:", err)
			continue
		}
		taskTemplates = append(taskTemplates, template)
	}

	// 获取所有兑换记录
	exchangeRows, err := db.Query(`
		SELECT er.id, er.player_id, er.item_id, i.name, i.cost, er.timestamp 
		FROM exchange_records er 
		JOIN items i ON er.item_id = i.id 
		ORDER BY er.timestamp DESC
	`)
	if err != nil {
		log.Println("查询兑换记录失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}
	defer exchangeRows.Close()

	var exchangeRecords []ExchangeRecord
	for exchangeRows.Next() {
		var record ExchangeRecord
		err := exchangeRows.Scan(&record.ID, &record.PlayerID, &record.ItemID, &record.ItemName, &record.Cost, &record.Timestamp)
		if err != nil {
			log.Println("扫描兑换记录数据失败:", err)
			continue
		}
		exchangeRecords = append(exchangeRecords, record)
	}

	// 传递数据到模板
	data := map[string]interface{}{
		"Tasks":          tasks,
		"TaskTemplates":  taskTemplates,
		"ExchangeRecords": exchangeRecords,
	}

	tmpl.Execute(w, data)
}

// 启动日常任务自动刷新机制
func startDailyTaskRefresh() {
	// 计算距离下一次零点的时间
	now := time.Now()
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	durationUntilMidnight := nextMidnight.Sub(now)
	
	// 第一次执行等待到零点
	time.Sleep(durationUntilMidnight)
	
	// 执行一次刷新
	refreshDailyTasks()
	
	// 之后每24小时执行一次
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
		refreshDailyTasks()
		}
	}
}

// 刷新日常任务
func refreshDailyTasks() {
	log.Println("开始刷新日常任务")
	
	// 获取今天和明天是星期几（0-6，0是周日）
	todayWeekday := int(time.Now().Weekday())
	tomorrowWeekday := (todayWeekday + 1) % 7 // 计算明天的星期几（处理周日的情况）
	
	// 转换为字符串用于比较
	todayStr := strconv.Itoa(todayWeekday)
	tomorrowStr := strconv.Itoa(tomorrowWeekday)
	
	// 从任务模板表中获取所有日常任务模板
	templates, err := db.Query(
		"SELECT id, title, description, difficulty, reward, repeat_days FROM task_templates WHERE type = 'daily'",
	)
	if err != nil {
		log.Println("查询日常任务模板失败:", err)
		return
	}
	defer templates.Close()
	
	// 遍历每个模板，检查是否需要创建今天或明天的任务
	for templates.Next() {
		var templateID int
		var title, description, difficulty, repeatDays string
		var reward int
		err := templates.Scan(&templateID, &title, &description, &difficulty, &reward, &repeatDays)
		if err != nil {
			log.Println("扫描日常任务模板失败:", err)
			continue
		}
		
		// 检查重复周期是否为空
		if repeatDays == "" {
			continue
		}
		
		days := strings.Split(repeatDays, ",")
		
		// 检查是否是今天的任务
		isTodayTask := false
		for _, day := range days {
			if day == todayStr {
				isTodayTask = true
				break
			}
		}
		
		// 检查是否是明天的任务
		isTomorrowTask := false
		for _, day := range days {
			if day == tomorrowStr {
				isTomorrowTask = true
				break
			}
		}
		
		// 处理今天的任务
		if isTodayTask {
			// 设置日常任务过期时间为当天23:59:59
			today := time.Now()
			expiryTime := time.Date(today.Year(), today.Month(), today.Day(), 23, 59, 59, 0, today.Location())
			expiryTimeStr := expiryTime.Format("2006-01-02 15:04:05")
			
			// 检查今天的任务是否已经存在
			var count int
			sqlQuery := "SELECT COUNT(*) FROM tasks WHERE type = 'daily' AND template_id = ? AND date(created_at) = date('now')"
			err = db.QueryRow(sqlQuery, templateID).Scan(&count)
			if err == nil && count == 0 {
				// 创建新的任务实例，包含created_at字段
				sqlInsert := "INSERT INTO tasks (title, description, difficulty, type, reward, expiry_time, status, repeat_days, template_id, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
				createdAt := time.Now().Format("2006-01-02 15:04:05")
				_, err = db.Exec(sqlInsert, title, description, difficulty, "daily", reward, expiryTimeStr, "available", repeatDays, templateID, createdAt)
				if err != nil {
					log.Printf("创建今天的日常任务 '%s' 失败: %v", title, err)
				} else {
					log.Printf("成功创建今天的日常任务 '%s'", title)
				}
			}
		}
		
		// 处理明天的任务（提前一天显示，让用户可以提前看到并领取）
		if isTomorrowTask {
			// 设置明天任务的过期时间为明天23:59:59
			tomorrow := time.Now().Add(24 * time.Hour)
			expiryTime := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 23, 59, 59, 0, tomorrow.Location())
			expiryTimeStr := expiryTime.Format("2006-01-02 15:04:05")
			
			// 检查明天的任务是否已经存在
			var count int
			sqlQuery := "SELECT COUNT(*) FROM tasks WHERE type = 'daily' AND template_id = ? AND date(expiry_time) = date('now', '+1 day')"
			err = db.QueryRow(sqlQuery, templateID).Scan(&count)
			if err == nil && count == 0 {
				// 创建明天的任务实例，设置created_at为明天的开始时间，使其显示在即将开始的任务列表中
				sqlInsert := "INSERT INTO tasks (title, description, difficulty, type, reward, expiry_time, status, repeat_days, template_id, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
				createdAt := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, tomorrow.Location()).Format("2006-01-02 15:04:05")
				_, err = db.Exec(sqlInsert, title, description, difficulty, "daily", reward, expiryTimeStr, "available", repeatDays, templateID, createdAt)
				if err != nil {
					log.Printf("创建明天的日常任务 '%s' 失败: %v", title, err)
				} else {
					log.Printf("成功创建明天的日常任务 '%s'", title)
				}
			}
		}
	}
	
	// 将过期的日常任务标记为已完成
	_, err = db.Exec("UPDATE tasks SET status = 'completed' WHERE type = 'daily' AND status = 'claimed' AND date(updated_at) < date('now')")
	if err != nil {
		log.Println("更新过期日常任务状态失败:", err)
	}
	
	log.Println("日常任务刷新完成")
}