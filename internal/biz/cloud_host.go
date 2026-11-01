package biz

import (
	"context"
	"errors"
	"fmt"
	"strings"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	ecs "github.com/alibabacloud-go/ecs-20140526/v7/client"
	"github.com/alibabacloud-go/tea/tea"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	huaweiecs "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2"
	huaweimodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/model"
	huaweiregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/region"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tencentcvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"

	"gops/internal/data/api"
	"gops/internal/data/model"
	"gops/internal/pkg/cryptox"
	"gorm.io/gorm"
)

func (b *AssetBiz) PreviewCloudHosts(ctx context.Context, req *api.CloudHostPreviewRequest) ([]*api.CloudHostPreviewResponse, error) {
	account, err := b.repo.GetCloudAccount(ctx, b.db, req.CloudAccountID)
	if err != nil {
		return nil, err
	}
	if !account.Status {
		return nil, errors.New("云账号已禁用")
	}
	accessKey, err := cryptox.Decrypt(b.aesKey, account.AccessKey)
	if err != nil {
		return nil, fmt.Errorf("解密云账号 AccessKey: %w", err)
	}
	secretKey, err := cryptox.Decrypt(b.aesKey, account.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("解密云账号 SecretKey: %w", err)
	}
	var rows []*api.CloudHostPreviewResponse
	switch account.Provider {
	case "aliyun":
		rows, err = listAliyunHosts(accessKey, secretKey, account.Region)
	case "tencent":
		rows, err = listTencentHosts(accessKey, secretKey, account.Region)
	case "aws":
		rows, err = listAWSHosts(ctx, accessKey, secretKey, account.Region)
	case "huawei":
		rows, err = listHuaweiHosts(accessKey, secretKey, account.Region)
	default:
		err = errors.New("暂不支持该云服务商")
	}
	if err != nil {
		return nil, fmt.Errorf("拉取云主机失败: %w", err)
	}
	return rows, nil
}

func (b *AssetBiz) ImportCloudHosts(ctx context.Context, req *api.CloudHostImportRequest) (*api.CloudHostImportResponse, error) {
	if _, err := b.repo.GetCloudAccount(ctx, b.db, req.CloudAccountID); err != nil {
		return nil, err
	}
	if _, err := b.repo.GetCredential(ctx, b.db, req.CredentialID); err != nil {
		return nil, errors.New("SSH 凭据不存在")
	}
	duplicates := 0
	identities := make(map[string]struct{}, len(req.List))
	for _, item := range req.List {
		name, ip := strings.TrimSpace(item.Name), strings.TrimSpace(item.IP)
		identity := fmt.Sprintf("%s\x00%d\x00%s\x00%d", name, req.CloudAccountID, ip, item.Port)
		if _, exists := identities[identity]; exists {
			return nil, errors.New("导入列表中存在重复主机")
		}
		identities[identity] = struct{}{}
		if _, err := b.repo.GetCloudHostByIdentity(ctx, b.db, name, req.CloudAccountID, ip, item.Port); err == nil {
			duplicates++
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	if duplicates > 0 && !req.ConfirmUpdate {
		return &api.CloudHostImportResponse{DuplicateCount: duplicates, NeedsConfirm: true}, nil
	}
	result := &api.CloudHostImportResponse{DuplicateCount: duplicates}
	err := b.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range req.List {
			name, ip, port := strings.TrimSpace(item.Name), strings.TrimSpace(item.IP), item.Port
			existing, err := b.repo.GetCloudHostByIdentity(ctx, tx, name, req.CloudAccountID, ip, port)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			credentialID := req.CredentialID
			if existing == nil {
				row := &model.Host{Name: name, GroupID: req.GroupID, HostType: "cloud", IP: ip, Port: &port, CredentialID: &credentialID, CloudAccountID: req.CloudAccountID, Description: item.Description, Status: -1}
				if err := b.repo.CreateHost(ctx, tx, row); err != nil {
					return err
				}
				result.Created++
				continue
			}
			existing.GroupID, existing.CredentialID, existing.Description = req.GroupID, &credentialID, item.Description
			if err := b.repo.SaveHost(ctx, tx, existing); err != nil {
				return err
			}
			result.Updated++
		}
		return nil
	})
	return result, err
}

func listAliyunHosts(accessKey, secretKey, region string) ([]*api.CloudHostPreviewResponse, error) {
	client, err := ecs.NewClient(&openapi.Config{AccessKeyId: tea.String(accessKey), AccessKeySecret: tea.String(secretKey), RegionId: tea.String(region), Endpoint: tea.String("ecs." + region + ".aliyuncs.com")})
	if err != nil {
		return nil, err
	}
	result := make([]*api.CloudHostPreviewResponse, 0)
	for page := int32(1); ; page++ {
		response, err := client.DescribeInstances(&ecs.DescribeInstancesRequest{RegionId: tea.String(region), PageNumber: tea.Int32(page), PageSize: tea.Int32(100)})
		if err != nil {
			return nil, err
		}
		if response.Body == nil || response.Body.Instances == nil {
			break
		}
		instances := response.Body.Instances.Instance
		for _, instance := range instances {
			ip := ""
			if instance.PublicIpAddress != nil {
				ip = firstString(instance.PublicIpAddress.IpAddress)
			}
			if ip == "" && instance.VpcAttributes != nil && instance.VpcAttributes.PrivateIpAddress != nil {
				ip = firstString(instance.VpcAttributes.PrivateIpAddress.IpAddress)
			}
			if ip == "" {
				continue
			}
			result = append(result, cloudHost(tea.StringValue(instance.InstanceId), tea.StringValue(instance.InstanceName), ip, tea.StringValue(instance.Description)))
		}
		if len(instances) < 100 {
			break
		}
	}
	return result, nil
}

