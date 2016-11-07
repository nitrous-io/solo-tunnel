package sshconfig

import (
	"io/ioutil"
	"net/url"

	"golang.org/x/crypto/ssh"
)

func Config(sshUrl, sshKeyPath string) (string, *ssh.ClientConfig, error) {
	u, err := url.Parse(sshUrl)
	if err != nil {
		return "", nil, err
	}

	key, err := ReadSSHKey(sshKeyPath)
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

func ReadSSHKey(sshKeyPath string) (ssh.Signer, error) {
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
