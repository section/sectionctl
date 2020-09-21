package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/crypto/ssh/terminal"
)

func stdin() {
	fmt.Printf("Password: ")
	passwd, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}
	fmt.Printf("\nPassword: %s\n", passwd)
}

func term() {
	c := exec.Command("echo", "hello\ns3cr3t\n")
	tty, err := pty.Start(c)
	if err != nil {
		log.Fatal(err)
	}
	defer tty.Close()
	fmt.Printf("%+v\n", tty)

	fd := int(tty.Fd())
	fmt.Println(fd) // 3

	fmt.Println("is terminal", terminal.IsTerminal(fd))

	//bufio.NewBuffer()
	fmt.Printf("Password: ")
	//fd := int(syscall.Stdin)
	passwd, err := terminal.ReadPassword(fd)
	if err != nil {
		log.Fatalf("%s\n", err)
	}
	fmt.Printf("\nPassword: %s\n", passwd)
}

// CloseHandler resets the terminal on exit
func CloseHandler(tty *os.File, state *terminal.State) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("restoring state")
		err := terminal.Restore(int(tty.Fd()), state)
		if err != nil {
			log.Fatalf("err: %s\n", err)
		}
		os.Exit(0)
	}()
}

var exitHandlers []exitHandler

type exitHandler struct {
	Name    string
	handler func()
}

// Exit gracefully exits the process, and runs any handlers
func Exit(c int) {
	for _, h := range exitHandlers {
		fmt.Printf("Running exit handler: %s", h.Name)
		h.handler()
	}
	fmt.Println("Exiting!")
	os.Exit(c)
}

func main() {
	c := exec.Command("echo", "spoons\nofdoom\n")
	tty, err := pty.Start(c)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		Exit(1)
	}
	defer tty.Close()
	fmt.Printf("tty: %+v\n", tty)

	fd := int(tty.Fd())
	fmt.Printf("fd: %s\n", fd) // 3

	tty = os.Stdin
	fmt.Println("is terminal?", terminal.IsTerminal(fd))

	fmt.Printf("%+v\n%+v\n", tty, os.Stdin)

	var restoreTerminal func()

	t := terminal.NewTerminal(tty, "")

	/*
		for i := 0; i < 10; i++ {
			time.Sleep(1 * time.Second)
			fmt.Println("Waiting...")
		}
	*/

	//t.Write([]byte("Username: "))
	fmt.Printf("Username: ")

	if tty == os.Stdin {
		oldState, err := terminal.MakeRaw(int(tty.Fd()))
		if err != nil {
			fmt.Printf("error: %s\n", err)
			Exit(1)
		}
		restoreTerminal = func() {
			err = terminal.Restore(int(tty.Fd()), oldState)
			if err != nil {
				fmt.Printf("error: %s\n", err)
				Exit(1)
			}
			fmt.Println()
		}
		handler := exitHandler{Name: "Restore Terminal", handler: restoreTerminal}
		exitHandlers = append(exitHandlers, handler)
	}

	username, err := t.ReadLine()

	if tty == os.Stdin {
		restoreTerminal()
	}

	if err != nil {
		fmt.Println()
		fmt.Printf("error: %s\n", err)
		Exit(1)
	}

	//t.Write([]byte("Password: "))
	fmt.Printf("Password: ")

	if tty == os.Stdin {
		oldState, err := terminal.MakeRaw(int(tty.Fd()))
		if err != nil {
			fmt.Printf("error: %s\n", err)
			Exit(1)
		}
		restoreTerminal = func() {
			err = terminal.Restore(int(tty.Fd()), oldState)
			if err != nil {
				fmt.Printf("error: %s\n", err)
				Exit(1)
			}
			fmt.Println()
		}
		handler := exitHandler{Name: "Restore Terminal", handler: restoreTerminal}
		exitHandlers = append(exitHandlers, handler)
	}

	password, err := t.ReadPassword("")

	if tty == os.Stdin {
		restoreTerminal()
	}

	if err != nil {
		fmt.Println()
		fmt.Printf("error: %s\n", err)
		Exit(1)
	}

	fmt.Printf("%s:%s\n", username, password)

	creds := map[string]string{"username": username, "password": password}

	j, err := json.Marshal(creds)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		Exit(1)
	}
	err = ioutil.WriteFile("password.json", j, 0600)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		Exit(1)
	}

	Exit(0)
}
