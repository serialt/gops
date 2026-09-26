package biz

import (
	"bytes"
	_ "embed"
	"html/template"
	texttemplate "text/template"
)

//go:embed templates/verification_email.html
var verificationEmailHTML string

var verificationEmailTemplate = template.Must(template.New("verification-email").Parse(verificationEmailHTML))

//go:embed templates/monitor_alert_email.html
var monitorAlertEmailHTML string

//go:embed templates/monitor_alert_dingtalk.md
var monitorAlertDingTalkMarkdown string

//go:embed templates/monitor_alert_feishu.txt
var monitorAlertFeishuText string

var monitorAlertEmailTemplate = template.Must(template.New("monitor-alert-email").Parse(monitorAlertEmailHTML))
var monitorAlertDingTalkTemplate = texttemplate.Must(texttemplate.New("monitor-alert-dingtalk").Parse(monitorAlertDingTalkMarkdown))
var monitorAlertFeishuTemplate = texttemplate.Must(texttemplate.New("monitor-alert-feishu").Parse(monitorAlertFeishuText))

type verificationEmailData struct {
	Title         string
	Code          string
	ExpireMinutes int
}

func renderVerificationEmail(title, code string, expireMinutes int) (string, error) {
	var output bytes.Buffer
	err := verificationEmailTemplate.Execute(&output, &verificationEmailData{
		Title: title, Code: code, ExpireMinutes: expireMinutes,
	})
	return output.String(), err
}

type monitorAlertTemplateData struct {
	Title      string
	Target     string
	AlertType  string
	Message    string
	OccurredAt string
}

func renderMonitorAlertEmail(data monitorAlertTemplateData) (string, error) {
	var output bytes.Buffer
	err := monitorAlertEmailTemplate.Execute(&output, data)
	return output.String(), err
}

func renderMonitorAlertDingTalk(data monitorAlertTemplateData) (string, error) {
	var output bytes.Buffer
	err := monitorAlertDingTalkTemplate.Execute(&output, data)
	return output.String(), err
}

func renderMonitorAlertFeishu(data monitorAlertTemplateData) (string, error) {
	var output bytes.Buffer
	err := monitorAlertFeishuTemplate.Execute(&output, data)
	return output.String(), err
}
