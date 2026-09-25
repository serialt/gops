package biz

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"strings"
	"time"

	"gops/internal/config"
	"gops/internal/data/api"
	"gops/internal/data/model"
	"gops/internal/data/repo"
	"gops/internal/pkg/collector"
	"gops/internal/pkg/cryptox"

	"gorm.io/gorm"

	mysql "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type AssetBiz struct {
	db     *gorm.DB
	repo   *repo.AssetRepo
	aesKey string
	log    *slog.Logger
}

var cloudProviderOptions = []*api.CloudProviderOptionResponse{
	{Value: "aliyun", Label: "阿里云", Regions: cloudRegions(
		"华东 1（杭州）|cn-hangzhou", "华东 2（上海）|cn-shanghai", "华北 1（青岛）|cn-qingdao", "华北 2（北京）|cn-beijing", "华北 3（张家口）|cn-zhangjiakou", "华北 5（呼和浩特）|cn-huhehaote", "华北 6（乌兰察布）|cn-wulanchabu", "华南 1（深圳）|cn-shenzhen", "华南 2（河源）|cn-heyuan", "华南 3（广州）|cn-guangzhou", "西南 1（成都）|cn-chengdu",
		"中国香港|cn-hongkong", "新加坡|ap-southeast-1", "澳大利亚（悉尼）|ap-southeast-2", "马来西亚（吉隆坡）|ap-southeast-3", "印度尼西亚（雅加达）|ap-southeast-5", "菲律宾（马尼拉）|ap-southeast-6", "泰国（曼谷）|ap-southeast-7", "日本（东京）|ap-northeast-1", "韩国（首尔）|ap-northeast-2", "美国（硅谷）|us-west-1", "美国（弗吉尼亚）|us-east-1", "德国（法兰克福）|eu-central-1", "英国（伦敦）|eu-west-1", "阿联酋（迪拜）|me-east-1", "沙特（利雅得）|me-central-1",
	)},
	{Value: "huawei", Label: "华为云", Regions: cloudRegions(
		"华北-北京一|cn-north-1", "华北-北京四|cn-north-4", "华北-乌兰察布一|cn-north-9", "华北-乌兰察布-汽车一|cn-north-11", "华北三|cn-north-12", "华东-上海二|cn-east-2", "华东-上海一|cn-east-3", "华东二|cn-east-4", "华东-青岛|cn-east-5", "华南-广州|cn-south-1", "华南-广州-友好用户环境|cn-south-4", "西南-贵阳一|cn-southwest-2",
		"中国-香港|ap-southeast-1", "亚太-曼谷|ap-southeast-2", "亚太-新加坡|ap-southeast-3", "亚太-雅加达|ap-southeast-4", "亚太-马尼拉|ap-southeast-5", "非洲-约翰内斯堡|af-south-1", "非洲-开罗|af-north-1", "中东-利雅得|me-east-1", "土耳其-伊斯坦布尔|tr-west-1", "拉美-墨西哥城一|na-mexico-1", "拉美-墨西哥城二|la-north-2", "拉美-圣保罗一|sa-brazil-1", "拉美-圣地亚哥|la-south-2",
	)},
	{Value: "tencent", Label: "腾讯云", Regions: cloudRegions(
		"华北地区（北京）|ap-beijing", "华东地区（上海）|ap-shanghai", "华东地区（南京）|ap-nanjing", "华南地区（广州）|ap-guangzhou", "西南地区（成都）|ap-chengdu", "西南地区（重庆）|ap-chongqing", "华南地区（深圳金融）|ap-shenzhen-fsi", "华东地区（上海金融）|ap-shanghai-fsi",
		"中国香港|ap-hongkong", "新加坡|ap-singapore", "马来西亚（新山）|ap-johorbahru", "印度尼西亚（雅加达）|ap-jakarta", "泰国（曼谷）|ap-bangkok", "日本（东京）|ap-tokyo", "日本（大阪）|ap-osaka", "韩国（首尔）|ap-seoul", "印度（孟买）|ap-mumbai", "美国（硅谷）|na-siliconvalley", "美国（弗吉尼亚）|na-ashburn", "德国（法兰克福）|eu-frankfurt", "俄罗斯（莫斯科）|eu-moscow",
	)},
	{Value: "aws", Label: "AWS", Regions: cloudRegions(
		"中国（北京）|cn-north-1", "中国（宁夏）|cn-northwest-1",
		"亚太（香港）|ap-east-1", "亚太（台北）|ap-east-2", "亚太（孟买）|ap-south-1", "亚太（海得拉巴）|ap-south-2", "亚太（新加坡）|ap-southeast-1", "亚太（悉尼）|ap-southeast-2", "亚太（雅加达）|ap-southeast-3", "亚太（墨尔本）|ap-southeast-4", "亚太（马来西亚）|ap-southeast-5", "亚太（新西兰）|ap-southeast-6", "亚太（曼谷）|ap-southeast-7", "亚太（东京）|ap-northeast-1", "亚太（首尔）|ap-northeast-2", "亚太（大阪）|ap-northeast-3", "美国东部（弗吉尼亚北部）|us-east-1", "美国东部（俄亥俄）|us-east-2", "美国西部（加利福尼亚北部）|us-west-1", "美国西部（俄勒冈）|us-west-2", "加拿大（中部）|ca-central-1", "加拿大西部（卡尔加里）|ca-west-1", "南美洲（圣保罗）|sa-east-1", "墨西哥（中部）|mx-central-1", "欧洲（爱尔兰）|eu-west-1", "欧洲（伦敦）|eu-west-2", "欧洲（巴黎）|eu-west-3", "欧洲（法兰克福）|eu-central-1", "欧洲（苏黎世）|eu-central-2", "欧洲（斯德哥尔摩）|eu-north-1", "欧洲（米兰）|eu-south-1", "欧洲（西班牙）|eu-south-2", "中东（巴林）|me-south-1", "中东（阿联酋）|me-central-1", "以色列（特拉维夫）|il-central-1", "非洲（开普敦）|af-south-1",
	)},
}

