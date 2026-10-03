package biz

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gops/internal/data/api"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

var hostExcelHeaders = []string{"主机名称", "IP", "SSH端口", "主机类型", "所属分组ID", "SSH凭据ID", "VNC凭据ID", "云账号ID", "标签", "备注"}
var dbInstanceExcelHeaders = []string{"数据库名称", "数据库类型", "数据库地址", "端口", "所属分组ID", "凭据ID", "标签", "备注"}

func (b *AssetBiz) ExportHosts(ctx context.Context, req *api.HostListRequest) (string, []byte, error) {
	book, sheet, err := newHostExcelBook()
	if err != nil {
		return "", nil, err
	}
	defer func() { _ = book.Close() }()
	pageRequest := *req
	pageRequest.Page, pageRequest.Limit = 1, api.MaxLimit
	rowIndex := 2
	for {
		page, listErr := b.ListHosts(ctx, &pageRequest)
		if listErr != nil {
			return "", nil, listErr
		}
		for _, row := range page.List {
			values := []any{row.Name, row.IP, row.Port, row.HostType, row.GroupID, row.CredentialID, row.VNCCredentialID, row.CloudAccountID, row.Tags, row.Description}
			cell, _ := excelize.CoordinatesToCellName(1, rowIndex)
			if err = book.SetSheetRow(sheet, cell, &values); err != nil {
				return "", nil, err
			}
			rowIndex++
		}
		if int64(rowIndex-2) >= page.Total || len(page.List) == 0 {
			break
		}
		pageRequest.Page++
	}
	return hostExcelResponse(book, "主机管理_"+time.Now().Format("20060102_150405")+".xlsx")
}

func (b *AssetBiz) HostImportTemplate() (string, []byte, error) {
	book, _, err := newHostExcelBook()
	if err != nil {
		return "", nil, err
	}
	defer func() { _ = book.Close() }()
	return hostExcelResponse(book, "主机导入模板.xlsx")
}

