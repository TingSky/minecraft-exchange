package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"minecraft-exchange/models"
	"minecraft-exchange/utils"
)

// 任务页面处理器
func TasksHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/tasks.html")
	if err != nil {
		http.Error(w, "无法加载模板", http.StatusInternalServerError)
		return
	}

	// 获取可用任务
	tasks, err := models.GetAvailableTasks()
	if err != nil {
		log.Println("查询任务失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 获取玩家已领取任务
	playerID, err := models.GetFirstPlayerID()
	if err != nil {
		log.Println("获取玩家ID失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	claimedTasks, err := models.GetPlayerClaimedTasks(playerID)
	if err != nil {
		log.Println("查询已领取任务失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 获取即将开始的任务
	upcomingTasks, err := models.GetUpcomingTasks()
	if err != nil {
		log.Println("查询即将开始任务失败:", err)
	}

	// 获取玩家信息
	player, err := models.GetPlayerInfo(playerID)
	if err != nil {
		log.Println("查询玩家信息失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 准备传递给模板的数据
	data := map[string]interface{}{
		"PlayerName":    player.Name,
		"Emeralds":      player.Emeralds,
		"Tasks":         tasks,
		"UpcomingTasks": upcomingTasks,
		"ClaimedTasks":  claimedTasks,
	}

	// 执行模板渲染
	tmpl.Execute(w, data)
}

// 领取任务处理器
func ClaimTaskHandler(w http.ResponseWriter, r *http.Request) {
	// 确保是POST请求
	if r.Method != "POST" {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	// 获取任务ID
	taskIDStr := r.FormValue("task_id")
	if taskIDStr == "" {
		http.Error(w, "任务ID不能为空", http.StatusBadRequest)
		return
	}

	// 转换任务ID为整数
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		log.Println("任务ID格式错误:", err)
		http.Error(w, "任务ID格式错误", http.StatusBadRequest)
		return
	}

	// 获取第一个玩家ID
	playerID, err := models.GetFirstPlayerID()
	if err != nil {
		log.Println("获取玩家ID失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 使用models包中的ClaimTask函数
	err = models.ClaimTask(taskID, playerID)
	if err != nil {
		log.Println("领取任务失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 领取成功后重定向回任务页面
	http.Redirect(w, r, "/tasks", http.StatusFound)
}

// 提交完成任务处理器
func CompleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	// 确保是POST请求
	if r.Method != "POST" {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	// 获取任务ID
	taskIDStr := r.FormValue("task_id")
	if taskIDStr == "" {
		http.Error(w, "任务ID不能为空", http.StatusBadRequest)
		return
	}

	// 转换任务ID为整数
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		log.Println("任务ID格式错误:", err)
		http.Error(w, "任务ID格式错误", http.StatusBadRequest)
		return
	}

	// 获取第一个玩家ID
	currentPlayerID, err := models.GetFirstPlayerID()
	if err != nil {
		log.Println("获取玩家ID失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 先获取任务信息，检查状态和所有权
	task, err := models.GetTaskByID(taskID)
	if err != nil {
		log.Println("查询任务信息失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	if task.Status != "claimed" || task.PlayerID == nil || *task.PlayerID != currentPlayerID {
		log.Printf("任务提交验证失败: ID=%d, 状态=%s, 任务玩家ID=%v, 当前玩家ID=%d", taskID, task.Status, task.PlayerID, currentPlayerID)
		http.Error(w, "你不能提交此任务", http.StatusBadRequest)
		return
	}

	// 使用models包中的CompleteTask函数
	err = models.CompleteTask(taskID)
	if err != nil {
		log.Println("完成任务失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 记录成功日志
	log.Printf("任务提交成功: ID=%d, 状态已更新为completed", taskID)

	// 如果该任务有模板ID，尝试根据模板创建新的任务实例
	if task.TemplateID != nil && *task.TemplateID > 0 {
		log.Printf("任务 %d 有模板ID %d，尝试创建新的任务实例", taskID, *task.TemplateID)
		// 调用创建任务实例的函数
		err = utils.CreateTaskInstancesFromTemplate(*task.TemplateID, "", "")
		if err != nil {
			// 创建失败不会影响任务完成流程，只记录日志
			log.Printf("根据模板 %d 创建新任务实例失败: %v", *task.TemplateID, err)
		} else {
			log.Printf("根据模板 %d 成功触发任务实例创建流程", *task.TemplateID)
		}
	}

	// 提交成功后重定向回任务页面
	http.Redirect(w, r, "/tasks", http.StatusFound)
}

// 验证任务完成并发放奖励处理器
func VerifyTaskHandler(w http.ResponseWriter, r *http.Request) {
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
	taskIDStr := r.FormValue("task_id")
	if taskIDStr == "" {
		http.Error(w, "任务ID不能为空", http.StatusBadRequest)
		return
	}

	// 转换任务ID为整数
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		log.Println("任务ID格式错误:", err)
		http.Error(w, "任务ID格式错误", http.StatusBadRequest)
		return
	}

	// 先获取任务信息，检查状态和所有权
	task, err := models.GetTaskByID(taskID)
	if err != nil {
		log.Println("查询任务信息失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 检查任务状态是否为completed
	if task.Status != "completed" {
		http.Error(w, "该任务未完成，无法验证", http.StatusBadRequest)
		return
	}

	// 获取玩家当前绿宝石数量
	player, err := models.GetPlayerInfo(*task.PlayerID)
	if err != nil {
		log.Println("查询玩家信息失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 增加玩家绿宝石数量
	newEmeralds := player.Emeralds + task.Reward
	err = models.UpdatePlayerEmeralds(*task.PlayerID, newEmeralds)
	if err != nil {
		log.Println("增加绿宝石失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 验证任务
	err = models.VerifyTask(taskID)
	if err != nil {
		log.Println("验证任务失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 验证成功后重定向回管理员页面
	http.Redirect(w, r, "/admin", http.StatusFound)
}

// 创建任务模板处理器
func CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
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
	startTime := r.FormValue("start_time")

	log.Printf("接收到创建任务请求: title=%s, type=%s, startTime=%s, expiryTime=%s", title, taskType, startTime, expiryTime)
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

	// 创建任务模板结构体
	template := models.TaskTemplate{
		Title:       title,
		Description: description,
		Difficulty:  difficulty,
		Type:        taskType,
		Reward:      reward,
		RepeatDays:  repeatDays,
	}

	// 使用models包中的CreateTaskTemplate函数
	templateID, err := models.CreateTaskTemplate(template)
	if err != nil {
		log.Println("创建任务模板失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 调用创建任务实例的函数，根据模板类型和规则生成相应的任务实例
	err = CreateTaskInstancesFromTemplate(int(templateID), expiryTime, startTime)
	if err != nil {
		log.Println("创建任务实例失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 创建成功后重定向回管理员页面
	http.Redirect(w, r, "/admin", http.StatusFound)
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
		return fmt.Errorf("查询任务模板失败: %w", err)
	}

	// 根据任务类型处理
	if taskType == "daily" {
		// 检查是否存在状态为'available'或'claimed'的派生实例
		var activeCount int
		activeQuery := "SELECT COUNT(*) FROM tasks WHERE template_id = ? AND status IN ('available', 'claimed')"
		err = models.DB.QueryRow(activeQuery, templateID).Scan(&activeCount)
		if err != nil {
			return fmt.Errorf("查询活跃任务实例失败: %w", err)
		}

		// 如果没有活跃实例，根据重复周几设置创建新实例
		if activeCount == 0 {
			// 获取当前时间
			now := time.Now()

			days := strings.Split(repeatDays, ",")

			// 查找未来7天内下一个符合重复周期的日期
			var targetDate time.Time
			found := false

			for i := 0; i < 7; i++ {
				checkDate := now.AddDate(0, 0, i)
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
					return fmt.Errorf("创建日常任务实例失败: %w", err)
				}
				log.Printf("成功创建日常任务 '%s' 实例，开始时间: %s", title, startTimeStr)
			}
		}
	} else if taskType == "limited" {
		// 限时任务：如果没有派生实例，则创建
		var count int
		sqlQuery := "SELECT COUNT(*) FROM tasks WHERE template_id = ?"
		err = models.DB.QueryRow(sqlQuery, templateID).Scan(&count)
		if err != nil {
			return fmt.Errorf("查询任务实例失败: %w", err)
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
					return fmt.Errorf("解析开始时间失败: %w, 原始值: %s", err, startTimeForLimited)
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
				return fmt.Errorf("创建限时任务实例失败: %w", err)
			}
			log.Printf("成功创建限时任务 '%s' 实例，开始时间: %s", title, startTimeStr)
		}
	}

	return nil
}

// 删除任务处理器
func DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
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
	taskIDStr := r.FormValue("task_id")
	if taskIDStr == "" {
		http.Error(w, "任务ID不能为空", http.StatusBadRequest)
		return
	}

	// 转换任务ID为整数
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		log.Println("任务ID格式错误:", err)
		http.Error(w, "任务ID格式错误", http.StatusBadRequest)
		return
	}

	// 使用models包中的DeleteTask函数
	err = models.DeleteTask(taskID)
	if err != nil {
		log.Println("删除任务失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 删除成功后重定向回管理员页面
	http.Redirect(w, r, "/admin", http.StatusFound)
}

// 删除任务模板处理器
func DeleteTaskTemplateHandler(w http.ResponseWriter, r *http.Request) {
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
	templateIDStr := r.FormValue("template_id")
	if templateIDStr == "" {
		http.Error(w, "任务模板ID不能为空", http.StatusBadRequest)
		return
	}

	// 转换任务模板ID为整数
	templateID, err := strconv.Atoi(templateIDStr)
	if err != nil {
		log.Println("任务模板ID格式错误:", err)
		http.Error(w, "任务模板ID格式错误", http.StatusBadRequest)
		return
	}

	// 使用models包中的DeleteTaskTemplate函数
	err = models.DeleteTaskTemplate(templateID)
	if err != nil {
		log.Println("删除任务模板失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 删除成功后重定向回管理员页面
	http.Redirect(w, r, "/admin", http.StatusFound)
}
