package auth

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/zalando/go-keyring"
)

var (
	// KeyringService is the service the credential belongs to
	KeyringService = "sectionctl"
	// KeyringUser is the user of the credential
	KeyringUser = "aperture.section.io"
)

// Setup ensures authentication is set up
func Setup(endpoint string) (err error) {
	if !IsCredentialRecorded(KeyringService, KeyringUser) {
		fmt.Printf("No API credentials recorded.\n\n")
		fmt.Printf("Let's get you authenticated to the Section API!\n\n")

		t, err := PromptForCredential(os.Stdin, os.Stdout)
		if err != nil {
			return fmt.Errorf("error when prompting for credential: %w", err)
		}

		err = WriteCredential(endpoint, t)
		if err != nil {
			return fmt.Errorf("unable to save credential: %s", err)
		}

	}

	return err
}

// IsCredentialRecorded returns if API credentials have been recorded
func IsCredentialRecorded(s, u string) bool {
	token, err := keyring.Get(s, u)
	if err != nil {
		return false
	}
	return len(token) > 0
}

// PromptForCredential interactively prompts the user for a credential to authenticate to the Section API
func PromptForCredential(in io.Reader, out io.Writer) (token string, err error) {
	fmt.Fprintf(out, "Token: ")

	reader := bufio.NewReader(in)
	token, err = reader.ReadString('\n')

	if err != nil {
		return token, fmt.Errorf("unable to read your response: %w", err)
	}
	token = strings.Replace(token, "\n", "", -1)
	token = strings.Replace(token, "\r", "", -1) // convert CRLF to LF

	return token, err
}

// WriteCredential saves Section API credentials to a persistent store
func WriteCredential(endpoint, token string) (err error) {
	err = keyring.Set(KeyringService, endpoint, token)
	return err
}

// GetCredential returns a token for authenticating to the Section API
func GetCredential(endpoint string) (token string, err error) {
	token, err = keyring.Get(KeyringService, endpoint)
	return token, err
}
