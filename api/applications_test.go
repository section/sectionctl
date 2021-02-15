package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"os"
	"bytes"
	"log"

	"github.com/stretchr/testify/assert"
)

func TestApplicationEnvironmentModuleUpdateSendsUpdateInArray(t *testing.T) {
	assert := assert.New(t)

	// Setup
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		to := r.Header.Get("section-token")
		assert.NotEmpty(to)
		w.Header().Add("Aperture-Tx-Id", "400400400400.400400")

		b, err := ioutil.ReadAll(r.Body)
		assert.NoError(err)
		var ups []EnvironmentUpdateCommand
		err = json.Unmarshal(b, &ups)
		assert.NoError(err)
		if assert.Equal(len(ups), 1) {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	url, err := url.Parse(ts.URL)
	assert.NoError(err)
	PrefixURI = url
	Token = "s3cr3t"

	// Invoke
	var ups = []EnvironmentUpdateCommand{
		EnvironmentUpdateCommand{Op: "replace", Value: map[string]string{"hello": "world"}},
	}
	err = ApplicationEnvironmentModuleUpdate(1, 1, "production", "hello/world.json", ups)

	// Test
	assert.NoError(err)
}

func TestApplicationEnvironmentModuleUpdateErrorsIfRequestFails(t *testing.T) {
	assert := assert.New(t)

	// Setup
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		to := r.Header.Get("section-token")
		assert.NotEmpty(to)
		w.Header().Add("Aperture-Tx-Id", "400400400400.400400")
		w.WriteHeader(http.StatusBadRequest)
	}))
	url, err := url.Parse(ts.URL)
	assert.NoError(err)
	PrefixURI = url
	Token = "s3cr3t"

	// Invoke
	var ups = []EnvironmentUpdateCommand{
		EnvironmentUpdateCommand{Op: "replace", Value: map[string]string{"hello": "world"}},
	}
	err = ApplicationEnvironmentModuleUpdate(1, 1, "production", "hello/world.json", ups)

	// Test
	assert.Error(err)
}

func TestAPIApplicationCreateReturnsUniqueErrorsOnFailure(t *testing.T) {
	assert := assert.New(t)

	// Setup
	var testCases = []struct {
		requestAppID     int
		requestOrigin    string
		requestHostname  string
		requestStackName string
		responseStatus   int
		responseBody     string
		responseError    error
	}{
		{123, "hello.example", "127.0.0.1", "nodejs", http.StatusUnauthorized, "", ErrStatusUnauthorized},
		{123, "hello.example", "127.0.0.1", "nodejs", http.StatusForbidden, "An application has already been created with domain name hello.example", ErrApplicationAlreadyCreated},
		{123, "hello.example", "127.0.0.1", "nodejs", http.StatusForbidden, "System limit exceeded. Contact support to increase this limit.", ErrSystemLimitExceeded},
		{123, "hello.example", "127.0.0.1", "nodejs", http.StatusForbidden, "An unhandled error", ErrStatusForbidden},
	}

	for _, tc := range testCases {
		n := fmt.Sprintf("%d-%s", tc.responseStatus, tc.responseBody)
		t.Run(n, func(t *testing.T) {
			// Setup
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.responseStatus)
				fmt.Fprint(w, tc.responseBody)
			}))
			url, err := url.Parse(ts.URL)
			assert.NoError(err)
			PrefixURI = url
			Token = "s3cr3t"

			// Invoke
			_, err = ApplicationCreate(tc.requestAppID, tc.requestHostname, tc.requestOrigin, tc.requestStackName)

			// Test
			assert.Error(err, tc.responseError)
		})
	}
}

func TestAPIApplicationDeleteReturnsUniqueErrorsOnFailure(t *testing.T) {
	assert := assert.New(t)

	// Setup
	var testCases = []struct {
		requestAccountID int
		requestAppID     int
		responseStatus   int
		responseBody     string
		responseError    error
	}{
		{456, 123, http.StatusBadRequest, "", ErrStatusBadRequest},
		{456, 123, http.StatusUnauthorized, "", ErrStatusUnauthorized},
		{456, 123, http.StatusForbidden, "", ErrSystemLimitExceeded},
		{456, 123, http.StatusNotFound, "", ErrStatusNotFound},
		{456, 123, http.StatusInternalServerError, "", ErrStatusInternalServerError},
	}

	for _, tc := range testCases {
		n := fmt.Sprintf("%d-%s", tc.responseStatus, tc.responseError)
		t.Run(n, func(t *testing.T) {
			// Setup
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.responseStatus)
				fmt.Fprint(w, tc.responseBody)
			}))
			url, err := url.Parse(ts.URL)
			assert.NoError(err)
			PrefixURI = url
			Token = "s3cr3t"

			// Invoke
			_, err = ApplicationDelete(tc.requestAccountID, tc.requestAppID)

			// Test
			assert.Error(err, tc.responseError)
		})
	}
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
		servConfBroken    bool
		servConfExist    bool
		pkgjsonBroken    bool
		pkgjsonExist		 bool
	}{
		{false, false, false, false}, // no files are broken or missing
		{true, false, true, false}, // both files are broken 
		{false, true, false, true}, // both files are missing
	}


	var stdout bytes.Buffer
	var stderr bytes.Buffer
	for _, tc := range testCases {
		n := fmt.Sprintf("")
		t.Run(n, func(t *testing.T) {
			// Setup
			err1 := os.Remove("package.json")
			err2 := os.Remove("server.conf")
			stackName := "nodejs-basic"
			force := false
			if err1 != nil || err2 != nil {
				log.Println("[ERROR] unable to remove files, perhaps they do not exist?")
			}

			if tc.servConfBroken {
				err := WriteFile("server.conf", `proxy_set_header X-Forwarded-Proto $http_x_forwarded_proto;
						proxy_set_header Host $host;
						include /etc/nginx/section.module/node.conf;
					}
					
					location ~ "/next-proxy-hop/" {
						proxy_set_header X-Forwarded-For $http_x_forwarded_for;
						proxy_set_header X-Forwa`)
				if err != nil {
					fmt.Println("server.conf creation failed")
				}
			} else if (tc.servConfExist) {
				err := WriteFile("server.conf", 	`location / {
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
					fmt.Println("server.conf creation failed")
				}
			}	
			if tc.pkgjsonBroken {
				err := WriteFile("package.json", `{
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
				err := WriteFile("package.json", `{
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
			err := InitializeNodeBasicApp(stdout, stderr)

			// Test
			assert.Error(err)
		})
	}
}
