package utils

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"database/sql"
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/mereith/nav/logger"
	"github.com/mereith/nav/types"
)

func CheckErr(err error) {
	if err != nil {
		logger.LogError("捕获到错误：%s, 堆栈信息：%s", err, string(debug.Stack()))
	}
}

func CheckTxErr(err error, tx *sql.Tx) {
	if err != nil {
		logger.LogError("出现事务异常，回滚事务: %s, 堆栈信息：%s", err, string(debug.Stack()))
		err2 := tx.Rollback()
		CheckErr(err2)
	}
}

func In(target string, str_array []string) bool {
	for _, element := range str_array {
		if target == element {
			return true
		}
	}
	return false
}

// looksLikeImage 用内容魔数判断是不是真图片。
// 目的：Cloudflare / WAF 拦截时服务器会返回 200 + HTML 挑战页，
// 以前这种响应会被原样 base64 存进 nav_img，前端拿到「标头是图片、内容是 HTML」的坏缓存，渲染成破图。
func looksLikeImage(data []byte) bool {
	if len(data) < 8 {
		return false
	}
	switch {
	case bytes.HasPrefix(data, []byte("\x89PNG\r\n\x1a\n")), // PNG
		bytes.HasPrefix(data, []byte("\xff\xd8\xff")), // JPEG
		bytes.HasPrefix(data, []byte("GIF87a")),
		bytes.HasPrefix(data, []byte("GIF89a")),
		bytes.HasPrefix(data, []byte("\x00\x00\x01\x00")), // ICO
		bytes.HasPrefix(data, []byte("BM")):               // BMP
		return true
	}
	if len(data) >= 12 && bytes.HasPrefix(data, []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")) {
		return true
	}
	// SVG 是文本，可能带 BOM / XML 声明 / DOCTYPE / 注释，取前 1KB 找根标签
	head := data
	if len(head) > 1024 {
		head = head[:1024]
	}
	return bytes.Contains(head, []byte("<svg"))
}

func GetImgBase64FromUrl(url string) string {
	imgUrl := url
	//获取远端图片
	req, err := http.NewRequest("GET", imgUrl, nil)
	if err != nil {
		CheckErr(err)
		return ""
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.88 Safari/537.36")
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // codeql[go/disabled-certificate-check] — favicon 获取需兼容自签名证书
		},
		Timeout: 10 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		CheckErr(err)
		return ""
	}
	defer res.Body.Close()

	// 非 2xx 直接放弃（重定向后仍失败的，抓到的多半是错误页）
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		logger.LogError("logo 抓取返回 HTTP %d: %s", res.StatusCode, url)
		return ""
	}
	// Content-Type 粗筛：明确非图片的跳过；application/octet-stream 之类的放到魔数校验里判断
	if ct := strings.ToLower(res.Header.Get("Content-Type")); ct != "" &&
		!strings.HasPrefix(ct, "image/") && !strings.Contains(ct, "octet-stream") {
		logger.LogError("logo 抓取返回非图片类型 %s: %s", ct, url)
		return ""
	}

	// 读取获取的[]byte数据（限制 5MB 防止 OOM）
	const maxImageSize = 5 * 1024 * 1024
	limitedReader := io.LimitReader(res.Body, maxImageSize+1)
	data, err := io.ReadAll(limitedReader)
	if err != nil || len(data) > maxImageSize {
		logger.LogError("图片过大或读取失败: %s", url)
		return ""
	}
	// 内容魔数兜底：挡掉「200 + HTML 挑战页」这种最常见的坏数据
	if !looksLikeImage(data) {
		logger.LogError("logo 抓取内容不是图片（疑似 HTML 错误页）: %s", url)
		return ""
	}

	imageBase64 := base64.StdEncoding.EncodeToString(data)
	return imageBase64
}

func GetSuffixFromUrl(url string) string {
	// 无点的 URL（如 localhost）直接返回空，避免 url[-1:] 越界 panic
	idx := strings.LastIndex(url, ".")
	if idx < 0 {
		return ""
	}
	return url[idx:]
}
func GetMIME(suffix string) string {
	var t string = "image/x-icon"
	if suffix == ".svg" {
		t = "image/svg+xml"
	}
	if suffix == ".png" {
		t = "image/png"
	}
	return t
}

func PathExistsOrCreate(path string) {
	_, err := os.Stat(path)
	if err == nil {
		return
	}
	os.Mkdir(path, os.ModePerm)
}

func GenerateId() int {
	// 生成一个随机 id，避免时间戳冲突
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		// 回退到时间戳（极少数情况）
		return int(time.Now().Unix())
	}
	// 确保正数
	id := int(b[0])<<24 | int(b[1])<<16 | int(b[2])<<8 | int(b[3])
	if id < 0 {
		id = -id
	}
	return id
}

func FilterHideTools(tools []types.Tool, cates []types.Catelog) []types.Tool {
	result := make([]types.Tool, 0)
	var hideCates []string
	// 提取出需要隐藏的分类
	for _, cate := range cates {
		if cate.Hide {
			hideCates = append(hideCates, cate.Name)
		}
	}
	// 过滤工具
	for _, tool := range tools {
		if !tool.Hide && !In(tool.Catelog, hideCates) {
			result = append(result, tool)
		}
	}
	return result
}

func FilterHideCates(cates []types.Catelog) []types.Catelog {
	result := make([]types.Catelog, 0)
	for _, cate := range cates {
		if !cate.Hide {
			result = append(result, cate)
		}
	}
	return result
}
