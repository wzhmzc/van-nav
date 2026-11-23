package handler

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/mereith/nav/service"
	"github.com/mereith/nav/utils"
)

// 预编译正则表达式（避免请求热路径上重复编译）
var (
	reTitle            = regexp.MustCompile(`(?i)<title[^>]*>([^<]+)</title>`)
	reCharset          = regexp.MustCompile(`(?i)charset=([^\s;]+)`)
	reMetaCharset      = regexp.MustCompile(`(?i)<meta[^>]+charset=["']?([^\s"'>]+)`)
	reMetaCharsetHTTP  = regexp.MustCompile(`(?i)<meta[^>]+http-equiv=["']?Content-Type["']?[^>]+content=["']?[^"]*charset=([^\s"'>]+)`)
	reMetaNameFirst    = regexp.MustCompile(`(?i)<meta[^>]+name=["']description["'][^>]+content=["']([^"']+)["']`)
	reMetaContentFirst = regexp.MustCompile(`(?i)<meta[^>]+content=["']([^"']+)["'][^>]+name=["']description["']`)
)

func GetAllHandler(c *gin.Context) {
	// 使用带 TTL 缓存的全量数据获取，降低 DB 查询频率
	cached, err := service.GetAllDataCached()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "errorMessage": err.Error()})
		return
	}
	tools := cached.Tools
	catelogs := cached.Catelogs
	isLogin := utils.IsLogin(c)
	if !isLogin {
		// 过滤掉隐藏工具
		tools = utils.FilterHideTools(tools, catelogs)
	}
	if !isLogin {
		// 过滤掉隐藏分类
		catelogs = utils.FilterHideCates(catelogs)
	}
	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"tools":      tools,
			"catelogs":   catelogs,
			"setting":    cached.Setting,
			"siteConfig": cached.SiteConfig,
		},
	})
}

// decodeRawDataURI 还原 data: URI 中可能出现的 percent 编码字符。
// 只覆盖 data URI 会遇到的字符集，避免与局部变量 url 的包名作用域冲突。

func GetAdminAllDataHandler(c *gin.Context) {
	// 管理员获取全部数据，还有个用户名。
	tools, err := service.GetAllTool()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "errorMessage": err.Error()})
		return
	}
	catelogs, err := service.GetAllCatelog()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "errorMessage": err.Error()})
		return
	}
	setting, err := service.GetSetting()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "errorMessage": err.Error()})
		return
	}
	siteConfig, err := service.GetSiteConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "errorMessage": err.Error()})
		return
	}
	tokens, err := service.GetApiTokens()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "errorMessage": err.Error()})
		return
	}
	userId, ok := c.Get("uid")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"errorMessage": "不存在该用户！",
		})
		return
	}
	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"tools":      tools,
			"catelogs":   catelogs,
			"setting":    setting,
			"siteConfig": siteConfig,
			"user": gin.H{
				"name": c.GetString("username"),
				"id":   userId,
			},
			"tokens": tokens,
		},
	})
}

func ManifestHandler(c *gin.Context) {

	setting, err := service.GetSetting()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "errorMessage": err.Error()})
		return
	}
	title := setting.Title

	var icons = []gin.H{}

	logo192 := setting.Logo192
	if logo192 == "" {
		logo192 = "logo192.png"
	}

	logo512 := setting.Logo512
	if logo512 == "" {
		logo512 = "logo512.png"
	}

	icons = append(icons, gin.H{
		"src":   logo192,
		"type":  "image/png",
		"sizes": "192x192",
	})
	icons = append(icons, gin.H{
		"src":   logo512,
		"type":  "image/png",
		"sizes": "512x512",
	})

	if title == "" {
		title = "Van nav"
	}
	c.JSON(200, gin.H{
		"short_name":       title,
		"name":             title,
		"icons":            icons,
		"start_url":        "/",
		"display":          "standalone",
		"scope":            "/",
		"theme_color":      "#000000",
		"background_color": "#ffffff",
	})
}
