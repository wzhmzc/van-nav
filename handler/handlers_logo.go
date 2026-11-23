package handler

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/mereith/nav/service"
	"github.com/mereith/nav/types"
	"golang.org/x/text/encoding/simplifiedchinese"
)

// decodeRawDataURI 还原 data: URI 中可能出现的 percent 编码字符。
// 只覆盖 data URI 会遇到的字符集，避免与局部变量 url 的包名作用域冲突。
func decodeRawDataURI(raw string) string {
	return strings.NewReplacer(
		"%2B", "+", "%2F", "/", "%3D", "=", "%3B", ";",
		"%3A", ":", "%2C", ",", "%26", "&", "%25", "%",
	).Replace(raw)
}

// serveDataURI 直接返回 data: URI 内嵌的图片，不经过 nav_img 缓存查找。
// 后台旧版前端会把 base64 logo 原样拼进 query（不做 URL 编码），
// 其中 + 经 URL 解码后变成空格，这里还原；否则会落到「缓存未命中 → 灰色占位图」。

// serveDataURI 直接返回 data: URI 内嵌的图片，不经过 nav_img 缓存查找。
// 后台旧版前端会把 base64 logo 原样拼进 query（不做 URL 编码），
// 其中 + 经 URL 解码后变成空格，这里还原；否则会落到「缓存未命中 → 灰色占位图」。
func serveDataURI(c *gin.Context, dataURI string) {
	comma := strings.IndexByte(dataURI, ',')
	if comma < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "errorMessage": "data URI 格式错误"})
		return
	}
	meta := dataURI[5:comma]
	payload := dataURI[comma+1:]
	mime := strings.Split(meta, ";")[0]
	if !strings.HasPrefix(mime, "image/") {
		mime = "image/png"
	}
	var raw []byte
	var err error
	if strings.Contains(meta, "base64") {
		payload = strings.ReplaceAll(payload, " ", "+")
		raw, err = base64.StdEncoding.DecodeString(payload)
		if err != nil {
			raw, err = base64.RawStdEncoding.DecodeString(strings.TrimRight(payload, "="))
		}
	} else {
		raw = []byte(payload)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "errorMessage": "图片解码失败"})
		return
	}
	if mime == "image/svg+xml" {
		c.Header("Cache-Control", "public, max-age=604800")
	}
	c.Data(http.StatusOK, mime, raw)
}

func GetLogoImgHandler(c *gin.Context) {
	url := c.Query("url")
	// Go 1.17+ 把 ; 当成非法 query 分隔符：未编码的 data: URI（形如 data:image/svg+xml;base64,…）
	// 会让整段 query 解析失败被丢弃，url 变空。这里从原始 RawQuery 兜底还原，兼容旧版前端。
	if url == "" {
		if rq := c.Request.URL.RawQuery; strings.HasPrefix(rq, "url=data:") {
			url = decodeRawDataURI(strings.TrimPrefix(rq, "url="))
		}
	}
	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"errorMessage": "URL参数不能为空",
		})
		return
	}
	// data: URI 单独处理，避免走缓存查询后返回占位图
	if strings.HasPrefix(url, "data:") {
		serveDataURI(c, url)
		return
	}
	img, err := service.GetImgFromDB(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "errorMessage": err.Error()})
		return
	}
	if img.Value == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"success":      false,
			"errorMessage": "未找到图片",
		})
		return
	}
	imgBuffer, err := base64.StdEncoding.DecodeString(img.Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":      false,
			"errorMessage": "图片解码失败",
		})
		return
	}
	l := strings.Split(url, ".")
	if len(l) == 0 || (len(l) == 1 && l[0] == "") {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"errorMessage": "无效的URL格式",
		})
		return
	}
	suffix := l[len(l)-1]
	t := "image/x-icon"
	if suffix == "svg" || strings.Contains(url, ".svg") {
		t = "image/svg+xml"
	} else if suffix == "png" {
		t = "image/png"
	}
	// 直接输出二进制数据，避免string转换导致的内存多分配
	c.Data(http.StatusOK, t, imgBuffer)
}

func GetFaviconFromApiHandler(c *gin.Context) {
	var req FaviconRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"errorMessage": "请提供有效的工具网址",
		})
		return
	}

	// 获取站点配置
	siteConfig, err := service.GetSiteConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "errorMessage": err.Error()})
		return
	}

	// 获取 API 地址模板，始终可用
	apiTemplate := siteConfig.FaviconApiTemplate
	if apiTemplate == "" {
		apiTemplate = "https://favicon.im/{domain}"
	}

	// 从 URL 中提取域名
	domain := extractDomain(req.URL)
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"errorMessage": "无法从网址中提取域名",
		})
		return
	}

	// 替换模板中的 {domain}
	logoUrl := replaceDomain(apiTemplate, domain)

	c.JSON(200, gin.H{
		"success":  true,
		"logoUrl":  logoUrl,
		"message":  "获取成功",
	})
}

