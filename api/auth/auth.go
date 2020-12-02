package auth

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jdxcode/netrc"
	"github.com/mitchellh/go-homedir"
	"github.com/zalando/go-keyring"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	// CredentialPath is the path on disk to where credential is recorded
	CredentialPath string
	// TTY is the terminal for reading credentials from users
	TTY *os.File
)

func init() {
	// Set tty for normal interactive use
	TTY = os.Stdin

	// Set CredentialPath for normal interactive use
	dir, err := homedir.Dir()
	if err != nil {
		return
	}
	CredentialPath = filepath.Join(dir, ".config", "section", "netrc")
}

// Setup ensures authentication is set up
func Setup(endpoint string) (err error) {
	if !IsCredentialRecorded() {
		Printf(TTY, "No API credentials recorded.\n\n")
		Printf(TTY, "Let's get you authenticated to the Section API!\n\n")

		u, p, err := PromptForCredential(endpoint)
		if err != nil {
			return fmt.Errorf("error when prompting for credential: %s", err)
		}

		err = WriteCredential(endpoint, u, p)
		if err != nil {
			return fmt.Errorf("unable to save credential: %s", err)
		}
		return err
	}

	return err
}

// IsCredentialRecorded returns if API credentials have been recorded
func IsCredentialRecorded() bool {
	service := "sectionctl"
	user := "local"
	token, err := keyring.Get(service, user)
	if err != nil {
		return false
	}
	return (len(token) > 0)
}

// PromptForCredential interactively prompts the user for a credential to authenticate to the Section API
func PromptForCredential() (u, p string, err error) {
	return u, p, err
}

// WriteCredential saves Section API credentials to disk
func WriteCredential(endpoint, username, password string) (err error) {
	service := "sectionctl"
	err = keyring.Set(service, endpoint, password)
	return err
}

// GetCredential returns Basic Auth credentials for authenticating to the Section API
func GetCredential(endpoint string) (u, p string, err error) {
	service := "sectionctl"
	u = "local"
	p, err = keyring.Get(service, u)
	return u, p, err
}

// Printf formats according to a format specifier and writes to an output.
// Output can be overridden for testing purposes by setting: auth.TTY
func Printf(out *os.File, str string, a ...interface{}) {
	s := fmt.Sprintf(str, a...)
	if out == os.Stdin {
		out.Write([]byte(s))
	} else {
		fmt.Printf("%s", s)
	}
}
