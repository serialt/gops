package biz

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewAuthBiz,
	NewDashboardBiz,
	NewUserBiz,
	NewRoleBiz,
	NewOrganizationBiz,
	NewAssetBiz,
	NewSystemConfigBiz,
	NewMonitorBiz,
	NewTaskBiz,
	NewSSLCertBiz,
	NewKubernetesBiz,
	NewAuditBiz,
)
