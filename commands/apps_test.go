package commands

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
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

	// Invoke
	err = cmd.Run()

	// Test
	assert.True(stackCalled)
	assert.Error(err)
	assert.Regexp("bad request: unable to find stack", err)
}
