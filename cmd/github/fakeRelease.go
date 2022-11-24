package github

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/bramvdbogaerde/go-scp"
	"github.com/schollz/progressbar/v3"
	"github.com/united-manufacturing-hub/autok3d/cmd/ssh"
	"github.com/united-manufacturing-hub/autok3d/cmd/tools"
	"golang.org/x/crypto/sha3"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

func MakeFakeRelease(gitBranchName *string) (bool, *semver.Version) {
	err, v := CreateFakeRelease(gitBranchName)
	if err != nil {
		tools.PrintErrorAndExit(err, "Error creating fake release", "", 1)
	}

	return gitBranchName != nil, v
}

func CreateFakeRelease(gitBranchName *string) (err error, version *semver.Version) {
	tools.PrintInfo("Creating fake release for branch %s", 0, *gitBranchName)
	if *gitBranchName == "" {
		return
	}

	var pullRequests GHPulls
	pullRequests, err = GetPullRequests("united-manufacturing-hub", "united-manufacturing-hub")
	if err != nil {
		return err, nil
	}

	var pullRequest *GHPull
	for _, pull := range pullRequests {
		if pull.Head.Ref == *gitBranchName {
			pullRequest = &pull
		}
	}
	if pullRequest == nil {
		return fmt.Errorf("no pull request found for branch %s", *gitBranchName), nil
	}

	version = semver.MustParse(fmt.Sprintf("%d.%d.%d", pullRequest.Number, 0, 0))

	// Create a temp folder

	var tempDir string
	tempDir, err = os.MkdirTemp("", "autok3d*")
	if err != nil {
		return err, nil
	}
	var repoPath string
	repoPath, err = DownloadBranch(*gitBranchName, tempDir)
	if err != nil {
		return err, nil
	}

	tools.PrintInfo("Creating release for branch %s", 1, *gitBranchName)

	fmt.Printf("repoPath: %s\n", repoPath)

	chartYamlPath := filepath.Join(repoPath, "deployment", "united-manufacturing-hub", "Chart.yaml")
	chartYamlContent, err := os.ReadFile(chartYamlPath)
	if err != nil {
		return err, nil
	}

	versionStr := fmt.Sprintf("%d.0.0", pullRequest.Number)

	var chartYaml Chart
	err = yaml.Unmarshal(chartYamlContent, &chartYaml)
	if err != nil {
		tools.PrintErrorAndExit(err, "Could not unmarshal chart yaml", "", 1)
	}
	chartYaml.Version = versionStr
	chartYaml.AppVersion = versionStr

	chartYamlContent, err = yaml.Marshal(chartYaml)
	if err != nil {
		tools.PrintErrorAndExit(err, "Could not marshal chart yaml", "", 1)
	}

	err = os.WriteFile(chartYamlPath, chartYamlContent, 0644)
	if err != nil {
		return err, nil
	}

	repoIp := "10.1.1.1"
	repoUrl := fmt.Sprintf("http://%s", repoIp)
	repoIpSSH := fmt.Sprintf("%s:22", repoIp)
	// Modify development.yaml
	development, err := os.ReadFile(path.Join(repoPath, "deployment", "helm-repo", "cloud-init", "development.yaml"))
	if err != nil {
		return err, nil
	}
	developmentStr := string(development)
	developmentStr = strings.ReplaceAll(developmentStr, "https://repo.umh.app", repoUrl)
	developmentStr = strings.ReplaceAll(
		developmentStr,
		"--set serialNumber=$(hostname) --kubeconfig /etc/rancher/k3s/k3s.yaml -n united-manufacturing-hub; then",
		fmt.Sprintf(
			"--set serialNumber=$(hostname) --kubeconfig /etc/rancher/k3s/k3s.yaml -n united-manufacturing-hub --version v%s; then",
			versionStr))

	err = os.WriteFile(
		path.Join(repoPath, "docs", "static", "examples", "development.yaml"),
		[]byte(developmentStr),
		0644)
	if err != nil {
		return err, nil
	}

	tools.PrintInfo("Uploading development.yaml to %s", 1, repoIpSSH)

	helmRepoSSHPass := "Shudder2-Luxurious-Stump-Suffrage"

	sshClientUpDevYaml := ssh.GetSSHClient(repoIpSSH, "root", helmRepoSSHPass, 10)
	var scpClientUpDevYaml scp.Client
	scpClientUpDevYaml, err = scp.NewClientBySSH(sshClientUpDevYaml.UnderlyingClient())
	if err != nil {
		return err, nil
	}

	h := sha3.New256()
	h.Write([]byte(developmentStr))
	hashHex := fmt.Sprintf("%x", h.Sum(nil))

	remoteName := fmt.Sprintf("%s.yaml", hashHex)
	devReader := bytes.NewReader([]byte(developmentStr))
	err = scpClientUpDevYaml.Copy(
		context.Background(),
		devReader,
		fmt.Sprintf("/www/testyamls/%s", remoteName),
		"0644",
		int64(len([]byte(developmentStr))))
	if err != nil {
		return err, nil
	}

	tools.PrintSuccess("Uploaded development.yaml to %s", 2, repoIpSSH)

	tools.PrintInfo("Packaging helm chart", 1)
	helmPackage := exec.Command("helm", "package", "../united-manufacturing-hub")
	helmPackage.Dir = path.Join(repoPath, "deployment", "helm-repo")
	var helmPackageOutput []byte
	helmPackageOutput, err = helmPackage.CombinedOutput()
	if err != nil {
		tools.PrintErrorAndExit(err, "Error packaging helm chart", string(helmPackageOutput), 1)
	}

	tools.PrintSuccess("Packaged helm chart", 2)

	tools.PrintInfo("Creating helm repo index", 1)
	helmRepoIndex := exec.Command("helm", "repo", "index", "--url", repoUrl, "--merge", "index.yaml", ".")
	helmRepoIndex.Dir = path.Join(repoPath, "deployment", "helm-repo")
	_, err = helmRepoIndex.Output()
	if err != nil {
		return err, nil
	}

	tools.PrintSuccess("Created helm repo index", 2)

	tools.PrintInfo("Uploading helm chart to %s", 1, repoIpSSH)
	sshClient1 := ssh.GetSSHClient(repoIpSSH, "root", helmRepoSSHPass, 10)
	var scpClient1 scp.Client
	scpClient1, err = scp.NewClientBySSH(sshClient1.UnderlyingClient())
	if err != nil {
		return err, nil
	}

	var serverIndex *os.File
	serverIndex, err = os.CreateTemp(os.TempDir(), "index.yaml")
	if err != nil {
		return err, nil
	}
	err = scpClient1.CopyFromRemote(context.Background(), serverIndex, "/www/index.yaml")
	if err != nil {
		return err, nil
	}
	defer func() {
		scpClient1.Close()
		_ = sshClient1.Close()
	}()

	err = serverIndex.Close()
	if err != nil {
		return err, nil
	}
	tools.PrintSuccess("Uploaded helm chart to %s", 2, repoIpSSH)

	tools.PrintInfo("Check if release is already in index", 1)
	// Cleanup server index.yaml after we're done
	defer func(name string) {
		_ = os.Remove(name)
	}(serverIndex.Name())

	serverIndexF, err := os.ReadFile(serverIndex.Name())
	if err != nil {
		return err, nil
	}
	var serverIndexYaml IndexYaml
	err = yaml.Unmarshal(serverIndexF, &serverIndexYaml)
	if err != nil {
		return err, nil
	}
	var present bool
	for _, version := range serverIndexYaml.Entries.UnitedManufacturingHub {
		if version.Version == versionStr {
			present = true
			break
		}
	}

	if !present {
		tools.PrintInfo("Adding fake release to server index.yaml", 2)
		sshClient2 := ssh.GetSSHClient(repoIpSSH, "root", helmRepoSSHPass, 10)
		var scpClient2 scp.Client
		scpClient2, err = scp.NewClientBySSH(sshClient2.UnderlyingClient())
		if err != nil {
			return err, nil
		}

		// Read local index
		var index []byte
		index, err = os.ReadFile(path.Join(repoPath, "deployment", "helm-repo", "index.yaml"))
		if err != nil {
			return err, nil
		}
		var localIndexYaml IndexYaml
		err = yaml.Unmarshal(index, &localIndexYaml)
		if err != nil {
			return err, nil
		}
		var local UnitedManufacturingHub
		var foundLocal bool
		for _, hub := range localIndexYaml.Entries.UnitedManufacturingHub {
			if hub.Version == versionStr {
				tools.PrintInfo("Found local version %s", 1, versionStr)
				local = hub
				foundLocal = true
			}
		}
		if !foundLocal {
			return fmt.Errorf("could not find local version %s", versionStr), nil
		}

		serverIndexYaml.Entries.UnitedManufacturingHub = append(serverIndexYaml.Entries.UnitedManufacturingHub, local)
		serverIndexYaml.Entries.FactorycubeServer = nil
		serverIndexYaml.Entries.FactorycubeEdge = nil

		var newServerIndexYaml []byte
		newServerIndexYaml, err = yaml.Marshal(serverIndexYaml)
		if err != nil {
			return err, nil
		}
		newServerIndexYaml = []byte(strings.ReplaceAll(string(newServerIndexYaml), "https://repo.umh.app", repoUrl))

		reader := bytes.NewReader(newServerIndexYaml)
		err = scpClient2.Copy(context.Background(), reader, "/www/index.yaml", "0644", int64(len(newServerIndexYaml)))
		if err != nil {
			return err, nil
		}

		defer func() {
			scpClient2.Close()
			_ = sshClient2.Close()
		}()
	} else {
		tools.PrintInfo("Version already present in server index.yaml", 2)
	}

	// Copy all tgz to server
	tools.PrintInfo("Copying tgz to server", 1)

	helmFiles, err := os.ReadDir(path.Join(repoPath, "deployment", "helm-repo"))
	if err != nil {
		return err, nil
	}
	await.Add(len(helmFiles))
	for _, file := range helmFiles {
		go CopySCP(file, repoIpSSH, helmRepoSSHPass, repoPath)
	}
	await.Wait()

	// Remove folder
	tools.PrintInfo("Removing downloaded repo", 1)
	err = os.RemoveAll(repoPath)
	if err != nil {
		tools.PrintWarning("Error removing downloaded repo", 2)
	}

	tools.PrintSuccess(fmt.Sprintf("Cloudconfig: http://%s/testyamls/%s.yaml", repoIp, hashHex), 2)

	return nil, version
}

