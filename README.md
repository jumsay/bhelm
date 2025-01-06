# Bhelm

BHELM is a command-line tool designed to simplify the installation and management of Kubernetes applications using Helm. It integrates with Artifact Hub to search for official Helm repositories and packages, ensuring users always get the most up-to-date and trusted software for their Kubernetes clusters.

## Features

- Simple Installation: Install Kubernetes applications with a single command using official or verified Helm charts from Artifact Hub.
Example:
``bash 
bhelm install <namespace> <package>
``
This automatically searches for the most appropriate Helm repository and installs the latest version of the software in the specified namespace.

- User Selection: When multiple packages are available, BHELM provides an interactive interface for users to choose the desired package.

## Requirements
- Go 1.18+
- Kubernetes cluster (configured with kubectl)
- Helm 3.0+

## Installation
- Clone the repository `git clone https://github.com/yourusername/bhelm.git`
- Build the executable:
  - linux :
  `go build -o bhelm main.go`
  - windows :
  `go build -o bhelm.exe main.go`
- (optional) Move the executable to a directory in your PATH:
  - Linux :
  `sudo mv bhelm /usr/local/bin`
  - Windows :
  ``

## Docs
`bhelm help`
`bhelm install <namespace> <package name>`
`bhelm official <package name>`
`bhelm official update`
`bhelm official list`

### contribution

