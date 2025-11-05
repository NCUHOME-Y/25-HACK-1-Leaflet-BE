package controller

import (
	"net/http"
	"us/config"
	"us/model"

	"github.com/gin-gonic/gin"
)

// 上传问题
func UploadProblem(c *gin.Context) {
	
	
	// 从上下文中获取用户信息
	claims, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户信息未授权"})
		return
	}
	var user model.User
	if err := config.DB.First(&user, claims.(uint)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户不存在"})
		return
	}
	req := model.Problem{
		SenderName: user.Username,
		Response:   0,
		UserID:    claims.(uint),
	}
	// 绑定请求体
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效请求"})
		return
	}
	// 保存问题到数据库
	if err := config.DB.Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "上传成功"})
}
