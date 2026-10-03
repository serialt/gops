package biz

import (
	"archive/zip"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gops/internal/data/api"
	"gops/internal/data/model"
	"gops/internal/pkg/cryptox"

	"github.com/pkg/sftp"
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
	recording    string
	commandCount int64
	mu           sync.Mutex
}

// ListSFTP lists the entries in a remote directory over the host's SSH connection.
func (b *AssetBiz) ListSFTP(req *api.SFTPListRequest) (*api.PageResponse[*api.SFTPEntryResponse], error) {
	_, client, err := b.dialHostSSH(req.HostID)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return nil, fmt.Errorf("create SFTP client: %w", err)
	}
	defer sftpClient.Close()
	directory := path.Clean(req.Path)
	if directory == "." || directory == "" {
		directory = "/"
	}
	entries, err := sftpClient.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("read SFTP directory: %w", err)
	}
	result := make([]*api.SFTPEntryResponse, 0, len(entries))
	for _, entry := range entries {
		entryType := "file"
		if entry.IsDir() {
			entryType = "directory"
		}
		result = append(result, &api.SFTPEntryResponse{
			Name: entry.Name(), Type: entryType, Size: entry.Size(),
			Modified: entry.ModTime().Format("2006/01/02 15:04:05"), Permissions: entry.Mode().String(),
		})
	}
	page, total := slicePage(result, req.Page, req.Limit)
	return &api.PageResponse[*api.SFTPEntryResponse]{List: page, Total: total}, nil
}

// UploadSFTP uploads files into a remote directory over the host's SSH connection.
func (b *AssetBiz) UploadSFTP(hostID int64, directory, transferID string, files []*multipart.FileHeader, relativePaths []string) (err error) {
	total := int64(0)
	for _, header := range files {
		total += header.Size
	}
	b.setSFTPTransferProgress(transferID, 0, total, "processing", "")
	defer func() {
		if err != nil {
			b.finishSFTPTransfer(transferID, "failed", err.Error())
		} else {
			b.finishSFTPTransfer(transferID, "success", "")
		}
	}()
	_, client, err := b.dialHostSSH(hostID)
	if err != nil {
		return err
	}
	defer client.Close()
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("create SFTP client: %w", err)
	}
	defer sftpClient.Close()
	directory = cleanSFTPPath(directory)
	uploadPaths := make([]string, len(files))
	directoryRoots := make(map[string]string)
	for index, header := range files {
		relativeName := header.Filename
		if index < len(relativePaths) && relativePaths[index] != "" {
			relativeName = relativePaths[index]
		}
		relativeName = cleanSFTPUploadPath(relativeName)
		parts := strings.SplitN(relativeName, "/", 2)
		if len(parts) == 2 {
			root := parts[0]
			mappedRoot, ok := directoryRoots[root]
			if !ok {
				availableRoot, availableErr := nextAvailableSFTPPath(sftpClient, path.Join(directory, root), true)
				if availableErr != nil {
					return availableErr
				}
				mappedRoot = path.Base(availableRoot)
				directoryRoots[root] = mappedRoot
			}
			relativeName = path.Join(mappedRoot, parts[1])
		}
		uploadPaths[index] = relativeName
	}
	var transferred int64
	for index, header := range files {
		source, err := header.Open()
		if err != nil {
			return fmt.Errorf("open upload %s: %w", header.Filename, err)
		}
		relativeName := uploadPaths[index]
		target := path.Join(directory, relativeName)
		if err = sftpClient.MkdirAll(path.Dir(target)); err == nil {
			if !strings.Contains(relativeName, "/") {
				target, err = nextAvailableSFTPPath(sftpClient, target, false)
			}
		}
		if err == nil {
			var destination *sftp.File
			destination, err = sftpClient.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL)
			if err == nil {
				_, err = io.Copy(sftpProgressWriter{writer: destination, onProgress: func(written int64) {
					transferred += written
					b.setSFTPTransferProgress(transferID, transferred, total, "processing", "")
				}}, source)
				_ = destination.Close()
			}
		}
		_ = source.Close()
		if err != nil {
			return fmt.Errorf("upload %s: %w", header.Filename, err)
		}
	}
	return nil
}

// SFTPTransferProgress returns the current server-side SFTP byte progress.
func (b *AssetBiz) SFTPTransferProgress(transferID string) (*api.SFTPTransferProgressResponse, bool) {
	b.sftpTransferMu.RLock()
	defer b.sftpTransferMu.RUnlock()
	progress, ok := b.sftpTransfers[transferID]
	if !ok {
		return nil, false
	}
	copy := *progress
	return &copy, true
}