func listTencentHosts(accessKey, secretKey, region string) ([]*api.CloudHostPreviewResponse, error) {
	client, err := tencentcvm.NewClient(common.NewCredential(accessKey, secretKey), region, profile.NewClientProfile())
	if err != nil {
		return nil, err
	}
	result := make([]*api.CloudHostPreviewResponse, 0)
	for offset := int64(0); ; offset += 100 {
		request := tencentcvm.NewDescribeInstancesRequest()
		request.Offset = &offset
		limit := int64(100)
		request.Limit = &limit
		response, err := client.DescribeInstances(request)
		if err != nil {
			return nil, err
		}
		instances := response.Response.InstanceSet
		for _, instance := range instances {
			ip := firstString(instance.PublicIpAddresses)
			if ip == "" {
				ip = firstString(instance.PrivateIpAddresses)
			}
			if ip == "" {
				continue
			}
			result = append(result, cloudHost(stringValue(instance.InstanceId), stringValue(instance.InstanceName), ip, ""))
		}
		if len(instances) < 100 {
			break
		}
	}
	return result, nil
}

func listAWSHosts(ctx context.Context, accessKey, secretKey, region string) ([]*api.CloudHostPreviewResponse, error) {
	config, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region), awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")))
	if err != nil {
		return nil, err
	}
	paginator := awsec2.NewDescribeInstancesPaginator(awsec2.NewFromConfig(config), &awsec2.DescribeInstancesInput{})
	result := make([]*api.CloudHostPreviewResponse, 0)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, reservation := range page.Reservations {
			for _, instance := range reservation.Instances {
				ip := stringValue(instance.PublicIpAddress)
				if ip == "" {
					ip = stringValue(instance.PrivateIpAddress)
				}
				if ip == "" {
					continue
				}
				name := awsTag(instance.Tags, "Name")
				if name == "" {
					name = stringValue(instance.InstanceId)
				}
				result = append(result, cloudHost(stringValue(instance.InstanceId), name, ip, ""))
			}
		}
	}
	return result, nil
}

func listHuaweiHosts(accessKey, secretKey, regionName string) ([]*api.CloudHostPreviewResponse, error) {
	credential, err := basic.NewCredentialsBuilder().WithAk(accessKey).WithSk(secretKey).SafeBuild()
	if err != nil {
		return nil, err
	}
	cloudRegion, err := huaweiregion.SafeValueOf(regionName)
	if err != nil {
		return nil, err
	}
	hcClient, err := huaweiecs.EcsClientBuilder().WithRegion(cloudRegion).WithCredential(credential).SafeBuild()
	if err != nil {
		return nil, err
	}
	client := huaweiecs.NewEcsClient(hcClient)
	result := make([]*api.CloudHostPreviewResponse, 0)
	limit := int32(1000)
	for page := int32(1); ; page++ {
		response, err := client.ListServersDetails(&huaweimodel.ListServersDetailsRequest{Limit: &limit, Offset: &page})
		if err != nil {
			return nil, err
		}
		if response.Servers == nil {
			break
		}
		instances := *response.Servers
		for _, instance := range instances {
			ip := ""
			for _, addresses := range instance.Addresses {
				for _, address := range addresses {
					if address.Version == "4" && address.OSEXTIPStype != nil && address.OSEXTIPStype.Value() == "floating" {
						ip = address.Addr
						break
					}
					if ip == "" && address.Version == "4" {
						ip = address.Addr
					}
				}
				if ip != "" {
					break
				}
			}
			if ip == "" {
				continue
			}
			description := ""
			if instance.Description != nil {
				description = *instance.Description
			}
			result = append(result, cloudHost(instance.Id, instance.Name, ip, description))
		}
		if len(instances) < int(limit) {
			break
		}
	}
	return result, nil
}

func cloudHost(id, name, ip, description string) *api.CloudHostPreviewResponse {
	if strings.TrimSpace(name) == "" {
		name = id
	}
	return &api.CloudHostPreviewResponse{InstanceID: id, Name: name, IP: ip, Port: 22, Description: description}
}
func firstString(values []*string) string {
	if len(values) > 0 && values[0] != nil {
		return *values[0]
	}
	return ""
}
func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
func awsTag(tags []types.Tag, key string) string {
	for _, tag := range tags {
		if stringValue(tag.Key) == key {
			return stringValue(tag.Value)
		}
	}
	return ""
}
