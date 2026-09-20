package biz

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/http"
	"net/smtp"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gops/internal/data/api"
)

// TestChannel sends a real test notification through one stored alert channel.
func (b *MonitorBiz) TestChannel(req *api.MonitorAlertChannelTestRequest) error {
	channel := &api.MonitorAlertChannelRequest{ChannelType: req.ChannelType, Config: req.Config}
	if err := validateAlertChannel(channel); err != nil {
		return err
	}
	config := req.Config
	switch req.ChannelType {
	case "email":
		if strings.TrimSpace(req.Recipient) == "" {
			return errors.New("请输入测试收件人")
		}
		return sendChannelTestEmail(config, strings.TrimSpace(req.Recipient))
	case "wechat":
		return b.postChannelJSON(config.WeChatWebhook, map[string]any{"msgtype": "text", "text": map[string]string{"content": "告警通道测试成功：企业微信通知已连接。"}})
	case "dingtalk":
		webhook, err := signedDingTalkURL(config.DingTalkWebhook, config.DingTalkSecret)
		if err != nil {
			return err
		}
		return b.postChannelJSON(webhook, map[string]any{"msgtype": "text", "text": map[string]string{"content": "告警通道测试成功：钉钉通知已连接。"}})
	case "feishu":
		return b.postChannelJSON(config.FeishuWebhook, map[string]any{"msg_type": "text", "content": map[string]string{"text": "告警通道测试成功：飞书通知已连接。"}})
	case "webhook":
		return b.postChannelJSON(config.WebhookURL, map[string]any{"event": "channel_test", "title": "告警通道测试", "message": "Webhook 告警通道连接成功", "timestamp": time.Now().Format(time.RFC3339)})
	default:
		return errors.New("不支持的告警通道类型")
	}
}

func (b *MonitorBiz) postChannelJSON(endpoint string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	request, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := b.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("告警通道返回 HTTP %d", response.StatusCode)
	}
	return nil
}

func signedDingTalkURL(webhook, secret string) (string, error) {
	if secret == "" {
		return webhook, nil
	}
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(timestamp + "\n" + secret))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	parsed, err := url.Parse(webhook)
	if err != nil {
		return "", err
	}
	query := parsed.Query()
	query.Set("timestamp", timestamp)
	query.Set("sign", signature)
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func sendChannelTestEmail(config *api.MonitorAlertChannelConfig, recipient string) error {
	address := net.JoinHostPort(config.SMTPHost, strconv.FormatInt(config.SMTPPort, 10))
	tlsEnabled := config.SMTPPort == 465 || config.SMTPPort == 587
	client, err := newSMTPClient(address, config.SMTPHost, tlsEnabled, int(config.SMTPPort))
	if err != nil {
		return err
	}
	defer client.Quit()
	if config.SMTPUser != "" {
		if err := client.Auth(smtp.PlainAuth("", config.SMTPUser, config.SMTPPassword, config.SMTPHost)); err != nil {
			return err
		}
	}
	from := config.FromEmail
	if from == "" {
		from = config.SMTPUser
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(recipient); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	sender := from
	if config.FromName != "" {
		sender = mime.QEncoding.Encode("UTF-8", config.FromName) + " <" + from + ">"
	}
	_, writeErr := fmt.Fprintf(w, "To: %s\r\nFrom: %s\r\nSubject: =?UTF-8?B?5ZGK6K2m6YCa6YGT5rWL6K+V?=\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n告警通道测试成功，SMTP 配置可用。\r\n", recipient, sender)
	closeErr := w.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}
