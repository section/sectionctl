package commands

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/section/sectionctl/api"
	"github.com/stretchr/testify/assert"
)

type MockGitService struct {
	Called bool
}

func (g *MockGitService) UpdateGitViaGit(ctx context.Context, c *DeployCmd, response UploadResponse) error {
	g.Called = true
	return nil
}

func helperLoadBytes(t *testing.T, name string) []byte {
	path := filepath.Join("testdata", name) // relative path
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

func TestCommandsDeployBuildFilelistIgnoresFiles(t *testing.T) {
	assert := assert.New(t)

	// Setup
	dir := filepath.Join("testdata", "deploy", "tree")
	ignores := []string{".git", "foo"}

	// Invoke
	paths, err := BuildFilelist(dir, ignores)

	// Test
	assert.NoError(err)
	assert.Greater(len(paths), 0)
	for _, p := range paths {
		for _, i := range ignores {
			assert.NotContains(p, i)
		}
	}
}

func TestCommandsDeployBuildFilelistErrorsOnNonExistentDirectory(t *testing.T) {
	assert := assert.New(t)

	// Setup
	testCases := []string{
		filepath.Join("testdata", "deploy", "non-existent-tree"),
		filepath.Join("testdata", "deploy", "file"),
	}
	var ignores []string

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			// Invoke
			paths, err := BuildFilelist(tc, ignores)

			// Test
			assert.Error(err)
			assert.Zero(len(paths))
		})
	}
}

func TestCommandsDeployCreateTarballAlwaysPutsAppAtRoot(t *testing.T) {
	assert := assert.New(t)

	// Setup
	var testCases = []struct {
		cwd    string // change into this directory
		target string // bundle up files from this directory
	}{
		{".", filepath.Join("testdata", "deploy", "valid-nodejs-app")},
		{filepath.Join("testdata", "deploy", "valid-nodejs-app"), "."},
	}
	var ignores []string

	for _, tc := range testCases {
		t.Run(tc.target, func(t *testing.T) {
			tempFile, err := ioutil.TempFile("", "sectionctl-deploy")
			assert.NoError(err)

			// Record where we were at the beginning of the test
			owd, err := os.Getwd()
			assert.NoError(err)

			err = os.Chdir(tc.cwd)
			assert.NoError(err)

			// Build the file list
			paths, err := BuildFilelist(tc.target, ignores)
			assert.NoError(err)

			// Create the tarball
			err = CreateTarball(tempFile, paths)
			assert.NoError(err)

			_, err = tempFile.Seek(0, 0)
			assert.NoError(err)

			// Extract the tarball
			tempDir, err := ioutil.TempDir("", "sectionctl-deploy")
			assert.NoError(err)
			err = untar(tempFile, tempDir)

			// Test
			assert.NoError(err)
			path := filepath.Join(tempDir, "package.json")
			_, err = os.Stat(path)
			assert.NoError(err)

			// Change back to where we were at the beginning of the test
			err = os.Chdir(owd)
			assert.NoError(err)
		})
	}
}

func untar(src io.Reader, dst string) (err error) {
	// ungzip
	zr, err := gzip.NewReader(src)
	if err != nil {
		return err
	}
	// untar
	tr := tar.NewReader(zr)

	// uncompress each element
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}

		// add dst + re-format slashes according to system
		target := filepath.Join(dst, header.Name)
		// if no join is needed, replace with ToSlash:
		// target = filepath.ToSlash(header.Name)

		// check the type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it (with 0755 permission)
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}
		// if it's a file create it (with same permission)
		case tar.TypeReg:
			fileToWrite, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			// copy over contents
			if _, err := io.Copy(fileToWrite, tr); err != nil {
				return err
			}
			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			fileToWrite.Close()
		}
	}
	return nil
}

func TestCommandsDeployValidatesPresenceOfNodeApp(t *testing.T) {
	assert := assert.New(t)

	// Setup
	var testCases = []struct {
		path  string
		valid bool
	}{
		{filepath.Join("testdata", "deploy", "tree"), false},
		{filepath.Join("testdata", "deploy", "valid-nodejs-app"), true},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			// Invoke
			errs := IsValidNodeApp(tc.path)

			// Test
			assert.Equal(len(errs) == 0, tc.valid)
			t.Logf("err: %v", errs)
		})
	}
}

func TestCommandsDeployUploadsTarball(t *testing.T) {
	assert := assert.New(t)

	// Setup
	type req struct {
		called    bool
		token     string
		accountID int
		file      []byte
	}
	var uploadReq req
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		to := r.Header.Get("section-token")
		assert.NotEmpty(to)

		switch r.URL.Path {
		case "/":
			uploadReq.called = true
			uploadReq.token = to

			r.ParseMultipartForm(MaxFileSize)

			file, _, err := r.FormFile("file")
			assert.NoError(err)
			b, err := ioutil.ReadAll(file)
			assert.NoError(err)
			uploadReq.file = b

			aid, err := strconv.Atoi(r.FormValue("account_id"))
			assert.NoError(err)
			uploadReq.accountID = aid

			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, string(helperLoadBytes(t, "deploy/upload.response.with_success.json")))
		default:
			assert.FailNowf("unhandled URL", "URL: %s", r.URL.Path)
		}

	}))

	url, err := url.Parse(ts.URL)
	assert.NoError(err)

	dir := filepath.Join("testdata", "deploy", "valid-nodejs-app")
	api.PrefixURI = url
	api.Token = "s3cr3t"

	mockGit := MockGitService{}
	globalGitService = &mockGit

	// Define Context
	ctx := context.Background()

	// Invoke
	c := DeployCmd{
		Directory:   dir,
		ServerURL:   url,
		AccountID:   100,
		AppID:       200,
		Environment: "dev",
		AppPath:     "nodejs",
	}
	err = c.Run(ctx)

	// Test
	assert.NoError(err)

	// upload request
	assert.True(uploadReq.called)
	assert.Equal(api.Token, uploadReq.token)
	assert.NotZero(len(uploadReq.file))
	assert.Equal([]byte{0x1f, 0x8b}, uploadReq.file[0:2]) // gzip header
	assert.Equal(c.AccountID, uploadReq.accountID)

	assert.True(mockGit.Called)
}
