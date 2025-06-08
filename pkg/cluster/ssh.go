package cluster

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ntentasd/iotkube/pkg/config"
	"golang.org/x/crypto/ssh"
)

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
		defer client.Close()
	}

	return nil
}
