package commands

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/section/sectionctl/api"
)

// MaxFileSize is the tarball file size allowed to be uploaded in bytes.
const MaxFileSize = 1073741824 // 1GB

// DeployCmd handles deploying an app to Section.
type DeployCmd struct {
	AccountID      int           `required short:"a" help:"AccountID to deploy application to."`
	AppID          int           `required short:"i" help:"AppID to deploy application to."`
	Environment    string        `short:"e" default:"Production" help:"Environment to deploy application to. (name of git branch ie: Production, staging, development)"`
	Directory      string        `short:"C" default:"." help:"Directory which contains the application to deploy."`
	ServerURL      *url.URL      `default:"https://aperture.section.io/new/code_upload/v1/upload" help:"URL to upload application to"`
	Timeout        time.Duration `default:"600s" help:"Timeout of individual HTTP requests."`
	SkipDelete     bool          `help:"Skip delete of temporary tarball created to upload app."`
	SkipValidation bool          `help:"Skip validation of the workload before pushing into Section. Use with caution."`
	AppPath        string        `default:"nodejs" help:"Path of NodeJS application in environment repository."`
}

// UploadResponse represents the response from a request to the upload service.
type UploadResponse struct {
	PayloadID string `json:"payloadID"`
}

// PayloadValue represents the value of a trigger update payload.
type PayloadValue struct {
	ID string `json:"section_payload_id"`
}

// Run deploys an app to Section's edge
func (c *DeployCmd) Run(ctx context.Context) (err error) {
	dir := c.Directory
	if dir == "." {
		abs, err := filepath.Abs(dir)
		if err == nil {
			dir = abs
		}
	}

	if !c.SkipValidation {
		errs := IsValidNodeApp(dir)
		if len(errs) > 0 {
			var se []string
			for _, err := range errs {
				se = append(se, fmt.Sprintf("- %s", err))
			}
			errstr := strings.Join(se, "\n")
			return fmt.Errorf("not a valid Node.js app: \n\n%s", errstr)
		}
	}

	s := NewSpinner(ctx, fmt.Sprintf("Packaging app in: %s", dir))
	s.Start()

	ignores := []string{".lint", ".git"}
	files, err := BuildFilelist(dir, ignores)
	if err != nil {
		s.Stop()
		return fmt.Errorf("unable to build file list: %s", err)
	}
	log.Println("[DEBUG] Archiving files:")
	for _, file := range files {
		log.Println("[DEBUG]", file)
	}

	tempFile, err := ioutil.TempFile("", "sectionctl-deploy.*.tar.gz")
	if err != nil {
		s.Stop()
		return fmt.Errorf("couldn't create a temp file: %v", err)
	}
	if c.SkipDelete {
		s.Stop()
		log.Println("[DEBUG] Temporary upload tarball location:", tempFile.Name())
		s.Start()
	} else {
		defer os.Remove(tempFile.Name())
	}

	err = CreateTarball(tempFile, files)
	if err != nil {
		s.Stop()
		return fmt.Errorf("failed to pack files: %v", err)
	}
	s.Stop()

	log.Println("[DEBUG] Temporary tarball path:", tempFile.Name())
	stat, err := tempFile.Stat()
	if err != nil {
		return fmt.Errorf("%s: could not stat, got error: %s", tempFile.Name(), err)
	}
	if stat.Size() > MaxFileSize {
		return fmt.Errorf("failed to upload tarball: file size (%d) is greater than (%d)", stat.Size(), MaxFileSize)
	}

	_, err = tempFile.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("unable to seek to beginning of tarball: %s", err)
	}

	req, err := newFileUploadRequest(c, tempFile)
	if err != nil {
		return fmt.Errorf("unable to build file upload: %s", err)
	}

	req.Header.Add("section-token", api.Token)

	log.Println("[DEBUG] Request URL:", req.URL)

	artifactSizeMB := stat.Size() / 1024 / 1024
	log.Printf("[DEBUG] Upload artifact is %dMB (%d bytes) large", artifactSizeMB, stat.Size())
	s = NewSpinner(ctx, fmt.Sprintf("Uploading app (%dMB)...", artifactSizeMB))
	s.Start()
	client := &http.Client{
		Timeout: c.Timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("upload request failed: %v", err)
	}
	defer resp.Body.Close()
	s.Stop()
	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return fmt.Errorf("upload failed with status: %s and transaction ID %s", resp.Status, resp.Header["Aperture-Tx-Id"][0])
	}

	var response UploadResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return fmt.Errorf("failed to decode response %v", err)
	}

	err = globalGitService.UpdateGitViaGit(ctx, c, response)
	if err != nil {
		return fmt.Errorf("failed to trigger app update: %v", err)
	}

	log.Println("[INFO] Done!")

	return nil
}

