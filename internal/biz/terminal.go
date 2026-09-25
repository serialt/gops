package biz

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"gops/internal/data/model"
	"gops/internal/pkg/cryptox"

	"golang.org/x/crypto/ssh"
)

const maxTerminalRecordingBytes = 1024 * 1024

// SSHConnection is the server-side state of one interactive SSH terminal.
type SSHConnection struct {
	hostID       int64
	userID       int64
	startedAt    time.Time
	client       *ssh.Client
	session      *ssh.Session
	input        io.WriteCloser
	output       io.Reader
	recording    strings.Builder
	commandCount int64
	mu           sync.Mutex
}

// StartSSH opens an SSH shell using the credential associated with a host.
func (b *AssetBiz) StartSSH(hostID, userID int64, cols, rows int) (*SSHConnection, error) {
	host, client, err := b.dialHostSSH(hostID)
	if err != nil {
		return nil, err
	}
	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("create SSH session: %w", err)
	}
	if cols < 20 {
		cols = 120
	}
	if rows < 5 {
		rows = 32
	}
	if err = session.RequestPty("xterm-256color", rows, cols, ssh.TerminalModes{ssh.ECHO: 1}); err != nil {
		session.Close()
		client.Close()
		return nil, fmt.Errorf("request terminal: %w", err)
	}
	input, err := session.StdinPipe()
	if err != nil {
		session.Close()
		client.Close()
		return nil, err
	}
	output, err := session.StdoutPipe()
	if err != nil {
		input.Close()
		session.Close()
		client.Close()
		return nil, err
	}
	if err = session.Shell(); err != nil {
		input.Close()
		session.Close()
		client.Close()
		return nil, fmt.Errorf("start shell: %w", err)
	}
	return &SSHConnection{hostID: host.ID, userID: userID, startedAt: time.Now(), client: client, session: session, input: input, output: output}, nil
}

func (b *AssetBiz) dialHostSSH(hostID int64) (*model.Host, *ssh.Client, error) {
	var host model.Host
	if err := b.db.First(&host, hostID).Error; err != nil {
		return nil, nil, ErrNotFound
	}
	if host.CredentialID == 0 {
		return nil, nil, fmt.Errorf("主机未配置凭证")
	}
	var credential model.Credential
	if err := b.db.First(&credential, host.CredentialID).Error; err != nil {
		return nil, nil, fmt.Errorf("加载主机凭证: %w", err)
	}
	password, err := cryptox.Decrypt(b.aesKey, credential.Password)
	if err != nil {
		return nil, nil, fmt.Errorf("解密凭证密码: %w", err)
	}
	privateKey, err := cryptox.Decrypt(b.aesKey, credential.PrivateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("解密凭证私钥: %w", err)
	}
	passphrase, err := cryptox.Decrypt(b.aesKey, credential.Passphrase)
	if err != nil {
		return nil, nil, fmt.Errorf("解密凭证口令: %w", err)
	}

	var method ssh.AuthMethod
	if credential.CredentialType == "key" {
		var signer ssh.Signer
		if passphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(privateKey), []byte(passphrase))
		} else {
			signer, err = ssh.ParsePrivateKey([]byte(privateKey))
		}
		if err != nil {
			return nil, nil, fmt.Errorf("解析 SSH 私钥: %w", err)
		}
		method = ssh.PublicKeys(signer)
	} else {
		method = ssh.Password(password)
	}
	sshUser := host.SSHUser
	if sshUser == "" {
		sshUser = credential.Username
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host.IP, host.Port), &ssh.ClientConfig{
		User: sshUser, Auth: []ssh.AuthMethod{method}, HostKeyCallback: ssh.InsecureIgnoreHostKey(), Timeout: 10 * time.Second,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("连接 SSH: %w", err)
	}
	return &host, client, nil
}

type sshCommandExecutor struct{ client *ssh.Client }

func (e sshCommandExecutor) Execute(command string) (string, error) {
	session, err := e.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	output, err := session.CombinedOutput(command)
	return string(output), err
}

func (c *SSHConnection) Output() io.Reader { return c.output }

func (c *SSHConnection) Write(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.recording.Len() < maxTerminalRecordingBytes {
		_, _ = c.recording.Write(data)
	}
	c.commandCount += int64(strings.Count(string(data), "\n"))
	_, err := c.input.Write(data)
	return err
}

func (c *SSHConnection) RecordOutput(data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.recording.Len() < maxTerminalRecordingBytes {
		_, _ = c.recording.Write(data)
	}
}

func (c *SSHConnection) Resize(cols, rows int) error {
	if cols < 20 || rows < 5 {
		return nil
	}
	return c.session.WindowChange(rows, cols)
}

// CloseSSH ends the connection and persists an audit record.
func (b *AssetBiz) CloseSSH(c *SSHConnection) {
	if c == nil {
		return
	}
	_ = c.input.Close()
	_ = c.session.Close()
	_ = c.client.Close()
	c.mu.Lock()
	recording, commands := c.recording.String(), c.commandCount
	c.mu.Unlock()
	now := time.Now()
	_ = b.db.Create(&model.TerminalSession{UserID: c.userID, HostID: c.hostID, SessionType: "ssh", StartedAt: c.startedAt, EndedAt: now, DurationSeconds: int64(now.Sub(c.startedAt).Seconds()), CommandCount: commands, Recording: recording}).Error
}
