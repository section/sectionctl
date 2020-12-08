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
func ApplicationStatus(accountID int, applicationID int) (as []AppStatus, err error) {
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
		"moduleName":    "nodejs",
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
