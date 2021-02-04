package commands

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/section/sectionctl/api"
)

// AppsCmd manages apps on Section
type AppsCmd struct {
	List   AppsListCmd   `cmd help:"List apps on Section." default:"1"`
	Info   AppsInfoCmd   `cmd help:"Show detailed app information on Section."`
	Create AppsCreateCmd `cmd help:"Create new app on Section."`
	Delete AppsDeleteCmd `cmd help:"Delete an existing app on Section."`
	Init   AppsInitCmd   `cmd help:"Initilize your project for deployment"`
}

// AppsListCmd handles listing apps running on Section
type AppsListCmd struct {
	AccountID int `required short:"a"`
}

// NewTable returns a table with sectionctl standard formatting
func NewTable(out io.Writer) (t *tablewriter.Table) {
	t = tablewriter.NewWriter(out)
	t.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	t.SetCenterSeparator("|")
	t.SetAlignment(tablewriter.ALIGN_LEFT)
	return t
}

// Run executes the command
func (c *AppsListCmd) Run() (err error) {
	s := NewSpinner("Looking up apps")
	s.Start()

	apps, err := api.Applications(c.AccountID)
	s.Stop()
	if err != nil {
		return err
	}

	table := NewTable(os.Stdout)
	table.SetHeader([]string{"App ID", "App Name"})

	for _, a := range apps {
		r := []string{strconv.Itoa(a.ID), a.ApplicationName}
		table.Append(r)
	}

	table.Render()
	return err
}

// AppsInfoCmd shows detailed information on an app running on Section
type AppsInfoCmd struct {
	AccountID int `required short:"a"`
	AppID     int `required short:"i"`
}

// Run executes the command
func (c *AppsInfoCmd) Run() (err error) {
	s := NewSpinner("Looking up app info")
	s.Start()

	app, err := api.Application(c.AccountID, c.AppID)
	s.Stop()
	if err != nil {
		return err
	}

	fmt.Printf("🌎🌏🌍\n")
	fmt.Printf("App Name: %s\n", app.ApplicationName)
	fmt.Printf("App ID: %d\n", app.ID)
	fmt.Printf("Environment count: %d\n", len(app.Environments))

	for i, env := range app.Environments {
		fmt.Printf("\n-----------------\n\n")
		fmt.Printf("Environment #%d: %s (ID:%d)\n\n", i+1, env.EnvironmentName, env.ID)
		fmt.Printf("💬 Domains (%d total)\n", len(env.Domains))

		for _, dom := range env.Domains {
			fmt.Println()

			table := NewTable(os.Stdout)
			table.SetHeader([]string{"Attribute", "Value"})
			table.SetAutoMergeCells(true)
			r := [][]string{
				[]string{"Domain name", dom.Name},
				[]string{"Zone name", dom.ZoneName},
				[]string{"CNAME", dom.CNAME},
				[]string{"Mode", dom.Mode},
			}
			table.AppendBulk(r)
			table.Render()
		}

		fmt.Println()
		mod := "modules"
		if len(env.Stack) == 1 {
			mod = "module"
		}
		fmt.Printf("🥞 Stack (%d %s total)\n", len(env.Stack), mod)
		fmt.Println()

		table := NewTable(os.Stdout)
		table.SetHeader([]string{"Name", "Image"})
		table.SetAutoMergeCells(true)
		for _, p := range env.Stack {
			r := []string{p.Name, p.Image}
			table.Append(r)
		}
		table.Render()
	}

	fmt.Println()

	return err
}

// AppsCreateCmd handles creating apps on Section
type AppsCreateCmd struct {
	AccountID int    `required short:"a" help:"ID of account to create the app under"`
	Hostname  string `required short:"d" help:"FQDN the app can be accessed at"`
	Origin    string `required short:"o" help:"URL to fetch the origin"`
	StackName string `required short:"s" help:"Name of stack to deploy"`
}

// Run executes the command
func (c *AppsCreateCmd) Run() (err error) {
	s := NewSpinner(fmt.Sprintf("Creating new app %s", c.Hostname))
	s.Start()

	api.Timeout = 120 * time.Second // this specific request can take a long time
	r, err := api.ApplicationCreate(c.AccountID, c.Hostname, c.Origin, c.StackName)
	s.Stop()
	if err != nil {
		return err
	}

	fmt.Printf("\nSuccess: created app '%s' with id '%d'\n", r.ApplicationName, r.ID)

	return err
}

// AppsDeleteCmd handles deleting apps on Section
type AppsDeleteCmd struct {
	AccountID int `required short:"a" help:"ID of account the app belongs to"`
	AppID     int `required short:"i" help:"ID of the app to delete"`
}

