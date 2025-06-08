package config

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

type NodeConfig struct {
	Address    string `yaml:"address" validate:"required"`
	Port       int32  `yaml:"port"`
	User       string `yaml:"user" validate:"required"`
	SSHKeyPath string `yaml:"ssh_key_path" validate:"required"`
	Role       string `yaml:"role" validate:"required"`
}

type ClusterConfig struct {
	Nodes      []NodeConfig `yaml:"nodes"`
	Networking struct {
		PodCIDR string `yaml:"pod_cidr"`
	} `yaml:"networking"`
	Kubernetes struct {
		Version string `yaml:"version"`
	} `yaml:"kubernetes"`
	Extensions []string `yaml:"extensions"`
}

func Parse(r io.Reader) (*ClusterConfig, error) {
	var cfg ClusterConfig
	dec := yaml.NewDecoder(r)
	if err := dec.Decode(&cfg); err != nil {
		return &ClusterConfig{}, err
	}

	ok, err := checkVersion(cfg.Kubernetes.Version)
	if err != nil {
		return &ClusterConfig{}, err
	}

	if !ok {
		return &ClusterConfig{}, fmt.Errorf("kubernetes version %q not found", cfg.Kubernetes.Version)
	}

	return &cfg, nil
}

func PrintYAML(v any) {
	out, err := yaml.Marshal(v)
	if err != nil {
		fmt.Println("error marshaling YAML:", err)
		return
	}
	fmt.Println(string(out))
}
