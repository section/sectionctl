package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// InitCmd creates the necessary files required to deploy an app (server.conf)
type InitCmd struct {
	StackName string `optional default:"nodejs-basic" short:"s" help:"Name of stack to deploy. Default is nodejs-basic"`
}

// Copy and paste of server.conf
func buildServerConf() []byte {
	return []byte(
		`location / {
	proxy_set_header X-Forwarded-For $http_x_forwarded_for;
	proxy_set_header X-Forwarded-Proto $http_x_forwarded_proto;
	proxy_set_header Host $host;
	include /etc/nginx/section.module/node.conf;

location ~ "/next-proxy-hop/" {
	proxy_set_header X-Forwarded-For $http_x_forwarded_for;
	proxy_set_header X-Forwarded-Proto $http_x_forwarded_proto;
	proxy_set_header Host $host;
	proxy_pass http://next-hop;
}`)
}

// Run executes the command
func (c *InitCmd) Run() (err error) {
	if c.StackName == "nodejs-basic" {
		fmt.Println("Checking to see if server.conf exists")
		checkServConf, err := os.Open("server.conf")
		if err != nil {
			fmt.Println("WARN: server.conf does not exist. Creating server.conf")
			f, err := os.Create("server.conf")
			if err != nil {
				panic(err)
			}
			b := buildServerConf()
			f.Write(b)
			defer f.Close()
		} else {
			fmt.Println("Validating server.conf")
			fileinfo, err := checkServConf.Stat()
			if err != nil {
				panic(err)
			}
			buff := make([]byte, fileinfo.Size())
			_, err = checkServConf.Read(buff)
			if err != nil {
				panic(err)
			}
			fileString := string(buff)
			if !strings.Contains(fileString, "location / {") {
				fmt.Println("WARN: default location unspecified. Edit or delete server.conf and rerun this command")
			}
		}
		defer checkServConf.Close()
		fmt.Println("Checking to see if package.json exists")
		checkPkgJSON, err := os.Open("package.json")
		if err != nil {
			fmt.Println("WARN: package.json does not exist. Creating package.json")
			cmd := exec.Command("npm", "init", "-y")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				fmt.Println("FATAL: There was an error creating package.json. Is node installed?")
			} else {
				fmt.Println("package.json created")
				fmt.Println("WARN: package.json does not have a 'start' script. This is required")
			}
		} else {
			fmt.Println("Validating package.json")
			fileinfo, err := checkPkgJSON.Stat()
			if err != nil {
				panic(err)
			}
			buff := make([]byte, fileinfo.Size())
			_, err = checkPkgJSON.Read(buff)
			if err != nil {
				panic(err)
			}
			fileString := string(buff)
			if !strings.Contains(fileString, "start") {
				fmt.Println("WARN: start script is required. Please add one to your package.json")
			}
		}
		defer checkPkgJSON.Close()
	} else {
		fmt.Printf("FATAL: Stack name %s does not have an initialization defined\n", c.StackName)
	}
	return err
}
