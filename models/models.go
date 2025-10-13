package models

import (
	"database/sql"
	"log"
	"os"
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
	TemplateID  *int      // 关联的任务模板ID
	CreatedAt   time.Time // 创建时间
	StartTime   string    // 任务开始时间
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
	Exchanged bool
}

// 玩家结构体
type Player struct {
	ID       int
	Name     string
	Emeralds int
}

var DB *sql.DB

// 初始化数据库
func InitDB() {
	var err error
	// 从环境变量获取数据库路径，默认为./minecraft_exchange.db
	DBPath := os.Getenv("DATABASE_PATH")
	if DBPath == "" {
		DBPath = "./minecraft_exchange.db"
	}

	// 检查数据库文件是否存在
	_, err = os.Stat(DBPath)
	if os.IsNotExist(err) {
		// 创建数据库文件
		file, err := os.Create(DBPath)
		if err != nil {
			log.Fatal("无法创建数据库文件:", err)
		}
		file.Close()
	}

	// 连接数据库
	DB, err = sql.Open("sqlite3", DBPath)
	if err != nil {
		log.Fatal("无法连接到数据库:", err)
	}

	// 创建表
	CreateTables()

	// 初始化一些示例数据
	InitSampleData()
}

// 创建数据库表
func CreateTables() {
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
			start_time TEXT,
			status TEXT DEFAULT 'available',
			player_id INTEGER,
			template_id INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (player_id) REFERENCES players(id)
		);`,
		// 任务模板表
		`CREATE TABLE IF NOT EXISTS task_templates (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			difficulty TEXT NOT NULL,
			type TEXT NOT NULL,
			reward INTEGER NOT NULL,
			repeat_days TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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
			exchanged BOOLEAN DEFAULT FALSE,
			exchanged_at TIMESTAMP,
			FOREIGN KEY (player_id) REFERENCES players(id),
			FOREIGN KEY (item_id) REFERENCES items(id)
		);`,
	}

	for _, table := range tables {
		_, err := DB.Exec(table)
		if err != nil {
			log.Fatal("无法创建表:", err)
		}
	}
}

// 初始化示例数据
func InitSampleData() {
	// 检查是否已有玩家数据
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM players").Scan(&count)
	if err != nil {
		log.Fatal("查询玩家数据失败:", err)
	}

	if count == 0 {
		// 插入玩家数据
		_, err = DB.Exec("INSERT INTO players (name, emeralds) VALUES (?, ?)", "张屹程", 10)
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
			_, err = DB.Exec(
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
			_, err = DB.Exec(
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
func GetFirstPlayerID() (int, error) {
	var playerID int
	err := DB.QueryRow("SELECT id FROM players ORDER BY id LIMIT 1").Scan(&playerID)
	if err != nil {
		return 0, err
	}
	return playerID, nil
}

// 获取玩家信息
func GetPlayerInfo(playerID int) (Player, error) {
	var player Player
	err := DB.QueryRow("SELECT id, name, emeralds FROM players WHERE id = ?", playerID).Scan(&player.ID, &player.Name, &player.Emeralds)
	if err != nil {
		return player, err
	}
	return player, nil
}

// 更新玩家绿宝石数量
func UpdatePlayerEmeralds(playerID int, emeralds int) error {
	_, err := DB.Exec("UPDATE players SET emeralds = ? WHERE id = ?", emeralds, playerID)
	return err
}

// 获取可用任务
func GetAvailableTasks() ([]Task, error) {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	rows, err := DB.Query(`SELECT id, title, description, difficulty, type, reward, expiry_time, created_at, start_time FROM tasks WHERE status = 'available' AND ((start_time IS NULL OR start_time <= ?) AND expiry_time > ?) ORDER BY created_at DESC`, currentTime, currentTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		var startTime sql.NullString
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Difficulty, &task.Type, &task.Reward, &task.ExpiryTime, &task.CreatedAt, &startTime)
		if err != nil {
			log.Println("扫描任务数据失败:", err)
			continue
		}
		if startTime.Valid {
			task.StartTime = startTime.String
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// 获取玩家已领取的任务
func GetPlayerClaimedTasks(playerID int) ([]Task, error) {
	rows, err := DB.Query("SELECT id, title, description, difficulty, type, reward, expiry_time, created_at, start_time, status FROM tasks WHERE status IN ('claimed', 'completed') AND player_id = ? ORDER BY updated_at DESC", playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		var startTime sql.NullString
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Difficulty, &task.Type, &task.Reward, &task.ExpiryTime, &task.CreatedAt, &startTime, &task.Status)
		if err != nil {
			log.Println("扫描已领取任务数据失败:", err)
			continue
		}
		if startTime.Valid {
			task.StartTime = startTime.String
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// 获取即将开始的任务
func GetUpcomingTasks() ([]Task, error) {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	rows, err := DB.Query(`SELECT id, title, description, difficulty, type, reward, expiry_time, created_at, start_time FROM tasks WHERE status = 'available' AND start_time > ? ORDER BY start_time ASC`, currentTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		var startTime sql.NullString
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Difficulty, &task.Type, &task.Reward, &task.ExpiryTime, &task.CreatedAt, &startTime)
		if err != nil {
			log.Println("扫描即将开始任务数据失败:", err)
			continue
		}
		if startTime.Valid {
			task.StartTime = startTime.String
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// 获取所有物品
func GetAllItems() ([]Item, error) {
	rows, err := DB.Query("SELECT id, name, description, cost, stock, expiry_time FROM items WHERE stock > 0 ORDER BY created_at DESC")
	if err != nil {
		return nil, err
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
	return items, nil
}

// 获取所有兑换记录
func GetAllExchangeRecords() ([]ExchangeRecord, error) {
	rows, err := DB.Query(`
		SELECT er.id, er.player_id, er.item_id, i.name, i.cost, er.timestamp, er.exchanged 
		FROM exchange_records er 
		JOIN items i ON er.item_id = i.id 
		ORDER BY er.timestamp DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exchangeRecords []ExchangeRecord
	for rows.Next() {
		var record ExchangeRecord
		err := rows.Scan(&record.ID, &record.PlayerID, &record.ItemID, &record.ItemName, &record.Cost, &record.Timestamp, &record.Exchanged)
		if err != nil {
			log.Println("扫描兑换记录数据失败:", err)
			continue
		}
		exchangeRecords = append(exchangeRecords, record)
	}
	return exchangeRecords, nil
}

