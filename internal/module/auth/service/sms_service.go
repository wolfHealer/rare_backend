// internal/module/auth/service/sms_service.go
package service

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"

	"rare_backend/internal/config"
	"rare_backend/internal/module/auth/domain"
	"rare_backend/internal/module/auth/repo"
)

type SMSService struct {
	repo *repo.SMSRepo
}

func NewSMSService(r *repo.SMSRepo) *SMSService {
	return &SMSService{repo: r}
}

// SendCode 发送验证码
func (s *SMSService) SendCode(phone, scene string) error {
	// 验证场景
	if scene != "register" && scene != "login" {
		return domain.ErrInvalidScene
	}

	// 限制发送频率（1分钟内最多发送1次，5分钟内最多发送3次）
	count1min, err := s.repo.CountRecent(phone, 1)
	if err != nil {
		return err
	}
	if count1min >= 1 {
		return domain.ErrTooManyRequests
	}

	count5min, err := s.repo.CountRecent(phone, 5)
	if err != nil {
		return err
	}
	if count5min >= 3 {
		return domain.ErrTooManyRequests
	}

	// 生成6位验证码
	code := s.generateCode()

	// 发送短信
	cfg := config.GetSMSConfig()
	err = s.sendSMS(phone, code, scene, cfg)
	if err != nil {
		return err
	}

	// 保存验证码记录（有效期5分钟）
	return s.repo.Create(phone, code, scene, 5)
}

// VerifyCode 验证验证码
func (s *SMSService) VerifyCode(phone, code, scene string) (bool, error) {
	ok, err := s.repo.Verify(phone, code, scene)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, domain.ErrInvalidCode
	}
	// 验证成功后标记为已使用
	err = s.repo.MarkUsed(phone, code, scene)
	return ok, err
}

// generateCode 生成6位数字验证码
func (s *SMSService) generateCode() string {
	rand.Seed(time.Now().UnixNano())
	return strconv.Itoa(rand.Intn(900000) + 100000)
}

// sendSMS 发送短信
func (s *SMSService) sendSMS(phone, code, scene string, cfg config.SMSConfig) error {
	client, err := dysmsapi.NewClientWithAccessKey("cn-hangzhou", cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return fmt.Errorf("创建短信客户端失败: %v", err)
	}

	request := dysmsapi.CreateSendSmsRequest()
	request.Scheme = "https"

	request.PhoneNumbers = phone
	request.SignName = cfg.SignName

	// 根据场景选择模板
	if scene == "register" {
		request.TemplateCode = cfg.RegisterTemplate
	} else {
		request.TemplateCode = cfg.LoginTemplate
	}

	// 设置模板参数
	request.TemplateParam = fmt.Sprintf(`{"code":"%s"}`, code)

	_, err = client.SendSms(request)
	if err != nil {
		return fmt.Errorf("发送短信失败: %v", err)
	}

	return nil
}
