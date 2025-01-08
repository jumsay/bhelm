package helm

import (
	"fmt"
	"os/exec"
)

var execCommand = exec.Command

func Install(namespace, software, repoURL, version, values string, verbose bool) error {
	if verbose {
		fmt.Println("Adding Helm repository...")
	}
	if err := execCommand("helm", "repo", "add", software, repoURL).Run(); err != nil {
		return fmt.Errorf("failed to add Helm repository: %v", err)
	}

	if verbose {
		fmt.Println("Updating Helm repositories...")
	}
	if err := execCommand("helm", "repo", "update").Run(); err != nil {
		return fmt.Errorf("failed to update Helm repositories: %v", err)
	}

	//@TODO  Add the folder name to the package name compatible for helm software/software
	args := []string{"install", software, fmt.Sprintf("%s/%s", software, software), "--namespace", namespace, "--create-namespace"}
	if version != "" {
		args = append(args, "--version", version)
	}
	if values != "" {
		args = append(args, "--values", values)
	}

	if verbose {
		fmt.Printf("Executing: helm %v\n", args)
	}

	if err := execCommand("helm", args...).Run(); err != nil {
		return fmt.Errorf("failed to install Helm package: %v", err)
	}

	return nil
}
