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
func WriteFile(loc string, data string) (err error) {
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
		servConfBroken bool
		servConfExist  bool
		pkgjsonBroken  bool
		pkgjsonExist   bool
	}{
		{false, true, false, true},   // no files are broken or missing
		{true, false, true, false},   // both files are broken
		{false, false, false, false}, // both files are missing
	}

	cmd := AppsInitCmd{
		StackName: "nodejs-basic",
		Force:     false,
	}
	for _, tc := range testCases {
		n := ""
		t.Run(n, func(t *testing.T) {
			// Setup
			err1 := os.Remove("testdata/package.json")
			err2 := os.Remove("testdata/server.conf")
			if err1 != nil || err2 != nil {
				log.Println("[ERROR] unable to remove files, perhaps they do not exist?")
			}

			if tc.servConfBroken {
				err := WriteFile("testdata/server.conf", `proxy_set_header X-Forwarded-Proto $http_x_forwarded_proto;
						proxy_set_header Host $host;
						include /etc/nginx/section.module/node.conf;
					}
					
					location ~ "/next-proxy-hop/" {
						proxy_set_header X-Forwarded-For $http_x_forwarded_for;
						proxy_set_header X-Forwa`)
				if err != nil {
					fmt.Println("server.conf creation failed")
				}
			} else if tc.servConfExist {
				err := WriteFile("testdata/server.conf", `location / {
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
				}`)
				if err != nil {
					fmt.Println("testdata/server.conf creation failed")
				}
			}
			if tc.pkgjsonBroken {
				err := WriteFile("testdata/package.json", `{
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
				}`)

				if err != nil {
					fmt.Println("package.json creation failed")
				}
			} else if tc.pkgjsonExist {
				err := WriteFile("testdata/package.json", `{
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
				}`)
				if err != nil {
					fmt.Println("package.json creation failed")
				}
			}
			// Invoke
			err := cmd.Run()

			// Test
			assert.NoError(err)
		})
		err1 := os.Remove("testdata/package.json")
		err2 := os.Remove("testdata/server.conf")
		if err1 != nil || err2 != nil {
			log.Println("[ERROR] unable to remove files, perhaps they do not exist?")
		}
	}
}
