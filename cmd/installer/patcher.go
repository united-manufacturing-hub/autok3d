package installer

import (
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/united-manufacturing-hub/autok3d/cmd/tools"
	"os/exec"
	"regexp"
	"strings"
)

func PatchRelease(branchName *string, version *semver.Version) {
	tools.PrintInfo(fmt.Sprintf("Patching release for branch %s and version %s", *branchName, version), 1)

	// Execute kubectl get deployment -n united-manufacturing-hub

	output, err := exec.Command("kubectl", "get", "deployment", "-n", "united-manufacturing-hub").CombinedOutput()
	if err != nil {
		tools.PrintErrorAndExit(err, "Error executing kubectl get deployment", string(output), 1)
	}

	umhRegex := regexp.MustCompile(`united-manufacturing-hub[-\w]*`)
	umhDockerRegex := regexp.MustCompile(`(unitedmanufacturinghub[-\w]*/[-\w]+):([.\-\w]*)`)
	// For each line in output, skipping the first
	// 	- Split line by space
	for i, s := range strings.Split(string(output), "\n") {
		if i == 0 {
			continue
		}

		deploymentNameRaw := umhRegex.FindStringSubmatch(s)
		if len(deploymentNameRaw) == 0 {
			continue
		}
		deploymentName := deploymentNameRaw[0]

		output, err = exec.Command(
			"kubectl",
			"get",
			"deployment",
			"-n",
			"united-manufacturing-hub",
			deploymentName,
			"-o",
			"jsonpath='{.spec.template.spec.containers[0].image}'").CombinedOutput()
		if err != nil {
			tools.PrintErrorAndExit(err, "Error executing kubectl get deployment", string(output), 1)
		}

		umhDockerVersion := umhDockerRegex.FindStringSubmatch(string(output))
		if len(umhDockerVersion) == 0 {
			continue
		}

		if umhDockerVersion[2] != version.String() {
			continue
		}

		tools.PrintSuccess(fmt.Sprintf("Patching deployment %s", deploymentName), 2)

		branchNameSplit := strings.Split(*branchName, "/")
		bN := &branchNameSplit[len(branchNameSplit)-1]

		output, err = exec.Command(
			"kubectl",
			"patch",
			"deployment",
			"-n",
			"united-manufacturing-hub",
			deploymentName,
			"--type=json",
			"-p",
			fmt.Sprintf(
				"[{\"op\": \"replace\", \"path\": \"/spec/template/spec/containers/0/image\", \"value\": \"%s:%s\"}]",
				umhDockerVersion[1],
				*bN)).CombinedOutput()
		if err != nil {
			tools.PrintErrorAndExit(err, "Error executing kubectl set image", string(output), 1)
		}

		// try patch init container
		output, err = exec.Command(
			"kubectl",
			"patch",
			"deployment",
			"-n",
			"united-manufacturing-hub",
			deploymentName,
			"--type=json",
			"-p",
			fmt.Sprintf(
				"[{\"op\": \"replace\", \"path\": \"/spec/template/spec/initContainers/0/image\", \"value\": \"%s:%s\"}]",
				umhDockerVersion[1],
				*bN)).CombinedOutput()
		if err != nil {
			tools.PrintWarning(
				fmt.Sprintf("Error executing kubectl set image for init container: %s", string(output)),
				2)
		}

	}

}