func (b *AssetBiz) setSFTPTransferProgress(transferID string, transferred, total int64, status, message string) {
	if transferID == "" {
		return
	}
	b.sftpTransferMu.Lock()
	b.sftpTransfers[transferID] = &api.SFTPTransferProgressResponse{Transferred: transferred, Total: total, Status: status, Message: message}
	b.sftpTransferMu.Unlock()
}

func (b *AssetBiz) finishSFTPTransfer(transferID, status, message string) {
	if transferID == "" {
		return
	}
	b.sftpTransferMu.Lock()
	if progress := b.sftpTransfers[transferID]; progress != nil {
		if status == "success" {
			progress.Transferred = progress.Total
		}
		progress.Status = status
		progress.Message = message
	}
	b.sftpTransferMu.Unlock()
	time.AfterFunc(5*time.Minute, func() {
		b.sftpTransferMu.Lock()
		delete(b.sftpTransfers, transferID)
		b.sftpTransferMu.Unlock()
	})
}

type sftpProgressWriter struct {
	writer     io.Writer
	onProgress func(int64)
}

func (writer sftpProgressWriter) Write(data []byte) (int, error) {
	written, err := writer.writer.Write(data)
	if written > 0 && writer.onProgress != nil {
		writer.onProgress(int64(written))
	}
	return written, err
}

func cleanSFTPUploadPath(value string) string {
	cleaned := filepath.ToSlash(filepath.Clean(value))
	for strings.HasPrefix(cleaned, "../") {
		cleaned = strings.TrimPrefix(cleaned, "../")
	}
	cleaned = strings.TrimPrefix(cleaned, "/")
	if cleaned == "" || cleaned == "." {
		return "upload"
	}
	return cleaned
}

func nextAvailableSFTPPath(client *sftp.Client, target string, isDirectory bool) (string, error) {
	if _, err := client.Lstat(target); os.IsNotExist(err) {
		return target, nil
	} else if err != nil {
		return "", fmt.Errorf("check upload target %s: %w", target, err)
	}
	for index := 1; ; index++ {
		candidateName := suffixedFilename(path.Base(target), index)
		if isDirectory {
			candidateName = fmt.Sprintf("%s_%d", path.Base(target), index)
		}
		candidate := path.Join(path.Dir(target), candidateName)
		if _, err := client.Lstat(candidate); os.IsNotExist(err) {
			return candidate, nil
		} else if err != nil {
			return "", fmt.Errorf("check upload target %s: %w", candidate, err)
		}
	}
}

func suffixedFilename(filename string, index int) string {
	extension := path.Ext(filename)
	stem := strings.TrimSuffix(filename, extension)
	if stem == "" {
		stem = filename
		extension = ""
	}
	return fmt.Sprintf("%s_%d%s", stem, index, extension)
}

// DownloadSFTP writes selected remote files as a ZIP archive.
func (b *AssetBiz) DownloadSFTP(req *api.SFTPDownloadRequest, output io.Writer) (err error) {
	_, client, err := b.dialHostSSH(req.HostID)
	if err != nil {
		return err
	}
	defer client.Close()
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("create SFTP client: %w", err)
	}
	defer sftpClient.Close()
	archive := zip.NewWriter(output)
	archiveClosed := false
	defer func() {
		if !archiveClosed {
			_ = archive.Close()
		}
	}()
	directory := cleanSFTPPath(req.Path)
	total := int64(0)
	for _, name := range req.Names {
		name = filepath.ToSlash(filepath.Clean(name))
		if name == "." || strings.HasPrefix(name, "../") || path.IsAbs(name) {
			return fmt.Errorf("invalid download path")
		}
		size, sizeErr := sftpRemoteSize(sftpClient, path.Join(directory, name))
		if sizeErr != nil {
			return sizeErr
		}
		total += size
	}
	b.setSFTPTransferProgress(req.TransferID, 0, total, "processing", "")
	defer func() {
		if err != nil {
			b.finishSFTPTransfer(req.TransferID, "failed", err.Error())
		} else {
			b.finishSFTPTransfer(req.TransferID, "success", "")
		}
	}()
	var transferred int64
	for _, name := range req.Names {
		name = filepath.ToSlash(filepath.Clean(name))
		if name == "." || strings.HasPrefix(name, "../") || path.IsAbs(name) {
			return fmt.Errorf("invalid download path")
		}
		remotePath := path.Join(directory, name)
		if err = addSFTPPathToArchive(sftpClient, archive, remotePath, name, func(written int64) {
			transferred += written
			b.setSFTPTransferProgress(req.TransferID, transferred, total, "processing", "")
		}); err != nil {
			return err
		}
	}
	if err = archive.Close(); err != nil {
		return fmt.Errorf("finish download archive: %w", err)
	}
	archiveClosed = true
	return nil
}

