package fs

import (
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

type SFTPConfig struct {
	User         string
	Host         string
	Port         int
	IdentityFile string
	Password     string
}

type SFTPFS struct {
	client    *sftp.Client
	sshClient *ssh.Client
}

func NewSFTPFS(cfg SFTPConfig) (*SFTPFS, error) {
	authMethods := buildAuthMethods(cfg)
	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no SSH authentication method available")
	}

	hostKeyCallback := ssh.InsecureIgnoreHostKey()
	if cb, err := knownHostsCallback(); err == nil {
		hostKeyCallback = cb
	}

	sshConfig := &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	sshClient, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("SSH connection failed: %w", err)
	}

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("SFTP session failed: %w", err)
	}

	return &SFTPFS{
		client:    sftpClient,
		sshClient: sshClient,
	}, nil
}

func buildAuthMethods(cfg SFTPConfig) []ssh.AuthMethod {
	var methods []ssh.AuthMethod

	if cfg.Password != "" {
		methods = append(methods, ssh.Password(cfg.Password))
	}

	if m := sshAgentAuth(); m != nil {
		methods = append(methods, m)
	}

	if cfg.IdentityFile != "" {
		if m := publicKeyAuth(cfg.IdentityFile); m != nil {
			methods = append(methods, m)
		}
	} else {
		for _, keyPath := range defaultKeyPaths() {
			if m := publicKeyAuth(keyPath); m != nil {
				methods = append(methods, m)
				break
			}
		}
	}

	return methods
}

func sshAgentAuth() ssh.AuthMethod {
	sock := os.Getenv("SSH_AUTH_SOCK")
	if sock == "" {
		return nil
	}
	conn, err := net.Dial("unix", sock)
	if err != nil {
		return nil
	}
	return ssh.PublicKeysCallback(agent.NewClient(conn).Signers)
}

func publicKeyAuth(keyPath string) ssh.AuthMethod {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(signer)
}

func defaultKeyPaths() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	return []string{
		home + "/.ssh/id_ed25519",
		home + "/.ssh/id_rsa",
	}
}

func knownHostsCallback() (ssh.HostKeyCallback, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return knownhosts.New(home + "/.ssh/known_hosts")
}

func (s *SFTPFS) Stat(p string) (FileInfo, error) {
	info, err := s.client.Stat(p)
	if err != nil {
		return FileInfo{}, err
	}
	return FileInfo{
		Name:    info.Name(),
		Size:    info.Size(),
		Mode:    info.Mode(),
		ModTime: info.ModTime(),
		IsDir:   info.IsDir(),
	}, nil
}

func (s *SFTPFS) Walk(root string, fn WalkFunc) error {
	walker := s.client.Walk(root)
	for walker.Step() {
		if err := walker.Err(); err != nil {
			if walkErr := fn(walker.Path(), FileInfo{}, err); walkErr != nil {
				return walkErr
			}
			continue
		}
		info := walker.Stat()
		if err := fn(walker.Path(), FileInfo{
			Name:    info.Name(),
			Size:    info.Size(),
			Mode:    info.Mode(),
			ModTime: info.ModTime(),
			IsDir:   info.IsDir(),
		}, nil); err != nil {
			return err
		}
	}
	return nil
}

func (s *SFTPFS) Open(p string) (io.ReadCloser, error) {
	return s.client.Open(p)
}

func (s *SFTPFS) Create(p string) (io.WriteCloser, error) {
	return s.client.Create(p)
}

func (s *SFTPFS) Remove(p string) error {
	return s.client.Remove(p)
}

func (s *SFTPFS) MkdirAll(p string, perm os.FileMode) error {
	return s.client.MkdirAll(p)
}

func (s *SFTPFS) Chmod(p string, mode os.FileMode) error {
	return s.client.Chmod(p, mode)
}

func (s *SFTPFS) Chtimes(p string, atime, mtime time.Time) error {
	return s.client.Chtimes(p, atime, mtime)
}

func (s *SFTPFS) Close() error {
	s.client.Close()
	return s.sshClient.Close()
}

// Join provides path joining for SFTP (uses forward slashes).
func (s *SFTPFS) Join(elem ...string) string {
	return path.Join(elem...)
}
