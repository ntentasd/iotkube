package install

import (
	"github.com/spf13/cobra"
)

var (
	dryRun     bool
	namespace  string
	InstallCmd = &cobra.Command{
		Use:   "install",
		Short: "Install a K8s extension",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()

			if dryRun {
				return
			}
		},
	}
)

func init() {
	InstallCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Only print the object that would be sent")
	InstallCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "The namespace to install the extension to")

	InstallCmd.AddCommand(ReleaseCmd)

	InstallCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	InstallCmd.PersistentFlags().BoolP("help", "h", false, "Help for install")
}
