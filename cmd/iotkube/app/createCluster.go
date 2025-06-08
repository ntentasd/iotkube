package app

import (
	"fmt"
	"os"

	"github.com/ntentasd/iotkube/pkg/cluster"
	"github.com/ntentasd/iotkube/pkg/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile          string
	dryRun           bool
	createClusterCmd = &cobra.Command{
		Use:   "create-cluster",
		Short: "Provision a K8s cluster",
		Run: func(cmd *cobra.Command, args []string) {
			fileName, err := cmd.Flags().GetString("config")
			if err != nil {
				cobra.CheckErr(err)
			}
			if fileName == "" {
				fmt.Fprintln(os.Stderr, "Error: --config (-f) flag is required")
				os.Exit(1)
			}

			file, err := os.Open(fileName)
			if err != nil {
				cobra.CheckErr(err)
			}

			cc, err := config.Parse(file)
			if err != nil {
				cobra.CheckErr(err)
			}

			if dryRun {
				config.PrintYAML(*cc)
				return
			}

			err = cluster.BootstrapCluster(cc)
			if err != nil {
				cobra.CheckErr(err)
			}

			config.PrintYAML(*cc)
		},
	}
)

func init() {
	createClusterCmd.PersistentFlags().StringVarP(&cfgFile, "config", "f", "", "IoTKube config file")
	createClusterCmd.MarkFlagRequired("config")

	createClusterCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Only print the object that would be sent")
}