// 获取所有任务模板
func GetAllTaskTemplates() ([]TaskTemplate, error) {
	rows, err := DB.Query("SELECT id, title, description, difficulty, type, reward, COALESCE(repeat_days, '') as repeat_days FROM task_templates ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var taskTemplates []TaskTemplate
	for rows.Next() {
		var template TaskTemplate
		err := rows.Scan(&template.ID, &template.Title, &template.Description, &template.Difficulty, &template.Type, &template.Reward, &template.RepeatDays)
		if err != nil {
			log.Println("扫描任务模板数据失败:", err)
			continue
		}
		taskTemplates = append(taskTemplates, template)
	}
	return taskTemplates, nil
}

// 获取所有任务
func GetAllTasks() ([]Task, error) {
	rows, err := DB.Query("SELECT id, title, description, difficulty, type, reward, expiry_time, status, player_id, COALESCE(template_id, 0) as template_id FROM tasks ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		var templateID int
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Difficulty, &task.Type, &task.Reward, &task.ExpiryTime, &task.Status, &task.PlayerID, &templateID)
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
	return tasks, nil
}

// 根据任务类型获取任务模板
func GetAllTaskTemplatesByType(taskType string) ([]TaskTemplate, error) {
	rows, err := DB.Query("SELECT id, title, description, difficulty, type, reward, COALESCE(repeat_days, '') as repeat_days FROM task_templates WHERE type = ? ORDER BY created_at DESC", taskType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var taskTemplates []TaskTemplate
	for rows.Next() {
		var template TaskTemplate
		err := rows.Scan(&template.ID, &template.Title, &template.Description, &template.Difficulty, &template.Type, &template.Reward, &template.RepeatDays)
		if err != nil {
			log.Println("扫描任务模板数据失败:", err)
			continue
		}
		taskTemplates = append(taskTemplates, template)
	}
	return taskTemplates, nil
}

// 获取物品信息
func GetItemInfo(itemID int) (Item, error) {
	var item Item
	err := DB.QueryRow("SELECT id, name, description, cost, stock FROM items WHERE id = ?", itemID).Scan(&item.ID, &item.Name, &item.Description, &item.Cost, &item.Stock)
	if err != nil {
		return item, err
	}
	return item, nil
}

// 创建物品
func CreateItem(name, description string, cost, stock int, expiryTime string) error {
	localTime := time.Now().Format("2006-01-02 15:04:05")
	_, err := DB.Exec(
		"INSERT INTO items (name, description, cost, stock, expiry_time, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		name, description, cost, stock, expiryTime, localTime,
	)
	return err
}

// 删除物品
func DeleteItem(itemID int) error {
	_, err := DB.Exec("DELETE FROM items WHERE id = ?", itemID)
	return err
}

// 更新物品库存
func UpdateItemStock(itemID int, stock int) error {
	_, err := DB.Exec("UPDATE items SET stock = ? WHERE id = ?", stock, itemID)
	return err
}

// 创建兑换记录
func CreateExchangeRecord(playerID int, itemID int) error {
	_, err := DB.Exec("INSERT INTO exchange_records (player_id, item_id) VALUES (?, ?)", playerID, itemID)
	return err
}

// 更新兑换记录状态
func UpdateExchangeRecordStatus(recordID int, exchanged bool) error {
	// 简化函数，只更新exchanged列，避免依赖不存在的exchanged_at列
	if exchanged {
		localTime := time.Now().Format("2006-01-02 15:04:05")
		_, err := DB.Exec("UPDATE exchange_records SET exchanged = ?, exchanged_at = ? WHERE id = ?", exchanged, localTime, recordID)
		return err
	} else {
		_, err := DB.Exec("UPDATE exchange_records SET exchanged = ? WHERE id = ?", exchanged, recordID)
		return err
	}
}

// 领取任务
func ClaimTask(taskID int, playerID int) error {
	localTime := time.Now().Format("2006-01-02 15:04:05")
	_, err := DB.Exec("UPDATE tasks SET status = 'claimed', player_id = ?, updated_at = ? WHERE id = ? AND status = 'available'", playerID, localTime, taskID)
	return err
}

// 完成任务
func CompleteTask(taskID int) error {
	localTime := time.Now().Format("2006-01-02 15:04:05")
	_, err := DB.Exec("UPDATE tasks SET status = 'completed', updated_at = ? WHERE id = ? AND status = 'claimed'", localTime, taskID)
	return err
}

// 验证任务
func VerifyTask(taskID int) error {
	localTime := time.Now().Format("2006-01-02 15:04:05")
	_, err := DB.Exec("UPDATE tasks SET status = 'verified', updated_at = ? WHERE id = ? AND status = 'completed'", localTime, taskID)
	return err
}

// 创建任务
func CreateTask(task Task) error {
	localTime := time.Now().Format("2006-01-02 15:04:05")
	_, err := DB.Exec(
		"INSERT INTO tasks (title, description, difficulty, type, reward, expiry_time, start_time, template_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		task.Title, task.Description, task.Difficulty, task.Type, task.Reward, task.ExpiryTime, task.StartTime, task.TemplateID, localTime, localTime,
	)
	return err
}

// 删除任务
func DeleteTask(taskID int) error {
	_, err := DB.Exec("DELETE FROM tasks WHERE id = ?", taskID)
	return err
}

// 删除任务模板
func DeleteTaskTemplate(templateID int) error {
	_, err := DB.Exec("DELETE FROM task_templates WHERE id = ?", templateID)
	return err
}

// 创建任务模板
func CreateTaskTemplate(template TaskTemplate) (int64, error) {
	localTime := time.Now().Format("2006-01-02 15:04:05")
	result, err := DB.Exec(
		"INSERT INTO task_templates (title, description, difficulty, type, reward, repeat_days, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		template.Title, template.Description, template.Difficulty, template.Type, template.Reward, template.RepeatDays, localTime, localTime,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// 根据ID获取任务
func GetTaskByID(taskID int) (Task, error) {
	var task Task
	err := DB.QueryRow("SELECT id, title, description, difficulty, type, reward, expiry_time, status, player_id, COALESCE(template_id, 0) as template_id FROM tasks WHERE id = ?", taskID).Scan(&task.ID, &task.Title, &task.Description, &task.Difficulty, &task.Type, &task.Reward, &task.ExpiryTime, &task.Status, &task.PlayerID, &task.TemplateID)
	if err != nil {
		return task, err
	}
	return task, nil
}