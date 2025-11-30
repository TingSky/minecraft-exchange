package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"minecraft-exchange/models"
	"minecraft-exchange/utils"
)

// 获取商店数据的JSON接口
func GetShopDataHandler(w http.ResponseWriter, r *http.Request) {
	// 查询物品列表
	items, err := models.GetAllItems()
	if err != nil {
		log.Println("查询物品失败:", err)
		utils.SendJSONResponse(w, http.StatusInternalServerError, utils.JSONResponse{
			Success: false,
			Message: "服务器错误",
		})
		return
	}

	// 获取第一个玩家ID
	playerID, err := models.GetFirstPlayerID()
	if err != nil {
		log.Println("获取玩家ID失败:", err)
		utils.SendJSONResponse(w, http.StatusInternalServerError, utils.JSONResponse{
			Success: false,
			Message: "服务器错误",
		})
		return
	}

	// 查询玩家信息
	player, err := models.GetPlayerInfo(playerID)
	if err != nil {
		log.Println("查询玩家信息失败:", err)
		utils.SendJSONResponse(w, http.StatusInternalServerError, utils.JSONResponse{
			Success: false,
			Message: "服务器错误",
		})
		return
	}

	// 返回JSON响应
	utils.SendJSONResponse(w, http.StatusOK, utils.JSONResponse{
		Success: true,
		Data: map[string]interface{}{
			"PlayerName": player.Name,
			"Emeralds":   player.Emeralds,
			"Items":      items,
		},
	})
}

