package utils

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"minecraft-exchange/models"
)

// JSONResponse 是通用的JSON响应结构体
type JSONResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message,omitempty"`
	Data     any    `json:"data,omitempty"`
	Redirect string `json:"redirect,omitempty"`
	Refresh  bool   `json:"refresh,omitempty"`
}

// SendJSONResponse 发送JSON响应
func SendJSONResponse(w http.ResponseWriter, statusCode int, response JSONResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// IsAJAXRequest 检查是否为AJAX请求
func IsAJAXRequest(r *http.Request) bool {
	// 检查X-Requested-With头
	if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
		return true
	}
	// 检查Accept头是否包含application/json
	if strings.Contains(r.Header.Get("Accept"), "application/json") {
		return true
	}
	return false
}

// 生成安全随机字符串的函数
func GenerateSecureToken(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

// 启动日常任务自动刷新机制
func StartDailyTaskRefresh() {
	// 计算下一个零点的时间
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Add(24 * time.Hour)
	duration := next.Sub(now)

	// 创建一个定时器，在下次零点触发
	timer := time.NewTimer(duration)
	go func() {
		for {
			select {
			case <-timer.C:
				// 刷新日常任务
				RefreshDailyTasks()

				// 设置下一个24小时的定时器
				timer.Reset(24 * time.Hour)
			}
		}
	}()
}

// 创建任务实例的函数，可被createTaskHandler和定时任务调用
func CreateTaskInstancesFromTemplate(templateID int, expiryTimeForLimited string, startTimeForLimited string) error {
	log.Printf("创建任务实例: templateID=%d, expiryTime=%s, startTime=%s", templateID, expiryTimeForLimited, startTimeForLimited)
	// 查询模板信息
	var title, description, difficulty, taskType, repeatDays string
	var reward int
	query := "SELECT title, description, difficulty, type, reward, repeat_days FROM task_templates WHERE id = ?"
	err := models.DB.QueryRow(query, templateID).Scan(&title, &description, &difficulty, &taskType, &reward, &repeatDays)
	if err != nil {
		return err
	}

	// 根据任务类型处理
	if taskType == "daily" {
		// 获取当前时间
		now := time.Now()

		// 查询该模板今天或之后创建的最新任务实例
		var latestInstanceDate sql.NullString
		latestQuery := "SELECT MAX(expiry_time) FROM tasks WHERE template_id = ? AND expiry_time >= ?"
		currentDateStr := now.Format("2006-01-02") + " 00:00:00"
		err = models.DB.QueryRow(latestQuery, templateID, currentDateStr).Scan(&latestInstanceDate)
		if err != nil {
			return err
		}

		days := strings.Split(repeatDays, ",")

		// 确定开始查找的日期：如果有今天或之后的实例，从该实例日期的下一天开始查找
		var startDate time.Time
		if latestInstanceDate.Valid && latestInstanceDate.String != "" {
			return nil
		} else {
			// 如果没有今天或之后的实例，从现在开始查找
			startDate = now
		}

		// 查找未来7天内下一个符合重复周期的日期
		var targetDate time.Time
		found := false

		for i := 0; i < 7; i++ {
			checkDate := startDate.AddDate(0, 0, i)
			checkWeekday := int(checkDate.Weekday())
			checkWeekdayStr := strconv.Itoa(checkWeekday)

			for _, day := range days {
				if day == checkWeekdayStr {
					targetDate = checkDate
					found = true
					break
				}
			}
			if found {
				break
			}
		}

		if found {
			// 设置任务过期时间为目标日期的23:59:59
			expiryTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 23, 59, 59, 0, targetDate.Location())
			expiryTimeStr := expiryTime.Format("2006-01-02 15:04:05")

			// 设置任务开始时间为目标日期的00:00:00
			startTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location())
			startTimeStr := startTime.Format("2006-01-02 15:04:05")

			// 创建任务结构体
			task := models.Task{
				Title:       title,
				Description: description,
				Difficulty:  difficulty,
				Type:        "daily",
				Reward:      reward,
				ExpiryTime:  expiryTimeStr,
				Status:      "available",
				TemplateID:  &templateID,
				CreatedAt:   time.Now(),
				StartTime:   startTimeStr,
			}

			// 使用models包中的CreateTask函数
			err = models.CreateTask(task)
			if err != nil {
				return err
			}
			log.Printf("成功创建日常任务 '%s' 实例，开始时间: %s", title, startTimeStr)
		}
	} else if taskType == "limited" {
		// 限时任务：如果没有派生实例，则创建
		var count int
		sqlQuery := "SELECT COUNT(*) FROM tasks WHERE template_id = ?"
		err = models.DB.QueryRow(sqlQuery, templateID).Scan(&count)
		if err != nil {
			return err
		}

		if count == 0 {
			// 为限时任务创建一个任务实例，包含created_at和start_time字段
			var startTimeStr string
			if startTimeForLimited != "" {
				// 使用用户设置的开始时间，格式化成2006-01-02 15:04:05类型
				// 尝试解析用户输入的时间字符串
				log.Printf("尝试解析用户提交的开始时间: %s", startTimeForLimited)

				// 尝试多种常见格式解析
				parsedTime, err := time.Parse("2006-01-02 15:04:05", startTimeForLimited)
				if err != nil {
					// 尝试带T的ISO格式
					parsedTime, err = time.Parse("2006-01-02T15:04:05", startTimeForLimited)
				}
				if err != nil {
					// 尝试datetime-local格式（不带秒）
					parsedTime, err = time.Parse("2006-01-02T15:04", startTimeForLimited)
				}

				if err != nil {
					// 如果所有解析都失败，记录详细错误信息
					log.Printf("解析开始时间失败: %v, 原始值: %s", err, startTimeForLimited)
					// 这里不自动使用当前时间，而是返回错误，确保用户知道开始时间设置有问题
					return err
				} else {
					// 解析成功，格式化为标准格式
					startTimeStr = parsedTime.Format("2006-01-02 15:04:05")
				}
				log.Printf("成功解析并格式化用户设置的开始时间: %s", startTimeStr)
			} else {
				// 如果没有设置开始时间，则使用当前时间
				startTimeStr = time.Now().Format("2006-01-02 15:04:05")
				log.Printf("没有设置开始时间，使用当前时间: %s", startTimeStr)
			}

			log.Printf("准备创建限时任务: title=%s, start_time=%s, expiry_time=%s", title, startTimeStr, expiryTimeForLimited)

			// 创建任务结构体
			task := models.Task{
				Title:       title,
				Description: description,
				Difficulty:  difficulty,
				Type:        taskType,
				Reward:      reward,
				ExpiryTime:  expiryTimeForLimited,
				Status:      "available",
				TemplateID:  &templateID,
				CreatedAt:   time.Now(),
				StartTime:   startTimeStr,
			}

			// 使用models包中的CreateTask函数
			err = models.CreateTask(task)
			if err != nil {
				return err
			}
			log.Printf("成功创建限时任务 '%s' 实例，开始时间: %s", title, startTimeStr)
		}
	}

	return nil
}

// 刷新日常任务的函数
// 刷新日常任务的函数
func RefreshDailyTasks() {
	log.Println("开始刷新日常任务")

	// 查询所有日常任务模板
	taskTemplates, err := models.GetAllTaskTemplatesByType("daily")
	if err != nil {
		log.Println("查询日常任务模板失败:", err)
		return
	}

	// 处理每个日常任务模板
	for _, template := range taskTemplates {
		// 调用创建任务实例的函数，根据模板创建新的日常任务实例
		err := CreateTaskInstancesFromTemplate(template.ID, "", "")
		if err != nil {
			log.Printf("刷新任务模板ID %d 失败: %v", template.ID, err)
		}
	}

	// 更新过期的任务状态 - 使用Go代码中的本地时间
	localTime := time.Now().Format("2006-01-02 15:04:05")
	_, err = models.DB.Exec("UPDATE tasks SET status = 'expired' WHERE expiry_time < ? AND status NOT IN ('completed', 'verified', 'expired')", localTime)
	if err != nil {
		log.Println("更新过期任务状态失败:", err)
		return
	}

	log.Println("日常任务刷新完成")
}
