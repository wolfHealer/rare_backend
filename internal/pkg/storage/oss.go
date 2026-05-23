package storage

import (
	"fmt"
	"io"
	"strings"

	"rare_backend/internal/config"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// ValidateOSSConfig 检查 OSS 必填项。
func ValidateOSSConfig(cfg config.OSSConfig) error {
	if strings.TrimSpace(cfg.Endpoint) == "" {
		return fmt.Errorf("OSS_ENDPOINT 未配置")
	}
	if strings.TrimSpace(cfg.AccessKeyID) == "" {
		return fmt.Errorf("OSS_ACCESS_KEY_ID 未配置")
	}
	if strings.TrimSpace(cfg.AccessKeySecret) == "" {
		return fmt.Errorf("OSS_ACCESS_KEY_SECRET 未配置")
	}
	if strings.TrimSpace(cfg.BucketName) == "" {
		return fmt.Errorf("OSS_BUCKET_NAME 未配置")
	}
	return nil
}

// NormalizeEndpoint 去掉 scheme 与尾部斜杠。
func NormalizeEndpoint(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")
	return strings.TrimSuffix(endpoint, "/")
}

// PublicURL 构建对象公网访问地址。
func PublicURL(cfg config.OSSConfig, objectKey string) string {
	base := strings.TrimSpace(cfg.PublicBaseURL)
	if base != "" {
		base = strings.TrimSuffix(base, "/")
		return base + "/" + strings.TrimPrefix(objectKey, "/")
	}
	endpoint := NormalizeEndpoint(cfg.Endpoint)
	return fmt.Sprintf("https://%s.%s/%s", cfg.BucketName, endpoint, strings.TrimPrefix(objectKey, "/"))
}

// Upload 上传对象到 OSS（使用 Bucket 默认 ACL，兼容「Bucket 所有者 enforced」）。
func Upload(cfg config.OSSConfig, objectKey string, reader io.Reader, contentType string) error {
	if err := ValidateOSSConfig(cfg); err != nil {
		return err
	}

	client, err := oss.New(NormalizeEndpoint(cfg.Endpoint), cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return fmt.Errorf("OSS 客户端初始化失败: %w", err)
	}
	bucket, err := client.Bucket(cfg.BucketName)
	if err != nil {
		return fmt.Errorf("获取 OSS Bucket 失败: %w", err)
	}

	opts := []oss.Option{}
	if contentType != "" {
		opts = append(opts, oss.ContentType(contentType))
	}
	if err := bucket.PutObject(objectKey, reader, opts...); err != nil {
		return fmt.Errorf("OSS 上传失败: %w", err)
	}
	return nil
}

// Delete 删除 OSS 对象（失败时仅返回 error，调用方可忽略）。
func Delete(cfg config.OSSConfig, objectKey string) error {
	if err := ValidateOSSConfig(cfg); err != nil {
		return err
	}
	client, err := oss.New(NormalizeEndpoint(cfg.Endpoint), cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return err
	}
	bucket, err := client.Bucket(cfg.BucketName)
	if err != nil {
		return err
	}
	return bucket.DeleteObject(objectKey)
}
