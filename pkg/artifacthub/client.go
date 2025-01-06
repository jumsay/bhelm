package artifacthub

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
)

const (
	artifactHubAPI        = "https://artifacthub.io/api/v1"
	officialReposFile     = "official_repos.json"
	rateLimitWaitTime     = 1 * time.Second
	questionConfirmChoice = "Do you want to proceed with this repository? (Y/N): "
	operationCancelled    = "operation cancelled by user"
)

type Repository struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Org      string `json:"organization_name"`
	Verified bool   `json:"verified_publisher"`
}

type Package struct {
	Name       string `json:"name"`
	Repository struct {
		Name     string `json:"name"`
		URL      string `json:"url"`
		Org      string `json:"organization_name"`
		Verified bool   `json:"verified_publisher"`
	} `json:"repository"`
}

type ArtifactHubResponse struct {
	Total    int       `json:"total"`
	Packages []Package `json:"packages"`
}

// SearchOfficialRepository searches for an official repository using org or user
func SearchOfficialRepository(org, user string) (string, error) {
	// check if local repository file exists
	if _, err := os.Stat(officialReposFile); os.IsNotExist(err) {
		fmt.Println("Local repository file not found. Updating repositories...")
		if err := UpdateOfficialRepositories(); err != nil {
			return "", fmt.Errorf("failed to update local repositories: %v", err)
		}
	}

	// search local repositories
	if org != "" {
		fmt.Printf("Searching for official repository locally for organization: %s...\n", org)
		repos, err := GetRepositoriesByOrganization(org)
		if err == nil && len(repos) > 0 {
			if len(repos) == 1 {
				fmt.Printf("Found local official repository: %s\n", repos[0].URL)
				return repos[0].URL, nil
			}

			// if multiple results, ask user to choose
			fmt.Println("Multiple repositories found. Please select one:")
			displayRepositoriesTable(repos)
			selectedIndex := promptUserSelection(len(repos))

			selectedRepo := repos[selectedIndex]
			fmt.Printf("You selected: %s (URL: %s)\n", selectedRepo.Name, selectedRepo.URL)

			// ask for confirmation before proceeding
			if confirmAction(questionConfirmChoice) {
				return selectedRepo.URL, nil
			} else {
				return "", errors.New(operationCancelled)
			}
		}
		fmt.Println("No local repository found. Falling back to remote search...")
	}

	// use the API if local repositories are not found
	if org == "" && user == "" {
		return "", errors.New("organization or user must be specified")
	}

	url := fmt.Sprintf("%s/repositories/search?offset=0&limit=10&kind=0&official=true&user=%s&org=%s", artifactHubAPI, user, org)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch repository: status code %d", resp.StatusCode)
	}

	var repos []Repository
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if err := json.Unmarshal(body, &repos); err != nil {
		return "", err
	}

	if len(repos) == 0 {
		return "", errors.New("no official repository found")
	}

	if len(repos) == 1 {
		fmt.Printf("Found remote official repository: %s\n", repos[0].URL)
		return repos[0].URL, nil
	}

	// ask user to choose if multiple repositories are found
	fmt.Println("Multiple repositories found remotely. Please select one:")
	displayRepositoriesTableRemote(repos)
	selectedIndex := promptUserSelection(len(repos))

	selectedRepo := repos[selectedIndex]
	fmt.Printf("You selected: %s (URL: %s)\n", selectedRepo.Name, selectedRepo.URL)

	// ask user to confirm before proceeding
	if confirmAction(questionConfirmChoice) {
		return selectedRepo.URL, nil
	}

	return "", errors.New(operationCancelled)
}

// displayRepositoriesTable displays the list of local repositories in a table format
func displayRepositoriesTable(repos []RepositoryInfo) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Index", "Name", "Organization", "Repository URL"})

	for i, repo := range repos {
		table.Append([]string{strconv.Itoa(i), repo.Name, repo.Org, repo.URL})
	}

	table.Render()
}

// displayRepositoriesTableRemote displays the list of remote repositories in a table format
func displayRepositoriesTableRemote(repos []Repository) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Index", "Name", "Organization", "Repository URL"})

	for i, repo := range repos {
		table.Append([]string{strconv.Itoa(i), repo.Name, repo.Org, repo.URL})
	}

	table.Render()
}

