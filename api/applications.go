package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// App represents an application deployed on Section
type App struct {
	ID              int           `json:"id"`
	Href            string        `json:"href"`
	ApplicationName string        `json:"application_name"`
	Environments    []Environment `json:"environments"`
}

// Environment represents an application's environments on Section
type Environment struct {
	ID              int      `json:"id"`
	Href            string   `json:"href"`
	EnvironmentName string   `json:"environment_name"`
	Domains         []Domain `json:"domains"`
	Stack           []Module `json:"stack"`
}

// Domain represents an applications environments' domains
type Domain struct {
	Name     string `json:"name"`
	ZoneName string `json:"zoneName"`
	CNAME    string `json:"cname"`
	Mode     string `json:"mode"`
}

// Module represents a proxy in the traffic delivery stack
type Module struct {
	Name  string `json:"name"`
	Image string `json:"image"`
	Href  string `json:"href"`
}

// AppStatus represents the status of an application deployed on Section
type AppStatus struct {
	InService    bool   `json:"inService"`
	State        string `json:"state"`
	InstanceName string `json:"instanceName"`
	PayloadID    string `json:"payloadID"`
	IsLatest     bool   `json:"isLatest"`
}

// AppLogs represents the logs from an application deployed on Section
type AppLogs struct {
	Timestamp    string `json:"timestamp"`
	InstanceName string `json:"instanceName"`
	Type         string `json:"type"`
	Message      string `json:"message"`
}

// Application returns detailed information about a given application.
func Application(accountID int, applicationID int) (a App, err error) {
	u := BaseURL()
	u.Path += fmt.Sprintf("/account/%d/application/%d", accountID, applicationID)

	resp, err := request(http.MethodGet, u, nil)
	if err != nil {
		return a, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return a, prettyTxIDError(resp)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return a, err
	}

	err = json.Unmarshal(body, &a)
	if err != nil {
		return a, err
	}

	envs, err := ApplicationEnvironments(accountID, applicationID)
	if err != nil {
		return a, err
	}
	a.Environments = envs

	for i, e := range a.Environments {
		stack, err := ApplicationEnvironmentStack(accountID, applicationID, e.EnvironmentName)
		if err != nil {
			return a, err
		}
		a.Environments[i].Stack = stack
	}

	return a, err
}

// ApplicationEnvironments returns environment information for a given application.
func ApplicationEnvironments(accountID int, applicationID int) (es []Environment, err error) {
	u := BaseURL()
	u.Path += fmt.Sprintf("/account/%d/application/%d/environment", accountID, applicationID)

	resp, err := request(http.MethodGet, u, nil)
	if err != nil {
		return es, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return es, prettyTxIDError(resp)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return es, err
	}

	err = json.Unmarshal(body, &es)
	if err != nil {
		return es, err
	}
	return es, err
}

// ApplicationEnvironmentStack returns the stack for a given application and environment.
func ApplicationEnvironmentStack(accountID int, applicationID int, environmentName string) (s []Module, err error) {
	u := BaseURL()
	u.Path += fmt.Sprintf("/account/%d/application/%d/environment/%s/stack", accountID, applicationID, environmentName)

	resp, err := request(http.MethodGet, u, nil)
	if err != nil {
		return s, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return s, prettyTxIDError(resp)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return s, err
	}

	err = json.Unmarshal(body, &s)
	if err != nil {
		return s, err
	}
	return s, err
}

