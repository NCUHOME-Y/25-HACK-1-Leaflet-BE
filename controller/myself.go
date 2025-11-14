package controller

import (
	"net/http"

	"github.com/NCUHOME-Y/25-HACK-1-Leaflet-BE/config"
	"github.com/NCUHOME-Y/25-HACK-1-Leaflet-BE/consts"
	"github.com/NCUHOME-Y/25-HACK-1-Leaflet-BE/model"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// 获取预设头像列表和用户当前头像（现在也不用这个功能了，之前是考虑存头像，但是前端那边直接弄完了，不需要过数据库了）
func GetProfilePicture(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var user model.User
	if err := config.DB.First(&user, currentUserID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户不存在"})
		return
	}

	// 查询用户当前头像
	var myself model.Myself
	err := config.DB.Where("user_id = ?", currentUserID).First(&myself).Error

	if err != nil {
		// 用户还没有设置头像，先设置成第一张
		c.JSON(http.StatusOK, gin.H{
			"current_avatar": gin.H{
				"id":  1,
				"url": consts.ProfilePictures[0].URL,
			},
			"profile_pictures": consts.ProfilePictures,
		})
		return
	}

	// 返回用户当前头像和所有预设头像
	c.JSON(http.StatusOK, gin.H{
		"current_avatar": gin.H{
			"id":  myself.ProfilePictureID,
			"url": myself.URL,
		},
		"profile_pictures": consts.ProfilePictures,
	})
}

// 更新头像
func UpdateProfilePicture(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var user model.User
	if err := config.DB.First(&user, currentUserID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户不存在"})
		return
	}

	var req struct {
		ProfilePictureID uint `json:"profile_picture_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 验证头像ID是否正确
	var avatarURL string
	found := false
	for _, avatar := range consts.ProfilePictures {
		if avatar.ID == req.ProfilePictureID {
			avatarURL = avatar.URL
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的头像ID"})
		consts.Logger.WithFields(logrus.Fields{
			"username":           user.Username,
			"user_id":            currentUserID,
			"profile_picture_id": req.ProfilePictureID,
			"action":             "invalid_avatar_id",
		}).Warn("用户尝试使用无效的头像ID")
		return
	}

	// 检查用户是否已有头像记录
	var myself model.Myself
	err := config.DB.Where("user_id = ?", currentUserID).First(&myself).Error

	if err != nil {
		// 用户还没有头像记录，创建新记录
		err = config.DB.Create(&model.Myself{
			UserID:           currentUserID.(uint),
			URL:              avatarURL,
			ProfilePictureID: req.ProfilePictureID,
		}).Error

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建头像记录失败", "details": err.Error()})
			consts.Logger.WithFields(logrus.Fields{
				"username": user.Username,
				"user_id":  currentUserID,
				"action":   "create_avatar",
				"error":    err.Error(),
			}).Error("创建用户头像记录失败")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":    "头像设置成功",
			"avatar_url": avatarURL,
		})

		consts.Logger.WithFields(logrus.Fields{
			"username":           user.Username,
			"user_id":            currentUserID,
			"profile_picture_id": req.ProfilePictureID,
			"action":             "create_avatar",
		}).Info("用户首次设置头像成功")
		return
	}

	// 用户已有头像记录，更新现有记录
	/*err = config.DB.Model(&myself).Updates(gin.H{
		"url":                avatarURL,
		"profile_picture_id": req.ProfilePictureID,
	}).Error8    这里是我的原来写的，然后会出问题，AI说这里是map (gin.H) 来更新记录时，GORM 需要处理这个 map 键值对。由于您在 model.Myself 中嵌入了 gorm.Model（包含 ID, CreatedAt, UpdatedAt, DeletedAt 四个字段），GORM 在处理 map 时有时会错误地尝试将 map 中的键映射到这些内部字段，或者在类型推断上发生错误*/

	err = config.DB.Model(&myself).Updates(struct {
		URL              string
		ProfilePictureID uint
	}{
		URL:              avatarURL,
		ProfilePictureID: req.ProfilePictureID,
	}).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新头像失败", "details": err.Error()})
		consts.Logger.WithFields(logrus.Fields{
			"username": user.Username,
			"user_id":  currentUserID,
			"action":   "update_avatar",
			"error":    err.Error(),
		}).Error("更新用户头像失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "头像更新成功",
		"avatar_url": avatarURL,
	})

	consts.Logger.WithFields(logrus.Fields{
		"username":           user.Username,
		"user_id":            currentUserID,
		"profile_picture_id": req.ProfilePictureID,
		"action":             "update_avatar",
	}).Info("用户更新头像成功")
}

// 头像部分一直到这，上面都是没用到的
func CalculateLevel(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var user model.User
	if err := config.DB.First(&user, currentUserID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户不存在"})
		return
	}

	// 计算用户等级
	var treelevel int64
	err := config.DB.Model(&model.Status{}).Where("user_id = ?", currentUserID).Count(&treelevel).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法计算等级", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"level": (treelevel / 3) + 1}) //返回等级
}

func UpdateName(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var req struct {
		UserName string `json:"user_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	var user model.User
	if err := config.DB.First(&user, currentUserID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户不存在"})
		return
	}

	oldUsername := user.Username
	user.Username = req.UserName
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户名失败", "details": err.Error()})
		consts.Logger.WithFields(logrus.Fields{
			"username":     oldUsername,
			"user_id":      currentUserID,
			"new_username": req.UserName,
			"action":       "update_username",
			"error":        err.Error(),
		}).Error("更新用户名失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户名更新成功"})

	// 记录成功事件
	consts.Logger.WithFields(logrus.Fields{
		"old_username": oldUsername,
		"new_username": req.UserName,
		"user_id":      currentUserID,
		"action":       "update_username",
	}).Info("用户更新用户名成功")
} //这个改名字也不用了
