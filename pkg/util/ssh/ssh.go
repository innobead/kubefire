package ssh

import (
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
)

func CreateClientConfig(keyPath string, user string, clientConfigCallback func(config *ssh.ClientConfig)) (*ssh.ClientConfig, error) {
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	clientConfig := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	if clientConfigCallback != nil {
		clientConfigCallback(clientConfig)
	}

	return clientConfig, nil
}
