package biz

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gops/internal/data/api"
	"gops/internal/data/model"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

var monitorExcelHeaders = []string{"域名", "端口", "请求客户端", "检测间隔(秒)", "抖动(秒)", "HTTPS", "证书校验", "启用告警", "响应阈值(毫秒)", "SSL到期提醒(天)", "状态", "响应时间(毫秒)", "SSL有效", "SSL到期时间", "根CA", "最后检测", "最近错误"}

func (b *MonitorBiz) ExportDomains(req *api.MonitorDomainListRequest) (*api.MonitorDomainExcelExportResponse, error) {
	rows, err := b.ListDomains(req)
	if err != nil {
		return nil, err
	}
	book := excelize.NewFile()
	defer func() { _ = book.Close() }()
	sheet := "域名监控"
	_ = book.SetSheetName("Sheet1", sheet)
	headerStyle, _ := book.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Color: "FFFFFF"}, Fill: excelize.Fill{Type: "pattern", Color: []string{"4F63E9"}, Pattern: 1}, Alignment: &excelize.Alignment{Vertical: "center"}})
	_ = book.SetSheetRow(sheet, "A1", &monitorExcelHeaders)
	_ = book.SetCellStyle(sheet, "A1", "Q1", headerStyle)
	_ = book.SetPanes(sheet, &excelize.Panes{Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})
	for index, row := range rows {
		values := []any{row.Domain, row.Port, row.RequestClient, row.CheckInterval, row.CheckJitter, yesNo(row.EnableSSL), yesNo(row.VerifyCertificate), yesNo(row.EnableAlert), row.ResponseThreshold, row.SSLExpiryDays, monitorStatusLabel(row.Status), row.ResponseTime, yesNo(row.SSLValid), excelTime(row.SSLExpiry), row.SSLRootCA, excelTime(row.LastCheck), row.LastError}
		cell, _ := excelize.CoordinatesToCellName(1, index+2)
		_ = book.SetSheetRow(sheet, cell, &values)
	}
	for column, width := range map[string]float64{"A": 28, "B": 10, "C": 18, "D": 15, "E": 12, "F": 10, "G": 12, "H": 12, "I": 18, "J": 18, "K": 12, "L": 18, "M": 12, "N": 22, "O": 24, "P": 22, "Q": 38} {
		_ = book.SetColWidth(sheet, column, column, width)
	}
	buffer, err := book.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return &api.MonitorDomainExcelExportResponse{FileName: "域名监控_" + time.Now().Format("20060102_150405") + ".xlsx", Content: base64.StdEncoding.EncodeToString(buffer.Bytes())}, nil
}

func (b *MonitorBiz) ImportDomains(req *api.MonitorDomainExcelImportRequest) (*api.MonitorDomainExcelImportResponse, error) {
	data, err := base64.StdEncoding.DecodeString(req.Content)
	if err != nil {
		return nil, errors.New("Excel 文件内容无效")
	}
	book, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, errors.New("无法解析 Excel 文件，请使用导出的 .xlsx 模板")
	}
	defer func() { _ = book.Close() }()
	sheets := book.GetSheetList()
	if len(sheets) == 0 {
		return nil, errors.New("Excel 中没有工作表")
	}
	rows, err := book.GetRows(sheets[0])
	if err != nil || len(rows) < 2 {
		return nil, errors.New("Excel 中没有可导入的数据")
	}
	header := make(map[string]int)
	for index, value := range rows[0] {
		header[strings.TrimSpace(value)] = index
	}
	for _, required := range monitorExcelHeaders[:10] {
		if _, ok := header[required]; !ok {
			return nil, fmt.Errorf("缺少列：%s", required)
		}
	}
	result := &api.MonitorDomainExcelImportResponse{}
	err = b.db.Transaction(func(tx *gorm.DB) error {
		for index, values := range rows[1:] {
			if rowCell(values, header["域名"]) == "" {
				continue
			}
			request, parseErr := monitorRequestFromExcel(values, header)
			if parseErr != nil {
				return fmt.Errorf("第 %d 行：%w", index+2, parseErr)
			}
			normalized := normalizeMonitorDomainRequest(request)
			var existing model.MonitorDomain
			lookup := tx.Where("domain = ? AND port = ?", normalized.Domain, normalized.Port).First(&existing)
			if errors.Is(lookup.Error, gorm.ErrRecordNotFound) {
				created := b.domainModel(normalized)
				created.NextCheck = time.Now()
				if err := tx.Create(&created).Error; err != nil {
					return err
				}
				result.Created++
			} else if lookup.Error != nil {
				return lookup.Error
			} else {
				existing.RequestClient = normalized.RequestClient
				existing.CheckInterval = normalized.CheckInterval
				existing.CheckJitter = normalized.CheckJitter
				existing.EnableSsl = normalized.EnableSSL
				existing.VerifyCertificate = normalized.VerifyCertificate
				existing.EnableAlert = normalized.EnableAlert
				existing.ResponseThreshold = normalized.ResponseThreshold
				existing.SslExpiryDays = normalized.SSLExpiryDays
				existing.NextCheck = time.Now()
				if err := tx.Save(&existing).Error; err != nil {
					return err
				}
				result.Updated++
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func monitorRequestFromExcel(row []string, header map[string]int) (*api.MonitorDomainRequest, error) {
	port, err := excelInt(rowCell(row, header["端口"]))
	if err != nil || port < 1 || port > 65535 {
		return nil, errors.New("端口必须是 1-65535 的整数")
	}
	interval, err := excelInt(rowCell(row, header["检测间隔(秒)"]))
	if err != nil || interval < 30 {
		return nil, errors.New("检测间隔不能小于 30 秒")
	}
	jitter, err := excelInt(rowCell(row, header["抖动(秒)"]))
	if err != nil || jitter < 0 {
		return nil, errors.New("抖动必须是非负整数")
	}
	threshold, err := excelInt(rowCell(row, header["响应阈值(毫秒)"]))
	if err != nil || threshold < 0 {
		return nil, errors.New("响应阈值必须是非负整数")
	}
	expiry, err := excelInt(rowCell(row, header["SSL到期提醒(天)"]))
	if err != nil || expiry < 0 {
		return nil, errors.New("SSL 到期提醒必须是非负整数")
	}
	return &api.MonitorDomainRequest{Domain: rowCell(row, header["域名"]), Port: port, RequestClient: rowCell(row, header["请求客户端"]), CheckInterval: interval, CheckJitter: jitter, EnableSSL: excelBool(rowCell(row, header["HTTPS"])), VerifyCertificate: excelBool(rowCell(row, header["证书校验"])), EnableAlert: excelBool(rowCell(row, header["启用告警"])), ResponseThreshold: threshold, SSLExpiryDays: expiry}, nil
}

func rowCell(row []string, index int) string {
	if index < 0 || index >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[index])
}
func excelInt(value string) (int64, error) {
	return strconv.ParseInt(strings.TrimSuffix(strings.TrimSpace(value), ".0"), 10, 64)
}
func excelBool(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "是" || value == "true" || value == "1" || value == "yes"
}
func yesNo(value bool) string {
	if value {
		return "是"
	}
	return "否"
}
func excelTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02 15:04:05")
}
func monitorStatusLabel(value string) string {
	switch value {
	case "normal":
		return "正常"
	case "abnormal":
		return "异常"
	case "paused":
		return "暂停"
	default:
		return "未知"
	}
}
