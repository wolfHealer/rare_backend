// internal/module/auth/service/sms_service.go
package service

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	openapiutil "github.com/alibabacloud-go/darabonba-openapi/v2/utils"
	dypnsapi "github.com/alibabacloud-go/dypnsapi-20170525/v3/client"
	"github.com/alibabacloud-go/tea/dara"
	// 公司资质短信（dysmsapi SendSms）
	// "github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"

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

	// 发送短信（个人资质：号码认证「短信认证」SendSmsVerifyCode）
	cfg := config.GetSMSConfig()
	err = s.sendSMSPersonal(phone, code, cfg)
	if err != nil {
		return err
	}

	// 公司资质短信（dysmsapi SendSms）
	// err = s.sendSMSCorporate(phone, code, scene, cfg)
	// if err != nil {
	// 	return err
	// }

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

// sendSMSPersonal 个人资质：号码认证「短信认证」SendSmsVerifyCode
// 需在控制台开通短信认证，使用赠送签名 + 赠送模板（SignName 与 TemplateCode 须配套）。
// 文档：https://help.aliyun.com/zh/pnvs/use-cases/sms-verify-for-individual-developers
func (s *SMSService) sendSMSPersonal(phone, code string, cfg config.SMSConfig) error {
	if cfg.AccessKeyID == "" || cfg.AccessKeySecret == "" {
		return fmt.Errorf("SMS AccessKey 未配置")
	}
	if cfg.PNVSSignName == "" || cfg.PNVSTemplateCode == "" {
		return fmt.Errorf("个人短信认证未配置：请设置 SMS_PNVS_SIGN_NAME 与 SMS_PNVS_TEMPLATE_CODE")
	}

	clientCfg := &openapiutil.Config{
		AccessKeyId:     dara.String(cfg.AccessKeyID),
		AccessKeySecret: dara.String(cfg.AccessKeySecret),
		Endpoint:        dara.String("dypnsapi.aliyuncs.com"),
	}
	client, err := dypnsapi.NewClient(clientCfg)
	if err != nil {
		return fmt.Errorf("创建短信认证客户端失败: %w", err)
	}

	// 传入业务侧生成的验证码；模板变量名需与控制台模板一致（常见为 code、min）
	templateParam, err := json.Marshal(map[string]string{
		"code": code,
		"min":  "5",
	})
	if err != nil {
		return fmt.Errorf("构造模板参数失败: %w", err)
	}

	req := &dypnsapi.SendSmsVerifyCodeRequest{
		PhoneNumber:  dara.String(phone),
		SignName:     dara.String(cfg.PNVSSignName),
		TemplateCode: dara.String(cfg.PNVSTemplateCode),
		TemplateParam: dara.String(string(templateParam)),
		ValidTime:    dara.Int64(300),
	}

	resp, err := client.SendSmsVerifyCode(req)
	if err != nil {
		return fmt.Errorf("发送短信验证码失败: %w", err)
	}
	if resp.Body == nil {
		return fmt.Errorf("发送短信验证码失败: 空响应")
	}
	if resp.Body.Success != nil && !dara.BoolValue(resp.Body.Success) {
		msg := dara.StringValue(resp.Body.Message)
		code := dara.StringValue(resp.Body.Code)
		if msg == "" {
			msg = "未知错误"
		}
		return fmt.Errorf("发送短信验证码失败: %s (%s)", msg, code)
	}
	if resp.Body.Code != nil && dara.StringValue(resp.Body.Code) != "OK" {
		return fmt.Errorf("发送短信验证码失败: %s (%s)",
			dara.StringValue(resp.Body.Message), dara.StringValue(resp.Body.Code))
	}
	return nil
}

// sendSMSCorporate 公司资质：短信服务 dysmsapi SendSms（需企业资质、签名与模板审核）
// func (s *SMSService) sendSMSCorporate(phone, code, scene string, cfg config.SMSConfig) error {
// 	client, err := dysmsapi.NewClientWithAccessKey("cn-hangzhou", cfg.AccessKeyID, cfg.AccessKeySecret)
// 	if err != nil {
// 		return fmt.Errorf("创建短信客户端失败: %v", err)
// 	}
//
// 	request := dysmsapi.CreateSendSmsRequest()
// 	request.Scheme = "https"
//
// 	request.PhoneNumbers = phone
// 	request.SignName = cfg.SignName
//
// 	// 根据场景选择模板
// 	if scene == "register" {
// 		request.TemplateCode = cfg.RegisterTemplate
// 	} else {
// 		request.TemplateCode = cfg.LoginTemplate
// 	}
//
// 	// 设置模板参数
// 	request.TemplateParam = fmt.Sprintf(`{"code":"%s"}`, code)
//
// 	response, err := client.SendSms(request)
// 	if err != nil {
// 		return fmt.Errorf("发送短信失败: %v", err)
// 	}
// 	if response.Code != "OK" {
// 		return fmt.Errorf("发送短信失败: %s (%s)", response.Message, response.Code)
// 	}
//
// 	return nil
// }
