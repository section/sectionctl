package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

/*
{
	  "issued": true,
	    "message": "The certificate has been renewed",
		  "expiry": "2021-01-26T12:31:38.000Z",
		    "renewFrom": "2020-12-27T12:31:38.000Z"
		}
*/

// RenewCertResponse represents an API response to POST /account/{accountId}/domain/{hostName}/renewCertificate
type RenewCertResponse struct {
	Issued    bool      `json:"issued"`
	Message   string    `json:"message"`
	Expiry    time.Time `json:"expiry"`
	RenewFrom time.Time `json:"renewFrom"`
}

// DomainsRenewCert handles renewing a certificate for a given account and domain.
func DomainsRenewCert(accountID int, hostname string) (r RenewCertResponse, err error) {
	u := BaseURL()
	u.Path += fmt.Sprintf("/account/%d/domain/%s/renewCertificate", accountID, hostname)

	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	resp, err := request(ctx, http.MethodPost, u, nil)
	if err != nil {
		return r, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case 401:
			return r, ErrStatusUnauthorized
		case 403:
			return r, ErrStatusForbidden
		default:
			return r, prettyTxIDError(resp)
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return r, err
	}

	err = json.Unmarshal(body, &r)
	if err != nil {
		return r, err
	}

	return r, err
}
