package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mereith/nav/logger"
	"github.com/mereith/nav/service"
	"github.com/mereith/nav/types"
)

// ==================== WebDAV 备份相关 Handler ====================

// GetBackupConfigHandler 获取备份配置
func GetBackupConfigHandler(c *gin.Context) {
	config, err := service.GetBackupConfig()
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

// SaveBackupConfigHandler 保存备份配置

// SaveBackupConfigHandler 保存备份配置
func SaveBackupConfigHandler(c *gin.Context) {
	var config types.BackupConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"errorMessage": err.Error(),
		})
		return
	}

	err := service.SaveBackupConfig(&config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":      false,
			"errorMessage": err.Error(),
		})
		return
	}

	// 更新定时调度
	service.UpdateBackupCron()

	c.JSON(200, gin.H{
		"success": true,
		"message": "备份配置已保存",
	})
}

// TestBackupConnectionHandler 测试 WebDAV 连接

// TestBackupConnectionHandler 测试 WebDAV 连接
func TestBackupConnectionHandler(c *gin.Context) {
	var config types.BackupConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"errorMessage": err.Error(),
		})
		return
	}

	err := service.TestWebDAVConnection(&config)
	if err != nil {
		c.JSON(200, gin.H{
			"success":      false,
			"errorMessage": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "WebDAV 连接成功",
	})
}

// BackupNowHandler 立即执行备份

// BackupNowHandler 立即执行备份
func BackupNowHandler(c *gin.Context) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.LogError("手动备份 panic: %v", r)
			}
		}()
		err := service.ExecuteBackup()
		if err != nil {
			logger.LogError("手动备份失败: %s", err)
		}
	}()

	c.JSON(200, gin.H{
		"success": true,
		"message": "备份任务已启动",
	})
}

// GetBackupStatusHandler 获取备份状态

// GetBackupStatusHandler 获取备份状态
func GetBackupStatusHandler(c *gin.Context) {
	status := service.GetBackupStatusForDisplay()
	c.JSON(200, gin.H{
		"success": true,
		"data":    status,
	})
}

// ListBackupFilesHandler 列出 WebDAV 上的备份文件

// ListBackupFilesHandler 列出 WebDAV 上的备份文件
func ListBackupFilesHandler(c *gin.Context) {
	files, err := service.ListBackupFiles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":      false,
			"errorMessage": err.Error(),
		})
		return
	}
	if files == nil {
		files = []service.BackupFileInfo{}
	}
	c.JSON(200, gin.H{
		"success": true,
		"data":    files,
	})
}

// RestoreBackupHandler 从指定备份文件恢复数据库

// RestoreBackupHandler 从指定备份文件恢复数据库
func RestoreBackupHandler(c *gin.Context) {
	var req struct {
		Filename string `json:"filename"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"errorMessage": err.Error(),
		})
		return
	}
	if req.Filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"errorMessage": "文件名不能为空",
		})
		return
	}

	err := service.RestoreFromBackup(req.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":      false,
			"errorMessage": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "数据库恢复成功，请刷新页面以查看最新数据",
	})
}

// ==================== 主题美化配置相关 ====================

// GetThemeConfigHandler 获取主题配置（公开接口，无token返回默认主题）
