package commands

import (
	"fmt"
	"os"
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
}
		
location ~ "/notnuxt/" {
	proxy_set_header X-Forwarded-For $http_x_forwarded_for;
	proxy_set_header X-Forwarded-Proto $http_x_forwarded_proto;
	proxy_set_header Host $host;
	proxy_pass http://next-hop;
}`)
}

// Run executes the command
func (c *InitCmd) Run() (err error) {
	if c.StackName == "nodejs-basic" {
		fmt.Printf("Checking to see if server.conf exists\n")
		checkServConf, err := os.Open("server.conf")
		if err != nil {
			fmt.Printf("WARN: server.conf does not exist. Creating server.conf\n")
			f, err := os.Create("server.conf")
			if err != nil {
				panic(err)
			}
			b := buildServerConf()
			f.Write(b)
			defer f.Close()
		} else {
			fmt.Printf("Validating server.conf\n")
			fileinfo, err := checkServConf.Stat()
			if err != nil {
				panic(err)
			}
			b1 := make([]byte, fileinfo.Size())
			checkServConf.Read(b1)
			fileString := string(b1)
			if !strings.Contains(fileString, "location / {") {
				fmt.Println("WARN: default location unspecified. Edit or delete server.conf and rerun this command")
			}
			if !strings.Contains(fileString, "/notnuxt/") {
				fmt.Println("WARN: notnuxt location unspecified. Edit or delete server.conf and rerun this command")
			}
		}
		defer checkServConf.Close()
	} else {
		fmt.Printf("Stack name %s does not have an initialization defined\n", c.StackName)
	}
	return err
}
