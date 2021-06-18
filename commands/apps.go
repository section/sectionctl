package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/olekukonko/tablewriter"
	"github.com/section/sectionctl/api"
)

// AppsCmd manages apps on Section
type AppsCmd struct {
	List   AppsListCmd   `cmd help:"List apps on Section." default:"1"`
	Info   AppsInfoCmd   `cmd help:"Show detailed app information on Section."`
	Create AppsCreateCmd `cmd help:"Create new app on Section."`
	Delete AppsDeleteCmd `cmd help:"DANGER ZONE. This deletes an existing app on Section."`
	Init   AppsInitCmd   `cmd help:"Initialize your project for deployment."`
	Stacks AppsStacksCmd `cmd help:"See the available stacks to create new apps with."`
}

// AppsListCmd handles listing apps running on Section
type AppsListCmd struct {
	AccountID int `short:"a" help:"Account ID to find apps under"`
}

// NewTable returns a table with sectionctl standard formatting
func NewTable(cli *CLI, out io.Writer) (t *tablewriter.Table) {
	if cli.Quiet {
		out = io.Discard
	}
	t = tablewriter.NewWriter(out)
	t.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
	t.SetCenterSeparator("|")
	t.SetAlignment(tablewriter.ALIGN_LEFT)
	return t
}

// Run executes the command
func (c *AppsListCmd) Run(cli *CLI, logWriters *LogWriters) (err error){
	s := NewSpinner("Looking up apps",logWriters)
	s.Start()

	accounts, err := api.Accounts()
	if err != nil {
		s. Stop()
		log.Error().Err(err).Msg("Unable to look up accounts");
		os.Exit(1)
	}
	s.Stop()
	if c.AccountID != 0{
		newAct := []api.Account{}
		for _, a := range accounts {
			if a.ID == c.AccountID{
				newAct = append(newAct, a)
			}
			accounts = newAct
		}
		if(len(newAct) == 0){
			log.Info().Int("Account ID",c.AccountID).Msg("Unable to find accounts where")
			os.Exit(1)
		}
	}
	fmt.Println()
	fmt.Println()
	for _, acc := range accounts {
		log.Info().Msg(fmt.Sprint(HiWhite("Account #"),HiWhite(strconv.Itoa(acc.ID))," - ", HiYellow(acc.AccountName)))
		table := NewTable(cli, os.Stdout)
		table.SetHeader([]string{"App ID", "App Name"})
		table.SetColumnColor(tablewriter.Colors{tablewriter.Normal,tablewriter.FgWhiteColor},
		tablewriter.Colors{tablewriter.Normal, tablewriter.FgHiGreenColor})
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetNoWhiteSpace(true)
		table.SetAutoMergeCells(true)
		table.SetRowLine(true)
		for _, app := range acc.Applications {
				r := []string{strconv.Itoa(app.ID), strings.Trim(app.ApplicationName,"\"")}
				table.Append(r)
		}
		table.Render()
		fmt.Println()
		fmt.Println()
	}
	return err
}

// AppsInfoCmd shows detailed information on an app running on Section
type AppsInfoCmd struct {
	AccountID int `required short:"a"`
	AppID     int `required short:"i"`
}

// Run executes the command
func (c *AppsInfoCmd) Run(cli *CLI, logWriters *LogWriters) (err error) {
	s := NewSpinner("Looking up app info", logWriters)
	s.Start()

	app, err := api.Application(c.AccountID, c.AppID)
	s.Stop()
	fmt.Println()
	if err != nil {
		return err
	}

	if !(cli.Quiet){
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

				table := NewTable(cli, os.Stdout)
				table.SetHeader([]string{"Attribute", "Value"})
				table.SetHeaderColor(tablewriter.Colors{tablewriter.Normal,tablewriter.FgWhiteColor},
					tablewriter.Colors{tablewriter.Normal, tablewriter.FgWhiteColor})
				table.SetColumnColor(tablewriter.Colors{tablewriter.Normal,tablewriter.FgWhiteColor},
					tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor})
				table.SetAutoMergeCells(true)
				r := [][]string{
					{"Domain name", dom.Name},
					{"Zone name", dom.ZoneName},
					{"CNAME", dom.CNAME},
					{"Mode", dom.Mode},
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

			table := NewTable(cli, os.Stdout)
			table.SetHeader([]string{"Name", "Image"})
			table.SetHeaderColor(tablewriter.Colors{tablewriter.Normal,tablewriter.FgWhiteColor},
				tablewriter.Colors{tablewriter.Normal, tablewriter.FgWhiteColor})
			table.SetColumnColor(tablewriter.Colors{tablewriter.Normal,tablewriter.FgWhiteColor},
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor})
			table.SetAutoMergeCells(true)
			for _, p := range env.Stack {
				r := []string{p.Name, p.Image}
				table.Append(r)
			}
			table.Render()
		}

		fmt.Println()
	}


	return err
}

