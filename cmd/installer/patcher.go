package installer

import (
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/united-manufacturing-hub/autok3d/cmd/tools"
	"os/exec"
	"regexp"
	"strings"
	"sync"
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
	splits := strings.Split(string(output), "\n")
	wg := sync.WaitGroup{}
	for i, s := range splits {
		if i == 0 {
			continue
		}

		deploymentNameRaw := umhRegex.FindStringSubmatch(s)
		if len(deploymentNameRaw) == 0 {
			continue
		}
		deploymentName := deploymentNameRaw[0]

		wg.Add(2)
		go patchContainer(deploymentName, umhDockerRegex, version, branchName, false, &wg)
		go patchContainer(deploymentName, umhDockerRegex, version, branchName, true, &wg)

	}

	wg.Wait()

}

func patchContainer(
	deploymentName string,
	umhDockerRegex *regexp.Regexp,
	version *semver.Version,
	branchName *string,
	isInitContainer bool,
	wg *sync.WaitGroup) {
	defer wg.Done()
	var output []byte
	var err error

	var c string
	if isInitContainer {
		c = "initContainers"
	} else {
		c = "containers"
	}

	output, err = exec.Command(
		"kubectl",
		"get",
		"deployment",
		"-n",
		"united-manufacturing-hub",
		deploymentName,
		"-o",
		fmt.Sprintf("jsonpath='{.spec.template.spec.%s[0].image}'", c)).CombinedOutput()
	if err != nil {
		tools.PrintErrorAndExit(err, "Error executing kubectl get deployment", string(output), 1)
	}

	umhDockerVersion := umhDockerRegex.FindStringSubmatch(string(output))
	if len(umhDockerVersion) == 0 {
		return
	}

	if umhDockerVersion[2] != version.String() {
		return
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
			"[{\"op\": \"replace\", \"path\": \"/spec/template/spec/%s/0/image\", \"value\": \"%s:%s\"}]",
			c,
			umhDockerVersion[1],
			*bN)).CombinedOutput()
	if err != nil {
		tools.PrintWarning("Error executing kubectl set image: %s | %s", 1, string(output), err)
	}
}
