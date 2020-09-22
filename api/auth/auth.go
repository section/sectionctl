package auth

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/jdxcode/netrc"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	// consentPath is the path on disk to where credential is recorded
	credentialPath string
	// tty is the terminal for reading credentials from users
	tty *os.File
)

func init() {
	// Set tty for normal interactive use
	tty = os.Stdin
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
		Printf(tty, "No API credentials recorded.\n\n")
		Printf(tty, "Let's get you authenticated to the Section API!\n\n")

		m, u, p, err := PromptForCredential()
		if err != nil {
			return fmt.Errorf("error when prompting for credential: %s", err)
		}

		err = WriteCredential(m, u, p)
		if err != nil {
			return fmt.Errorf("unable to save credential: %s", err)
		}
		return err
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

// PromptForCredential interactively prompts the user for a credential to authenticate to the Section API
func PromptForCredential() (m, u, p string, err error) {
	m = "aperture.section.io"
	//u = "jane@section.example"
	//p = "s3cr3t"

	var restoreTerminal func()
	if tty == os.Stdin {
		oldState, err := terminal.MakeRaw(int(tty.Fd()))
		if err != nil {
			return m, u, p, fmt.Errorf("unable to set up terminal: %s", err)
		}
		restoreTerminal = func() {
			err = terminal.Restore(int(tty.Fd()), oldState)
			if err != nil {
				fmt.Printf("unable to restore terminal: %s\n", err)
				os.Exit(1)
			}
			fmt.Println()
		}
	}

	t := terminal.NewTerminal(tty, "")

	Printf(tty, "Username: ")
	u, err = t.ReadLine()
	if err != nil {
		restoreTerminal()
		return m, u, p, fmt.Errorf("unable to read username: %s", err)
	}

	Printf(tty, "Password: ")
	p, err = t.ReadPassword("")
	if err != nil {
		restoreTerminal()
		return m, u, p, fmt.Errorf("unable to read password: %s", err)
	}

	return m, u, p, err
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

// Printf formats according to a format specifier and writes to an output.
// Output can be overridden for testing purposes by setting: auth.tty
func Printf(tty *os.File, str string, a ...interface{}) {
	s := fmt.Sprintf(str, a...)
	if tty == os.Stdin {
		tty.Write([]byte(s))
	} else {
		fmt.Printf("%s", s)
	}
}
