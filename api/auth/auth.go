package auth

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/jdxcode/netrc"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	// consentPath is the path on disk to where credential is recorded
	credentialPath string
	// out is the buffer to write feedback to users
	out io.Writer
	// in is the buffer to read feedback from users
	in *os.File
)

func init() {
	// Set in + out for normal interactive use
	in = os.Stdin
	out = os.Stdout
}

// GetBasicAuth returns credentials for authenticating to the Section API
func GetBasicAuth() (u, p string, err error) {
	n, err := netrc.Parse(credentialPath)
	if err != nil {
		return u, p, err
	}
	if n.Machine("aperture.section.io") == nil {
		return u, p, fmt.Errorf("invalid credentials file at %s", credentialPath)
	}
	u = n.Machine("aperture.section.io").Get("login")
	p = n.Machine("aperture.section.io").Get("password")
	return u, p, err
}

// Setup ensures authentication is set up
func Setup() (err error) {
	if !IsCredentialRecorded() {
		Printf("No API credentials recorded.\n\n")
		Printf("Let's get you authenticated to the Section API!\n\n")
		err = PromptForAndSaveCredential()
	}

	return err
}

// IsCredentialRecorded returns if API credentials have been recorded
func IsCredentialRecorded() bool {
	if len(credentialPath) == 0 {
		usr, err := user.Current()
		if err != nil {
			return false
		}
		credentialPath = filepath.Join(usr.HomeDir, ".config", "section", "netrc")
	}

	n, err := netrc.Parse(credentialPath)
	if err != nil {
		return false
	}
	if n.Machine("aperture.section.io") == nil {
		return false
	}

	u := n.Machine("aperture.section.io").Get("login")
	p := n.Machine("aperture.section.io").Get("password")
	return (len(u) > 0 && len(p) > 0)
}

func readInput() (text string, err error) {
	reader := bufio.NewReader(in)
	text, err = reader.ReadString('\n')
	if err != nil {
		return text, fmt.Errorf("unable to read your response: %s", err)
	}
	text = strings.Replace(text, "\n", "", -1)
	text = strings.Replace(text, "\r", "", -1) // convert CRLF to LF
	return text, err
}

func readPasswordInput() (text string, err error) {
	fd := int(in.Fd())
	password, err := terminal.ReadPassword(fd)
	if err != nil {
		return text, fmt.Errorf("unable to read your response: %s", err)
	}
	return string(password), err
}

// PromptForAndSaveCredential interactively prompts the user for a credential to authenticate to the Section API
func PromptForAndSaveCredential() (err error) {
	machine := "aperture.section.io"
	fmt.Println("machine:", machine)

	Printf("Username: ")
	username, err := readInput()
	if err != nil {
		return fmt.Errorf("unable to read username: %s", err)
	}
	fmt.Println("username:", username)

	Printf("Password: ")
	password, err := readPasswordInput()
	if err != nil {
		return fmt.Errorf("unable to read password: %s", err)
	}
	fmt.Println("password:", password)

	err = WriteCredential(machine, username, password)
	if err != nil {
		fmt.Errorf("unable to save credential: %s", err)
		return err
	}
	return err
}

// WriteCredential saves Section API credentials to disk
func WriteCredential(machine, username, password string) (err error) {
	_, err = os.Stat(credentialPath)
	if os.IsNotExist(err) {
		file, err := os.Create(credentialPath)
		if err != nil {
			return err
		}
		file.Close()
	}
	if err := os.Chmod(credentialPath, 0600); err != nil {
		return err
	}

	n, err := netrc.Parse(credentialPath)
	if err != nil {
		return err
	}
	n.AddMachine(machine, username, password)
	err = n.Save()
	return err
}

// Println formats using the default formats for its operands and writes to output.
// Output can be overridden for testing purposes by setting: analytics.out
// It returns the number of bytes written and any write error encountered.
func Println(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(out, a...)
}

// Printf formats according to a format specifier and writes to standard output.
// Output can be overridden for testing purposes by setting: analytics.out
// It returns the number of bytes written and any write error encountered.
func Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(out, format, a...)
}
