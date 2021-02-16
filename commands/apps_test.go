package commands

import (
	"fmt"
	"log"
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

	// Invoke
	err = cmd.Run()

	// Test
	assert.True(stackCalled)
	assert.Error(err)
	assert.Regexp("bad request: unable to find stack", err)
}

// Looks like nested functions are not yet supported by the compiler
func OverwriteFile(loc string, data string) (err error) {
	err = os.Remove(loc)
	if err != nil {
		log.Println("[ERROR] unable to remove files, perhaps they do not exist?")
	}
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

func TestInitNodejsBasicAppEncounteringPossibleFailureStates(t *testing.T) {
	assert := assert.New(t)

	// Setup
	var testCases = []struct {
		servConf string
		pkgJson  string
		isFatal  bool
	}{
		{`location / {
			proxy_set_header X-Forwarded-For $http_x_forwarded_for;
			proxy_set_header X-Forwarded-Proto $http_x_forwarded_proto;
			proxy_set_header Host $host;
			include /etc/nginx/section.module/node.conf;
		}
		
		location ~ "/next-proxy-hop/" {
			proxy_set_header X-Forwarded-For $http_x_forwarded_for;
			proxy_set_header X-Forwarded-Proto $http_x_forwarded_proto;
			proxy_set_header Host $host;
			proxy_pass http://next-hop;
		}`, `{
			"name": "api-routes",
			"version": "1.0.0",
			"scripts": {
				"dev": "next",
				"build": "next build",
				"start": "next start -p 8080",
				"predeploy": "npm install && npm run build",
				"deploy": "sectionctl deploy -a 1887 -i 7749"
			},
			"dependencies": {
				"next": "latest",
				"react": "^16.8.6",
				"react-dom": "^16.8.6",
				"swr": "0.1.18"
			},
			"license": "MIT"
		}`, false}, // no files are broken or missing
		{`proxy_set_header X-Forwarded-Proto $http_x_forwarded_proto;
		proxy_set_header Host $host;
		include /etc/nginx/section.module/node.conf;
	}
	
	location ~ "/next-proxy-hop/" {
		proxy_set_header X-Forwarded-For $http_x_forwarded_for;
		proxy_set_header X-Forwa`, `{
			"name": "api-routes",
			"version": "1.0.0",
			"scripts": {
				"dev": "next",
				"build": "next build",
				"predeploy": "npm install && npm run build",
				"deploy": "sectionctl deploy -a 1887 -i 7749"
			},
			"dependencies": {
				"next": "latest",
				"react": "^16.8.6",
				"react-dom": "^16.8.6",
				"swr": "0.1.18"
			},
			"license": "MIT"
		}`, false}, // both files are broken
		{``, ``, false}, // both files are empty
		{``, `{
			"name": "api-routes",
			"version": "1.0.0",
			"scripts": {
				"dev": "next",
				"build": "next build"
				"predeploy": "npm install887 -i 7749"
			"dependencies": {
				"next": "latest",
				"react": "^16.8.6",
				"react-dom": "^16.8.6",
				"swr": "0.1.18"
			},
			"license": "MIT"`, true}, // package.json can not be unmarshalled
	}

	cmd := AppsInitCmd{
		StackName: "nodejs-basic",
		Force:     false,
	}
	for _, tc := range testCases {
		n := ""
		t.Run(n, func(t *testing.T) {
			// Setup
			err := OverwriteFile("server.conf", tc.servConf)
			if err != nil {
				fmt.Println("server.conf creation failed")
			}

			err = OverwriteFile("package.json", tc.pkgJson)
			if err != nil {
				fmt.Println("server.conf creation failed")
			}

			// Invoke
			err = cmd.Run()

			// Test
			if tc.isFatal {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
		err1 := os.Remove("package.json")
		err2 := os.Remove("server.conf")
		if err1 != nil || err2 != nil {
			log.Println("[ERROR] unable to remove files, perhaps they do not exist?")
		}
	}
}