var databaseTypeOptions = []*api.DatabaseTypeOptionResponse{
	{Value: "postgresql", Label: "PostgreSQL", DefaultPort: 5432},
	{Value: "mysql", Label: "MySQL", DefaultPort: 3306},
	{Value: "oracle", Label: "Oracle", DefaultPort: 1521},
	{Value: "sqlserver", Label: "SQL Server", DefaultPort: 1433},
	{Value: "mariadb", Label: "MariaDB", DefaultPort: 3306},
}

func cloudRegions(values ...string) []*api.CloudRegionOptionResponse {
	regions := make([]*api.CloudRegionOptionResponse, 0, len(values))
	for _, value := range values {
		parts := strings.SplitN(value, "|", 2)
		regions = append(regions, &api.CloudRegionOptionResponse{Label: parts[0], Value: parts[1]})
	}
	return regions
}

func NewAssetBiz(db *gorm.DB, assetRepo *repo.AssetRepo, cfg *config.Config, log *slog.Logger) *AssetBiz {
	return &AssetBiz{db: db, repo: assetRepo, aesKey: cfg.Server.AESKey, log: log}
}

// CloudProviderOptions 返回后端定义的云服务商与区域枚举。
func (b *AssetBiz) CloudProviderOptions() []*api.CloudProviderOptionResponse {
	return cloudProviderOptions
}

// DatabaseTypeOptions returns backend-defined database type options.
func (b *AssetBiz) DatabaseTypeOptions() []*api.DatabaseTypeOptionResponse {
	return databaseTypeOptions
}

func (b *AssetBiz) ListGroups(ctx context.Context, req *api.AssetGroupListRequest) (*api.PageResponse[*api.AssetGroupResponse], error) {
	rows, total, err := b.repo.ListGroups(ctx, b.db, req.Keyword, req.Offset(), req.Limit)
	if err != nil {
		return nil, err
	}
	result := make([]*api.AssetGroupResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, assetGroupResponse(*row))
	}
	return &api.PageResponse[*api.AssetGroupResponse]{List: result, Total: total}, nil
}

