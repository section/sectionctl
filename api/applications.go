package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
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

	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	resp, err := request(ctx, http.MethodGet, u, nil)
	if err != nil {
		return a, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case 401:
			return a, ErrStatusUnauthorized
		case 403:
			return a, ErrStatusForbidden
		default:
			return a, prettyTxIDError(resp)
		}
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

	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	resp, err := request(ctx, http.MethodGet, u, nil)
	if err != nil {
		return es, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case 401:
			return es, ErrStatusUnauthorized
		case 403:
			return es, ErrStatusForbidden
		default:
			return es, prettyTxIDError(resp)
		}
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

	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	resp, err := request(ctx, http.MethodGet, u, nil)
	if err != nil {
		return s, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case 401:
			return s, ErrStatusUnauthorized
		case 403:
			return s, ErrStatusForbidden
		default:
			return s, prettyTxIDError(resp)
		}
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

	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	resp, err := request(ctx, http.MethodGet, u, nil)
	if err != nil {
		return as, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case 401:
			return as, ErrStatusUnauthorized
		case 403:
			return as, ErrStatusForbidden
		default:
			return as, prettyTxIDError(resp)
		}
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
	log.Debug().Msg(fmt.Sprintf(" JSON payload: %s\n", b))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute) // because these requests can take a long time to complete on Section's side
	defer cancel()
	headers := map[string][]string{"filepath": {filePath}}

	resp, err := request(ctx, http.MethodPatch, u, bytes.NewBuffer(b), headers)
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

	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()
	data, err := json.Marshal(requestData)
	resp, err := request(ctx, http.MethodPost, u, bytes.NewBuffer(data))
	if err != nil {
		return as, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case 401:
			return as, ErrStatusUnauthorized
		case 403:
			return as, ErrStatusForbidden
		default:
			return as, prettyTxIDError(resp)
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return as, fmt.Errorf("could not read response body: %s", err)
	}

	log.Debug().Msg(fmt.Sprintf(" RESPONSE: %s\n", string(body)))

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
func ApplicationLogs(accountID int, applicationID int, moduleName string, instanceName string, number int, startTimestampRfc3339 string) (al []AppLogs, err error) {
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
		"length":        number,
		// "startTimestamp": startTimestamp,
		// "endTimestamp": endTimestamp,
	}

	log.Debug().Msg(fmt.Sprintf(" requestData: %v\n", requestData.Variables))

	requestData.Query = "query Logs($moduleName: String!, $environmentId: Int!, $instanceName: String, $length: Int){logs(moduleName:$moduleName, environmentId:$environmentId, instanceName:$instanceName, length:$length){timestamp instanceName message type}}"

	if startTimestampRfc3339 != "" {
		requestData.Variables["startTimestampRfc3339"] = startTimestampRfc3339
		requestData.Query = "query Logs($moduleName: String!, $environmentId: Int!, $instanceName: String, $length: Int, $startTimestampRfc3339: String){logs(moduleName:$moduleName, environmentId:$environmentId, instanceName:$instanceName, length:$length, startTimestampRfc3339:$startTimestampRfc3339){timestamp instanceName message type}}"
	}

	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()
	data, err := json.Marshal(requestData)
	resp, err := request(ctx, http.MethodPost, u, bytes.NewBuffer(data))
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

	log.Debug().Msg(fmt.Sprintf(" RESPONSE: %s\n", string(body)))

	var responseBody struct {
		Data struct {
			Logs []AppLogs `json:"logs"`
		} `json:"data"`
	}

	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		return al, err
	}

	al = responseBody.Data.Logs
	return al, nil
}

// ApplicationCreateResponse represents an API response for application create requests
type ApplicationCreateResponse struct {
	ID              int      `json:"id"`
	Href            string   `json:"href"`
	ApplicationName string   `json:"application_name"`
	PathPrefix      string   `json:"path_prefix"`
	PathPrefixes    []string `json:"path_prefixes"`
	Message         string   `json:"message"` // for errors
}

// ApplicationCreate creates an application on the Section platform.
func ApplicationCreate(accountID int, hostname, origin, stackName string) (r ApplicationCreateResponse, err error) {
	u := BaseURL()
	u.Path += fmt.Sprintf("/account/%d/application/create", accountID)

	appCreateReq := struct {
		Hostname  string `json:"hostname"`
		Origin    string `json:"origin"`
		StackName string `json:"stackName"`
	}{
		hostname,
		origin,
		stackName,
	}

	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	data, err := json.Marshal(appCreateReq)
	resp, err := request(ctx, http.MethodPost, u, bytes.NewBuffer(data))
	if err != nil {
		return r, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return r, err
	}

	err = json.Unmarshal(body, &r)
	if err != nil {
		return r, err
	}

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case 401:
			return r, ErrStatusUnauthorized
		case 403:
			return r, disambiguateForbiddenApplicationCreate(hostname, r)
		default:
			return r, prettyTxIDError(resp)
		}
	}

	return r, err
}

var (
	// ErrApplicationAlreadyCreated indicates an app has already been created with that FQDN.
	ErrApplicationAlreadyCreated = errors.New("an application has already been created with that domain name")
	// ErrSystemLimitExceeded indicates you have hit a soft limit on your account. Contact Section support to increase the limit.
	ErrSystemLimitExceeded = errors.New("system limit exceeded")
)

// The Section API returns multiple distinct errors with a HTTP 403 when
// creating applications.
//
// This function attempts to disambiguate those responses and return usable errors.
func disambiguateForbiddenApplicationCreate(hostname string, r ApplicationCreateResponse) (err error) {
	switch r.Message {
	case "An application has already been created with domain name " + hostname:
		return fmt.Errorf("%s - %w", hostname, ErrApplicationAlreadyCreated)
	case "System limit exceeded. Contact support to increase this limit.":
		return fmt.Errorf("%w: contact support to increase this limit", ErrSystemLimitExceeded)
	default:
		return ErrStatusForbidden
	}
}

var (
	// ErrStatusBadRequest indicates the request was malformed
	ErrStatusBadRequest = errors.New("malformed request")
	// ErrStatusNotFound indicates the requested resource was not found
	ErrStatusNotFound = errors.New("application not found")
	// ErrStatusInternalServerError indicates the server errored when handling your request
	ErrStatusInternalServerError = errors.New("unexpected condition when handling request")
)

// ApplicationDeleteResponse respresents an API response for application delete requests
type ApplicationDeleteResponse struct {
	Message string `json:"message"` // for errors
}

// ApplicationDelete deletes an application on the Section platform.
func ApplicationDelete(accountID, appID int) (r ApplicationDeleteResponse, err error) {
	u := BaseURL()
	u.Path += fmt.Sprintf("/account/%d/application/%d", accountID, appID)

	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	resp, err := request(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return r, fmt.Errorf("unable to perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return r, fmt.Errorf("unable to read response body: %w", err)
		}

		err = json.Unmarshal(body, &r)
		if err != nil {
			return r, fmt.Errorf("unable to unmarshal JSON: %w", err)
		}

		switch resp.StatusCode {
		case http.StatusBadRequest:
			return r, fmt.Errorf("%s: %w", r.Message, ErrStatusBadRequest)
		case http.StatusUnauthorized:
			return r, ErrStatusUnauthorized
		case http.StatusForbidden:
			return r, ErrStatusForbidden
		case http.StatusNotFound:
			return r, ErrStatusNotFound
		case http.StatusInternalServerError:
			return r, fmt.Errorf("error occurred during deletion: %s: %w", r.Message, ErrStatusInternalServerError)
		default:
			return r, prettyTxIDError(resp)
		}
	}

	return r, err
}
