package biz

import (
	"bytes"
	_ "embed"
	"html/template"
)

//go:embed templates/verification_email.html
var verificationEmailHTML string

var verificationEmailTemplate = template.Must(template.New("verification-email").Parse(verificationEmailHTML))

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
