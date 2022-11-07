package main

import (
	"github.com/Masterminds/semver"
	"github.com/akamensky/argparse"
	"github.com/united-manufacturing-hub/autok3d/cmd/checks"
	"github.com/united-manufacturing-hub/autok3d/cmd/installer"
	"github.com/united-manufacturing-hub/autok3d/cmd/tools"
	"os"
	"time"
)

func main() {
	parser := argparse.NewParser("umh", "United Manufacturing Hub")
	chartVersion := parser.String(
		"v",
		"version",
		&argparse.Options{Help: "Version of the chart to install", Required: false})

	forceOverwrite := parser.Flag(
		"f",
		"force",
		&argparse.Options{Help: "Force overwrite of existing cluster", Required: false})

	k3dUseLocalNetwork := parser.Flag(
		"",
		"k3d-local-network",
		&argparse.Options{Help: "Enables --api-port 127.0.0.1:6443 for k3d cluster", Required: false})

	err := parser.Parse(os.Args)
	if err != nil {
		tools.PrintErrorAndExit(err, "Error parsing arguments", "", 0)
	}

	var chartSemver *semver.Version
	if *chartVersion != "" {
		chartSemver, err = semver.NewVersion(*chartVersion)
		if err != nil {
			tools.PrintErrorAndExit(err, "Error parsing chart version", "", 0)
		}
	}

	checks.CheckIfToolsExist()
	installer.CheckIfAlreadyInstalled(forceOverwrite)
	installer.CreateK3dCluster(k3dUseLocalNetwork)
	installer.CreateNamespace()
	installer.AddUMHRepo()
	installer.UpdateHelmRepo()
	installer.InstallHelmRelease(chartSemver)

	tools.PrintSuccess("Installation completed successfully", 0)
	time.Sleep(5 * time.Second)
}
