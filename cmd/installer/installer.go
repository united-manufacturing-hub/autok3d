package installer

import (
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/united-manufacturing-hub/autok3d/cmd/tools"
	"os/exec"
	"strings"
)

func InstallHelmRelease(version *semver.Version) {
	tools.PrintInfo("️ Installing Helm release...", 0)
	var output []byte
	var err error

	if version == nil {
		output, err = exec.Command(
			"helm",
			"install",
			"united-manufacturing-hub",
			"united-manufacturing-hub/united-manufacturing-hub",
			"--namespace", "united-manufacturing-hub",
		).CombinedOutput()
	} else {
		tools.PrintInfo("Using version %s", 1, version.String())
		/* #nosec G204 semver validation should prevent command injection */
		output, err = exec.Command(
			"helm",
			"install",
			"united-manufacturing-hub",
			"united-manufacturing-hub/united-manufacturing-hub",
			"--namespace", "united-manufacturing-hub",
			"--version", fmt.Sprintf("v%s", version.String()),
		).CombinedOutput()
	}
	if err != nil {
		tools.PrintErrorAndExit(err, "Error installing Helm release: ", string(output), 1)
	}
	tools.PrintSuccess("Helm release installed", 1)
}

func UpdateHelmRepo() {
	tools.PrintInfo("️ Updating Helm repository...", 0)
	output, err := exec.Command("helm", "repo", "update").CombinedOutput()
	if err != nil {
		tools.PrintErrorAndExit(err, "Error updating Helm repository: ", string(output), 1)
	}
	tools.PrintSuccess("Helm repository updated", 1)
}

func AddUMHRepo(useFakeRepo bool) {
	tools.PrintInfo("️ Adding UMH Helm repository...", 0)
	// Remove old repo if it exists
	tools.PrintInfo("Removing old UMH Helm repository...", 1)
	output, err := exec.Command("helm", "repo", "remove", "united-manufacturing-hub").CombinedOutput()
	if err != nil {
		if !strings.Contains(string(output), "no repo named \"united-manufacturing-hub\" found") {
			tools.PrintErrorAndExit(err, "Error removing old UMH Helm repository: ", string(output), 1)
		}
	}
	// Add new repo
	tools.PrintInfo("Adding new UMH Helm repository...", 1)
	if useFakeRepo {
		output, err = exec.Command(
			"helm",
			"repo",
			"add",
			"united-manufacturing-hub",
			"https://test-repo.umh.app").CombinedOutput()
	} else {
		output, err = exec.Command(
			"helm",
			"repo",
			"add",
			"united-manufacturing-hub",
			"https://repo.umh.app/").CombinedOutput()
	}
	if err != nil {
		tools.PrintErrorAndExit(err, "Error adding UMH Helm repository: ", string(output), 1)
	}
	tools.PrintSuccess("UMH Helm repository added", 1)
}

func CreateNamespace() {
	tools.PrintInfo("️ Creating namespace...", 0)
	// Remove old namespace if it exists
	tools.PrintInfo("Removing old namespace...", 1)
	output, err := exec.Command("kubectl", "delete", "namespace", "united-manufacturing-hub").CombinedOutput()
	if err != nil {
		if !strings.Contains(
			string(output),
			"Error from server (NotFound): namespaces \"united-manufacturing-hub\" not found") {
			tools.PrintErrorAndExit(err, "Error removing old namespace: ", string(output), 1)
		}
	}

	// Create new namespace
	tools.PrintInfo("Creating new namespace...", 1)
	output, err = exec.Command("kubectl", "create", "namespace", "united-manufacturing-hub").CombinedOutput()
	if err != nil {
		tools.PrintErrorAndExit(err, "Error creating namespace: ", string(output), 1)
	}
	tools.PrintSuccess("Namespace created", 1)
}

func CreateK3dCluster(useLocalNamespace *bool, exposeNodePorts *bool) {
	tools.PrintInfo("️ (Re-)creating k3d cluster...", 0)

	// Remove old cluster if it exists
	tools.PrintInfo("Removing old k3d cluster...", 1)
	output, err := exec.Command("k3d", "cluster", "delete", "united-manufacturing-hub").CombinedOutput()
	if err != nil {
		tools.PrintErrorAndExit(err, "Error removing old k3d cluster: ", string(output), 1)
	}

	// Create new cluster
	tools.PrintInfo("Creating new k3d cluster...", 1)
	args := []string{
		"cluster",
		"create",
		"united-manufacturing-hub",
	}

	if *useLocalNamespace {
		tools.PrintInfo("Using local network for k3d cluster", 1)
		args = append(args, "--api-port", "127.0.0.1:6443")
	} else {
		tools.PrintInfo("Using default network for k3d cluster", 1)
	}
	output, err = exec.Command("k3d", args...).CombinedOutput()

	if *exposeNodePorts {
		tools.PrintInfo("Exposing node ports...", 1)
		args = append(args, "--agents", "3", "-p", "\"30000-32767:30000-32767\"")
	}

	if err != nil {
		tools.PrintErrorAndExit(err, "Error creating k3d cluster: ", string(output), 1)
	}
	tools.PrintSuccess("k3d cluster created", 1)
}

func CheckIfAlreadyInstalled(forceOverwrite *bool) {
	tools.PrintInfo("️ Checking if already installed...", 0)
	if *forceOverwrite {
		tools.PrintWarning("Force overwrite enabled, skipping check", 1)
		return
	}
	output, err := exec.Command("kubectl", "get", "namespace", "united-manufacturing-hub").CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "NotFound") {
			tools.PrintInfo("Not installed yet", 1)
			return
		}
		if strings.Contains(string(output), "Unable to connect to the server") {
			tools.PrintInfo("Not installed yet", 1)
			return
		}
		if strings.Contains(
			string(output),
			"Error in configuration: context was not found for specified context: k3d-united-manufacturing-hub") {
			tools.PrintInfo("Not installed yet", 1)
			return
		}
		tools.PrintErrorAndExit(err, "Error checking if already installed: ", string(output), 1)
	}
	tools.PrintErrorAndExit(nil, "Already installed. Use --force to overwrite", "", 1)
}