// promptUserSelection prompts the user to select a repository by index
func promptUserSelection(max int) int {
	var selection int
	for {
		fmt.Printf("Enter the index of the repository to select (0-%d): ", max-1)
		_, err := fmt.Scan(&selection)
		if err == nil && selection >= 0 && selection < max {
			break
		}
		fmt.Println("Invalid selection. Please try again.")
	}
	return selection
}

// confirmAction prompts the user for confirmation (Y/N)
func confirmAction(prompt string) bool {
	var response string
	for {
		fmt.Print(prompt)
		_, err := fmt.Scan(&response)
		if err != nil {
			fmt.Println("Invalid input. Please enter Y or N.")
			continue
		}

		response = strings.TrimSpace(strings.ToUpper(response))
		if response == "Y" {
			return true
		} else if response == "N" {
			return false
		} else {
			fmt.Println("Invalid input. Please enter Y or N.")
		}
	}
}

// SearchPackageFallback searches for a package and allows user selection if needed
func SearchPackageFallback(packageName string) (string, error) {
	url := fmt.Sprintf("%s/packages/search?ts_query_web=%s&kind=0&verified_publisher=true&official=true&deprecated=false", artifactHubAPI, packageName)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch package: status code %d", resp.StatusCode)
	}

	var result ArtifactHubResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if len(result.Packages) == 0 {
		return "", errors.New("no package found")
	}

	// filter packages by package name
	var filteredPackages []Package
	for _, pkg := range result.Packages {
		if strings.Contains(pkg.Name, packageName) {
			filteredPackages = append(filteredPackages, pkg)
		}
	}

	if len(filteredPackages) == 1 {
		fmt.Printf("Match found: %s (Repository URL: %s)\n", filteredPackages[0].Name, filteredPackages[0].Repository.URL)
		// ask user to confirm before proceeding
		if confirmAction(questionConfirmChoice) {
			return filteredPackages[0].Repository.URL, nil
		} else {
			return "", errors.New(operationCancelled)
		}

	}

	if len(filteredPackages) > 1 {
		fmt.Println("Multiple matches found. Please select a package from the list below:")
		displayPackagesTable(filteredPackages)
		selectedIndex := promptUserSelection(len(filteredPackages))
		return filteredPackages[selectedIndex].Repository.URL, nil
	}

	// if no exact match found, display all packages
	fmt.Println("No exact or partial match found. Please select a package from the list below:")
	displayPackagesTable(result.Packages)

	selectedIndex := promptUserSelection(len(result.Packages))
	return result.Packages[selectedIndex].Repository.URL, nil
}

// // filterOfficialPackages filters packages from official repositories
// func filterOfficialPackages(packages []Package) []Package {
// 	var official []Package
// 	for _, pkg := range packages {
// 		if isOfficialRepository(pkg.Repository.URL) {
// 			official = append(official, pkg)
// 		}
// 	}
// 	return official
// }

// // isOfficialRepository checks if a repository URL is considered official
// func isOfficialRepository(repoURL string) bool {
// 	data, err := os.ReadFile(officialReposFile)
// 	if err != nil {
// 		fmt.Printf("Warning: failed to read official repositories file: %v\n", err)
// 		return false
// 	}

// 	var repositories []RepositoryInfo
// 	if err := json.Unmarshal(data, &repositories); err != nil {
// 		fmt.Printf("Warning: failed to parse official repositories file: %v\n", err)
// 		return false
// 	}

// 	for _, repo := range repositories {
// 		if repo.URL == repoURL {
// 			return true
// 		}
// 	}

// 	return false
// }

// displayPackagesTable displays the list of packages in a table format
func displayPackagesTable(packages []Package) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Index", "Name", "Repository URL"})

	for i, pkg := range packages {
		table.Append([]string{strconv.Itoa(i), pkg.Name, pkg.Repository.URL})
	}

	table.Render()
}

// GetRepositoryURL returns the repository URL, falling back to package search if needed
func GetRepositoryURL(software, org, user string) (string, error) {
	if org != "" || user != "" {
		repoURL, err := SearchOfficialRepository(org, user)
		if err == nil {
			fmt.Printf("Official repository found: %s\n", repoURL)
			return repoURL, nil
		}
		fmt.Println("No official repository found, falling back to package search...")
	}

	return SearchPackageFallback(software)
}
