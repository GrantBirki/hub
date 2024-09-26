package commands

import (
	"fmt"
	"os"

	"github.com/github/hub/github"
	"github.com/github/hub/utils"
)

var cmdUser = &Command{
	Run:   user,
	Usage: "user",
	Short: "Show the default user name",
	Long: `Show which user name will be used by hub commands.`,
}

func init() {
	CmdRunner.Use(cmdUser)
}

/*
  $ gh user
  YOUR_USER
*/
func user(cmd *Command, args *Args) {
    var host *github.Host
    config := github.CurrentConfig()
    host, err := config.DefaultHost()
    if err != nil {
            utils.Check(github.FormatError("reading config", err))
    }
    fmt.Print(host.User + "\n")
    os.Exit(0)
}
