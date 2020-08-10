# How to add new subcommands

1. Create a new file in `commands/`
1. Ensure that it:

   - Is part of the module package
   - Exports a type for your command
   - Has a `Run` function

   It should look like this:

   ```
   # commands/example.go
   package commands

   import (
     "fmt"
   )

   // MyNewCmd is an example command
   type ExampleCmd struct{}

   // Run executes the `login` command
   func (a *ExampleCmd) Run() (err error) {
     fmt.Printf("Hello, world!")
     return err
   }
   ```
1. Add any tests at `commands/...\_test.go`
1. Include your new type in the `CLI` struct in `./section.go`
