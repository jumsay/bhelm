package cmd

import (
	"bhelm/pkg/artifacthub"
	"bhelm/pkg/helm"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var orgFlag string
var userFlag string
var versionFlag string
var valuesFlag string
var verboseFlag bool

// rootCmd represents the main command
var rootCmd = &cobra.Command{
	Use:   "bhelm",
	Short: "bhelm simplifies Kubernetes application installation with Helm",
}

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install <namespace> <software>",
	Short: "Install a Kubernetes application using Helm",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		namespace := args[0]
		software := args[1]

		repoURL, err := artifacthub.GetRepositoryURL(software, orgFlag, userFlag)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		err = helm.Install(namespace, software, repoURL, versionFlag, valuesFlag, verboseFlag)
		if err != nil {
			fmt.Println("Installation failed:", err)
			os.Exit(1)
		}

		fmt.Println("Installation completed successfully!")
	},
}

func init() {
	installCmd.Flags().StringVarP(&orgFlag, "org", "o", "", "Specify the organization (optional)")
	installCmd.Flags().StringVarP(&userFlag, "user", "u", "", "Specify the user (optional)")
	installCmd.Flags().StringVarP(&versionFlag, "version", "v", "", "Specify the version of the software (optional)")
	installCmd.Flags().StringVarP(&valuesFlag, "values", "", "", "Specify a values file (optional)")
	installCmd.Flags().BoolVarP(&verboseFlag, "verbose", "", false, "Enable detailed logs (optional)")
	rootCmd.AddCommand(installCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
