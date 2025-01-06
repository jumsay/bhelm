package artifacthub

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
)

// const artifactHubAPI = "https://artifacthub.io/api/v1"
//const officialReposFile = "official_repos.json"

type RepositoryInfo struct {
	Name string `json:"name"`
	Org  string `json:"organization_name"`
	URL  string `json:"url"`
}

type APIResponse struct {
	Packages []Package `json:"packages"`
}

func startLoadingAnimation(done chan bool, pause chan bool) {
	chars := []string{"|", "/", "-", "\\"}
	animating := true

	for animating {
		select {
		case <-done:
			fmt.Fprint(os.Stderr, "\r") // delete the loading animation
			animating = false
		case <-pause:
			<-pause // wait for animation
		default:
			for _, char := range chars {
				fmt.Fprintf(os.Stderr, "\rUpdating repositories... %s", char)
				time.Sleep(200 * time.Millisecond)
			}
		}
	}
}

// FetchAndWriteRepositories fetches official repositories from Artifact Hub API and writes them to a local JSON file
func FetchAndWriteRepositories() error {
	repositoriesMap := make(map[string]RepositoryInfo)
	const pageSize = 50
	var page int

	for {
		url := fmt.Sprintf("%s/packages/search?kind=0&verified_publisher=true&official=true&limit=%d&offset=%d", artifactHubAPI, pageSize, page*pageSize)
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("failed to fetch packages: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == 429 {
			fmt.Println("Rate limit reached, waiting...")
			time.Sleep(rateLimitWaitTime)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to fetch packages: status code %d", resp.StatusCode)
		}

		var apiResponse APIResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
			return fmt.Errorf("failed to decode response: %v", err)
		}

		if len(apiResponse.Packages) == 0 {
			break
		}

		for _, pkg := range apiResponse.Packages {
			repo := pkg.Repository
			if repo.Verified && repo.Org != "" {
				repositoriesMap[repo.Name] = RepositoryInfo{
					Name: repo.Name,
					Org:  repo.Org,
					URL:  repo.URL,
				}
			}
		}

		page++
		time.Sleep(1 * time.Second) // Respect the rate limit
	}

	return writeRepositoriesToFile(repositoriesMap)
}

func writeRepositoriesToFile(repositoriesMap map[string]RepositoryInfo) error {
	repositories := make([]RepositoryInfo, 0, len(repositoriesMap))
	for _, repo := range repositoriesMap {
		repositories = append(repositories, repo)
	}

	data, err := json.MarshalIndent(repositories, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal repositories: %v", err)
	}

	err = os.WriteFile(officialReposFile, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write to file: %v", err)
	}

	fmt.Println("Finished writing unique repositories to file.")
	return nil
}

func UpdateOfficialRepositories() error {
	done := make(chan bool)
	pause := make(chan bool)

	// Démarrer l'animation de chargement
	go startLoadingAnimation(done, pause)

	err := FetchAndWriteRepositories()
	if err != nil {
		done <- true // stop animation if error
		close(done)
		close(pause)
		return err
	}

	done <- true // stop animation
	close(done)
	close(pause)
	fmt.Println("\nOfficial repositories list updated successfully.")
	return nil
}

// GetRepositoriesByOrganization returns repositories for a given organization
func GetRepositoriesByOrganization(orgName string) ([]RepositoryInfo, error) {
	data, err := os.ReadFile(officialReposFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read official repositories file: %v", err)
	}

	var repositories []RepositoryInfo
	if err := json.Unmarshal(data, &repositories); err != nil {
		return nil, fmt.Errorf("failed to parse official repositories file: %v", err)
	}

	// filter repositories by repository organization
	var filteredRepos []RepositoryInfo
	for _, repo := range repositories {
		if repo.Org == orgName {
			filteredRepos = append(filteredRepos, repo)
		}
	}

	if len(filteredRepos) == 0 {
		return nil, fmt.Errorf("no repositories found for organization: %s", orgName)
	}

	return filteredRepos, nil
}

// DisplayRepositories shows the repositories in a table
func DisplayRepositories(repositories []RepositoryInfo) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Organization", "Repository URL"})

	for _, repo := range repositories {
		table.Append([]string{repo.Name, repo.Org, repo.URL})
	}

	table.Render()
}

// ListOfficialRepositories lists the official repositories from local JSON file
func ListOfficialRepositories() error {
	data, err := os.ReadFile(officialReposFile)
	if err != nil {
		return fmt.Errorf("failed to read official repositories file: %v", err)
	}

	var repositories []RepositoryInfo
	if err := json.Unmarshal(data, &repositories); err != nil {
		return fmt.Errorf("failed to parse official repositories file: %v", err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Organization", "Repository URL"})

	for _, repo := range repositories {
		table.Append([]string{repo.Name, repo.Org, repo.URL})
	}

	table.Render()

	return nil
}