// Run executes the command
func (c *AppsDeleteCmd) Run() (err error) {
	s := NewSpinner(fmt.Sprintf("Deleting app with id '%d'", c.AppID))
	s.Start()

	api.Timeout = 120 * time.Second // this specific request can take a long time
	_, err = api.ApplicationDelete(c.AccountID, c.AppID)
	s.Stop()
	if err != nil {
		return err
	}

	fmt.Printf("\nSuccess: deleted app with id '%d'\n", c.AppID)

	return err
}

// AppsInitCmd creates and validates server.conf and package.json to prepare an app for deployment
type AppsInitCmd struct {
	StackName string `optional default:"nodejs-basic" short:"s" help:"Name of stack to deploy. Default is nodejs-basic"`
	Force     bool   `optional short:"f" help:"Resets deployment specific files to their default configuration"`
}

func (c *AppsInitCmd) buildServerConf() []byte {
	return []byte(
		`location / {
	proxy_set_header X-Forwarded-For $http_x_forwarded_for;
	proxy_set_header X-Forwarded-Proto $http_x_forwarded_proto;
	proxy_set_header Host $host;
	include /etc/nginx/section.module/node.conf;
}

location ~ "/next-proxy-hop/" {
	proxy_set_header X-Forwarded-For $http_x_forwarded_for;
	proxy_set_header X-Forwarded-Proto $http_x_forwarded_proto;
	proxy_set_header Host $host;
	proxy_pass http://next-hop;
}`)
}

// Run executes the command
func (c *AppsInitCmd) Run() (err error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	switch c.StackName {
	case "nodejs-basic":
		err := c.InitializeNodeBasicApp(stdout, stderr)
		if err != nil {
			panic(err)
		}
	default:
		log.Printf("[ERROR]: Stack name %s does not have an initialization defined\n", c.StackName)
	}
	return err
}

// InitializeNodeBasicApp initializes a basic node app.
func (c *AppsInitCmd) InitializeNodeBasicApp(stdout, stderr bytes.Buffer) (err error) {
	if c.Force {
		log.Println("[INFO] Removing old versions of server.conf and package.json")
		err1 := os.Remove("package.json")
		err2 := os.Remove("server.conf")
		if err1 != nil || err2 != nil {
			log.Println("[ERROR] unable to remove files, perhaps they do not exist?")
		} else {
			log.Println("[DEBUG] Files successfully removed")
		}
	}
	log.Println("[DEBUG] Checking to see if server.conf exists")
	checkServConf, err := os.Open("server.conf")
	if err != nil {
		log.Println("[WARN] server.conf does not exist. Creating server.conf")
		f, err := os.Create("server.conf")
		if err != nil {
			panic(err)
		}
		b := c.buildServerConf()
		f.Write(b)
		defer f.Close()
	} else {
		log.Println("[INFO] Validating server.conf")
		fileinfo, err := checkServConf.Stat()
		if err != nil {
			panic(err)
		}
		buf := make([]byte, fileinfo.Size())
		_, err = checkServConf.Read(buf)
		if err != nil {
			panic(err)
		}
		fStr := string(buf)
		if !strings.Contains(fStr, "location / {") {
			log.Println("[WARN] default location unspecified. Edit or delete server.conf and rerun this command")
		}
	}
	defer checkServConf.Close()
	log.Println("[DEBUG] Checking to see if package.json exists")
	checkPkgJSON, err := os.Open("package.json")
	if err != nil {
		log.Println("[WARN] package.json does not exist. Creating package.json")
		cmd := exec.Command("npm", "init", "-y")
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			log.Println("[ERROR] There was an error creating package.json. Is node installed?")
			panic(err)
		} else {
			log.Println("[INFO] package.json created")
		}
	}
	defer checkPkgJSON.Close()
	validPkgJSON, err := os.OpenFile("package.json", os.O_RDWR, 0777)
	if err != nil {
		panic(err)
	}
	log.Println("[INFO] Validating package.json")
	finfo, err := validPkgJSON.Stat()
	if err != nil {
		panic(err)
	}
	buf := make([]byte, finfo.Size())
	_, err = validPkgJSON.Read(buf)
	if err != nil {
		panic(err)
	}
	fStr := string(buf)
	if !strings.Contains(fStr, `"start": "`) {
		replace := `"scripts": {
    "start": "node YOUR_SERVER_HERE.js",`
		fStr = strings.Replace(fStr, `"scripts": {`, replace, 1)
		err = validPkgJSON.Truncate(0)
		if err != nil {
			panic(err)
		}
		_, err = validPkgJSON.WriteString(fStr)
		if err != nil {
			log.Println("[ERROR] unable to add start script placeholder")
		}
	}
	if strings.Contains(fStr, `YOUR_SERVER_HERE.js`) {
		log.Println("[ERROR] start script is required. Please edit the placeholder in package.json")
	}
	defer validPkgJSON.Close()
	return err
}