// extractDomain 从 URL 中提取主域名

// extractDomain 从 URL 中提取主域名
func extractDomain(urlStr string) string {
	// 如果 URL 没有协议头，添加 http://
	if len(urlStr) > 0 && !(urlStr[:7] == "http://" || urlStr[:8] == "https://") {
		urlStr = "http://" + urlStr
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	host := u.Host
	if host == "" {
		return ""
	}

	// 如果包含端口，去掉端口
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	return host
}

// replaceDomain 替换模板中的 {domain}

// replaceDomain 替换模板中的 {domain}
func replaceDomain(template, domain string) string {
	result := strings.ReplaceAll(template, "{domain}", domain)
	return result
}

// FetchPageInfoHandler 获取页面标题和描述

// FetchPageInfoHandler 获取页面标题和描述
func FetchPageInfoHandler(c *gin.Context) {
	urlStr := c.Query("url")
	if urlStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"errorMessage": "url 参数不能为空",
		})
		return
	}

	// 如果 URL 没有协议头，添加 https://
	if len(urlStr) > 0 && !(urlStr[:7] == "http://" || urlStr[:8] == "https://") {
		urlStr = "https://" + urlStr
	}

	// SSRF 防护：校验目标地址不是内网/私有地址
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"errorMessage": "无效的 URL 格式",
		})
		return
	}
	host := parsedURL.Hostname()
	ip := net.ParseIP(host)
	if ip != nil && (ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()) {
		c.JSON(http.StatusForbidden, gin.H{
			"success":      false,
			"errorMessage": "禁止请求内网地址",
		})
		return
	}

	// 浏览器 User-Agent，模拟真实浏览器请求
	browserUA := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // codeql[go/disabled-certificate-check] — 页面信息抓取需兼容自签名证书
	}

	// 最多重试 1 次
	maxRetries := 1
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest("GET", urlStr, nil)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success":      false,
				"errorMessage": "构建请求失败: " + err.Error(),
			})
			return
		}
		req.Header.Set("User-Agent", browserUA)
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			// 网络错误可以重试
			if attempt < maxRetries {
				time.Sleep(time.Second)
				continue
			}
			c.JSON(http.StatusOK, gin.H{
				"success":      false,
				"errorMessage": "请求失败: " + err.Error(),
			})
			return
		}
		defer resp.Body.Close()

		// 处理 429 限速
		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter := resp.Header.Get("Retry-After")
			waitSeconds := 2 // 默认等待 2 秒
			if retryAfter != "" {
				if s, err := strconv.Atoi(retryAfter); err == nil {
					waitSeconds = s
				}
			}
			if attempt < maxRetries {
				time.Sleep(time.Duration(waitSeconds) * time.Second)
				continue
			}
			c.JSON(http.StatusOK, gin.H{
				"success":      false,
				"errorMessage": "请求过于频繁，请稍后再试",
			})
			return
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("状态码 %d", resp.StatusCode)
			c.JSON(http.StatusOK, gin.H{
				"success":      false,
				"errorMessage": "请求失败，状态码: " + strconv.Itoa(resp.StatusCode),
			})
			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success":      false,
				"errorMessage": "读取响应失败: " + err.Error(),
			})
			return
		}

		// 识别网页编码并转换
		html := decodeHTMLBody(body, resp.Header.Get("Content-Type"))

		// 提取 title
		title := extractTitle(html)

		// 检测反爬页面（如验证码页面）
		if isAntiCrawlPage(title) != "" {
			c.JSON(http.StatusOK, gin.H{
				"success":      false,
				"errorMessage": "目标网站存在反爬限制，请手动填写描述",
			})
			return
		}

		// 提取 description
		desc := extractMetaContent(html, "description")

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": types.FetchPageInfoResponse{
				Title:       title,
				Description: desc,
			},
		})
		return
	}

	// 兜底错误
	c.JSON(http.StatusOK, gin.H{
		"success":      false,
		"errorMessage": "请求失败: " + lastErr.Error(),
	})
}

// extractMetaContent 从 HTML 中提取 meta 标签内容

