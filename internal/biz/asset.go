package biz

import (
	"errors"

	"gops/internal/data/api"
	"gops/internal/data/model"
	"gops/internal/pkg/cryptox"

	"gorm.io/gorm"
)

type AssetBiz struct {
	db     *gorm.DB
	aesKey string
}

func NewAssetBiz(db *gorm.DB, aesKey string) *AssetBiz { return &AssetBiz{db: db, aesKey: aesKey} }

func (b *AssetBiz) ListGroups(req *api.AssetGroupListRequest) ([]*api.AssetGroupResponse, error) {
	var rows []model.AssetGroup
	query := b.db.Order("parent_id ASC, sort_order ASC, id ASC")
	if req.Keyword != "" {
		query = query.Where("name ILIKE ? OR code ILIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]*api.AssetGroupResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, assetGroupResponse(row))
	}
	return result, nil
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
	if err := b.db.First(&row, req.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	row.Name, row.Code, row.ParentID, row.Description, row.SortOrder, row.Status = req.Name, req.Code, req.ParentID, req.Description, req.SortOrder, *req.Status
	if err := b.db.Save(&row).Error; err != nil {
		return nil, err
	}
	return assetGroupResponse(row), nil
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

func (b *AssetBiz) ListCredentials(req *api.CredentialListRequest) ([]*api.CredentialResponse, error) {
	var rows []model.Credential
	q := b.db.Order("id DESC")
	if req.Keyword != "" {
		q = q.Where("name ILIKE ? OR username ILIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]*api.CredentialResponse, 0, len(rows))
	for _, r := range rows {
		result = append(result, credentialResponse(r))
	}
	return result, nil
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

func (b *AssetBiz) ListCloudAccounts(req *api.CloudAccountListRequest) ([]*api.CloudAccountResponse, error) {
	var rows []model.CloudAccount
	q := b.db.Order("id DESC")
	if req.Keyword != "" {
		q = q.Where("name ILIKE ? OR provider ILIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]*api.CloudAccountResponse, 0, len(rows))
	for _, r := range rows {
		result = append(result, cloudAccountResponse(r))
	}
	return result, nil
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
	accessKey, err := cryptox.Encrypt(b.aesKey, req.AccessKey)
	if err != nil {
		return model.CloudAccount{}, err
	}
	secretKey, err := cryptox.Encrypt(b.aesKey, req.SecretKey)
	if err != nil {
		return model.CloudAccount{}, err
	}
	return model.CloudAccount{Name: req.Name, Provider: req.Provider, AccessKey: accessKey, SecretKey: secretKey, Region: req.Region, Description: req.Description, Status: *req.Status}, nil
}
func cloudAccountResponse(row model.CloudAccount) *api.CloudAccountResponse {
	return &api.CloudAccountResponse{ID: row.ID, Name: row.Name, Provider: row.Provider, Region: row.Region, Description: row.Description, Status: row.Status, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
}

func (b *AssetBiz) ListHosts(req *api.HostListRequest) ([]*api.HostResponse, error) {
	var rows []model.Host
	q := b.db.Order("id DESC")
	if req.Keyword != "" {
		q = q.Where("name ILIKE ? OR ip ILIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if req.GroupID > 0 {
		q = q.Where("group_id = ?", req.GroupID)
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]*api.HostResponse, 0, len(rows))
	for _, r := range rows {
		result = append(result, hostResponse(r))
	}
	return result, nil
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
func hostModel(req *api.HostRequest) model.Host {
	return model.Host{Name: req.Name, GroupID: req.GroupID, HostType: req.HostType, IP: req.IP, Port: req.Port, SSHUser: req.SSHUser, CredentialID: req.CredentialID, CloudAccountID: req.CloudAccountID, Tags: req.Tags, Description: req.Description, Status: req.Status}
}
func hostResponse(r model.Host) *api.HostResponse {
	return &api.HostResponse{ID: r.ID, Name: r.Name, GroupID: r.GroupID, HostType: r.HostType, IP: r.IP, Port: r.Port, SSHUser: r.SSHUser, CredentialID: r.CredentialID, CloudAccountID: r.CloudAccountID, Tags: r.Tags, Description: r.Description, Status: r.Status, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

func (b *AssetBiz) ListDBInstances(req *api.DbInstanceListRequest) ([]*api.DbInstanceResponse, error) {
	var rows []model.DbInstance
	q := b.db.Order("id DESC")
	if req.Keyword != "" {
		q = q.Where("name ILIKE ? OR host ILIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]*api.DbInstanceResponse, 0, len(rows))
	for _, r := range rows {
		result = append(result, dbInstanceResponse(r))
	}
	return result, nil
}
func (b *AssetBiz) CreateDBInstance(req *api.DbInstanceRequest) (*api.DbInstanceResponse, error) {
	row, err := b.dbInstanceModel(req)
	if err != nil {
		return nil, err
	}
	if err = b.db.Create(&row).Error; err != nil {
		return nil, err
	}
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
	if req.Password == "" {
		updated.Password = row.Password
	}
	if err = b.db.Save(&updated).Error; err != nil {
		return nil, err
	}
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
func (b *AssetBiz) dbInstanceModel(req *api.DbInstanceRequest) (model.DbInstance, error) {
	password, err := cryptox.Encrypt(b.aesKey, req.Password)
	if err != nil {
		return model.DbInstance{}, err
	}
	return model.DbInstance{Name: req.Name, DbType: req.DbType, Host: req.Host, Port: req.Port, Username: req.Username, Password: password, Version: req.Version, CredentialID: req.CredentialID, Tags: req.Tags, Description: req.Description, Status: *req.Status}, nil
}
func dbInstanceResponse(r model.DbInstance) *api.DbInstanceResponse {
	return &api.DbInstanceResponse{ID: r.ID, Name: r.Name, DbType: r.DbType, Host: r.Host, Port: r.Port, Username: r.Username, Version: r.Version, CredentialID: r.CredentialID, Tags: r.Tags, Description: r.Description, Status: r.Status, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

func (b *AssetBiz) ListTerminalSessions(req *api.TerminalSessionListRequest) ([]*api.TerminalSessionResponse, error) {
	var rows []model.TerminalSession
	q := b.db.Order("started_at DESC")
	if req.SessionType != "" {
		q = q.Where("session_type = ?", req.SessionType)
	}
	if req.Keyword != "" {
		q = q.Where("CAST(user_id AS TEXT) ILIKE ? OR CAST(host_id AS TEXT) ILIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
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
