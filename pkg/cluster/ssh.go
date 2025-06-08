package cluster

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ntentasd/iotkube/pkg/config"
	"golang.org/x/crypto/ssh"
)

type Client struct {
	client *ssh.Client
}

func checkNodes(nodes []config.NodeConfig) error {
	for _, node := range nodes {
		var path string

		if strings.HasPrefix(node.SSHKeyPath, "~/") {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			path = filepath.Join(home, node.SSHKeyPath[2:])
		} else {
			path = node.SSHKeyPath
		}

		key, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return err
		}

		config := &ssh.ClientConfig{
			User: node.User,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		if node.Port == 0 {
			node.Port = 22
		}
		client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", node.Address, node.Port), config)
		if err != nil {
			return err
		}

		c := NewSess(client)

		hostname, err := c.execute("hostname")
		if err != nil {
			return err
		}

		enabled, err := c.checkSwap()
		if err != nil {
			return err
		}
		if enabled {
			fmt.Fprintf(os.Stderr, "Please disable swap on the machine at %s (%s)\n", node.Address,
				strings.TrimSpace(string(hostname)))
		}

		defer client.Close()
	}

	return nil
}

func NewSess(client *ssh.Client) *Client {
	return &Client{
		client,
	}
}

func (c *Client) execute(cmd string) ([]byte, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return nil, err
	}

	return session.Output(cmd)
}

func (c *Client) checkSwap() (bool, error) {
	out, err := c.execute("swapon --show")
	if err != nil {
		return false, err
	}

	if len(strings.TrimSpace(string(out))) > 0 {
		return true, nil
	}

	return false, nil
}
