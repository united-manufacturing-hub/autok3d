package github

import (
	"bytes"
	"fmt"
	"github.com/cristalhq/base64"
	jsoniter "github.com/json-iterator/go"
	"github.com/united-manufacturing-hub/autok3d/cmd/tools"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const httpPass = "Sandstorm-Stranger-Nanometer-Wing3-Clicker"

type upFile struct {
	Filename string `json:"filename"`
	Filebody string `json:"filebody"`
	Password string `json:"password"`
}

func uploadFile(fp os.DirEntry, prefix string, group *sync.WaitGroup, baseDir string) {
	defer group.Done()
	if fp.IsDir() {
		return
	}
	if !strings.HasSuffix(fp.Name(), ".tgz") {
		return
	}
	// Read filepath
	file, err := os.Open(filepath.Join(baseDir, fp.Name()))
	if err != nil {
		tools.PrintErrorAndExit(err, "error opening file", "", 1)
	}
	defer file.Close()

	// Read the file
	filebody, err := io.ReadAll(file)
	if err != nil {
		tools.PrintErrorAndExit(err, "error reading file", "", 1)
	}
	upload(fp.Name(), filebody)
}

func upload(filename string, filebody []byte) {
	// Do a post request to /upload with file as json.
	// Filebody is base64 encoded.

	filebodyBase64 := base64.StdEncoding.EncodeToString(filebody)

	f := upFile{
		Filename: filename,
		Filebody: filebodyBase64,
		Password: httpPass,
	}

	// JSON encode the file
	fJSON, err := json.Marshal(f)
	if err != nil {
		tools.PrintErrorAndExit(err, "error encoding file to json", "", 1)
	}

	// Create a new HTTP request
	req, err := http.NewRequest("POST", "https://test-repo.umh.app/upload", bytes.NewBuffer(fJSON))
	if err != nil {
		tools.PrintErrorAndExit(err, "error creating the post request", "", 1)
	}

	// Set the content type of the request to "application/json"
	req.Header.Set("Content-Type", "application/json")

	// Send the request and get the response
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		tools.PrintErrorAndExit(err, "error sending the post request", "", 1)
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		tools.PrintErrorAndExit(err, "error reading the response body", "", 1)
	}

	// Print the response body
	tools.PrintSuccess(fmt.Sprintf("Uploaded file to https://test-repo.umh.app/%s (%s)", filename, string(respBody)), 2)

}

func download(filename string, outfile *os.File) bool {
	defer outfile.Close()
	// Download a file from the server

	// Send a GET request to the URL
	resp, err := http.Get(fmt.Sprintf("https://test-repo.umh.app/%s", filename))
	if err != nil {
		tools.PrintErrorAndExit(err, "error sending the get request", "", 1)
	}

	// Check response status code
	if resp.StatusCode != 200 {
		return false
	}

	defer resp.Body.Close()

	// Copy the response body to the file
	_, err = io.Copy(outfile, resp.Body)
	if err != nil {
		tools.PrintErrorAndExit(err, "error copying the response body to the file", "", 1)
	}
	return true
}
