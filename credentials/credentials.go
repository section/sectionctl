package credentials

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
func Setup(endpoint string) (token string, err error) {
	if !IsCredentialRecorded(KeyringService, KeyringUser) {
		fmt.Printf("No API credentials recorded.\n\n")
		fmt.Printf("Let's get you authenticated to the Section API!\n\n")

		_, err := PromptAndWrite(os.Stdin, os.Stdout, endpoint)
		if err != nil {
			return token, err
		}
	}

	return Read(endpoint)
}

// PromptAndWrite prompts for a credential then writes it to a store
func PromptAndWrite(in io.Reader, out io.Writer, endpoint string) (token string, err error) {
	token, err = Prompt(in, out)
	if err != nil {
		return token, fmt.Errorf("unable to prompt for credential: %w", err)
	}

	err = Write(endpoint, token)
	if err != nil {
		return token, fmt.Errorf("unable to save credential: %s", err)
	}

	return token, err
}

// IsCredentialRecorded returns if API credentials have been recorded
func IsCredentialRecorded(s, u string) bool {
	token, err := keyring.Get(s, u)
	if err != nil {
		return false
	}
	return len(token) > 0
}

// Prompt interactively prompts the user for a credential to authenticate to the Section API
func Prompt(in io.Reader, out io.Writer) (token string, err error) {
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

// Write saves Section API credentials to a persistent store
func Write(endpoint, token string) (err error) {
	err = keyring.Set(KeyringService, endpoint, token)
	return err
}

// Read returns a token for authenticating to the Section API
func Read(endpoint string) (token string, err error) {
	token, err = keyring.Get(KeyringService, endpoint)
	return token, err
}

// Delete deletes a previously stored credential for the Section API
func Delete(endpoint string) error {
	return keyring.Delete(KeyringService, endpoint)
}
