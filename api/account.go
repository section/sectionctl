package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sort"
)

// Account represents an account on Section
type Account struct {
	ID          int    `json:"id"`
	Href        string `json:"href"`
	AccountName string `json:"account_name"`
	IsAdmin     bool   `json:"is_admin"`
	BillingUser int    `json:"billing_user"`
	Requires2FA bool   `json:"requires_2fa"`
}

// Accounts returns a list of account the current user has access to.
func Accounts() (as []Account, err error) {
	u := BaseURL()
	u.Path += "/account"

	resp, err := request(http.MethodGet, u, nil)
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
	sort.Slice(as, func(i, j int) bool {
		return as[i].ID < as[j].ID
	})
	return as, err
}
