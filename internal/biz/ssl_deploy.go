package biz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"gops/internal/data/model"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type nginxSSHDeployTarget struct {
	HostID        int64  `json:"host_id"`
	CertPath      string `json:"cert_path"`
	KeyPath       string `json:"key_path"`
	ChainPath     string `json:"chain_path"`
	NginxBin      string `json:"nginx_bin"`
	ReloadCommand string `json:"reload_command"`
	BackupEnabled bool   `json:"backup_enabled"`
	BackupPath    string `json:"backup_path"`
}

type kubernetesSecretDeployTarget struct {
	ClusterID      int64             `json:"cluster_id"`
	Namespace      string            `json:"namespace"`
	SecretName     string            `json:"secret_name"`
	CertKey        string            `json:"cert_key"`
	KeyKey         string            `json:"key_key"`
	Labels         map[string]string `json:"labels"`
	TriggerRollout bool              `json:"trigger_rollout"`
	Deployments    []string          `json:"deployments"`
}

func (b *SSLCertBiz) testDeployTarget(ctx context.Context, config *model.SslDeployConfig) error {
	switch config.DeployType {
	case "nginx_ssh":
		target, err := parseNginxTarget(config.TargetConfig)
		if err != nil {
			return err
		}
		_, client, err := b.assets.dialHostSSH(target.HostID)
		if err != nil {
			return fmt.Errorf("连接 SSH 主机: %w", err)
		}
		defer client.Close()
		nginxBin := defaultString(target.NginxBin, "nginx")
		return runSSHCommand(client, "command -v "+shellQuote(nginxBin))
	case "k8s_secret":
		target, err := parseKubernetesTarget(config.TargetConfig)
		if err != nil {
			return err
		}
		client, err := b.kubernetes.client(target.ClusterID)
		if err != nil {
			return fmt.Errorf("连接 Kubernetes 集群: %w", err)
		}
		_, err = client.CoreV1().Namespaces().Get(ctx, target.Namespace, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("访问命名空间: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("不支持的部署类型 %q", config.DeployType)
	}
}

func (b *SSLCertBiz) runDeployTask(taskID, configID int64) {
	ctx := context.Background()
	task, err := b.repo.GetTask(ctx, b.db, taskID)
	if err != nil {
		return
	}
	startedAt := time.Now()
	task.Status, task.StartedAt, task.UpdatedAt = "running", startedAt, startedAt
	_ = b.repo.SaveTask(ctx, b.db, task)

	config, err := b.repo.GetDeployConfig(ctx, b.db, configID)
	if err == nil {
		var cert *model.SslCertificate
		cert, err = b.repo.GetCertificate(ctx, b.db, config.CertificateID)
		if err == nil {
			if strings.TrimSpace(cert.Certificate) == "" || strings.TrimSpace(cert.PrivateKey) == "" {
				err = errors.New("证书或私钥内容为空")
			} else {
				err = b.deployCertificate(ctx, cert, config)
			}
		}
	}

	finishedAt := time.Now()
	task.FinishedAt, task.UpdatedAt = finishedAt, finishedAt
	if err != nil {
		task.Status, task.ErrorMessage = "failed", err.Error()
	} else {
		task.Status, task.ErrorMessage = "success", ""
	}
	_ = b.repo.SaveTask(ctx, b.db, task)
	if config != nil {
		config.LastDeployAt, config.LastDeployOk = finishedAt, err == nil
		if err != nil {
			config.LastError = err.Error()
		} else {
			config.LastError = ""
		}
		_ = b.repo.SaveDeployConfig(ctx, b.db, config)
	}
}

func (b *SSLCertBiz) runUnsupportedRenewTask(taskID int64) {
	ctx := context.Background()
	task, err := b.repo.GetTask(ctx, b.db, taskID)
	if err != nil {
		return
	}
	now := time.Now()
	task.StartedAt, task.FinishedAt, task.UpdatedAt = now, now, now
	task.Status = "failed"
	task.ErrorMessage = "当前版本尚未配置 ACME 签发执行器，请先手动导入证书"
	_ = b.repo.SaveTask(ctx, b.db, task)
}

func (b *SSLCertBiz) deployCertificate(ctx context.Context, cert *model.SslCertificate, config *model.SslDeployConfig) error {
	switch config.DeployType {
	case "nginx_ssh":
		return b.deployToNginx(cert, config.TargetConfig)
	case "k8s_secret":
		return b.deployToKubernetes(ctx, cert, config.TargetConfig)
	default:
		return fmt.Errorf("不支持的部署类型 %q", config.DeployType)
	}
}

func (b *SSLCertBiz) deployToNginx(cert *model.SslCertificate, raw string) error {
	target, err := parseNginxTarget(raw)
	if err != nil {
		return err
	}
	_, client, err := b.assets.dialHostSSH(target.HostID)
	if err != nil {
		return fmt.Errorf("连接 SSH 主机: %w", err)
	}
	defer client.Close()
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("创建 SFTP 客户端: %w", err)
	}
	defer sftpClient.Close()
	for _, filePath := range []string{target.CertPath, target.KeyPath, target.ChainPath} {
		if filePath != "" {
			if err = sftpClient.MkdirAll(path.Dir(filePath)); err != nil {
				return fmt.Errorf("创建远程目录: %w", err)
			}
		}
	}
	if target.BackupEnabled {
		backupDir := defaultString(target.BackupPath, path.Join(path.Dir(target.CertPath), "backup"))
		if err = sftpClient.MkdirAll(backupDir); err != nil {
			return fmt.Errorf("创建备份目录: %w", err)
		}
		stamp := time.Now().Format("20060102150405")
		for _, filePath := range []string{target.CertPath, target.KeyPath, target.ChainPath} {
			if filePath == "" {
				continue
			}
			if _, statErr := sftpClient.Stat(filePath); statErr == nil {
				ext := path.Ext(filePath)
				backupName := strings.TrimSuffix(path.Base(filePath), ext) + "_" + stamp + ext
				if err = copyRemoteFile(sftpClient, filePath, path.Join(backupDir, backupName), 0); err != nil {
					return fmt.Errorf("备份远程证书: %w", err)
				}
			}
		}
	}
	certificate := cert.Certificate
	if strings.TrimSpace(cert.CertChain) != "" && target.ChainPath == "" {
		certificate = strings.TrimSpace(certificate) + "\n" + strings.TrimSpace(cert.CertChain) + "\n"
	}
	if err = writeRemoteFile(sftpClient, target.CertPath, certificate, 0o644); err != nil {
		return fmt.Errorf("上传证书: %w", err)
	}
	if err = writeRemoteFile(sftpClient, target.KeyPath, cert.PrivateKey, 0o600); err != nil {
		return fmt.Errorf("上传私钥: %w", err)
	}
	if target.ChainPath != "" && strings.TrimSpace(cert.CertChain) != "" {
		if err = writeRemoteFile(sftpClient, target.ChainPath, cert.CertChain, 0o644); err != nil {
			return fmt.Errorf("上传证书链: %w", err)
		}
	}
	nginxBin := defaultString(target.NginxBin, "nginx")
	if err = runSSHCommand(client, shellQuote(nginxBin)+" -t"); err != nil {
		return fmt.Errorf("Nginx 配置检查失败: %w", err)
	}
	reloadCommand := target.ReloadCommand
	if strings.TrimSpace(reloadCommand) == "" {
		reloadCommand = shellQuote(nginxBin) + " -s reload"
	}
	if err = runSSHCommand(client, reloadCommand); err != nil {
		return fmt.Errorf("重载 Nginx: %w", err)
	}
	return nil
}

func (b *SSLCertBiz) deployToKubernetes(ctx context.Context, cert *model.SslCertificate, raw string) error {
	target, err := parseKubernetesTarget(raw)
	if err != nil {
		return err
	}
	client, err := b.kubernetes.client(target.ClusterID)
	if err != nil {
		return fmt.Errorf("连接 Kubernetes 集群: %w", err)
	}
	certKey, keyKey := defaultString(target.CertKey, corev1.TLSCertKey), defaultString(target.KeyKey, corev1.TLSPrivateKeyKey)
	certificate := cert.Certificate
	if strings.TrimSpace(cert.CertChain) != "" {
		certificate = strings.TrimSpace(certificate) + "\n" + strings.TrimSpace(cert.CertChain) + "\n"
	}
	secrets := client.CoreV1().Secrets(target.Namespace)
	secret, err := secrets.Get(ctx, target.SecretName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = secrets.Create(ctx, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: target.SecretName, Namespace: target.Namespace, Labels: target.Labels}, Type: corev1.SecretTypeTLS, Data: map[string][]byte{certKey: []byte(certificate), keyKey: []byte(cert.PrivateKey)}}, metav1.CreateOptions{})
	} else if err == nil {
		secret.Type, secret.Labels = corev1.SecretTypeTLS, target.Labels
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}
		secret.Data[certKey], secret.Data[keyKey] = []byte(certificate), []byte(cert.PrivateKey)
		_, err = secrets.Update(ctx, secret, metav1.UpdateOptions{})
	}
	if err != nil {
		return fmt.Errorf("创建或更新 Secret: %w", err)
	}
	if target.TriggerRollout {
		for _, name := range target.Deployments {
			deployment, getErr := client.AppsV1().Deployments(target.Namespace).Get(ctx, name, metav1.GetOptions{})
			if getErr != nil {
				return fmt.Errorf("读取 Deployment %s: %w", name, getErr)
			}
			if deployment.Spec.Template.Annotations == nil {
				deployment.Spec.Template.Annotations = make(map[string]string)
			}
			deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
			if _, err = client.AppsV1().Deployments(target.Namespace).Update(ctx, deployment, metav1.UpdateOptions{}); err != nil {
				return fmt.Errorf("滚动更新 Deployment %s: %w", name, err)
			}
		}
	}
	return nil
}

