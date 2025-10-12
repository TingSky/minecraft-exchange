package handlers

import (
	"html/template"
	"log"
	"net/http"

	"minecraft-exchange/models"
)

// 首页处理器
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "无法加载模板", http.StatusInternalServerError)
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
	}

	// 执行模板渲染
	tmpl.Execute(w, data)
}