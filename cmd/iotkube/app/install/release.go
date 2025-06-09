package install

import (
	"fmt"
	"os"

	"github.com/ntentasd/iotkube/pkg/helm"
	"github.com/spf13/cobra"
)

var (
	releaseName string
	version     string
	repository  string
	ReleaseCmd  = &cobra.Command{
		Use:   "release",
		Short: "Install a Helm release",
		Run: func(cmd *cobra.Command, args []string) {
			namespace, err := cmd.Flags().GetString("namespace")
			if err != nil {
				cobra.CheckErr(err)
			}

			if releaseName == "" {
				fmt.Fprintf(os.Stderr, "Error: --release flag is required\n\n")
				cmd.Help()
			}

			release := helm.NewRelease(releaseName, namespace, version, repository)

			err = release.Install()
			if err != nil {
				cobra.CheckErr(err)
			}

			fmt.Printf("Successfully installed %s\n", releaseName)
		},
	}
)

func init() {
	ReleaseCmd.PersistentFlags().StringVar(&releaseName, "release", "", "The Helm release name to install")
	ReleaseCmd.PersistentFlags().StringVar(&version, "version", "", "The Helm release version to install")
	ReleaseCmd.PersistentFlags().StringVar(&repository, "repository", "", "The Helm repository to install from")

	ReleaseCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	ReleaseCmd.PersistentFlags().BoolP("help", "h", false, "Help for install release")
}
