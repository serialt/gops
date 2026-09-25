package biz

import (
	"context"
	"fmt"
	"gops/internal/data/api"

	"github.com/xuri/excelize/v2"
)

var hostExcelHeaders = []string{"名称", "主机", "状态", "类型", "CPU", "内存", "磁盘", "标签", "系统信息", "创建时间", "更新时间"}

func (a *AssetBiz) ExportHosts(req *api.HostListRequest) (resp []*api.HostResponse, err error) {
	req.Page, req.Limit = 1, api.MaxLimit
	page, err := a.ListHosts(context.Background(), req)

	if err != nil {
		a.log.Error("ExportHosts: get list failed", "err", err)
		return
	}
	book := excelize.NewFile()
	defer func() { _ = book.Close() }()
	sheet := "主机管理"
	_ = book.SetSheetName("Sheet1", sheet)
	headerStyle, _ := book.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Color: "FFFFFF"}, Fill: excelize.Fill{Type: "pattern", Color: []string{"4F63E9"}, Pattern: 1}, Alignment: &excelize.Alignment{Vertical: "center"}})
	_ = book.SetSheetRow(sheet, "A1", &monitorExcelHeaders)
	_ = book.SetCellStyle(sheet, "A1", "Q1", headerStyle)
	_ = book.SetPanes(sheet, &excelize.Panes{Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})
	for index, row := range page.List {
		values := []any{row.Name, fmt.Sprintf("%v:%v", row.IP, row.Port), row.Status, row.HostType}
		cell, _ := excelize.CoordinatesToCellName(1, index+2)
		_ = book.SetSheetRow(sheet, cell, &values)
	}
	for column, width := range map[string]float64{"A": 28, "B": 10, "C": 18, "D": 15, "E": 12, "F": 10, "G": 12, "H": 12, "I": 18, "J": 18, "K": 12, "L": 18, "M": 12, "N": 22, "O": 24, "P": 22, "Q": 38} {
		_ = book.SetColWidth(sheet, column, column, width)
	}

	return
}
