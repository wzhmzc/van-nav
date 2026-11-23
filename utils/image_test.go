package utils

import (
	"testing"
)

// TestLooksLikeImage 覆盖 looksLikeImage 的所有分支。
// 这条校验是防止「Cloudflare 返回 200 + HTML 挑战页」被当成图片存进缓存的唯一防线，
// 一旦被改坏，前端会出现大量破图且没有报错 —— 值得用表驱动锁死。
func TestLooksLikeImage(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{"PNG", []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR"), true},
		{"JPEG", []byte("\xff\xd8\xff\xe0\x00\x10JFIF\x00"), true},
		{"GIF87a", []byte("GIF87a\x01\x00\x01\x00"), true},
		{"GIF89a", []byte("GIF89a\x01\x00\x01\x00"), true},
		{"ICO", []byte("\x00\x00\x01\x00\x01\x00\x20\x20"), true},
		{"BMP", []byte("BM\x36\x00\x00\x00\x00\x00\x00\x00"), true},
		{"WEBP", []byte("RIFF\x24\x00\x00\x00WEBPVP8 "), true},
		{"SVG 带 XML 声明", []byte(`<?xml version="1.0" encoding="UTF-8"?><svg xmlns="http://www.w3.org/2000/svg"></svg>`), true},
		{"SVG 裸标签", []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="24"></svg>`), true},
		{"SVG 前面有 1KB 内注释", append(make([]byte, 600), []byte(`<svg xmlns="http://www.w3.org/2000/svg"></svg>`)...), true},

		// 以下是必须挡掉的
		{"Cloudflare 挑战页", []byte("<!DOCTYPE html>\n<!--[if lt IE 7]> <html class=\"no-js ie6 oldie\" lang=\"en-US\"> <![endif]--><html><head><title>Just a moment...</title></head></html>"), false},
		{"普通 HTML 错误页", []byte("<html><body>404 Not Found</body></html>"), false},
		{"JSON 错误响应", []byte(`{"success":false,"errorMessage":"图片解码失败"}`), false},
		{"纯文本", []byte("ok"), false},
		{"空数据", []byte(""), false},
		{"不足 8 字节", []byte("abc"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := looksLikeImage(tt.data); got != tt.want {
				t.Errorf("looksLikeImage(%s) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestGetMIME(t *testing.T) {
	cases := []struct {
		suffix string
		want   string
	}{
		{".svg", "image/svg+xml"},
		{".png", "image/png"},
		{".ico", "image/x-icon"},
		{".jpg", "image/x-icon"}, // 未显式支持的后缀回落到默认值
		{"", "image/x-icon"},
	}
	for _, c := range cases {
		if got := GetMIME(c.suffix); got != c.want {
			t.Errorf("GetMIME(%q) = %q, want %q", c.suffix, got, c.want)
		}
	}
}

func TestGetSuffixFromUrl(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"https://a.com/x.svg", ".svg"},
		{"https://a.com/x.png?v=1", ".png?v=1"},
		{"https://a.com/plain", ".com/plain"}, // 取最后一个点之后的所有内容，含路径
		{"localhost", ""},                     // 无点时返回空，不再 panic
	}
	for _, c := range cases {
		if got := GetSuffixFromUrl(c.in); got != c.want {
			t.Errorf("GetSuffixFromUrl(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
