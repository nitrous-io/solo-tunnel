package sshconfig

import (
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"os/user"
	"path/filepath"

	"github.com/Sirupsen/logrus"

	"golang.org/x/crypto/ssh"
)

var (
	DefaultSSHKeyPaths []string
	ErrNoSSHKey        = errors.New("SSH Key could not be found")
)

func init() {
	currentUser, err := user.Current()
	if err != nil {
		logrus.Fatalf("Could not get current user: %v", err)
	}

	DefaultSSHKeyPaths = append(DefaultSSHKeyPaths, filepath.Join(currentUser.HomeDir, ".ssh", "id_rsa"))
	DefaultSSHKeyPaths = append(DefaultSSHKeyPaths, filepath.Join(currentUser.HomeDir, ".ssh", "id_dsa"))
}

func Config(sshUrl, sshKeyPath string) (string, *ssh.ClientConfig, error) {
	u, err := url.Parse(sshUrl)
	if err != nil {
		return "", nil, err
	}

	var key ssh.Signer
	if sshKeyPath != "" {
		_, err := os.Stat(sshKeyPath)
		if err != nil {
			if os.IsNotExist(err) {
				return "", nil, ErrNoSSHKey
			}
			return "", nil, err
		}

	} else {
		for i := 0; i < len(DefaultSSHKeyPaths); i++ {
			_, err := os.Stat(DefaultSSHKeyPaths[i])
			if err != nil {
				if !os.IsNotExist(err) {
					return "", nil, err
				}
			} else {
				sshKeyPath = DefaultSSHKeyPaths[i]
				break
			}
		}

		if sshKeyPath == "" {
			return "", nil, ErrNoSSHKey
		}
	}

	key, err = readSSHKey(sshKeyPath)
	if err != nil {
		return "", nil, err
	}

	config := &ssh.ClientConfig{
		User: u.User.Username(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
	}

	return u.Host, config, nil
}

func readSSHKey(sshKeyPath string) (ssh.Signer, error) {
	buf, err := ioutil.ReadFile(sshKeyPath)
	if err != nil {
		return nil, err
	}

	key, err := ssh.ParsePrivateKey(buf)
	if err != nil {
		return nil, err
	}
	return key, err
}
