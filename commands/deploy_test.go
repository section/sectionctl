package commands

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"

	"github.com/section/sectionctl/api/auth"
	"github.com/stretchr/testify/assert"
)

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

func TestCommandsDeployUploadsTarball(t *testing.T) {
	assert := assert.New(t)

	// Setup
	type req struct {
		called    bool
		username  string
		password  string
		body      []byte
		accountID int
		file      []byte
	}
	var uploadReq req
	var triggerUpdateReq req
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		assert.True(ok)

		switch r.URL.Path {
		case "/":
			uploadReq.called = true
			uploadReq.username = u
			uploadReq.password = p

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
		case "/api/v1/account/100/application/200/environment/production/update":
			triggerUpdateReq.called = true
			triggerUpdateReq.username = u
			triggerUpdateReq.password = p
			b, err := ioutil.ReadAll(r.Body)
			assert.NoError(err)
			triggerUpdateReq.body = b
			w.WriteHeader(http.StatusOK)
		default:
			assert.FailNow("unhandled URL %s", r.URL.Path)
		}

	}))

	url, err := url.Parse(ts.URL)
	assert.NoError(err)

	auth.CredentialPath = newCredentialTempfile(t)
	endpoint := url.Host
	username := "hello"
	password := "s3cr3t"
	auth.WriteCredential(endpoint, username, password)

	dir := filepath.Join("testdata", "deploy", "tree")

	// Invoke
	c := DeployCmd{
		Directory:        dir,
		ServerURL:        url,
		ApertureURL:      url.String() + "/api/v1",
		AccountID:        100,
		AppID:            200,
		EnvUpdatePathFmt: "/account/%d/application/%d/environment/%s/update",
	}
	err = c.Run()

	// Test
	assert.NoError(err)

	// upload request
	assert.True(uploadReq.called)
	assert.Equal(username, uploadReq.username)
	assert.Equal(password, uploadReq.password)
	assert.NotZero(len(uploadReq.file))
	assert.Equal([]byte{0x1f, 0x8b}, uploadReq.file[0:2]) // gzip header
	assert.Equal(c.AccountID, uploadReq.accountID)

	// trigger update request
	assert.True(triggerUpdateReq.called)
	assert.Equal(username, triggerUpdateReq.username)
	assert.Equal(password, triggerUpdateReq.password)
	assert.NotZero(len(triggerUpdateReq.body))
}

func TestCommandsDeployDoesNotBalloonMemory(t *testing.T) {
	assert := assert.New(t)

	PrintMemUsage()
	s0 := SnapshotMemUsage()

	// Setup
	url, err := url.Parse("http://127.0.0.1/")
	assert.NoError(err)

	dir := filepath.Join("testdata", "deploy", "tree")
	c := DeployCmd{
		Directory: dir,
		ServerURL: url,
	}

	ignores := []string{".lint/", ".git/"}
	files, err := BuildFilelist(c.Directory, ignores)
	assert.NoError(err)

	tempFile, err := ioutil.TempFile("", "section")
	assert.NoError(err)
	err = CreateTarball(tempFile, files)
	assert.NoError(err)
	_, err = tempFile.Seek(0, 0)
	assert.NoError(err)
	PrintMemUsage()
	s1 := SnapshotMemUsage()

	// Invoke
	_, err = newFileUploadRequest(&c, tempFile)
	assert.NoError(err)

	// Test
	PrintMemUsage()
	s2 := SnapshotMemUsage()

	t.Logf("%+v\n%+v\n%+v\n", s0.Alloc, s1.Alloc, s2.Alloc)
	assert.Less(s2.Alloc, uint64(10000000))
}

func SnapshotMemUsage() (m runtime.MemStats) {
	runtime.ReadMemStats(&m)
	return m
}

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