// AppsCreateCmd handles creating apps on Section
type AppsCreateCmd struct {
	AccountID int    `required short:"a" help:"ID of account to create the app under"`
	Hostname  string `required short:"d" help:"FQDN the app can be accessed at"`
	Origin    string `required short:"o" help:"URL to fetch the origin"`
	StackName string `required short:"s" help:"Name of stack to deploy. Try, for example, nodejs-basic"`
}

// Run executes the command
func (c *AppsCreateCmd) Run(logWriters *LogWriters) (err error) {	
	s := NewSpinner(fmt.Sprintf("Creating new app %s", c.Hostname),logWriters)
	s.Start()

	api.Timeout = 120 * time.Second // this specific request can take a long time
	r, err := api.ApplicationCreate(c.AccountID, c.Hostname, c.Origin, c.StackName)
	s.Stop()
	fmt.Println()
	if err != nil {
		if err == api.ErrStatusForbidden {
			stacks, herr := api.Stacks()
			if herr != nil {
				return fmt.Errorf("unable to query stacks: %w", herr)
			}
			for _, s := range stacks {
				if s.Name == c.StackName {
					return err
				}
			}
			return fmt.Errorf("bad request: unable to find stack %s", c.StackName)
		}
		return err
	}

	log.Info().Msg(fmt.Sprintf("\nSuccess: created app '%s' with id '%d'\n", r.ApplicationName, r.ID))

	return err
}

// AppsDeleteCmd handles deleting apps on Section
type AppsDeleteCmd struct {
	AccountID int `required short:"a" help:"ID of account the app belongs to"`
	AppID     int `required short:"i" help:"ID of the app to delete"`
}

// Run executes the command
func (c *AppsDeleteCmd) Run(logWriters *LogWriters) (err error) {
	s := NewSpinner(fmt.Sprintf("Deleting app with id '%d'", c.AppID),logWriters)
	s.Start()

	api.Timeout = 120 * time.Second // this specific request can take a long time
	_, err = api.ApplicationDelete(c.AccountID, c.AppID)
	s.Stop()
	if err != nil {
		return err
	}
	
	log.Info().Msg(fmt.Sprintf("\nSuccess: deleted app with id '%d'\n", c.AppID))

	return err
}

// AppsInitCmd creates and validates package.json to prepare an app for deployment
type AppsInitCmd struct {
	StackName string `optional default:"nodejs-basic" short:"s" help:"Name of stack to deploy. Default is nodejs-basic"`
	Force     bool   `optional short:"f" help:"Resets deployment specific files to their default configuration"`
}

// Run executes the command
func (c *AppsInitCmd) Run() (err error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	switch c.StackName {
	case "nodejs-basic":
		err := c.InitializeNodeBasicApp(stdout, stderr)
		if err != nil {
			return fmt.Errorf("[ERROR]: init completed with error %x", err)
		}
	default:
		log.Error().Msg(fmt.Sprintf(": Stack name %s does not have an initialization defined\n", c.StackName))
	}
	return err
}

// Create package.json
func (c *AppsInitCmd) CreatePkgJSON(stdout, stderr bytes.Buffer) (err error) {
	cmd := exec.Command("npm", "init", "-y")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	return err
}