func (b *AssetBiz) CreateGroup(req *api.AssetGroupRequest) (*api.AssetGroupResponse, error) {
	if req.ParentID > 0 {
		var parent model.AssetGroup
		if err := b.db.First(&parent, req.ParentID).Error; err != nil {
			return nil, err
		}
	}
	row := model.AssetGroup{Name: req.Name, Code: req.Code, ParentID: req.ParentID, Description: req.Description, SortOrder: req.SortOrder, Status: *req.Status}
	if err := b.db.Create(&row).Error; err != nil {
		return nil, err
	}
	return assetGroupResponse(row), nil
}

func (b *AssetBiz) UpdateGroup(req *api.AssetGroupRequest) (*api.AssetGroupResponse, error) {
	if req.ID == 0 || req.ID == req.ParentID {
		return nil, ErrInvalid
	}
	var row model.AssetGroup
	if err := b.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&row, req.ID).Error; err != nil {
			return err
		}

		wasEnabled := row.Status
		row.Name, row.Code, row.ParentID, row.Description, row.SortOrder, row.Status = req.Name, req.Code, req.ParentID, req.Description, req.SortOrder, *req.Status
		if err := tx.Save(&row).Error; err != nil {
			return err
		}

		if wasEnabled && !row.Status {
			return updateAssetGroupDescendantStatus(tx, row.ID, false)
		}
		if !wasEnabled && row.Status {
			return updateAssetGroupAncestorStatus(tx, row.ParentID, true)
		}
		return nil
	}); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return assetGroupResponse(row), nil
}

// updateAssetGroupDescendantStatus 同步更新指定分组的所有子孙分组状态。
func updateAssetGroupDescendantStatus(tx *gorm.DB, groupID int64, status bool) error {
	var groups []model.AssetGroup
	if err := tx.Find(&groups).Error; err != nil {
		return err
	}
	childIDs := make(map[int64][]int64, len(groups))
	for _, group := range groups {
		childIDs[group.ParentID] = append(childIDs[group.ParentID], group.ID)
	}
	ids := make([]int64, 0)
	queue := append([]int64(nil), childIDs[groupID]...)
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		ids = append(ids, id)
		queue = append(queue, childIDs[id]...)
	}
	if len(ids) == 0 {
		return nil
	}
	return tx.Model(&model.AssetGroup{}).Where("id IN ?", ids).Update("status", status).Error
}

// updateAssetGroupAncestorStatus 同步更新指定分组的所有父级分组状态。
func updateAssetGroupAncestorStatus(tx *gorm.DB, parentID int64, status bool) error {
	if parentID == 0 {
		return nil
	}
	var groups []model.AssetGroup
	if err := tx.Find(&groups).Error; err != nil {
		return err
	}
	parentIDs := make(map[int64]int64, len(groups))
	for _, group := range groups {
		parentIDs[group.ID] = group.ParentID
	}
	ids := make([]int64, 0)
	seen := make(map[int64]struct{})
	for parentID != 0 {
		if _, ok := seen[parentID]; ok {
			return ErrInvalid
		}
		seen[parentID] = struct{}{}
		ids = append(ids, parentID)
		parentID = parentIDs[parentID]
	}
	return tx.Model(&model.AssetGroup{}).Where("id IN ?", ids).Update("status", status).Error
}

