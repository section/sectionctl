package commands

import (
	"fmt"
	"os"
)

// DeployCmd handles deploying an app to Section
type DeployCmd struct{}

// Run deploys an app to Section's edge
func (a *DeployCmd) Run() (err error) {
	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Packaging %s\n", path)
	fmt.Println("Deploying...")
	return
}
