package api

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// Stack represents a deployable application stack on Section
type Stack struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

// Stacks returns the available deployable stacks
func Stacks() (s []Stack, err error) {
	ur := BaseURL()
	ur.Path += "/stack"

	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	resp, err := request(ctx, http.MethodGet, ur, nil)
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
