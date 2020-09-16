package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

// Application represents an application deployed on Section
type Application struct {
	ID              int    `json:"id"`
	Href            string `json:"href"`
	ApplicationName string `json:"application_name"`
}

// Applications returns a list of applications on a given account.
func Applications(accountID int) (as []Application, err error) {
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