// IsValidNodeApp detects if a Node.js app is present in a given directory
func IsValidNodeApp(dir string) (errs []error) {
	packageJSONPath := filepath.Join(dir, "package.json")
	if _, err := os.Stat(packageJSONPath); os.IsNotExist(err) {
		errs = append(errs, fmt.Errorf("%s is not a file", packageJSONPath))
	}

	nodeModulesPath := filepath.Join(dir, "node_modules")
	fi, err := os.Stat(nodeModulesPath)
	if os.IsNotExist(err) {
		errs = append(errs, fmt.Errorf("%s is not a directory", nodeModulesPath))
	} else {
		if !fi.IsDir() {
			errs = append(errs, fmt.Errorf("%s is not a directory", nodeModulesPath))
		}
	}

	return errs
}

// Split helps differentiate between different directory delimiters. / or \
func Split(r rune) bool {
	return r == '\\' || r == '/'
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
		for _, i := range ignores {
			location := strings.FieldsFunc(path, Split) // split by subdirectory or filename
			for _, loc := range location {
				if strings.Contains(loc, i) {
					return nil
				}
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

	prefix := filePaths[0]
	for _, filePath := range filePaths {
		err := addFileToTarWriter(filePath, tarWriter, prefix)
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Could not add file '%s', to tarball, got error '%s'", filePath, err.Error()))
		}
	}

	return nil
}

func addFileToTarWriter(filePath string, tarWriter *tar.Writer, prefix string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Could not open file '%s', got error '%s'", filePath, err.Error()))
	}
	defer file.Close()

	stat, err := os.Lstat(filePath)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Could not get stat for file '%s', got error '%s'", filePath, err.Error()))
	}

	baseFilePath := strings.TrimPrefix(filePath, prefix)
	header, err := tar.FileInfoHeader(stat, baseFilePath)
	if err != nil {
		return err
	}
	if stat.Mode()&os.ModeSymlink == os.ModeSymlink {
		link, err := os.Readlink(filePath)
		if err != nil {
			return err
		}
		header.Linkname = link
	}

	// must provide real name
	// (see https://golang.org/src/archive/tar/common.go?#L626)
	header.Name = filepath.ToSlash(baseFilePath)
	// ensure windows provides filemodes for binaries in node_modules/.bin
	if runtime.GOOS == "windows" {
		match := strings.Contains(baseFilePath, "node_modules\\.bin")
		if match {
			header.Mode = 0o755
		}
	}
	err = tarWriter.WriteHeader(header)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Could not write header for file '%s', got error '%s'", baseFilePath, err.Error()))
	}

	if !stat.IsDir() && stat.Mode()&os.ModeSymlink != os.ModeSymlink {
		_, err = io.Copy(tarWriter, file)
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Could not copy the file '%s' data to the tarball, got error '%s'", baseFilePath, err.Error()))
		}
	}

	return nil
}

// newFileUploadRequest builds a HTTP request for uploading an app and the account + app it belongs to
func newFileUploadRequest(c *DeployCmd, f *os.File) (r *http.Request, err error) {
	contents, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filepath.Base(f.Name()))
	if err != nil {
		return nil, err
	}
	_, err = part.Write(contents)
	if err != nil {
		return nil, err
	}

	err = writer.WriteField("account_id", strconv.Itoa(c.AccountID))
	if err != nil {
		return nil, err
	}
	err = writer.WriteField("app_id", strconv.Itoa(c.AppID))
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, c.ServerURL.String(), &body)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload URL: %v", err)
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())

	return req, err
}
