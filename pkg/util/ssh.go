package util

import (
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
)

func CreateSSHClientConfig(keyPath string, user string, sshConfigCB func(config *ssh.ClientConfig)) (*ssh.ClientConfig, error) {
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	sshConfig := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	if sshConfigCB != nil {
		sshConfigCB(sshConfig)
	}

	return sshConfig, nil
}

func CreateSSHSession(address string, sshConfig *ssh.ClientConfig) (*ssh.Session, error) {
	client, err := ssh.Dial("tcp", address+":22", sshConfig)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	return session, nil
}