func parseNginxTarget(raw string) (*nginxSSHDeployTarget, error) {
	var target nginxSSHDeployTarget
	if err := json.Unmarshal([]byte(raw), &target); err != nil {
		return nil, fmt.Errorf("解析 Nginx 部署配置: %w", err)
	}
	if target.HostID <= 0 || strings.TrimSpace(target.CertPath) == "" || strings.TrimSpace(target.KeyPath) == "" {
		return nil, errors.New("Nginx 部署配置必须包含 host_id、cert_path 和 key_path")
	}
	return &target, nil
}

func parseKubernetesTarget(raw string) (*kubernetesSecretDeployTarget, error) {
	var target kubernetesSecretDeployTarget
	if err := json.Unmarshal([]byte(raw), &target); err != nil {
		return nil, fmt.Errorf("解析 Kubernetes 部署配置: %w", err)
	}
	if target.ClusterID <= 0 || strings.TrimSpace(target.Namespace) == "" || strings.TrimSpace(target.SecretName) == "" {
		return nil, errors.New("Kubernetes 部署配置必须包含 cluster_id、namespace 和 secret_name")
	}
	return &target, nil
}

func runSSHCommand(client interface{ NewSession() (*ssh.Session, error) }, command string) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	output, err := session.CombinedOutput(command)
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func writeRemoteFile(client *sftp.Client, filePath, content string, mode os.FileMode) error {
	file, err := client.Create(filePath)
	if err != nil {
		return err
	}
	if _, err = io.WriteString(file, content); err != nil {
		_ = file.Close()
		return err
	}
	if err = file.Chmod(mode); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func copyRemoteFile(client *sftp.Client, sourcePath, targetPath string, mode os.FileMode) error {
	source, err := client.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()
	target, err := client.Create(targetPath)
	if err != nil {
		return err
	}
	defer target.Close()
	if _, err = io.Copy(target, source); err != nil {
		return err
	}
	if mode != 0 {
		return target.Chmod(mode)
	}
	return nil
}