// Applications returns a list of applications on a given account.
func Applications(accountID int) (as []App, err error) {
	u := BaseURL()
	u.Path += fmt.Sprintf("/account/%d/application", accountID)

	resp, err := request(http.MethodGet, u, nil)
	if err != nil {
		return as, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return as, fmt.Errorf("request failed with status %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return as, err
	}

	err = json.Unmarshal(body, &as)
	if err != nil {
		return as, err
	}
	return as, err
}

// EnvironmentUpdateCommand is a blah
type EnvironmentUpdateCommand struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

// ApplicationEnvironmentModuleUpdate updates a module's configuration
func ApplicationEnvironmentModuleUpdate(accountID int, applicationID int, env string, filePath string, up []EnvironmentUpdateCommand) (err error) {
	u := BaseURL()
	u.Path += fmt.Sprintf("/account/%d/application/%d/environment/%s/update", accountID, applicationID, env)

	b, err := json.Marshal(up)
	if err != nil {
		return fmt.Errorf("failed to encode json payload: %v", err)
	}
	log.Printf("[DEBUG] JSON payload: %s\n", b)
	headers := map[string][]string{"filepath": []string{filePath}}
	resp, err := request(http.MethodPatch, u, bytes.NewBuffer(b), headers)
	if err != nil {
		return fmt.Errorf("failed to execute trigger request: %v", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not read response body: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		var objmap map[string]interface{}
		if err := json.Unmarshal(body, &objmap); err != nil {
			nerr := fmt.Errorf("unable to decode error message: %s", err)
			return fmt.Errorf("trigger update failed with status: %s and transaction ID %s\n. Error received: \n%s", resp.Status, resp.Header["Aperture-Tx-Id"][0], nerr)
		}
		return fmt.Errorf("trigger update failed with status: %s and transaction ID %s\n. Error received: \n%s", resp.Status, resp.Header["Aperture-Tx-Id"][0], objmap["message"])
	}
	return nil
}

// getEnvironmentID returns the environment ID for a given account, application and environment name
func getEnvironmentID(accountID int, applicationID int, environmentName string) (int, error) {
	envs, err := ApplicationEnvironments(accountID, applicationID)
	if err != nil {
		return 0, err
	}

	for _, e := range envs {
		if e.EnvironmentName == environmentName {
			return e.ID, nil
		}
	}
	return 0, fmt.Errorf("could not find %s environment", environmentName)
}

// ApplicationStatus returns a module's current status on Section's delivery platform
func ApplicationStatus(accountID int, applicationID int, moduleName string) (as []AppStatus, err error) {
	u := BaseURL()
	u.Path = "/new/authorized/graphql_api/query"

	// Hard coding to Production environment for now.
	// Can be changed later for multiple environment support on the same application.
	environmentID, err := getEnvironmentID(accountID, applicationID, "Production")
	if err != nil {
		return as, err
	}

	var requestData struct {
		OperationName string                 `json:"operationName"`
		Variables     map[string]interface{} `json:"variables"`
		Query         string                 `json:"query"`
	}

	requestData.Variables = map[string]interface{}{
		"moduleName":    moduleName,
		"environmentID": environmentID,
	}
	requestData.Query = "query DeploymentStatus($moduleName: String!, $environmentID: Int!){deploymentStatus(moduleName:$moduleName, environmentID:$environmentID){inService state instanceName payloadID}}"

	data, err := json.Marshal(requestData)
	resp, err := request(http.MethodPost, u, bytes.NewBuffer(data))
	if err != nil {
		return as, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return as, prettyTxIDError(resp)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return as, fmt.Errorf("could not read response body: %s", err)
	}

	log.Printf("[DEBUG] RESPONSE: %s\n", string(body))

	var responseBody struct {
		Data struct {
			DeploymentStatus []AppStatus `json:"deploymentStatus"`
		} `json:"data"`
	}

	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		return as, err
	}

	as = responseBody.Data.DeploymentStatus
	return as, nil
}

// ApplicationLogs returns a module's logs from Section's delivery platform
func ApplicationLogs(accountID int, applicationID int, moduleName string, instanceName string, length int) (al []AppLogs, err error) {
	u := BaseURL()
	u.Path = "/new/authorized/graphql_api/query"

	// Hard coding to Production environment for now.
	// Can be changed later for multiple environment support on the same application.
	environmentID, err := getEnvironmentID(accountID, applicationID, "Production")
	if err != nil {
		return al, err
	}

	var requestData struct {
		OperationName string                 `json:"operationName"`
		Variables     map[string]interface{} `json:"variables"`
		Query         string                 `json:"query"`
	}

	requestData.Variables = map[string]interface{}{
		"environmentId": environmentID,
		"moduleName":    moduleName,
		"instanceName":  instanceName,
		"length":        length,
		// "startTimestamp": startTimestamp,
		// "endTimestamp": endTimestamp,
	}

	log.Printf("[DEBUG] requestData: %v\n", requestData.Variables)

	requestData.Query = "query Logs($moduleName: String!, $environmentId: Int!, $instanceName: String, $length: Int){logs(moduleName:$moduleName, environmentId:$environmentId, instanceName:$instanceName, length:$length){timestamp instanceName message type}}"

	data, err := json.Marshal(requestData)
	resp, err := request(http.MethodPost, u, bytes.NewBuffer(data))
	if err != nil {
		return al, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return al, prettyTxIDError(resp)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return al, fmt.Errorf("could not read response body: %s", err)
	}

	log.Printf("[DEBUG] RESPONSE: %s\n", string(body))

	var responseBody struct {
		Data struct {
			Logs []AppLogs `json:"logs"`
		} `json:"data"`
	}

	// Stubbed response for testing, uncomment to use
	// body = []byte(`{"data":{"logs":[{"timestamp":"2020-12-22T11:24:45.335Z","source":"testjs-d97hd-4k4", "type":"app","message":"Redirecting access.log to STDOUT and error.log to STDERR."},{"timestamp":"2020-12-22T11:24:45.335Z","source":"testjs-d97hd-4k4", "type":"app","message":"Listening on http://0.0.0.0:9000/metrics"},{"timestamp":"2020-12-22T11:24:45.335Z","source":"testjs-d97hd-4k4", "type":"app","message":"[WARN] Updating environment"},{"timestamp":"2020-12-22T11:24:47.335Z","source":"testjs-d97hd-4k4", "type":"app","message":"> learn-starter@0.1.0 start /opt/section/node"},{"timestamp":"2020-12-22T11:24:47.335Z","source":"testjs-d97hd-4k4", "type":"app","message":"> next start /opt/section/node -p 8080"},{"timestamp":"2020-12-22T11:24:47.335Z","source":"testjs-d97hd-4k4", "type":"app","message":"info  - Loaded env from /opt/section/node/.env.production"},{"timestamp":"2020-12-22T11:24:47.335Z","source":"testjs-d97hd-4k4", "type":"app","message":"ready - started server on http://localhost:8080"},{"timestamp":"2020-12-22T11:24:49.335Z","source":"testjs-d97hd-4k4", "type":"app","message":"[WARN] The node app is already being served."},{"timestamp":"2020-12-22T11:24:49.335Z","source":"testjs-d97hd-4k4", "type":"app","message":"[WARN] Initialising Nginx"},{"timestamp":"2020-12-22T11:24:49.335Z","source":"testjs-d97hd-4k4", "type":"app","message":"[WARN] Updating custom Nginx server.conf"},{"timestamp":"2020-12-22T11:24:15.335Z","source":"testjs-d97hd-4k4", "type":"app","message":" ERROR  Error: Request failed with status code 403"},{"timestamp":"2020-12-22T11:24:15.335Z","source":"testjs-d97hd-4k4", "type":"app", "message":"    at settle (/opt/section/node/node_modules/axios/lib/core/settle.js:17:12)"},{"timestamp":"2020-12-22T11:24:15.335Z","source":"testjs-d97hd-4k4", "type":"app","message":"    at endReadableNT (_stream_readable.js:1092:12)"},{"timestamp":"2020-12-22T11:24:15.335Z","source":"testjs-d97hd-4k4", "type":"app","message":"    at IncomingMessage.emit (events.js:187:15)"},{"timestamp":"2020-12-22T11:24:15.335Z","source":"testjs-d97hd-4k4", "type":"app","message":"    at createError (/opt/section/node/node_modules/axios/lib/core/createError.js:16:15)"},{"timestamp":"2020-12-22T11:24:15.335Z","source":"testjs-d97hd-4k4", "type":"app","message":"    at IncomingMessage.handleStreamEnd (/opt/section/node/node_modules/axios/lib/adapters/http.js:236:11)"},{"timestamp":"2020-12-22T11:24:49.335Z","source":"testjs-d97hd-4k4", "type":"access","message":"method=GET path=/_next/static/chunks/main.js time_taken_ms=18 status=200 bytes_sent=7405 host=www.varnishdemo.com section_io_id=e47a44a74364e984eb4b75c3c0ec908"},{"timestamp":"2020-12-22T11:24:51.335Z","source":"testjs-d97hd-4k4", "type":"access","message":"method=GET path=/ time_taken_ms=1800 status=200 bytes_sent=1024 host=www.varnishdemo.com section_io_id=52ad30f0f33a6ed11b5693c63e5b3ea4"},{"timestamp":"2020-12-22T11:24:52.335Z","source":"testjs-d97hd-4k4", "type":"access","message":"method=GET path=/_next/static/chunks/framework.js time_taken_ms=11 status=200 bytes_sent=3405 host=www.varnishdemo.com section_io_id=981d39903a281a5f896517a364b82a80"}]}}`)
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		return al, err
	}

	al = responseBody.Data.Logs
	return al, nil
}
