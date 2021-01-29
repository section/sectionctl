package commands

import (
	"fmt"
	"os"
)

// InitCmd creates the necessary files required to deploy an app (server.conf)
type InitCmd struct {
	StackName string `default:"nodejs-basic" optional short:"s" help:"Name of stack to deploy. Default is nodejs-basic"`
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
	}
	b1 := make([]byte, 100)
	fStream, err := checkServConf.Read(b1)
	fmt.Printf("%d", fStream)
	return err
}
