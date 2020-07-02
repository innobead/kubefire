package ssh

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Commander interface {
	Init() error
	Run(before Callback, after Callback, cmds ...string) error
	Download(remotePath string, destPath string) error
}

type Callback func(session *ssh.Session) bool

type Client struct {
	name    string
	keyPath string
	user    string
	address string

	sshClientConfig *ssh.ClientConfig
	sshClient       *ssh.Client
	log             *logrus.Entry
}

func NewClient(name string, keyPath string, user string, address string, sshClientConfigCallback func(config *ssh.ClientConfig)) (*Client, error) {
	clientConfig, err := CreateClientConfig(keyPath, user, sshClientConfigCallback)
	if err != nil {
		return nil, err
	}

	if !strings.Contains(address, ":") {
		address += ":22"
	}

	client := &Client{
		name:            name,
		keyPath:         keyPath,
		user:            user,
		sshClientConfig: clientConfig,
		address:         address,
	}
	client.log = logrus.WithField("node", client.name)

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
		c.log.Infof("running %s", cmd)

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

func (c *Client) Download(remotePath string, destPath string) error {
	session, err := c.createSSHSession()
	if err != nil {
		return err
	}
	defer session.Close()

	buf := &bytes.Buffer{}
	session.Stdout = buf

	if err := session.Run(fmt.Sprintf("cat %s", remotePath)); err != nil {
		return errors.WithStack(err)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil && err != os.ErrExist {
		return errors.WithStack(err)
	}

	if err := ioutil.WriteFile(destPath, buf.Bytes(), 0755); err != nil {
		return errors.WithStack(err)
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
