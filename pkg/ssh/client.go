package ssh

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

func PrepareNodes(cc *config.ClusterConfig) error {
	nodes := cc.Nodes

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

		enabled, err = c.checkPacketForwarding()
		if !enabled {
			fmt.Println("Need to enable packet forwarding")
			err = c.enablePacketForwarding()
			if err != nil {
				return err
			}
		}

		err = c.installKubeadm(cc.Kubernetes.Version)
		if err != nil {
			return err
		}

		err = c.kubeadmInit(cc.Networking.PodCIDR)
		if err != nil {
			return err
		}
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
	defer session.Close()

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

func (c *Client) checkPacketForwarding() (bool, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return false, err
	}
	defer session.Close()

	out, err := session.Output("sysctl net.ipv4.ip_forward")
	if err != nil {
		return false, err
	}

	// Last byte is '\n', second to last has to be '1'
	if string(out[len(out)-2]) == "1" {
		return true, nil
	}

	return false, nil
}

func (c *Client) enablePacketForwarding() error {
	password, err := promptSudoPassword()
	if err != nil {
		return fmt.Errorf("failed to get password: %w", err)
	}

	cmd := `sudo -S tee /etc/sysctl.d/k8s.conf > /dev/null`
	_, err = c.remoteWithSudo(password, cmd, "net.ipv4.ip_forward=1\n")
	if err != nil {
		return fmt.Errorf("failed to set sysctl packet forwarding params: %w", err)
	}

	cmd = "sudo -S sysctl --system"
	_, err = c.remoteWithSudo(password, cmd, "")
	if err != nil {
		return fmt.Errorf("failed to apply sysctl params: %w", err)
	}

	return nil
}

func (c *Client) installKubeadm(version string) error {
	out, err := c.execute("uname -m")
	if err != nil {
		return err
	}

	arch := strings.TrimSpace(string(out))
	if arch == "aarch64" {
		arch = "arm64"
	}

	binaries := []string{"kubeadm", "kubelet"}
	for _, binary := range binaries {
		// Temporary solution, must fix after binary move
		exists, err := c.checkFile("/usr/local/bin/" + binary)
		if err != nil {
			return err
		}
		// TODO: Check file execution bit
		// TODO: Install systemd services for binaries
		// TODO: Move binaries to appropriate location
		if !exists {
			cmd := fmt.Sprintf("wget --quiet --show-progress --https-only --retry-connrefused --waitretry=2 --tries=5 https://dl.k8s.io/release/%s/bin/linux/%s/%s -O %s", version, arch, binary, binary)
			_, err = c.execute(cmd)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Client) checkFile(filename string) (bool, error) {
	_, err := c.execute(fmt.Sprintf("stat %s", filename))
	if err != nil {
		if _, ok := err.(*ssh.ExitError); ok {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (c *Client) kubeadmInit(podCidr string) error {
	password, err := promptSudoPassword()
	if err != nil {
		return fmt.Errorf("failed to get password: %w", err)
	}

	cmd := fmt.Sprintf("sudo -S kubeadm init --pod-network-cidr %s", podCidr)
	out, err := c.remoteWithSudo(password, cmd)
	if err != nil {
		return err
	}

	fmt.Println(out)

	return nil
}
