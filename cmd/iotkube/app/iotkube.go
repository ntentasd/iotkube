package app

import (
	"fmt"
	"os"

	"github.com/ntentasd/iotkube/cmd/iotkube/app/install"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "iotkube",
	Short: "IoTKube: Opinionated Kubernetes for edge, IoT, and resilient data platforms.",
}

func init() {
	rootCmd.AddCommand(CreateClusterCmd)
	rootCmd.AddCommand(install.InstallCmd)
}

func Run() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
