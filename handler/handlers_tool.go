package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mereith/nav/logger"
	"github.com/mereith/nav/service"
	"github.com/mereith/nav/types"
	"github.com/mereith/nav/utils"
)

func ExportToolsHandler(c *gin.Context) {
	tools, err := service.GetAllTool()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "errorMessage": err.Error()})
		return
	}
	c.JSON(200, gin.H{
		"success": true,
		"message": "导出工具成功",
		"data":    tools,
	})
}

func ImportToolsHandler(c *gin.Context) {
	var tools []types.Tool
	err := c.ShouldBindJSON(&tools)
	if err != nil {
		utils.CheckErr(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"errorMessage": err.Error(),
		})
		return
	}
	// 导入所有工具，返回统计信息
	result := service.ImportTools(tools)
	c.JSON(200, gin.H{
		"success":        true,
		"message":        "导入工具成功",
		"tools_imported": result.Imported,
		"tools_skipped":  result.Skipped,
		"categories":     result.Categories,
	})
}

func AddToolHandler(c *gin.Context) {
	// 添加工具
	var data types.AddToolDto
	if err := c.ShouldBindJSON(&data); err != nil {
		utils.CheckErr(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"errorMessage": err.Error(),
		})
		return
	}

	logger.LogInfo("%s 获取 logo: %s", data.Name, data.Logo)
	id, err := service.AddTool(data)
	if err != nil {
		utils.CheckErr(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"errorMessage": err.Error(),
		})
		return
	}
	if data.Logo == "" {
		go service.LazyFetchLogo(data.Url, id)
	}
	c.JSON(200, gin.H{
		"success": true,
		"message": "添加成功",
	})
}

func DeleteToolHandler(c *gin.Context) {
	// 删除工具
	id := c.Param("id")
	err := service.DeleteTool(id)
	if err != nil {
		utils.CheckErr(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":      false,
			"errorMessage": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"success": true,
		"message": "删除成功",
	})
}

func UpdateToolHandler(c *gin.Context) {
	// 更新工具
	var data types.UpdateToolDto
	if err := c.ShouldBindJSON(&data); err != nil {
		utils.CheckErr(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"errorMessage": err.Error(),
		})
		return
	}
	service.UpdateTool(data)
	if data.Logo == "" {
		logger.LogInfo("%s 获取 logo: %s", data.Name, data.Logo)
		go service.LazyFetchLogo(data.Url, int64(data.Id))
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "更新成功",
	})
}

// UpdateToolDescOnlyHandler 只更新工具描述，不修改其他字段

// UpdateToolDescOnlyHandler 只更新工具描述，不修改其他字段
func UpdateToolDescOnlyHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"errorMessage": "无效的ID",
		})
		return
	}

	var body struct {
		Desc string `json:"desc"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"errorMessage": err.Error(),
		})
		return
	}

	err = service.UpdateToolDesc(id, body.Desc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":      false,
			"errorMessage": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "更新描述成功",
	})
}

func UpdateToolsSortHandler(c *gin.Context) {
	var updates []types.UpdateToolsSortDto
	if err := c.ShouldBindJSON(&updates); err != nil {
		utils.CheckErr(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"errorMessage": err.Error(),
		})
		return
	}

	err := service.UpdateToolsSort(updates)
	if err != nil {
		utils.CheckErr(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":      false,
			"errorMessage": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "更新排序成功",
	})
}

// ==================== 搜索引擎相关处理函数 ====================

// 获取所有搜索引擎

// GetMaxSortHandler 获取工具表最大排序值
func GetMaxSortHandler(c *gin.Context) {
	maxSort, err := service.GetMaxSort()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":      false,
			"errorMessage": "获取最大排序失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": types.MaxSortResponse{
			MaxSort: maxSort,
		},
	})
}

// ==================== 导入导出相关 Handler ====================

// ExportConfigHandler 导出所有配置

// ==================== 导入导出相关 Handler ====================

// ExportConfigHandler 导出所有配置
func ExportConfigHandler(c *gin.Context) {
	resp, err := service.ExportFullConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":      false,
			"errorMessage": "导出配置失败: " + err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    resp,
	})
}

// ImportConfigHandler 导入配置

// ImportConfigHandler 导入配置
func ImportConfigHandler(c *gin.Context) {
	var req types.ImportConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"errorMessage": "请求格式无效: " + err.Error(),
		})
		return
	}

	result := service.ImportFullConfig(req)

	c.JSON(200, gin.H{
		"success": true,
		"data":    result,
	})
}

// ==================== 网站健康检测 Handler ====================

// CheckLinksHandler 并发检测所有链接的存活状态

// ==================== 网站健康检测 Handler ====================

// CheckLinksHandler 并发检测所有链接的存活状态
func CheckLinksHandler(c *gin.Context) {
	results, aliveCount, deadCount := service.CheckAllLinks()
	if results == nil {
		c.JSON(200, gin.H{
			"success": true,
			"data": types.LinkCheckResponse{
				Total:   0,
				Alive:   0,
				Dead:    0,
				Results: []types.LinkCheckResult{},
			},
		})
		return
	}

	linkResults := make([]types.LinkCheckResult, len(results))
	for i, r := range results {
		linkResults[i] = types.LinkCheckResult{
			Id:         r.Id,
			Url:        r.Url,
			Title:      r.Title,
			StatusCode: r.StatusCode,
			Alive:      r.Alive,
			Error:      r.Error,
		}
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": types.LinkCheckResponse{
			Total:   len(results),
			Alive:   aliveCount,
			Dead:    deadCount,
			Results: linkResults,
		},
	})
}

// OrganizeDeadLinksHandler 先刷新检测，再将失效链接移至末尾，最后返回更新后的完整数据

// OrganizeDeadLinksHandler 先刷新检测，再将失效链接移至末尾，最后返回更新后的完整数据
func OrganizeDeadLinksHandler(c *gin.Context) {
	affected, tools, catelogs, setting, siteConfig, err := service.OrganizeDeadLinks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":      false,
			"errorMessage": "整理失败: " + err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"affected":   affected,
			"tools":      tools,
			"catelogs":   catelogs,
			"setting":    setting,
			"siteConfig": siteConfig,
		},
		"message": fmt.Sprintf("已整理，%d 条失效链接已移至末尾", affected),
	})
}

// UpdateLanguageHandler 仅更新语言设置，不覆盖其他配置
