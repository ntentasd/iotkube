package cluster

import (
	"github.com/ntentasd/iotkube/pkg/config"
	"github.com/ntentasd/iotkube/pkg/ssh"
)

func BootstrapCluster(cc *config.ClusterConfig) error {
	err := ssh.PrepareNodes(cc)
	if err != nil {
		return err
	}

	return nil
}
