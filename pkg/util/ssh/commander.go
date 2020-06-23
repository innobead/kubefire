package ssh

import (
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"os"
	"strings"
)

type Commander interface {
	Init() error
	Run(before Callback, after Callback, cmds ...string) error
}

type Callback func(session *ssh.Session) bool

type Client struct {
	keyPath string
	user    string
	address string

	sshClientConfig *ssh.ClientConfig
	sshClient       *ssh.Client
}

func NewClient(keyPath string, user string, address string, sshClientConfigCallback func(config *ssh.ClientConfig)) (*Client, error) {
	clientConfig, err := CreateClientConfig(keyPath, user, sshClientConfigCallback)
	if err != nil {
		return nil, err
	}

	if !strings.Contains(address, ":") {
		address += ":22"
	}

	client := &Client{
		keyPath:         keyPath,
		user:            user,
		sshClientConfig: clientConfig,
		address:         address,
	}

	if err := client.Init(); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) Init() error {
	client, err := ssh.Dial("tcp", c.address, c.sshClientConfig)
	if err != nil {
		return errors.WithStack(err)
	}

	c.sshClient = client

	return nil
}

func (c *Client) Run(before Callback, after Callback, cmds ...string) error {
	for _, cmd := range cmds {
		session, err := c.createSSHSession()
		if err != nil {
			return err
		}

		err = func() error {
			defer session.Close()

			if before != nil && !before(session) {
				return nil
			}

			if err := session.Run(cmd); err != nil {
				return err
			}

			if after != nil {
				_ = after(session)
			}

			return nil
		}()
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (c *Client) Close() error {
	if c.sshClient != nil {
		return c.sshClient.Close()
	}

	return nil
}

func (c *Client) createSSHSession() (*ssh.Session, error) {
	session, err := c.sshClient.NewSession()
	if err != nil {
		return nil, err
	}
	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	return session, nil
}