// extractMetaContent 从 HTML 中提取 meta 标签内容
func extractMetaContent(html, name string) string {
	// 优先使用预编译正则（覆盖 description 这个 99% 调用场景）
	if name == "description" {
		if m := reMetaNameFirst.FindStringSubmatch(html); len(m) > 1 && strings.TrimSpace(m[1]) != "" {
			return strings.TrimSpace(m[1])
		}
		if m := reMetaContentFirst.FindStringSubmatch(html); len(m) > 1 && strings.TrimSpace(m[1]) != "" {
			return strings.TrimSpace(m[1])
		}
		return ""
	}
	// 其他 name 动态编译（极少触发）
	p1 := fmt.Sprintf(`<meta[^>]+name=["']%s["'][^>]+content=["']([^"']+)["']`, name)
	p2 := fmt.Sprintf(`<meta[^>]+content=["']([^"']+)["'][^>]+name=["']%s["']`, name)
	for _, p := range []string{p1, p2} {
		re := regexp.MustCompile("(?i)" + p)
		if m := re.FindStringSubmatch(html); len(m) > 1 && strings.TrimSpace(m[1]) != "" {
			return strings.TrimSpace(m[1])
		}
	}
	return ""
}

// extractTitle 从 HTML 中提取 title 标签内容

// extractTitle 从 HTML 中提取 title 标签内容
func extractTitle(html string) string {
	matches := reTitle.FindStringSubmatch(html)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// decodeHTMLBody 根据 Content-Type 或 HTML meta 标签识别编码并转换为 UTF-8

// decodeHTMLBody 根据 Content-Type 或 HTML meta 标签识别编码并转换为 UTF-8
func decodeHTMLBody(body []byte, contentType string) string {
	// 1. 尝试从 Content-Type header 获取编码
	charset := extractCharsetFromContentType(contentType)

	// 2. 如果 header 没有，尝试从 HTML meta 标签获取
	if charset == "" {
		charset = extractCharsetFromMeta(string(body))
	}

	// 3. 根据编码转换
	if charset != "" {
		encodingName := strings.ToLower(charset)
		switch encodingName {
		case "gbk", "gb2312", "gb18030":
			// 使用 simplifiedchinese 将 GBK 转换为 UTF-8
			result := convertGBKToUTF8(body)
			if result != "" {
				return result
			}
		case "big5":
			// Big5 编码需要额外处理，这里暂不支持
			return string(body)
		case "utf-8", "utf8":
			return string(body)
		}
	}

	// 4. 默认：尝试用 UTF-8 解码，如果失败则尝试 GBK
	// 检查是否有效的 UTF-8
	if validUTF8(body) {
		return string(body)
	}
	// 不是有效 UTF-8，尝试 GBK
	result := convertGBKToUTF8(body)
	if result != "" {
		return result
	}

	// 5. 兜底：直接返回原始内容
	return string(body)
}

// validUTF8 检查是否为有效的 UTF-8 编码

// validUTF8 检查是否为有效的 UTF-8 编码
func validUTF8(data []byte) bool {
	return utf8.Valid(data)
}

// extractCharsetFromContentType 从 Content-Type header 提取 charset

// extractCharsetFromContentType 从 Content-Type header 提取 charset
func extractCharsetFromContentType(contentType string) string {
	matches := reCharset.FindStringSubmatch(contentType)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// extractCharsetFromMeta 从 HTML meta 标签提取编码

// extractCharsetFromMeta 从 HTML meta 标签提取编码
func extractCharsetFromMeta(html string) string {
	matches := reMetaCharset.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}
	matches = reMetaCharsetHTTP.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// convertGBKToUTF8 手动转换 GBK 到 UTF-8

// convertGBKToUTF8 手动转换 GBK 到 UTF-8
func convertGBKToUTF8(body []byte) string {
	// 使用 golang.org/x/text/encoding/simplifiedchinese
	GBK := simplifiedchinese.GB18030
	decoder := GBK.NewDecoder()
	result, err := decoder.Bytes(body)
	if err == nil {
		return string(result)
	}
	return ""
}

// isAntiCrawlPage 检测是否为反爬页面（如验证码页面）

// isAntiCrawlPage 检测是否为反爬页面（如验证码页面）
func isAntiCrawlPage(title string) string {
	if title == "" {
		return ""
	}

	// 常见的反爬关键词
	antiCrawlKeywords := []string{
		"验证码",
		"captcha",
		"验证",
		"安全验证",
		"人机验证",
		"atk",
		"安全中心",
	}

	titleLower := strings.ToLower(title)
	for _, keyword := range antiCrawlKeywords {
		if strings.Contains(titleLower, strings.ToLower(keyword)) {
			return keyword
		}
	}
	return ""
}

// GetMaxSortHandler 获取工具表最大排序值