var await = sync.WaitGroup{}

func CopySCP(file os.DirEntry, repoIpSSH string, helmRepoSSHPass string, dir string) {
	defer await.Done()
	if file.IsDir() {
		return
	}
	if !strings.HasSuffix(file.Name(), ".tgz") {
		return
	}
	tools.PrintInfo("Copying %s", 2, file.Name())
	sshClientTgz := ssh.GetSSHClient(repoIpSSH, "root", helmRepoSSHPass, 10)
	var scpClientTgz scp.Client
	var err error
	scpClientTgz, err = scp.NewClientBySSH(sshClientTgz.UnderlyingClient())
	if err != nil {
		tools.PrintErrorAndExit(err, "Could not create scp client", "", 2)
	}
	var reader *os.File
	reader, err = os.Open(path.Join(dir, "deployment", "helm-repo", file.Name()))
	if err != nil {
		tools.PrintErrorAndExit(err, "Could not open file", "", 2)
	}
	err = scpClientTgz.CopyFile(context.Background(), reader, fmt.Sprintf("/www/%s", file.Name()), "0644")
	if err != nil {
		tools.PrintErrorAndExit(err, "Could not copy file", "", 2)
	}
	scpClientTgz.Close()
	err = sshClientTgz.Close()
	if err != nil {
		tools.PrintErrorAndExit(err, "Could not close ssh client", "", 2)
	}

}

