package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	parser := argparse.NewParser("umh", "United Manufacturing Hub")
	k3dUseLocalNetwork := parser.Flag(
		"",
		"k3d-local-network",
		&argparse.Options{Help: "Enables --api-port 127.0.0.1:6443 for k3d cluster", Required: false})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Println("❌ Error parsing arguments: ", err)
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}

	preliminaryChecks()
	createK3dCluster(k3dUseLocalNetwork)
	createNamespace()
	addUMHRepo()
	updateHelmRepo()
	installHelmRelease()
}

func installHelmRelease() {
	fmt.Println("ℹ️ Installing Helm release...")
	output, err := exec.Command(
		"helm",
		"install",
		"united-manufacturing-hub",
		"united-manufacturing-hub/united-manufacturing-hub",
		"--namespace", "united-manufacturing-hub",
	).CombinedOutput()
	if err != nil {
		fmt.Println("\t❌ Error installing Helm release: ", err)
		fmt.Println("\t⚠️ Output: ", string(output))
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
	fmt.Println("\t✔ Helm release installed")
}

func updateHelmRepo() {
	fmt.Println("ℹ️ Updating Helm repository...")
	output, err := exec.Command("helm", "repo", "update").CombinedOutput()
	if err != nil {
		fmt.Println("\t❌ Error updating Helm repository: ", err)
		fmt.Println("\t⚠️ Output: ", string(output))
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
	fmt.Println("\t✔ Helm repository updated")
}

func addUMHRepo() {
	fmt.Println("ℹ️ Adding UMH Helm repository...")
	// Remove old repo if it exists
	fmt.Println("\tℹ️ Removing old UMH Helm repository...")
	output, err := exec.Command("helm", "repo", "remove", "united-manufacturing-hub").CombinedOutput()
	if err != nil {
		if !strings.Contains(string(output), "no repo named \"united-manufacturing-hub\" found") {
			fmt.Println("\t❌ Error removing old repo: ", err)
			fmt.Println("\t⚠️ Output: ", string(output))
			time.Sleep(5 * time.Second)
			os.Exit(1)
		}
	}
	// Add new repo
	fmt.Println("\tℹ️ Adding new UMH Helm repository...")
	output, err = exec.Command(
		"helm",
		"repo",
		"add",
		"united-manufacturing-hub",
		"https://repo.umh.app/").CombinedOutput()
	if err != nil {
		fmt.Println("\t❌ Error adding repo: ", err)
		fmt.Println("\t⚠️ Output: ", string(output))
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
	fmt.Println("\t✔ UMH Helm repository added")
}

func createNamespace() {
	fmt.Println("ℹ️ Creating namespace...")
	// Remove old namespace if it exists
	fmt.Println("\tℹ️ Removing old namespace...")
	output, err := exec.Command("kubectl", "delete", "namespace", "united-manufacturing-hub").CombinedOutput()
	if err != nil {
		if !strings.Contains(
			string(output),
			"Error from server (NotFound): namespaces \"united-manufacturing-hub\" not found") {
			fmt.Println("\t❌ Error deleting old namespace: ", err)
			fmt.Println("\t⚠️ Output: ", string(output))
			time.Sleep(5 * time.Second)
			os.Exit(1)
		}
	}

	// Create new namespace
	fmt.Println("\tℹ️ Creating new namespace...")
	output, err = exec.Command("kubectl", "create", "namespace", "united-manufacturing-hub").CombinedOutput()
	if err != nil {
		fmt.Println("\t❌ Error creating namespace: ", err)
		fmt.Println("\t⚠️ Output: ", string(output))
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
	fmt.Println("\t✔ Namespace created")
}

func createK3dCluster(useLocalNamespace *bool) {
	fmt.Println("ℹ️ (Re-)creating k3d cluster...")

	// Remove old cluster if it exists
	fmt.Println("\tℹ️ Removing old k3d cluster...")
	output, err := exec.Command("k3d", "cluster", "delete", "united-manufacturing-hub").CombinedOutput()
	if err != nil {
		fmt.Println("\t❌ Error deleting old cluster: ", err)
		fmt.Println("\t⚠️ Output: ", string(output))
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}

	// Create new cluster
	fmt.Println("\tℹ️ Creating new k3d cluster...")
	if *useLocalNamespace {
		fmt.Println("\tℹ️ Using local network for k3d cluster")
		output, err = exec.Command(
			"k3d",
			"cluster",
			"create",
			"united-manufacturing-hub",
			"--api-port",
			"127.0.0.1:6443").CombinedOutput()
	} else {
		fmt.Println("\tℹ️ Using default network for k3d cluster")
		output, err = exec.Command("k3d", "cluster", "create", "united-manufacturing-hub").CombinedOutput()
	}

	if err != nil {
		fmt.Println("\t❌ Error creating cluster: ", err)
		fmt.Println("\t⚠️ Output: ", string(output))
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
	fmt.Println("\t✔ k3d cluster created")
}

func preliminaryChecks() {
	fmt.Println("ℹ️ Checking for required tools...")
	// Check if the user is running as root
	// Check if the user is running on a supported OS
	// Check if the user is running on a supported architecture
	if !IsAdmin() {
		runMeElevated()
		fmt.Print("Any text behind this is fine ! ")
		os.Exit(0)
	}

	// Check if the user has the required dependencies installed

	var success bool
	var chocolateyInstalled bool
	success = true
	chocolateyInstalled = true

	// choco
	if !CommandExists("choco") {
		fmt.Println("\t❌ chocolatey")
		success = false
		chocolateyInstalled = false
	}
	fmt.Println("\t✔ chocolatey")

	if !chocolateyInstalled {
		fmt.Println("❌ Chocolatey is not installed. Please install it and run this script again.")
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}

	// k3d
	if !CommandExists("k3d") {
		fmt.Println("\t❌ k3d")
		installK3d()
		success = false
	}
	fmt.Println("\t✔ k3d")

	// kubectl
	if !CommandExists("kubectl") {
		fmt.Println("\t❌ kubectl")
		installKubectl()
		success = false
	}
	fmt.Println("\t✔ kubectl")

	// helm
	if !CommandExists("helm") {
		fmt.Println("\t❌ helm")
		installHelm()
		success = false
	}
	fmt.Println("\t✔ helm")

	if !success {
		fmt.Println("\t❌ Please run this script again.")
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}

}

func installHelm() {
	err := executeChocoInstaller("kubernetes-helm")
	if err != nil {
		fmt.Println("\t❌ Error installing helm: ", err)
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
}

func installKubectl() {
	err := executeChocoInstaller("kubernetes-cli")
	if err != nil {
		fmt.Println("\t❌ Error installing kubectl: ", err)
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
}

func installK3d() {
	err := executeChocoInstaller("k3d")
	if err != nil {
		fmt.Println("\t❌ Error installing k3d: ", err)
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
}

func executeChocoInstaller(packageName string) error {
	// Install packageName using chocolatey
	err := exec.Command("choco", "install", packageName, "-y").Run()
	if err != nil {
		return err
	}
	return nil
}
