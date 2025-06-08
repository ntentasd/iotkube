package cluster

import "github.com/ntentasd/iotkube/pkg/config"

func BootstrapCluster(cc *config.ClusterConfig) error {
	err := checkNodes(cc.Nodes)
	if err != nil {
		return err
	}

	return nil
}
