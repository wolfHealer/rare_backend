package storage

import (
	"mime"
	"path/filepath"
	"strings"
)

var allowedImageTypes = map[string]bool{
	"image/jpeg":               true,
	"image/png":                true,
	"image/webp":               true,
	"image/gif":                true,
	"application/octet-stream": true, // 微信上传常见，需结合扩展名判断
}

var extToMIME = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".webp": "image/webp",
	".gif":  "image/gif",
}

// ResolveImageContentType 根据 multipart Content-Type 与文件名推断图片 MIME。
// 微信小程序上传常不带 Content-Type 或使用 application/octet-stream。
func ResolveImageContentType(headerType, filename string) (string, bool) {
	headerType = strings.TrimSpace(strings.Split(headerType, ";")[0])
	ext := strings.ToLower(filepath.Ext(filename))

	if headerType == "application/octet-stream" || headerType == "" {
		if mimeType, ok := extToMIME[ext]; ok {
			return mimeType, true
		}
		return "", false
	}

	if allowedImageTypes[headerType] && headerType != "application/octet-stream" {
		return headerType, true
	}

	if mimeType, ok := extToMIME[ext]; ok {
		return mimeType, true
	}
	if mimeType := mime.TypeByExtension(ext); allowedImageTypes[mimeType] && mimeType != "application/octet-stream" {
		return mimeType, true
	}
	return "", false
}

// NormalizeImageExt 根据 MIME 返回标准扩展名。
func NormalizeImageExt(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ".jpg"
	}
}