// 商店页面处理器
func ShopHandler(w http.ResponseWriter, r *http.Request) {
	// 检查是否为AJAX请求
	if utils.IsAJAXRequest(r) {
		GetShopDataHandler(w, r)
		return
	}

	tmpl, err := template.ParseFiles("templates/shop.html")
	if err != nil {
		http.Error(w, "无法加载模板", http.StatusInternalServerError)
		return
	}

	// 查询物品列表
	items, err := models.GetAllItems()
	if err != nil {
		log.Println("查询物品失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 获取第一个玩家ID
	playerID, err := models.GetFirstPlayerID()
	if err != nil {
		log.Println("获取玩家ID失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 查询玩家信息
	player, err := models.GetPlayerInfo(playerID)
	if err != nil {
		log.Println("查询玩家信息失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 准备传递给模板的数据
	data := map[string]interface{}{
		"PlayerName": player.Name,
		"Emeralds":   player.Emeralds,
		"Items":      items,
	}

	// 执行模板渲染
	tmpl.Execute(w, data)
}

// 兑换处理器
func ExchangeHandler(w http.ResponseWriter, r *http.Request) {
	// 确保是POST请求
	if r.Method != "POST" {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	// 获取表单数据
	itemIDStr := r.FormValue("item_id")
	if itemIDStr == "" {
		http.Error(w, "物品ID不能为空", http.StatusBadRequest)
		return
	}

	// 转换物品ID为整数
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		log.Println("物品ID格式错误:", err)
		http.Error(w, "物品ID格式错误", http.StatusBadRequest)
		return
	}

	// 获取第一个玩家ID
	playerID, err := models.GetFirstPlayerID()
	if err != nil {
		log.Println("获取玩家ID失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 事务处理兑换物品
	tx, err := models.DB.Begin()
	if err != nil {
		log.Println("开始事务失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// 查询物品信息
	item, err := models.GetItemInfo(itemID)
	if err != nil {
		log.Println("查询物品信息失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 检查物品库存
	if item.Stock <= 0 {
		http.Error(w, "物品库存不足", http.StatusBadRequest)
		return
	}

	// 查询玩家信息
	player, err := models.GetPlayerInfo(playerID)
	if err != nil {
		log.Println("查询玩家信息失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 检查绿宝石是否足够
	if player.Emeralds < item.Cost {
		http.Error(w, "绿宝石不足", http.StatusBadRequest)
		return
	}

	// 扣减玩家绿宝石
	newEmeralds := player.Emeralds - item.Cost
	err = models.UpdatePlayerEmeralds(playerID, newEmeralds)
	if err != nil {
		log.Println("扣减绿宝石失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 减少物品库存
	newStock := item.Stock - 1
	err = models.UpdateItemStock(itemID, newStock)
	if err != nil {
		log.Println("减少物品库存失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 记录兑换记录
	err = models.CreateExchangeRecord(playerID, itemID)
	if err != nil {
		log.Println("记录兑换记录失败:", err)
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

	// 检查是否为AJAX请求
	if utils.IsAJAXRequest(r) {
		utils.SendJSONResponse(w, http.StatusOK, utils.JSONResponse{
			Success: true,
			Message: "物品兑换成功",
			Refresh: true,
		})
	} else {
		// 重定向到商店页面
		http.Redirect(w, r, "/shop", http.StatusFound)
	}
}

// 创建物品处理器
func CreateItemHandler(w http.ResponseWriter, r *http.Request) {
	// 检查是否已登录
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		// 未登录，检查是否为AJAX请求
		if utils.IsAJAXRequest(r) {
			utils.SendJSONResponse(w, http.StatusUnauthorized, utils.JSONResponse{
				Success: false,
				Message: "未登录，请先登录",
				Redirect: "/login",
			})
		} else {
			http.Redirect(w, r, "/login", http.StatusFound)
		}
		return
	}

	// 确保是POST请求
	if r.Method != "POST" {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	// 获取表单数据
	name := r.FormValue("name")
	description := r.FormValue("description")
	costStr := r.FormValue("cost")
	stockStr := r.FormValue("stock")
	expiryTimeStr := r.FormValue("expiry_time")

	// 验证表单数据
	if name == "" || costStr == "" || stockStr == "" {
		http.Error(w, "名称、价格和库存不能为空", http.StatusBadRequest)
		return
	}

	// 转换数值
	cost, err := strconv.Atoi(costStr)
	if err != nil || cost <= 0 {
		http.Error(w, "价格必须是正整数", http.StatusBadRequest)
		return
	}

	stock, err := strconv.Atoi(stockStr)
	if err != nil || stock < 0 {
		http.Error(w, "库存必须是非负整数", http.StatusBadRequest)
		return
	}

	// 处理过期时间
	expiryTime := ""
	if expiryTimeStr != "" {
		expiryTime = expiryTimeStr
	} else {
		// 如果未设置过期时间，默认设置为30天后
		expiryTime = time.Now().Add(30 * 24 * time.Hour).Format("2006-01-02 15:04:05")
	}

	// 创建物品
	err = models.CreateItem(name, description, cost, stock, expiryTime)
	if err != nil {
		log.Println("创建物品失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 检查是否为AJAX请求
	if utils.IsAJAXRequest(r) {
		utils.SendJSONResponse(w, http.StatusOK, utils.JSONResponse{
			Success: true,
			Message: "物品创建成功",
			Refresh: true,
		})
	} else {
		// 重定向到管理员页面
		http.Redirect(w, r, "/admin", http.StatusFound)
	}
}

// 删除物品处理器
func DeleteItemHandler(w http.ResponseWriter, r *http.Request) {
	// 检查是否已登录
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		// 未登录，检查是否为AJAX请求
		if utils.IsAJAXRequest(r) {
			utils.SendJSONResponse(w, http.StatusUnauthorized, utils.JSONResponse{
				Success: false,
				Message: "未登录，请先登录",
				Redirect: "/login",
			})
		} else {
			http.Redirect(w, r, "/login", http.StatusFound)
		}
		return
	}

	// 确保是POST请求
	if r.Method != "POST" {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	// 获取物品ID
	itemIDStr := r.FormValue("item_id")
	if itemIDStr == "" {
		http.Error(w, "物品ID不能为空", http.StatusBadRequest)
		return
	}

	// 转换物品ID为整数
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		log.Println("物品ID格式错误:", err)
		http.Error(w, "物品ID格式错误", http.StatusBadRequest)
		return
	}

	// 删除物品
	err = models.DeleteItem(itemID)
	if err != nil {
		log.Println("删除物品失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 检查是否为AJAX请求
	if utils.IsAJAXRequest(r) {
		utils.SendJSONResponse(w, http.StatusOK, utils.JSONResponse{
			Success: true,
			Message: "物品删除成功",
			Refresh: true,
		})
	} else {
		// 重定向到管理员页面
		http.Redirect(w, r, "/admin", http.StatusFound)
	}
}

// 兑换奖励处理器
func ExchangeRewardHandler(w http.ResponseWriter, r *http.Request) {
	// 确保是POST请求
	if r.Method != "POST" {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	// 获取兑换记录ID
	exchangeIDStr := r.FormValue("exchange_id")
	if exchangeIDStr == "" {
		http.Error(w, "兑换记录ID不能为空", http.StatusBadRequest)
		return
	}

	// 转换兑换记录ID为整数
	exchangeID, err := strconv.Atoi(exchangeIDStr)
	if err != nil {
		log.Println("兑换记录ID格式错误:", err)
		http.Error(w, "兑换记录ID格式错误", http.StatusBadRequest)
		return
	}

	// 更新兑换记录状态为已兑换
	err = models.UpdateExchangeRecordStatus(exchangeID, true)
	if err != nil {
		log.Println("更新兑换记录失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 检查是否为AJAX请求
	if utils.IsAJAXRequest(r) {
		utils.SendJSONResponse(w, http.StatusOK, utils.JSONResponse{
			Success: true,
			Message: "奖励兑换成功",
			Refresh: true,
		})
	} else {
		// 处理成功后重定向回管理员页面
		http.Redirect(w, r, "/admin", http.StatusFound)
	}
}

// 更新物品处理器
func UpdateItemHandler(w http.ResponseWriter, r *http.Request) {
	// 检查是否已登录
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		// 未登录，检查是否为AJAX请求
		if utils.IsAJAXRequest(r) {
			utils.SendJSONResponse(w, http.StatusUnauthorized, utils.JSONResponse{
				Success: false,
				Message: "未登录，请先登录",
				Redirect: "/login",
			})
		} else {
			http.Redirect(w, r, "/login", http.StatusFound)
		}
		return
	}

	// 确保是POST请求
	if r.Method != "POST" {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	// 获取表单数据
	itemIDStr := r.FormValue("item_id")
	name := r.FormValue("name")
	description := r.FormValue("description")
	costStr := r.FormValue("cost")
	stockStr := r.FormValue("stock")
	expiryTimeStr := r.FormValue("expiry_time")

	// 验证表单数据
	if itemIDStr == "" || name == "" || costStr == "" || stockStr == "" {
		http.Error(w, "物品ID、名称、价格和库存不能为空", http.StatusBadRequest)
		return
	}

	// 转换数值
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		log.Println("物品ID格式错误:", err)
		http.Error(w, "物品ID格式错误", http.StatusBadRequest)
		return
	}

	cost, err := strconv.Atoi(costStr)
	if err != nil || cost <= 0 {
		http.Error(w, "价格必须是正整数", http.StatusBadRequest)
		return
	}

	stock, err := strconv.Atoi(stockStr)
	if err != nil || stock < 0 {
		http.Error(w, "库存必须是非负整数", http.StatusBadRequest)
		return
	}

	// 处理过期时间
	expiryTime := ""
	if expiryTimeStr != "" {
		expiryTime = expiryTimeStr
	} else {
		// 如果未设置过期时间，默认设置为30天后
		expiryTime = time.Now().Add(30 * 24 * time.Hour).Format("2006-01-02 15:04:05")
	}

	// 更新物品
	err = models.UpdateItem(itemID, name, description, cost, stock, expiryTime)
	if err != nil {
		log.Println("更新物品失败:", err)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 检查是否为AJAX请求
	if utils.IsAJAXRequest(r) {
		utils.SendJSONResponse(w, http.StatusOK, utils.JSONResponse{
			Success: true,
			Message: "物品更新成功",
			Refresh: true,
		})
	} else {
		// 重定向到管理员页面
		http.Redirect(w, r, "/admin", http.StatusFound)
	}
}