func sftpRemoteSize(client *sftp.Client, remotePath string) (int64, error) {
	info, err := client.Lstat(remotePath)
	if err != nil {
		return 0, fmt.Errorf("read remote path %s: %w", remotePath, err)
	}
	if !info.IsDir() {
		if !info.Mode().IsRegular() {
			return 0, fmt.Errorf("unsupported remote file type: %s", remotePath)
		}
		return info.Size(), nil
	}
	children, err := client.ReadDir(remotePath)
	if err != nil {
		return 0, fmt.Errorf("read remote directory %s: %w", remotePath, err)
	}
	var total int64
	for _, child := range children {
		size, sizeErr := sftpRemoteSize(client, path.Join(remotePath, child.Name()))
		if sizeErr != nil {
			return 0, sizeErr
		}
		total += size
	}
	return total, nil
}

func addSFTPPathToArchive(client *sftp.Client, archive *zip.Writer, remotePath, archivePath string, onProgress func(int64)) error {
	info, err := client.Lstat(remotePath)
	if err != nil {
		return fmt.Errorf("read remote path %s: %w", remotePath, err)
	}
	if info.IsDir() {
		children, readErr := client.ReadDir(remotePath)
		if readErr != nil {
			return fmt.Errorf("read remote directory %s: %w", remotePath, readErr)
		}
		if len(children) == 0 {
			if _, createErr := archive.Create(strings.TrimSuffix(archivePath, "/") + "/"); createErr != nil {
				return createErr
			}
		}
		for _, child := range children {
			if err = addSFTPPathToArchive(client, archive, path.Join(remotePath, child.Name()), path.Join(archivePath, child.Name()), onProgress); err != nil {
				return err
			}
		}
		return nil
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("unsupported remote file type: %s", remotePath)
	}
	file, err := client.Open(remotePath)
	if err != nil {
		return fmt.Errorf("open %s: %w", remotePath, err)
	}
	defer file.Close()
	entry, err := archive.Create(archivePath)
	if err != nil {
		return err
	}
	if _, err = io.Copy(sftpProgressWriter{writer: entry, onProgress: onProgress}, file); err != nil {
		return fmt.Errorf("download %s: %w", remotePath, err)
	}
	return nil
}

func cleanSFTPPath(value string) string {
	cleaned := path.Clean("/" + value)
	if cleaned == "." {
		return "/"
	}
	return cleaned
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
	startedAt := time.Now()
	return &SSHConnection{hostID: host.ID, userID: userID, startedAt: startedAt, client: client, session: session, input: input, output: output, recording: newAsciinemaRecording(startedAt, cols, rows)}, nil
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
	client, err := b.dialSSH(&host, &credential)
	if err != nil {
		return nil, nil, err
	}
	return &host, client, nil
}

func (b *AssetBiz) dialSSH(host *model.Host, credential *model.Credential) (*ssh.Client, error) {
	if credential.Username == "" {
		return nil, fmt.Errorf("凭证未配置用户名")
	}
	password, err := cryptox.Decrypt(b.aesKey, credential.Password)
	if err != nil {
		return nil, fmt.Errorf("解密凭证密码: %w", err)
	}
	privateKey, err := cryptox.Decrypt(b.aesKey, credential.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("解密凭证私钥: %w", err)
	}
	passphrase, err := cryptox.Decrypt(b.aesKey, credential.Passphrase)
	if err != nil {
		return nil, fmt.Errorf("解密凭证口令: %w", err)
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
			return nil, fmt.Errorf("解析 SSH 私钥: %w", err)
		}
		method = ssh.PublicKeys(signer)
	} else {
		method = ssh.Password(password)
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host.IP, host.Port), &ssh.ClientConfig{
		User: credential.Username, Auth: []ssh.AuthMethod{method}, HostKeyCallback: ssh.InsecureIgnoreHostKey(), Timeout: 10 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("连接 SSH: %w", err)
	}
	return client, nil
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
	appendAsciinemaEvent(&c.recording, c.startedAt, "i", data)
	c.commandCount += int64(strings.Count(string(data), "\n"))
	_, err := c.input.Write(data)
	return err
}

func (c *SSHConnection) RecordOutput(data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	appendAsciinemaEvent(&c.recording, c.startedAt, "o", data)
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
	recording, commands := c.recording, c.commandCount
	c.mu.Unlock()
	now := time.Now()
	_ = b.db.Create(&model.TerminalSession{UserID: c.userID, HostID: c.hostID, SessionType: "ssh", StartedAt: c.startedAt, EndedAt: now, DurationSeconds: int64(now.Sub(c.startedAt).Seconds()), CommandCount: commands, Recording: recording}).Error
}
