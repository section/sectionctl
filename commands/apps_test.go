package commands

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/section/sectionctl/api"
	"github.com/stretchr/testify/assert"
)

func TestCommandsAppsCreateAttemptsToValidateStackOnError(t *testing.T) {
	assert := assert.New(t)

	// Setup
	var stackCalled bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("test request: %s\n", r.URL.Path)
		switch r.URL.Path {
		case "/api/v1/account/0/application/create":
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, "{}")
		case "/api/v1/stack":
			stackCalled = true
			fmt.Fprint(w, string(helperLoadBytes(t, "apps/create.response.with_success.json")))
		default:
			assert.FailNowf("unhandled URL", "URL: %s", r.URL.Path)
		}
	}))
	ur, err := url.Parse(ts.URL)
	assert.NoError(err)
	api.PrefixURI = ur

	cmd := AppsCreateCmd{
		StackName: "helloworld-1.0.0",
	}

	ctx := context.Background()

	// Invoke
	err = cmd.Run(ctx)

	// Test
	assert.True(stackCalled)
	assert.Error(err)
	assert.Regexp("bad request: unable to find stack", err)
}

func TestCommandsAppsInitHandlesErrors(t *testing.T) {
	assert := assert.New(t)

	// Setup
	var testCases = []struct {
		servConf string
		pkgJSON  string
		isFatal  bool
	}{
		{"ABSENT", string(helperLoadBytes(t, "apps/init.good.package.json")), false},                                                 // server.conf is missing
		{string(helperLoadBytes(t, "apps/init.good.server.conf")), string(helperLoadBytes(t, "apps/init.good.package.json")), false}, // no files are broken or missing
		{string(helperLoadBytes(t, "apps/init.bad.server.conf")), string(helperLoadBytes(t, "apps/init.bad.package.json")), false},   // both files are broken
		{``, ``, false}, // both files are empty
		{``, string(helperLoadBytes(t, "apps/init.broken.package.json")), true}, // package.json can not be unmarshalled
	}

	cmd := AppsInitCmd{
		StackName: "nodejs-basic",
		Force:     false,
	}

	OverwriteFile := func(loc string, data string) (err error) {
		f, err := os.Create(loc)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.WriteString(data)
		if err != nil {
			return err
		}
		return err
	}

	owd, err := os.Getwd()
	assert.NoError(err)

	for _, tc := range testCases {
		n := ""
		t.Run(n, func(t *testing.T) {
			// Setup
			// Create a temp directory
			tempDir, err := ioutil.TempDir("", "sectionctl-apps-init")
			assert.NoError(err)
			err = os.Chdir(tempDir)
			assert.NoError(err)

			if tc.servConf != "ABSENT" {
				err = OverwriteFile("server.conf", tc.servConf)
				if err != nil {
					fmt.Println("server.conf creation failed")
				}
			}

			err = OverwriteFile("package.json", tc.pkgJSON)
			if err != nil {
				fmt.Println("server.conf creation failed")
			}

			ctx := context.Background()

			// Invoke
			err = cmd.Run(ctx)

			// Test
			if tc.isFatal {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}

			// Return to base directory
			err = os.Chdir(owd)
			assert.NoError(err)
		})
	}
}