func (b *AssetBiz) DeleteGroup(req *api.AssetGroupIDRequest) error {
	var children int64
	if err := b.db.Model(&model.AssetGroup{}).Where("parent_id = ?", req.ID).Count(&children).Error; err != nil {
		return err
	}
	if children > 0 {
		return errors.New("remove child asset groups first")
	}
	var hosts int64
	if err := b.db.Model(&model.Host{}).Where("group_id = ?", req.ID).Count(&hosts).Error; err != nil {
		return err
	}
	if hosts > 0 {
		return errors.New("remove hosts from asset group first")
	}
	result := b.db.Delete(&model.AssetGroup{}, req.ID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func assetGroupResponse(row model.AssetGroup) *api.AssetGroupResponse {
	return &api.AssetGroupResponse{ID: row.ID, Name: row.Name, Code: row.Code, ParentID: row.ParentID, Description: row.Description, SortOrder: row.SortOrder, Status: row.Status, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
}

func (b *AssetBiz) ListCredentials(ctx context.Context, req *api.CredentialListRequest) (*api.PageResponse[*api.CredentialResponse], error) {
	rows, total, err := b.repo.ListCredentials(ctx, b.db, req.Keyword, req.Offset(), req.Limit)
	if err != nil {
		return nil, err
	}
	result := make([]*api.CredentialResponse, 0, len(rows))
	for _, r := range rows {
		result = append(result, credentialResponse(*r))
	}
	return &api.PageResponse[*api.CredentialResponse]{List: result, Total: total}, nil
}
func (b *AssetBiz) CreateCredential(req *api.CredentialRequest) (*api.CredentialResponse, error) {
	row, err := b.credentialModel(req)
	if err != nil {
		return nil, err
	}
	if err = b.db.Create(&row).Error; err != nil {
		return nil, err
	}
	return credentialResponse(row), nil
}
func (b *AssetBiz) UpdateCredential(req *api.CredentialRequest) (*api.CredentialResponse, error) {
	if req.ID == 0 {
		return nil, ErrInvalid
	}
	var row model.Credential
	if err := b.db.First(&row, req.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	updated, err := b.credentialModel(req)
	if err != nil {
		return nil, err
	}
	updated.ID = row.ID
	updated.CreatedAt = row.CreatedAt
	// 编辑时未输入敏感字段表示保持原凭证，避免前端不回显密文时误清空 SSH 登录信息。
	if req.Password == "" {
		updated.Password = row.Password
	}
	if req.PrivateKey == "" {
		updated.PrivateKey = row.PrivateKey
	}
	if req.Passphrase == "" {
		updated.Passphrase = row.Passphrase
	}
	if err = b.db.Save(&updated).Error; err != nil {
		return nil, err
	}
	return credentialResponse(updated), nil
}
func (b *AssetBiz) DeleteCredential(req *api.CredentialIDRequest) error {
	result := b.db.Delete(&model.Credential{}, req.ID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
func (b *AssetBiz) credentialModel(req *api.CredentialRequest) (model.Credential, error) {
	password, err := cryptox.Encrypt(b.aesKey, req.Password)
	if err != nil {
		return model.Credential{}, err
	}
	privateKey, err := cryptox.Encrypt(b.aesKey, req.PrivateKey)
	if err != nil {
		return model.Credential{}, err
	}
	passphrase, err := cryptox.Encrypt(b.aesKey, req.Passphrase)
	if err != nil {
		return model.Credential{}, err
	}
	return model.Credential{Name: req.Name, CredentialType: req.CredentialType, Username: req.Username, Password: password, PrivateKey: privateKey, Passphrase: passphrase, Description: req.Description}, nil
}
func credentialResponse(row model.Credential) *api.CredentialResponse {
	return &api.CredentialResponse{ID: row.ID, Name: row.Name, CredentialType: row.CredentialType, Username: row.Username, Description: row.Description, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
}

func (b *AssetBiz) ListCloudAccounts(ctx context.Context, req *api.CloudAccountListRequest) (*api.PageResponse[*api.CloudAccountResponse], error) {
	rows, total, err := b.repo.ListCloudAccounts(ctx, b.db, req.Keyword, req.Offset(), req.Limit)
	if err != nil {
		return nil, err
	}
	result := make([]*api.CloudAccountResponse, 0, len(rows))
	for _, r := range rows {
		result = append(result, cloudAccountResponse(*r))
	}
	return &api.PageResponse[*api.CloudAccountResponse]{List: result, Total: total}, nil
}
func (b *AssetBiz) CreateCloudAccount(req *api.CloudAccountRequest) (*api.CloudAccountResponse, error) {
	row, err := b.cloudAccountModel(req)
	if err != nil {
		return nil, err
	}
	if err = b.db.Create(&row).Error; err != nil {
		return nil, err
	}
	return cloudAccountResponse(row), nil
}
func (b *AssetBiz) UpdateCloudAccount(req *api.CloudAccountRequest) (*api.CloudAccountResponse, error) {
	if req.ID == 0 {
		return nil, ErrInvalid
	}
	var row model.CloudAccount
	if err := b.db.First(&row, req.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	updated, err := b.cloudAccountModel(req)
	if err != nil {
		return nil, err
	}
	updated.ID = row.ID
	updated.CreatedAt = row.CreatedAt
	if req.AccessKey == "" {
		updated.AccessKey = row.AccessKey
	}
	if req.SecretKey == "" {
		updated.SecretKey = row.SecretKey
	}
	if err = b.db.Save(&updated).Error; err != nil {
		return nil, err
	}
	return cloudAccountResponse(updated), nil
}
func (b *AssetBiz) DeleteCloudAccount(req *api.CloudAccountIDRequest) error {
	result := b.db.Delete(&model.CloudAccount{}, req.ID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
func (b *AssetBiz) cloudAccountModel(req *api.CloudAccountRequest) (model.CloudAccount, error) {
	provider, ok := findCloudProviderOption(req.Provider)
	if !ok {
		return model.CloudAccount{}, errors.New("unsupported cloud provider")
	}
	region := req.Region
	if region == "" {
		region = provider.Regions[0].Value
	}
	if !cloudProviderHasRegion(provider, region) {
		return model.CloudAccount{}, errors.New("unsupported cloud region")
	}
	accessKey, err := cryptox.Encrypt(b.aesKey, req.AccessKey)
	if err != nil {
		return model.CloudAccount{}, err
	}
	secretKey, err := cryptox.Encrypt(b.aesKey, req.SecretKey)
	if err != nil {
		return model.CloudAccount{}, err
	}
	return model.CloudAccount{Name: req.Name, Provider: provider.Value, AccessKey: accessKey, SecretKey: secretKey, Region: region, Description: req.Description, Status: *req.Status}, nil
}

func findCloudProviderOption(value string) (*api.CloudProviderOptionResponse, bool) {
	for _, provider := range cloudProviderOptions {
		if provider.Value == value {
			return provider, true
		}
	}
	return nil, false
}

func cloudProviderHasRegion(provider *api.CloudProviderOptionResponse, value string) bool {
	for _, region := range provider.Regions {
		if region.Value == value {
			return true
		}
	}
	return false
}
func cloudAccountResponse(row model.CloudAccount) *api.CloudAccountResponse {
	return &api.CloudAccountResponse{ID: row.ID, Name: row.Name, Provider: row.Provider, Region: row.Region, Description: row.Description, Status: row.Status, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
}

func (b *AssetBiz) ListHosts(ctx context.Context, req *api.HostListRequest) (*api.PageResponse[*api.HostResponse], error) {
	rows, total, err := b.repo.ListHosts(ctx, b.db, req.Keyword, req.GroupID, req.Offset(), req.Limit)
	if err != nil {
		return nil, err
	}
	result := make([]*api.HostResponse, 0, len(rows))
	for _, r := range rows {
		result = append(result, hostResponse(*r))
	}
	return &api.PageResponse[*api.HostResponse]{List: result, Total: total}, nil
}
func (b *AssetBiz) CreateHost(req *api.HostRequest) (*api.HostResponse, error) {
	row := hostModel(req)
	if err := b.db.Create(&row).Error; err != nil {
		return nil, err
	}
	return hostResponse(row), nil
}
func (b *AssetBiz) UpdateHost(req *api.HostRequest) (*api.HostResponse, error) {
	if req.ID == 0 {
		return nil, ErrInvalid
	}
	var row model.Host
	if err := b.db.First(&row, req.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	updated := hostModel(req)
	updated.ID = row.ID
	updated.CreatedAt = row.CreatedAt
	updated.LastSeenAt, updated.Os, updated.Kernel, updated.Arch = row.LastSeenAt, row.Os, row.Kernel, row.Arch
	updated.CPUInfo, updated.CPUCores, updated.CPUUsage = row.CPUInfo, row.CPUCores, row.CPUUsage
	updated.MemoryTotal, updated.MemoryUsed, updated.MemoryUsage = row.MemoryTotal, row.MemoryUsed, row.MemoryUsage
	updated.DiskTotal, updated.DiskUsed, updated.DiskUsage = row.DiskTotal, row.DiskUsed, row.DiskUsage
	updated.Uptime, updated.Hostname = row.Uptime, row.Hostname
	if err := b.db.Save(&updated).Error; err != nil {
		return nil, err
	}
	return hostResponse(updated), nil
}
func (b *AssetBiz) DeleteHost(req *api.HostIDRequest) error {
	result := b.db.Delete(&model.Host{}, req.ID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (b *AssetBiz) CollectHostInfo(req *api.HostIDRequest) (*api.HostResponse, error) {
	host, client, err := b.dialHostSSH(req.ID)
	if err != nil {
		_ = b.db.Model(&model.Host{}).Where("id = ?", req.ID).Update("status", 0).Error
		return nil, err
	}
	defer client.Close()
	info, err := collector.New(sshCommandExecutor{client: client}).CollectAll()
	if err != nil {
		_ = b.db.Model(host).Update("status", 0).Error
		return nil, err
	}
	now := time.Now()
	updates := map[string]any{"status": 1, "last_seen_at": now, "os": info.OS, "kernel": info.Kernel, "arch": info.Arch, "cpu_info": info.CPU.JSON(), "cpu_cores": info.CPU.Cores, "cpu_usage": info.CPU.Usage, "memory_total": info.Memory.Total, "memory_used": info.Memory.Used, "memory_usage": info.Memory.Usage, "disk_total": info.Disk.Total, "disk_used": info.Disk.Used, "disk_usage": info.Disk.Usage, "uptime": info.Uptime, "hostname": info.Hostname}
	if err := b.db.Model(host).Updates(updates).Error; err != nil {
		return nil, err
	}
	if err := b.db.First(host, host.ID).Error; err != nil {
		return nil, err
	}
	return hostResponse(*host), nil
}
func hostModel(req *api.HostRequest) model.Host {
	return model.Host{Name: req.Name, GroupID: req.GroupID, HostType: req.HostType, IP: req.IP, Port: req.Port, SSHUser: req.SSHUser, CredentialID: req.CredentialID, CloudAccountID: req.CloudAccountID, Tags: req.Tags, Description: req.Description, Status: req.Status}
}
func hostResponse(r model.Host) *api.HostResponse {
	return &api.HostResponse{ID: r.ID, Name: r.Name, GroupID: r.GroupID, HostType: r.HostType, IP: r.IP, Port: r.Port, SSHUser: r.SSHUser, CredentialID: r.CredentialID, CloudAccountID: r.CloudAccountID, Tags: r.Tags, Description: r.Description, Status: r.Status, LastSeenAt: r.LastSeenAt, OS: r.Os, Kernel: r.Kernel, Arch: r.Arch, CPUInfo: r.CPUInfo, CPUCores: r.CPUCores, CPUUsage: r.CPUUsage, MemoryTotal: r.MemoryTotal, MemoryUsed: r.MemoryUsed, MemoryUsage: r.MemoryUsage, DiskTotal: r.DiskTotal, DiskUsed: r.DiskUsed, DiskUsage: r.DiskUsage, Uptime: r.Uptime, Hostname: r.Hostname, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

func (b *AssetBiz) ListDBInstances(ctx context.Context, req *api.DbInstanceListRequest) (*api.PageResponse[*api.DbInstanceResponse], error) {
	rows, total, err := b.repo.ListDBInstances(ctx, b.db, req.Keyword, req.GroupID, req.Offset(), req.Limit)
	if err != nil {
		return nil, err
	}
	result := make([]*api.DbInstanceResponse, 0, len(rows))
	for _, r := range rows {
		result = append(result, dbInstanceResponse(*r))
	}
	return &api.PageResponse[*api.DbInstanceResponse]{List: result, Total: total}, nil
}
func (b *AssetBiz) CreateDBInstance(req *api.DbInstanceRequest) (*api.DbInstanceResponse, error) {
	row, err := b.dbInstanceModel(req)
	if err != nil {
		return nil, err
	}
	if err = b.db.Create(&row).Error; err != nil {
		return nil, err
	}
	b.collectDBInstanceAsync(row.ID)
	return dbInstanceResponse(row), nil
}
func (b *AssetBiz) UpdateDBInstance(req *api.DbInstanceRequest) (*api.DbInstanceResponse, error) {
	if req.ID == 0 {
		return nil, ErrInvalid
	}
	var row model.DbInstance
	if err := b.db.First(&row, req.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	updated, err := b.dbInstanceModel(req)
	if err != nil {
		return nil, err
	}
	updated.ID = row.ID
	updated.CreatedAt = row.CreatedAt
	if err = b.db.Save(&updated).Error; err != nil {
		return nil, err
	}
	b.collectDBInstanceAsync(updated.ID)
	return dbInstanceResponse(updated), nil
}
func (b *AssetBiz) DeleteDBInstance(req *api.DbInstanceIDRequest) error {
	result := b.db.Delete(&model.DbInstance{}, req.ID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (b *AssetBiz) CollectDBInstanceInfo(ctx context.Context, req *api.DbInstanceIDRequest) (*api.DbInstanceResponse, error) {
	instance, credential, err := b.repo.GetDBInstanceAndCredential(ctx, b.db, req.ID)
	if err != nil {
		return nil, err
	}
	if credential.CredentialType != "password" {
		return nil, fmt.Errorf("数据库采集仅支持用户名密码凭据")
	}
	username, password := credential.Username, ""
	password, err = cryptox.Decrypt(b.aesKey, credential.Password)
	if err != nil {
		return nil, fmt.Errorf("解密数据库凭据: %w", err)
	}
	collectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	version, err := collectDatabaseVersion(collectCtx, instance, username, password)
	if err != nil {
		_, _ = b.repo.UpdateDBInstanceCollectedInfo(ctx, b.db, instance.ID, false, "")
		return nil, err
	}
	updated, err := b.repo.UpdateDBInstanceCollectedInfo(ctx, b.db, instance.ID, true, version)
	if err != nil {
		return nil, err
	}
	return dbInstanceResponse(*updated), nil
}

func (b *AssetBiz) TestDBInstanceConnection(ctx context.Context, req *api.DbInstanceTestRequest) (*api.DbInstanceTestResponse, error) {
	credential, err := b.repo.GetCredential(ctx, b.db, req.CredentialID)
	if err != nil {
		return nil, fmt.Errorf("读取数据库凭据失败: %w", err)
	}
	if credential.CredentialType != "password" {
		return nil, fmt.Errorf("数据库连接仅支持用户名密码凭据")
	}
	password, err := cryptox.Decrypt(b.aesKey, credential.Password)
	if err != nil {
		return nil, fmt.Errorf("解密数据库凭据: %w", err)
	}
	testCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	instance := &model.DbInstance{DbType: req.DbType, Host: req.Host, Port: req.Port}
	version, err := collectDatabaseVersion(testCtx, instance, credential.Username, password)
	if err != nil {
		return nil, err
	}
	return &api.DbInstanceTestResponse{Version: version}, nil
}

func (b *AssetBiz) collectDBInstanceAsync(id int64) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if _, err := b.CollectDBInstanceInfo(ctx, &api.DbInstanceIDRequest{ID: id}); err != nil {
			b.log.Warn("异步采集数据库信息失败", "database_id", id, "error", err)
		}
	}()
}

func collectDatabaseVersion(ctx context.Context, instance *model.DbInstance, username, password string) (string, error) {
	address := net.JoinHostPort(instance.Host, fmt.Sprintf("%d", instance.Port))
	var driver, dsn, versionQuery string
	switch instance.DbType {
	case "mysql", "mariadb":
		cfg := mysql.NewConfig()
		cfg.User, cfg.Passwd, cfg.Net, cfg.Addr, cfg.DBName = username, password, "tcp", address, ""
		cfg.Timeout, cfg.ReadTimeout, cfg.WriteTimeout = 8*time.Second, 8*time.Second, 8*time.Second
		driver, dsn, versionQuery = "mysql", cfg.FormatDSN(), "SELECT VERSION()"
	case "postgresql":
		u := &url.URL{Scheme: "postgres", User: url.UserPassword(username, password), Host: address, Path: "/postgres"}
		queryValues := u.Query()
		queryValues.Set("sslmode", "disable")
		u.RawQuery = queryValues.Encode()
		driver, dsn, versionQuery = "pgx", u.String(), "SHOW server_version"
	default:
		connection, err := (&net.Dialer{Timeout: 8 * time.Second}).DialContext(ctx, "tcp", address)
		if err != nil {
			return "", fmt.Errorf("数据库连接失败: %w", err)
		}
		_ = connection.Close()
		return instance.Version, nil
	}
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return "", fmt.Errorf("创建数据库连接: %w", err)
	}
	defer db.Close()
	var version string
	if err = db.QueryRowContext(ctx, versionQuery).Scan(&version); err != nil {
		return "", fmt.Errorf("采集数据库信息失败: %w", err)
	}
	return strings.TrimSpace(version), nil
}
func (b *AssetBiz) dbInstanceModel(req *api.DbInstanceRequest) (model.DbInstance, error) {
	var databaseType *api.DatabaseTypeOptionResponse
	for _, option := range databaseTypeOptions {
		if option.Value == req.DbType {
			databaseType = option
			break
		}
	}
	if databaseType == nil || req.Port < 1 || req.CredentialID <= 0 {
		return model.DbInstance{}, ErrInvalid
	}
	return model.DbInstance{Name: req.Name, GroupID: req.GroupID, DbType: req.DbType, Host: req.Host, Port: req.Port, Version: req.Version, CredentialID: req.CredentialID, Tags: req.Tags, Description: req.Description, Status: *req.Status}, nil
}
func dbInstanceResponse(r model.DbInstance) *api.DbInstanceResponse {
	return &api.DbInstanceResponse{ID: r.ID, GroupID: r.GroupID, Name: r.Name, DbType: r.DbType, Host: r.Host, Port: r.Port, Version: r.Version, CredentialID: r.CredentialID, Tags: r.Tags, Description: r.Description, Status: r.Status, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

func (b *AssetBiz) ListTerminalSessions(req *api.TerminalSessionListRequest) ([]*api.TerminalSessionResponse, error) {
	var rows []model.TerminalSession
	q := b.db.Order("started_at DESC")
	if req.SessionType != "" {
		q = q.Where("session_type = ?", req.SessionType)
	}
	if req.Keyword != "" {
		q = q.Where("CAST(user_id AS TEXT) LIKE ? OR CAST(host_id AS TEXT) LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]*api.TerminalSessionResponse, 0, len(rows))
	for _, r := range rows {
		result = append(result, terminalSessionResponse(r))
	}
	return result, nil
}
func (b *AssetBiz) GetTerminalSessionRecording(req *api.TerminalSessionIDRequest) (*api.TerminalSessionRecordingResponse, error) {
	var row model.TerminalSession
	if err := b.db.First(&row, req.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &api.TerminalSessionRecordingResponse{ID: row.ID, Recording: row.Recording}, nil
}
func (b *AssetBiz) DeleteTerminalSession(req *api.TerminalSessionIDRequest) error {
	result := b.db.Delete(&model.TerminalSession{}, req.ID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
func terminalSessionResponse(r model.TerminalSession) *api.TerminalSessionResponse {
	return &api.TerminalSessionResponse{ID: r.ID, UserID: r.UserID, HostID: r.HostID, DbInstanceID: r.DbInstanceID, SessionType: r.SessionType, StartedAt: r.StartedAt, EndedAt: r.EndedAt, DurationSeconds: r.DurationSeconds, CommandCount: r.CommandCount}
}
