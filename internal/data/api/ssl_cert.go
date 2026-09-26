package api

import "time"

type SSLCertificateListRequest struct {
	PageRequest
	Keyword string `json:"keyword"`
}
type SSLCertificateRequest struct {
	ID              int64  `json:"id"`
	Name            string `json:"name" binding:"required"`
	Domain          string `json:"domain" binding:"required"`
	SourceType      string `json:"source_type"`
	CaProvider      string `json:"ca_provider"`
	AutoRenew       bool   `json:"auto_renew"`
	RenewDaysBefore int64  `json:"renew_days_before"`
}
type SSLCertificateResponse struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	Domain          string    `json:"domain"`
	SourceType      string    `json:"source_type"`
	CaProvider      string    `json:"ca_provider"`
	Status          string    `json:"status"`
	AutoRenew       bool      `json:"auto_renew"`
	RenewDaysBefore int64     `json:"renew_days_before"`
	NotAfter        time.Time `json:"not_after"`
}