func (b *AssetBiz) ImportHosts(ctx context.Context, fileName string, content io.Reader) (*api.HostExcelImportResponse, error) {
	extension := strings.ToLower(filepath.Ext(fileName))
	if extension != ".xlsx" {
		return nil, errors.New("仅支持 .xlsx 格式的 Excel 文件")
	}
	book, err := excelize.OpenReader(content)
	if err != nil {
		return nil, errors.New("无法解析 Excel 文件，请使用主机导入模板")
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
	header := make(map[string]int, len(rows[0]))
	for index, value := range rows[0] {
		header[strings.TrimSpace(value)] = index
	}
	for _, name := range hostExcelHeaders[:6] {
		if _, ok := header[name]; !ok {
			return nil, fmt.Errorf("缺少列：%s", name)
		}
	}
	requests := make([]*api.HostRequest, 0, len(rows)-1)
	for index, values := range rows[1:] {
		if hostExcelCell(values, header["主机名称"]) == "" && hostExcelCell(values, header["IP"]) == "" {
			continue
		}
		request, parseErr := hostRequestFromExcel(values, header)
		if parseErr != nil {
			return nil, fmt.Errorf("第 %d 行：%w", index+2, parseErr)
		}
		requests = append(requests, request)
	}
	if len(requests) == 0 {
		return nil, errors.New("Excel 中没有可导入的数据")
	}
	result := &api.HostExcelImportResponse{}
	err = b.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, request := range requests {
			existing, lookupErr := b.repo.GetHostByIP(ctx, tx, request.IP)
			if lookupErr != nil && !errors.Is(lookupErr, gorm.ErrRecordNotFound) {
				return lookupErr
			}
			if errors.Is(lookupErr, gorm.ErrRecordNotFound) {
				row := hostModel(request)
				if row.HostType == "" {
					row.HostType = "self"
				}
				row.Status = -1
				if err := b.repo.CreateHost(ctx, tx, &row); err != nil {
					return err
				}
				result.Created++
				continue
			}
			updated := hostModel(request)
			updated.ID, updated.CreatedAt = existing.ID, existing.CreatedAt
			updated.LastSeenAt, updated.Os, updated.Kernel, updated.Arch = existing.LastSeenAt, existing.Os, existing.Kernel, existing.Arch
			updated.CPUInfo, updated.CPUCores, updated.CPUUsage = existing.CPUInfo, existing.CPUCores, existing.CPUUsage
			updated.MemoryTotal, updated.MemoryUsed, updated.MemoryUsage = existing.MemoryTotal, existing.MemoryUsed, existing.MemoryUsage
			updated.DiskTotal, updated.DiskUsed, updated.DiskUsage = existing.DiskTotal, existing.DiskUsed, existing.DiskUsage
			updated.Uptime, updated.Hostname, updated.Status = existing.Uptime, existing.Hostname, existing.Status
			if err := b.repo.SaveHost(ctx, tx, &updated); err != nil {
				return err
			}
			result.Updated++
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (b *AssetBiz) ExportDBInstances(ctx context.Context, req *api.DbInstanceListRequest) (string, []byte, error) {
	book, sheet, err := newDBInstanceExcelBook()
	if err != nil {
		return "", nil, err
	}
	defer func() { _ = book.Close() }()
	pageRequest := *req
	pageRequest.Page, pageRequest.Limit = 1, api.MaxLimit
	rowIndex := 2
	for {
		page, listErr := b.ListDBInstances(ctx, &pageRequest)
		if listErr != nil {
			return "", nil, listErr
		}
		for _, row := range page.List {
			values := []any{row.Name, row.DbType, row.Host, row.Port, row.GroupID, row.CredentialID, row.Tags, row.Description}
			cell, _ := excelize.CoordinatesToCellName(1, rowIndex)
			if err = book.SetSheetRow(sheet, cell, &values); err != nil {
				return "", nil, err
			}
			rowIndex++
		}
		if int64(rowIndex-2) >= page.Total || len(page.List) == 0 {
			break
		}
		pageRequest.Page++
	}
	return hostExcelResponse(book, "数据库管理_"+time.Now().Format("20060102_150405")+".xlsx")
}

func (b *AssetBiz) DBInstanceImportTemplate() (string, []byte, error) {
	book, _, err := newDBInstanceExcelBook()
	if err != nil {
		return "", nil, err
	}
	defer func() { _ = book.Close() }()
	return hostExcelResponse(book, "数据库导入模板.xlsx")
}

func (b *AssetBiz) ImportDBInstances(ctx context.Context, fileName string, content io.Reader) (*api.DbInstanceExcelImportResponse, error) {
	if strings.ToLower(filepath.Ext(fileName)) != ".xlsx" {
		return nil, errors.New("仅支持 .xlsx 格式的 Excel 文件")
	}
	book, err := excelize.OpenReader(content)
	if err != nil {
		return nil, errors.New("无法解析 Excel 文件，请使用数据库导入模板")
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
	header := make(map[string]int, len(rows[0]))
	for index, value := range rows[0] {
		header[strings.TrimSpace(value)] = index
	}
	for _, name := range dbInstanceExcelHeaders[:6] {
		if _, ok := header[name]; !ok {
			return nil, fmt.Errorf("缺少列：%s", name)
		}
	}
	requests := make([]*api.DbInstanceRequest, 0, len(rows)-1)
	for index, values := range rows[1:] {
		if hostExcelCell(values, header["数据库名称"]) == "" && hostExcelCell(values, header["数据库地址"]) == "" {
			continue
		}
		request, parseErr := dbInstanceRequestFromExcel(values, header)
		if parseErr != nil {
			return nil, fmt.Errorf("第 %d 行：%w", index+2, parseErr)
		}
		requests = append(requests, request)
	}
	if len(requests) == 0 {
		return nil, errors.New("Excel 中没有可导入的数据")
	}
	result := &api.DbInstanceExcelImportResponse{}
	collectedIDs := make([]int64, 0, len(requests))
	err = b.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, request := range requests {
			existing, lookupErr := b.repo.GetDBInstanceByEndpoint(ctx, tx, request.DbType, request.Host, request.Port)
			if lookupErr != nil && !errors.Is(lookupErr, gorm.ErrRecordNotFound) {
				return lookupErr
			}
			row, modelErr := b.dbInstanceModel(request)
			if modelErr != nil {
				return modelErr
			}
			if errors.Is(lookupErr, gorm.ErrRecordNotFound) {
				if err := b.repo.CreateDBInstance(ctx, tx, &row); err != nil {
					return err
				}
				result.Created++
			} else {
				row.ID, row.CreatedAt, row.Version, row.Status = existing.ID, existing.CreatedAt, existing.Version, existing.Status
				if err := b.repo.SaveDBInstance(ctx, tx, &row); err != nil {
					return err
				}
				result.Updated++
			}
			collectedIDs = append(collectedIDs, row.ID)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	for _, id := range collectedIDs {
		b.collectDBInstanceAsync(id)
	}
	return result, nil
}

func newDBInstanceExcelBook() (*excelize.File, string, error) {
	book := excelize.NewFile()
	sheet := "数据库管理"
	if err := book.SetSheetName("Sheet1", sheet); err != nil {
		_ = book.Close()
		return nil, "", err
	}
	if err := book.SetSheetRow(sheet, "A1", &dbInstanceExcelHeaders); err != nil {
		_ = book.Close()
		return nil, "", err
	}
	style, err := book.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Color: "FFFFFF"}, Fill: excelize.Fill{Type: "pattern", Color: []string{"4F63E9"}, Pattern: 1}, Alignment: &excelize.Alignment{Vertical: "center"}})
	if err != nil {
		_ = book.Close()
		return nil, "", err
	}
	if err = book.SetCellStyle(sheet, "A1", "H1", style); err != nil {
		_ = book.Close()
		return nil, "", err
	}
	_ = book.SetPanes(sheet, &excelize.Panes{Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})
	for column, width := range map[string]float64{"A": 24, "B": 16, "C": 22, "D": 12, "E": 16, "F": 12, "G": 24, "H": 36} {
		_ = book.SetColWidth(sheet, column, column, width)
	}
	return book, sheet, nil
}

func dbInstanceRequestFromExcel(row []string, header map[string]int) (*api.DbInstanceRequest, error) {
	name := hostExcelCell(row, header["数据库名称"])
	dbType := strings.ToLower(hostExcelCell(row, header["数据库类型"]))
	host := hostExcelCell(row, header["数据库地址"])
	if name == "" || dbType == "" || host == "" {
		return nil, errors.New("数据库名称、类型和地址不能为空")
	}
	port, err := strconv.ParseInt(strings.TrimSuffix(hostExcelCell(row, header["端口"]), ".0"), 10, 64)
	if err != nil || port < 1 || port > 65535 {
		return nil, errors.New("端口必须是 1-65535 的整数")
	}
	groupID, err := hostExcelOptionalID(row, header, "所属分组ID")
	if err != nil {
		return nil, errors.New("所属分组ID必须是非负整数")
	}
	credentialID, err := hostExcelOptionalID(row, header, "凭据ID")
	if err != nil || credentialID <= 0 {
		return nil, errors.New("凭据ID必须是正整数")
	}
	status := true
	request := &api.DbInstanceRequest{Name: name, DbType: dbType, Host: host, Port: port, GroupID: groupID, CredentialID: credentialID, Tags: hostExcelValue(row, header, "标签"), Description: hostExcelValue(row, header, "备注"), Status: &status}
	if err := validateDatabaseType(request.DbType); err != nil {
		return nil, err
	}
	return request, nil
}

func validateDatabaseType(value string) error {
	for _, option := range databaseTypeOptions {
		if option.Value == value {
			return nil
		}
	}
	return fmt.Errorf("不支持的数据库类型：%s", value)
}

func newHostExcelBook() (*excelize.File, string, error) {
	book := excelize.NewFile()
	sheet := "主机管理"
	if err := book.SetSheetName("Sheet1", sheet); err != nil {
		_ = book.Close()
		return nil, "", err
	}
	if err := book.SetSheetRow(sheet, "A1", &hostExcelHeaders); err != nil {
		_ = book.Close()
		return nil, "", err
	}
	style, err := book.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Color: "FFFFFF"}, Fill: excelize.Fill{Type: "pattern", Color: []string{"4F63E9"}, Pattern: 1}, Alignment: &excelize.Alignment{Vertical: "center"}})
	if err != nil {
		_ = book.Close()
		return nil, "", err
	}
	if err = book.SetCellStyle(sheet, "A1", "J1", style); err != nil {
		_ = book.Close()
		return nil, "", err
	}
	_ = book.SetPanes(sheet, &excelize.Panes{Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})
	for column, width := range map[string]float64{"A": 24, "B": 20, "C": 12, "D": 14, "E": 16, "F": 14, "G": 14, "H": 14, "I": 24, "J": 36} {
		_ = book.SetColWidth(sheet, column, column, width)
	}
	return book, sheet, nil
}

