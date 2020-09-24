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
	// CredentialPath is the path on disk to where credential is recorded
	CredentialPath string
	// TTY is the terminal for reading credentials from users
	TTY *os.File
)

func init() {
	// Set tty for normal interactive use
	TTY = os.Stdin

	// Set CredentialPath for normal interactive use
	usr, err := user.Current()
	if err != nil {
		return
	}
	CredentialPath = filepath.Join(usr.HomeDir, ".config", "section", "netrc")
}

// GetBasicAuth returns credentials for authenticating to the Section API
func GetBasicAuth() (u, p string, err error) {
	n, err := netrc.Parse(CredentialPath)
	if err != nil {
		return u, p, err
	}
	if n.Machine("aperture.section.io") == nil {
		return u, p, fmt.Errorf("invalid credentials file at %s", CredentialPath)
	}
	u = n.Machine("aperture.section.io").Get("login")
	p = n.Machine("aperture.section.io").Get("password")
	return u, p, err
}

// Setup ensures authentication is set up
func Setup() (err error) {
	if !IsCredentialRecorded() {
		Printf(TTY, "No API credentials recorded.\n\n")
		Printf(TTY, "Let's get you authenticated to the Section API!\n\n")

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
	n, err := netrc.Parse(CredentialPath)
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

	restoreTerminal := func() {}
	if TTY == os.Stdin {
		oldState, err := terminal.MakeRaw(int(TTY.Fd()))
		if err != nil {
			return m, u, p, fmt.Errorf("unable to set up terminal: %s", err)
		}
		restoreTerminal = func() {
			err = terminal.Restore(int(TTY.Fd()), oldState)
			if err != nil {
				fmt.Printf("unable to restore terminal: %s\n", err)
				os.Exit(1)
			}
			fmt.Println()
		}
	}

	t := terminal.NewTerminal(TTY, "")

	Printf(TTY, "Username: ")
	u, err = t.ReadLine()
	if err != nil {
		restoreTerminal()
		return m, u, p, fmt.Errorf("unable to read username: %s", err)
	}

	Printf(TTY, "Password: ")
	p, err = t.ReadPassword("")
	if err != nil {
		restoreTerminal()
		return m, u, p, fmt.Errorf("unable to read password: %s", err)
	}

	restoreTerminal()
	return m, u, p, err
}

// WriteCredential saves Section API credentials to disk
func WriteCredential(machine, username, password string) (err error) {
	_, err = os.Stat(CredentialPath)
	if os.IsNotExist(err) {
		file, err := os.Create(CredentialPath)
		if err != nil {
			return err
		}
		file.Close()
	}
	if err := os.Chmod(CredentialPath, 0600); err != nil {
		return err
	}

	n, err := netrc.Parse(CredentialPath)
	if err != nil {
		return err
	}
	n.AddMachine(machine, username, password)
	err = n.Save()
	return err
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