// InitializeNodeBasicApp initializes a basic node app.
func (c *AppsInitCmd) InitializeNodeBasicApp(stdout bytes.Buffer, stderr bytes.Buffer) (err error) {
	if c.Force {
		log.Info().Msg(fmt.Sprintln("Removing old version of package.json"))
		err = os.Remove("package.json")
		if err != nil {
			if !os.IsNotExist(err) {
				log.Error().Msg(fmt.Sprintln("unable to remove package.json file"))
				return err
			}
		} else {
			log.Debug().Msg(fmt.Sprintln("Files successfully removed"))
		}
	}
	log.Debug().Msg(fmt.Sprintln("Checking to see if server.conf exists"))
	checkServConf, err := os.Open("server.conf")
	if err != nil {
		if !os.IsNotExist(err) {
			log.Error().Msg(fmt.Sprintln("server.conf file could not be opened"))
			return err
		}
	}
	defer checkServConf.Close()
	if err == nil {
		log.Info().Msg(fmt.Sprintln("Validating server.conf"))
		fileinfo, err := checkServConf.Stat()
		if err != nil {
			return fmt.Errorf("error in finding stat of server.conf %w", err)
		}
		buf := make([]byte, fileinfo.Size())
		_, err = checkServConf.Read(buf)
		if err != nil {
			return fmt.Errorf("error in size stat of server.conf %w", err)
		}
		fStr := string(buf)
		if !strings.Contains(fStr, "location / {") {
			log.Warn().Msg(fmt.Sprintln("default location unspecified. Edit or delete server.conf and rerun this command"))
		}
	}
	log.Debug().Msg(fmt.Sprintln("Checking to see if package.json exists"))
	checkPkgJSON, err := os.Open("package.json")
	if err != nil {
		log.Warn().Msg(fmt.Sprintln("package.json does not exist. Creating package.json"))
		err := c.CreatePkgJSON(stdout, stderr)
		if err != nil {
			return fmt.Errorf("there was an error creating package.json. Is node installed? %w", err)
		}
		log.Info().Msg(fmt.Sprintln("package.json created"))
	}
	defer checkPkgJSON.Close()
	validPkgJSON, err := os.OpenFile("package.json", os.O_RDWR, 0777)
	if err != nil {
		return fmt.Errorf("failed to open package.json %w", err)
	}
	defer validPkgJSON.Close()
	log.Info().Msg(fmt.Sprintln("Validating package.json"))
	buf, err := os.ReadFile("package.json")
	if err != nil {
		return fmt.Errorf("failed to read package.json %w", err)
	}
	fStr := string(buf)
	if len(fStr) == 0 {
		err := os.Remove("package.json")
		if err != nil {
			log.Error().Msg(fmt.Sprintln("unable to remove empty package.json"))
		}
		log.Warn().Msg(fmt.Sprintln("package.json is empty. Creating package.json"))
		err = c.CreatePkgJSON(stdout, stderr)
		if err != nil {
			return fmt.Errorf("there was an error creating package.json. Is node installed? %w", err)
		}
		log.Info().Msg(fmt.Sprintln("package.json created from empty file"))
		buf, err = os.ReadFile("package.json")
		if err != nil {
			return fmt.Errorf("failed to read package.json %w", err)
		}
		fStr = string(buf)
	}
	jsonMap := make(map[string]interface{})
	err = json.Unmarshal(buf, &jsonMap)
	if err != nil {
		return fmt.Errorf("package.json is not valid JSON %w", err)
	}
	lv := jsonMap["scripts"]
	jsonToStrMap, ok := lv.(map[string]interface{})
	if !ok {
		return fmt.Errorf("json unable to be read as map[string]interface %w", err)
	}
	_, ok = jsonToStrMap["start"]
	if !ok {
		jsonToStrMap["start"] = "node YOUR_SERVER_HERE.js"
		jsonMap["scripts"] = jsonToStrMap
		err = os.Truncate("package.json", 0)
		if err != nil {
			return fmt.Errorf("failed to empty package.json %w", err)
		}
		set, err := json.MarshalIndent(jsonMap, "", "  ")
		if err != nil {
			log.Error().Msg(fmt.Sprintln("unable to add start script placeholder"))
		}
		_, err = validPkgJSON.Write(set)
		if err != nil {
			log.Error().Msg(fmt.Sprintln("unable to add start script placeholder"))
		}
	}
	if strings.Contains(fStr, `YOUR_SERVER_HERE.js`) {
		log.Error().Msg(fmt.Sprintln("start script is required. Please edit the placeholder in package.json"))
	}
	return err
}

// AppsStacksCmd lists available stacks to create new apps with
type AppsStacksCmd struct{}

// Run executes the command
func (c *AppsStacksCmd) Run(cli *CLI, logWriters *LogWriters) (err error) {
	s := NewSpinner("Looking up stacks",logWriters)
	s.Start()
	k, err := api.Stacks()
	s.Stop()
	if err != nil {
		return fmt.Errorf("unable to look up stacks: %w", err)
	}

	table := NewTable(cli, os.Stdout)
	table.SetHeader([]string{"Name", "Label", "Description", "Type"})

	for _, s := range k {
		r := []string{s.Name, s.Label, s.Description, s.Type}
		table.Append(r)
	}

	table.Render()
	return err
}
