package checks

import (
	"errors"
	"github.com/united-manufacturing-hub/autok3d/cmd/admin"
	"github.com/united-manufacturing-hub/autok3d/cmd/command"
	"github.com/united-manufacturing-hub/autok3d/cmd/tools"
	"os"
	"os/exec"
)

func CheckIfToolsExist() {
	tools.PrintInfo("️ Checking for required tools...", 0)
	// Check if the user is running as root
	// Check if the user is running on a supported OS
	// Check if the user is running on a supported architecture
	if !admin.IsAdmin() {
		admin.RunMeElevated()
		tools.PrintInfo("Any text behind this is fine ! ", 0)
		os.Exit(0)
	}

	// Check if the user has the required dependencies installed

	var success bool
	var chocolateyInstalled bool
	success = true
	chocolateyInstalled = true

	// choco
	if !command.Exists("choco") {
		tools.PrintNotExisting("chocolatey", 1)
		success = false
		chocolateyInstalled = false
	}
	tools.PrintSuccess("chocolatey", 1)

	if !chocolateyInstalled {
		tools.PrintErrorAndExit(
			errors.New("chocolatey is not installed"),
			" Please install it and run this script again.",
			"Not found",
			1)
	}

	// k3d
	if !command.Exists("k3d") {
		tools.PrintNotExisting("k3d", 1)
		installK3d()
		success = false
	} else {
		tools.PrintSuccess("k3d", 1)
	}

	// kubectl
	if !command.Exists("kubectl") {
		tools.PrintNotExisting("kubectl", 1)
		installKubectl()
		success = false
	} else {
		tools.PrintSuccess("kubectl", 1)
	}

	// helm
	if !command.Exists("helm") {
		tools.PrintNotExisting("helm", 1)
		installHelm()
		success = false
	} else {
		tools.PrintSuccess("helm", 1)
	}

	if !success {
		tools.PrintNotExisting("Please run this script again.", 1)
	}

}

func installHelm() {
	output, err := executeChocoInstaller("kubernetes-helm")
	if err != nil {
		tools.PrintErrorAndExit(err, "Error installing helm", output, 1)
	}
}

func installKubectl() {
	output, err := executeChocoInstaller("kubernetes-cli")
	if err != nil {
		tools.PrintErrorAndExit(err, "Error installing kubectl", output, 1)
	}
}

func installK3d() {
	output, err := executeChocoInstaller("k3d")
	if err != nil {
		tools.PrintErrorAndExit(err, "Error installing k3d", output, 1)
	}
}

func executeChocoInstaller(packageName string) (string, error) {
	// Install packageName using chocolatey
	output, err := exec.Command("choco", "install", packageName, "-y").CombinedOutput()
	return string(output), err
}
