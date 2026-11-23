package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mereith/nav/logger"
	"github.com/mereith/nav/service"
	"github.com/mereith/nav/types"
	"github.com/mereith/nav/utils"
)

func UpdateSettingHandler(c *gin.Context) {
	var data types.Setting
	if err := c.ShouldBindJSON(&data); err != nil {
		utils.CheckErr(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"errorMessage": err.Error(),
		})
		return
	}
	logger.LogInfo("更新配置: %+v", data)
	err := service.UpdateSetting(data)
	if err != nil {
		utils.CheckErr(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"errorMessage": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"success": true,
		"message": "更新配置成功",
	})
}

func UpdateUserHandler(c *gin.Context) {
	var data types.UpdateUserDto
	if err := c.ShouldBindJSON(&data); err != nil {
		utils.CheckErr(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"errorMessage": err.Error(),
		})
		return
	}
	service.UpdateUser(data)
	c.JSON(200, gin.H{
		"success": true,
		"message": "更新用户成功",
	})
}

func UpdateSiteConfigHandler(c *gin.Context) {
	var data types.SiteConfig
	if err := c.ShouldBindJSON(&data); err != nil {
		utils.CheckErr(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"errorMessage": err.Error(),
		})
		return
	}
	logger.LogInfo("更新网站配置: %+v", data)
	err := service.UpdateSiteConfig(data)
	if err != nil {
		utils.CheckErr(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"errorMessage": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"success": true,
		"message": "更新网站配置成功",
	})
}

// UpdateLanguageHandler 仅更新语言设置，不覆盖其他配置
func UpdateLanguageHandler(c *gin.Context) {
	var req struct {
		Language string `json:"language"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"errorMessage": err.Error(),
		})
		return
	}
	if req.Language != "zh-CN" && req.Language != "en-US" {
		req.Language = "zh-CN"
	}
	err := service.UpdateLanguage(req.Language)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":      false,
			"errorMessage": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"success": true,
		"message": "语言设置已更新",
	})
}

// GetDeploymentVersionHandler 获取当前部署版本号

// GetDeploymentVersionHandler 获取当前部署版本号
func GetDeploymentVersionHandler(c *gin.Context) {
	version, err := service.GetDeploymentVersion()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "errorMessage": err.Error()})
		return
	}
	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"version": version,
		},
	})
}

// IncrementDeploymentVersionHandler 递增部署版本号（供部署 agent 调用）

// IncrementDeploymentVersionHandler 递增部署版本号（供部署 agent 调用）
func IncrementDeploymentVersionHandler(c *gin.Context) {
	newVersion, err := service.IncrementDeploymentVersion()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":      false,
			"errorMessage": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"version": newVersion,
		},
		"message": "版本号已递增",
	})
}

// ==================== WebDAV 备份相关 Handler ====================

// GetBackupConfigHandler 获取备份配置

// ==================== 主题美化配置相关 ====================

// GetThemeConfigHandler 获取主题配置（公开接口，无token返回默认主题）
func GetThemeConfigHandler(c *gin.Context) {
	config, err := service.GetThemeConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":      false,
			"errorMessage": err.Error(),
		})
		return
	}
	
	c.JSON(200, gin.H{
		"success": true,
		"data":    config,
	})
}

// UpdateThemeConfigHandler 更新主题配置（管理员接口）

// UpdateThemeConfigHandler 更新主题配置（管理员接口）
func UpdateThemeConfigHandler(c *gin.Context) {
	var config types.ThemeConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"errorMessage": "请求参数格式错误: " + err.Error(),
		})
		return
	}
	
	if err := service.SaveThemeConfig(config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"errorMessage": err.Error(),
		})
		return
	}
	
	c.JSON(200, gin.H{
		"success": true,
		"message": "主题配置已保存",
	})
}

// ResetThemeConfigHandler 重置主题配置为默认值（管理员接口）

// ResetThemeConfigHandler 重置主题配置为默认值（管理员接口）
func ResetThemeConfigHandler(c *gin.Context) {
	if err := service.ResetThemeConfig(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":      false,
			"errorMessage": err.Error(),
		})
		return
	}
	
	c.JSON(200, gin.H{
		"success": true,
		"message": "已重置为默认主题",
	})
}
