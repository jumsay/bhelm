package cmd

import (
	"fmt"
	"os"

	"bhelm/pkg/artifacthub"

	"github.com/spf13/cobra"
)

// officialCmd represents the official command group
var officialCmd = &cobra.Command{
	Use:   "official [organisation_name]",
	Short: "Manage the list of official repositories",
	Long:  "The 'official' command allows you to list, update, and query official repositories from Artifact Hub.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// Si aucun argument n'est passé, afficher l'aide
			_ = cmd.Help()
			return
		}

		orgName := args[0]

		// Mettre à jour les repositories officiels
		fmt.Println("Updating official repositories...")
		err := artifacthub.UpdateOfficialRepositories()
		if err != nil {
			fmt.Printf("Error updating repositories: %v\n", err)
			os.Exit(1)
		}

		// Récupérer les repositories correspondant à l'organisation
		repos, err := artifacthub.GetRepositoriesByOrganization(orgName)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Afficher les repositories sous forme de tableau
		artifacthub.DisplayRepositories(repos)
	},
}

// listCmd represents the 'official list' command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List the official repositories",
	Run: func(cmd *cobra.Command, args []string) {
		if err := artifacthub.ListOfficialRepositories(); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

// updateCmd represents the 'official update' command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the list of official repositories",
	Run: func(cmd *cobra.Command, args []string) {
		if err := artifacthub.UpdateOfficialRepositories(); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

func init() {
	officialCmd.AddCommand(listCmd)
	officialCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(officialCmd)
}
