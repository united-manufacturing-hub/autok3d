package github

import (
	"archive/zip"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"github.com/united-manufacturing-hub/autok3d/cmd/tools"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func MakeFakeRelease(gitBranchName *string) (url string) {
	url, err := CreateFakeRelease(gitBranchName)
	if err != nil {
		tools.PrintErrorAndExit(err, "Error creating fake release", "", 1)
	}
	return url
}

func CreateFakeRelease(gitBranchName *string) (url string, err error) {
	tools.PrintInfo("Creating fake release for branch %s", 0, *gitBranchName)
	if *gitBranchName == "" {
		return
	}

	var pullRequests GHPulls
	pullRequests, err = GetPullRequests("united-manufacturing-hub", "united-manufacturing-hub")
	if err != nil {
		return "", err
	}

	var pullRequest *GHPull
	for _, pull := range pullRequests {
		if pull.Head.Ref == *gitBranchName {
			pullRequest = &pull
		}
	}
	if pullRequest == nil {
		return "", fmt.Errorf("no pull request found for branch %s", *gitBranchName)
	}

	// Create a temp folder

	var tempDir string
	tempDir, err = os.MkdirTemp("", "autok3d*")
	if err != nil {
		return "", err
	}
	var fullPath string
	fullPath, err = DownloadBranch(*gitBranchName, tempDir)
	if err != nil {
		return "", err
	}

	tools.PrintInfo("Creating release for branch %s", 1, *gitBranchName)

	fmt.Printf("fullPath: %s", fullPath)

	return "", nil
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
	defer tempFile.Close()

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
