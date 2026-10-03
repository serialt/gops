package vncproxy

import (
	"crypto/des"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

const binaryMessage = 2

// WebSocket is the subset of a WebSocket connection used by the VNC proxy.
type WebSocket interface {
	NextReader() (messageType int, reader io.Reader, err error)
	WriteMessage(messageType int, data []byte) error
}

type webSocketReader struct {
	socket WebSocket
	reader io.Reader
}

func (r *webSocketReader) Read(buffer []byte) (int, error) {
	for r.reader == nil {
		messageType, reader, err := r.socket.NextReader()
		if err != nil {
			return 0, err
		}
		if messageType != 1 && messageType != binaryMessage {
			continue
		}
		r.reader = reader
	}
	n, err := r.reader.Read(buffer)
	if errors.Is(err, io.EOF) {
		r.reader = nil
		if n > 0 {
			return n, nil
		}
		return r.Read(buffer)
	}
	return n, err
}

// Bridge authenticates the upstream RFB server with the stored password and
// exposes an authenticated, password-free RFB stream to the browser.
func Bridge(remote net.Conn, socket WebSocket, password string) error {
	client := &webSocketReader{socket: socket}
	if err := negotiate(remote, client, socket, password); err != nil {
		return err
	}
	errorsChannel := make(chan error, 2)
	go func() {
		buffer := make([]byte, 32*1024)
		for {
			n, err := remote.Read(buffer)
			if n > 0 {
				if writeErr := socket.WriteMessage(binaryMessage, append([]byte(nil), buffer[:n]...)); writeErr != nil {
					errorsChannel <- writeErr
					return
				}
			}
			if err != nil {
				errorsChannel <- err
				return
			}
		}
	}()
	go func() {
		_, err := io.Copy(remote, client)
		errorsChannel <- err
	}()
	err := <-errorsChannel
	if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
		return nil
	}
	return err
}

func negotiate(remote io.ReadWriter, client io.Reader, socket WebSocket, password string) error {
	version := make([]byte, 12)
	if _, err := io.ReadFull(remote, version); err != nil {
		return fmt.Errorf("读取 VNC 协议版本: %w", err)
	}
	major, minor, err := parseVersion(string(version))
	if err != nil || major != 3 || minor < 3 {
		return fmt.Errorf("不支持的 VNC 协议版本 %q", strings.TrimSpace(string(version)))
	}
	if err = socket.WriteMessage(binaryMessage, version); err != nil {
		return err
	}
	clientVersion := make([]byte, 12)
	if _, err = io.ReadFull(client, clientVersion); err != nil {
		return err
	}
	if _, err = remote.Write(clientVersion); err != nil {
		return err
	}
	if minor == 3 {
		return negotiate33(remote, socket, password)
	}
	return negotiate37(remote, client, socket, password, minor)
}

func parseVersion(version string) (int, int, error) {
	if len(version) != 12 || !strings.HasPrefix(version, "RFB ") {
		return 0, 0, errors.New("invalid RFB version")
	}
	major, err := strconv.Atoi(version[4:7])
	if err != nil {
		return 0, 0, err
	}
	minor, err := strconv.Atoi(version[8:11])
	return major, minor, err
}

func negotiate33(remote io.ReadWriter, socket WebSocket, password string) error {
	security := make([]byte, 4)
	if _, err := io.ReadFull(remote, security); err != nil {
		return err
	}
	securityType := binary.BigEndian.Uint32(security)
	if securityType == 0 {
		return readFailureReason(remote)
	}
	if securityType == 1 {
		return socket.WriteMessage(binaryMessage, security)
	}
	if securityType != 2 {
		return fmt.Errorf("VNC 服务器不支持密码认证")
	}
	if err := authenticatePassword(remote, password); err != nil {
		return err
	}
	return socket.WriteMessage(binaryMessage, []byte{0, 0, 0, 1})
}

func negotiate37(remote io.ReadWriter, client io.Reader, socket WebSocket, password string, minor int) error {
	var count [1]byte
	if _, err := io.ReadFull(remote, count[:]); err != nil {
		return err
	}
	if count[0] == 0 {
		return readFailureReason(remote)
	}
	types := make([]byte, int(count[0]))
	if _, err := io.ReadFull(remote, types); err != nil {
		return err
	}
	selected := byte(0)
	for _, value := range types {
		if value == 2 {
			selected = 2
			break
		}
		if value == 1 {
			selected = 1
		}
	}
	if selected == 0 {
		return fmt.Errorf("VNC 服务器不支持标准密码认证")
	}
	if err := socket.WriteMessage(binaryMessage, []byte{1, 1}); err != nil {
		return err
	}
	var clientSelection [1]byte
	if _, err := io.ReadFull(client, clientSelection[:]); err != nil {
		return err
	}
	if clientSelection[0] != 1 {
		return fmt.Errorf("浏览器选择了不支持的 VNC 认证方式")
	}
	if _, err := remote.Write([]byte{selected}); err != nil {
		return err
	}
	if selected == 2 {
		if err := authenticatePassword(remote, password); err != nil {
			return err
		}
	} else if minor >= 8 {
		result := make([]byte, 4)
		if _, err := io.ReadFull(remote, result); err != nil {
			return err
		}
		if binary.BigEndian.Uint32(result) != 0 {
			return readFailureReason(remote)
		}
	}
	if minor >= 8 {
		return socket.WriteMessage(binaryMessage, []byte{0, 0, 0, 0})
	}
	return nil
}

func authenticatePassword(remote io.ReadWriter, password string) error {
	challenge := make([]byte, 16)
	if _, err := io.ReadFull(remote, challenge); err != nil {
		return err
	}
	response, err := challengeResponse(challenge, password)
	if err != nil {
		return err
	}
	if _, err = remote.Write(response); err != nil {
		return err
	}
	result := make([]byte, 4)
	if _, err = io.ReadFull(remote, result); err != nil {
		return err
	}
	if binary.BigEndian.Uint32(result) != 0 {
		return errors.New("VNC 密码认证失败")
	}
	return nil
}

func challengeResponse(challenge []byte, password string) ([]byte, error) {
	if len(challenge) != 16 {
		return nil, errors.New("invalid VNC challenge")
	}
	var key [8]byte
	copy(key[:], []byte(password))
	for index := range key {
		key[index] = reverseBits(key[index])
	}
	block, err := des.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	response := make([]byte, 16)
	block.Encrypt(response[:8], challenge[:8])
	block.Encrypt(response[8:], challenge[8:])
	return response, nil
}

func reverseBits(value byte) byte {
	value = value>>4 | value<<4
	value = (value&0xcc)>>2 | (value&0x33)<<2
	return (value&0xaa)>>1 | (value&0x55)<<1
}

func readFailureReason(reader io.Reader) error {
	var lengthBytes [4]byte
	if _, err := io.ReadFull(reader, lengthBytes[:]); err != nil {
		return errors.New("VNC 认证失败")
	}
	length := binary.BigEndian.Uint32(lengthBytes[:])
	if length == 0 || length > 4096 {
		return errors.New("VNC 认证失败")
	}
	reason := make([]byte, length)
	if _, err := io.ReadFull(reader, reason); err != nil {
		return errors.New("VNC 认证失败")
	}
	return fmt.Errorf("VNC 认证失败: %s", string(reason))
}
