package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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
}

// Domain represents an applications environments' domains
type Domain struct {
	Name     string `json:"name"`
	ZoneName string `json:"zoneName"`
	CNAME    string `json:"cname"`
	Mode     string `json:"mode"`
}

// Application returns detailed information about a given application.
func Application(accountID int, applicationID int) (a App, err error) {
	u, err := url.Parse(BaseURL)
	if err != nil {
		return a, err
	}
	u.Path += fmt.Sprintf("/account/%d/application/%d", accountID, applicationID)

	resp, err := request(http.MethodGet, u.String(), nil)
	if err != nil {
		return a, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return a, fmt.Errorf("request failed with status %s", resp.Status)
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

	return a, err
}

// ApplicationEnvironments returns environment information for a given application.
func ApplicationEnvironments(accountID int, applicationID int) (es []Environment, err error) {
	u, err := url.Parse(BaseURL)
	if err != nil {
		return es, err
	}
	u.Path += fmt.Sprintf("/account/%d/application/%d/environment", accountID, applicationID)

	resp, err := request(http.MethodGet, u.String(), nil)
	if err != nil {
		return es, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return es, fmt.Errorf("request failed with status %s", resp.Status)
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

// Applications returns a list of applications on a given account.
func Applications(accountID int) (as []App, err error) {
	u, err := url.Parse(BaseURL)
	if err != nil {
		log.Fatal(err)
	}
	u.Path += fmt.Sprintf("/account/%d/application", accountID)

	resp, err := request(http.MethodGet, u.String(), nil)
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