func hostExcelResponse(book *excelize.File, fileName string) (string, []byte, error) {
	buffer, err := book.WriteToBuffer()
	if err != nil {
		return "", nil, err
	}
	return fileName, bytes.Clone(buffer.Bytes()), nil
}

func hostRequestFromExcel(row []string, header map[string]int) (*api.HostRequest, error) {
	name, ip := hostExcelCell(row, header["主机名称"]), hostExcelCell(row, header["IP"])
	if name == "" || ip == "" {
		return nil, errors.New("主机名称和 IP 不能为空")
	}
	port, err := strconv.ParseInt(strings.TrimSuffix(hostExcelCell(row, header["SSH端口"]), ".0"), 10, 64)
	if err != nil || port < 1 || port > 65535 {
		return nil, errors.New("SSH端口必须是 1-65535 的整数")
	}
	credentialID, err := hostExcelOptionalID(row, header, "SSH凭据ID")
	if err != nil || credentialID <= 0 {
		return nil, errors.New("SSH凭据ID必须是正整数")
	}
	vncCredentialID, err := hostExcelOptionalID(row, header, "VNC凭据ID")
	if err != nil {
		return nil, errors.New("VNC凭据ID必须是非负整数")
	}
	groupID, err := hostExcelOptionalID(row, header, "所属分组ID")
	if err != nil {
		return nil, errors.New("所属分组ID必须是非负整数")
	}
	cloudAccountID, err := hostExcelOptionalID(row, header, "云账号ID")
	if err != nil {
		return nil, errors.New("云账号ID必须是非负整数")
	}
	hostType := hostExcelValue(row, header, "主机类型")
	if hostType == "" {
		hostType = "self"
	}
	if hostType != "self" && hostType != "cloud" {
		return nil, errors.New("主机类型必须是 self 或 cloud")
	}
	return &api.HostRequest{Name: name, IP: ip, Port: port, HostType: hostType, GroupID: groupID, CredentialID: credentialID, VNCCredentialID: vncCredentialID, CloudAccountID: cloudAccountID, Tags: hostExcelValue(row, header, "标签"), Description: hostExcelValue(row, header, "备注")}, nil
}

func hostExcelOptionalID(row []string, header map[string]int, name string) (int64, error) {
	index, exists := header[name]
	if !exists || hostExcelCell(row, index) == "" {
		return 0, nil
	}
	value := strings.TrimSuffix(hostExcelCell(row, index), ".0")
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id < 0 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}

func hostExcelCell(row []string, index int) string {
	if index < 0 || index >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[index])
}

func hostExcelValue(row []string, header map[string]int, name string) string {
	index, exists := header[name]
	if !exists {
		return ""
	}
	return hostExcelCell(row, index)
}
