package commands

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/section/sectionctl/api/auth"
)

// MaxFileSize is the tarball file size allowed to be uploaded in bytes.
const MaxFileSize = 1073741824 // 1GB

// DeployCmd handles deploying an app to Section.
type DeployCmd struct {
	AccountID        int `default:"4322"`  // harc-coded until authentication is implemented
	AppID            int `default:"65443"` // hard-coded for now until authentication is implmented
	Debug            bool
	Directory        string   `default:"."`
	ServerURL        *url.URL `default:"https://aperture.section.io/new/code_upload/v1"`
	ApertureURL      string   `default:"https://aperture.section.io/api"`
	EnvUpdatePathFmt string   `default:"/account/%d/application/%d/environment/%s/update"`
}

// UploadResponse represents the response from a request to the upload service.
type UploadResponse struct {
	PayloadID string `json:"payloadID"`
}

// Run deploys an app to Section's edge
func (c *DeployCmd) Run() (err error) {
	if c.Debug {
		fmt.Println("Server URL:", c.ServerURL.String())
	}

	ignores := []string{".lint/", ".git/"}
	files, err := BuildFilelist(c.Directory, ignores)
	if c.Debug {
		fmt.Println("Archiving files:")
		for _, file := range files {
			fmt.Println(file)
		}
	}

	dir := c.Directory
	if dir == "." {
		abs, err := filepath.Abs(dir)
		if err == nil {
			dir = abs
		}
	}
	fmt.Printf("Packaging app in: %s\n", dir)

	tempFile, err := ioutil.TempFile("", "section")
	if err != nil {
		return fmt.Errorf("couldn't create a temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	err = CreateTarball(tempFile, files)
	if err != nil {
		return fmt.Errorf("failed to pack files: %v", err)
	}
	stat, err := tempFile.Stat()
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Could not get stat for file '%s', got error '%s'", tempFile.Name(), err.Error()))
	}
	if stat.Size() > MaxFileSize {
		return fmt.Errorf("failed to upload tarball: file size (%d) is greater than (%d)", stat.Size(), MaxFileSize)
	}

	_, err = tempFile.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("unable to seek to beginning of tarball: %s", err)
	}

	fmt.Printf("Pushing %d bytes...\n", stat.Size())

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	req, err := http.NewRequest(http.MethodPost, c.ServerURL.String(), tempFile)
	if err != nil {
		return fmt.Errorf("failed to create upload URL: %v", err)
	}

	username, password, err := auth.GetCredential(c.ServerURL.Host)
	if err != nil {
		return fmt.Errorf("unable to read credentials: %s", err)
	}
	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("upload request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return fmt.Errorf("upload failed with status: %s and transaction ID %s", resp.Status, resp.Header["Aperture-Tx-Id"][0])
	}

	var response UploadResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return fmt.Errorf("failed to decode response %v", err)
	}
	svcURL := c.ApertureURL + fmt.Sprintf(c.EnvUpdatePathFmt, c.AccountID, c.AppID, "production")
	err = triggerUpdate(c.AccountID, c.AppID, response.PayloadID, svcURL, client)
	if err != nil {
		return fmt.Errorf("failed to trigger app update %v", err)
	}

	fmt.Println("Done.")

	return nil
}

// BuildFilelist builds a list of files to be tarballed, with optional ignores.
func BuildFilelist(dir string, ignores []string) (files []string, err error) {
	var fi os.FileInfo
	if fi, err = os.Stat(dir); os.IsNotExist(err) {
		return files, err
	}
	if !fi.IsDir() {
		return files, fmt.Errorf("specified path is not a directory: %s", dir)
	}

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		for _, i := range ignores {
			if strings.Contains(path, i) {
				return nil

			}
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return files, fmt.Errorf("failed to walk path: %v", err)
	}
	return files, err
}

// CreateTarball creates a tarball containing all the files in filePaths and writes it to w.
func CreateTarball(w io.Writer, filePaths []string) error {
	gzipWriter := gzip.NewWriter(w)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	for _, filePath := range filePaths {
		err := addFileToTarWriter(filePath, tarWriter)
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Could not add file '%s', to tarball, got error '%s'", filePath, err.Error()))
		}
	}

	return nil
}

func addFileToTarWriter(filePath string, tarWriter *tar.Writer) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Could not open file '%s', got error '%s'", filePath, err.Error()))
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Could not get stat for file '%s', got error '%s'", filePath, err.Error()))
	}

	header := &tar.Header{
		Name:    filePath,
		Size:    stat.Size(),
		Mode:    int64(stat.Mode()),
		ModTime: stat.ModTime(),
	}

	err = tarWriter.WriteHeader(header)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Could not write header for file '%s', got error '%s'", filePath, err.Error()))
	}

	_, err = io.Copy(tarWriter, file)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Could not copy the file '%s' data to the tarball, got error '%s'", filePath, err.Error()))
	}

	return nil
}

// PayloadValue represents the value of a trigger update payload.
type PayloadValue struct {
	ID string `json:"section_payload_id"`
}

func triggerUpdate(accountID, appID int, payloadID, serviceURL string, c *http.Client) error {
	var b bytes.Buffer
	payload := []struct {
		Op    string       `json:"op"`
		Path  string       `json:"path"`
		Value PayloadValue `json:"value"`
	}{
		{
			Op: "replace",
			Value: PayloadValue{
				ID: payloadID,
			},
		},
	}

	err := json.NewEncoder(&b).Encode(payload)
	req, err := http.NewRequest(http.MethodPatch, serviceURL, &b)
	if err != nil {
		return fmt.Errorf("failed to create trigger request: %v", err)
	}
	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute trigger request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return fmt.Errorf("trigger update failed with status %s", resp.Status)
	}
	return nil
}