func DownloadBranch(gitBranchName string, dir string) (path string, err error) {
	tools.PrintInfo("Downloading branch %s", 1, gitBranchName)
	// Download the branch as a zip file
	//https://github.com/united-manufacturing-hub/united-manufacturing-hub/archive/refs/heads/feat/opcuasimulator_amine.zip

	branchUrl := fmt.Sprintf(
		"https://github.com/united-manufacturing-hub/united-manufacturing-hub/archive/refs/heads/%s.zip",
		gitBranchName)
	var req *http.Request
	req, err = http.NewRequest(
		"GET",
		branchUrl,
		nil)

	if err != nil {
		return path, err
	}

	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return path, err
	}

	// Safe file to disk
	var tempFile *os.File
	tempFile, err = os.CreateTemp("", "autok3d*.zip")
	if err != nil {
		return path, err
	}

	defer func() {
		tempFile.Close()
		// remove zip file
		err = os.Remove(tempFile.Name())
		if err != nil {
			tools.PrintErrorAndExit(err, "Could not remove zip file", "", 2)
		}
	}()

	var fileLength int64
	fileLength, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return path, err
	}

	_, err = tempFile.Seek(0, 0)
	if err != nil {
		return path, err
	}

	var reader *zip.Reader
	reader, err = zip.NewReader(tempFile, fileLength)

	if err != nil {
		return path, err
	}

	bar := progressbar.Default(int64(len(reader.File)))
	for _, file := range reader.File {
		err = unzipFile(file, dir)
		if err != nil {
			_ = bar.Close()
			return path, err
		}
		_ = bar.Add(1)
	}
	_ = bar.Close()

	fullPath := filepath.Join(
		dir,
		fmt.Sprintf("united-manufacturing-hub-%s", strings.Replace(gitBranchName, "/", "-", -1)))

	return fullPath, nil
}

func unzipFile(f *zip.File, destination string) error {
	// 4. Check if file paths are not vulnerable to Zip Slip
	filePath := filepath.Join(destination, f.Name)
	if !strings.HasPrefix(filePath, filepath.Clean(destination)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", filePath)
	}

	// 5. Create directory tree
	if f.FileInfo().IsDir() {
		if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
			return err
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}

	// 6. Create a destination file for unzipped content
	destinationFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	// 7. Unzip the content of a file and copy it to the destination file
	zippedFile, err := f.Open()
	if err != nil {
		return err
	}
	defer zippedFile.Close()

	if _, err := io.Copy(destinationFile, zippedFile); err != nil {
		return err
	}
	return nil
}
