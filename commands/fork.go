package commands

import (
	"fmt"

	"github.com/github/hub/v2/github"
	"github.com/github/hub/v2/ui"
	"github.com/github/hub/v2/utils"
)

var cmdFork = &Command{
	Run:   fork,
	Usage: "fork [--no-remote] [--remote-name <REMOTE>] [--org <ORGANIZATION>]",
	Long: `Fork the current repository on GitHub and add a git remote for it.

## Options:
	--no-remote
		Skip adding a git remote for the fork.

	--remote-name <REMOTE>
		Set the name for the new git remote.

	--org <ORGANIZATION>
		Fork the repository within this organization.

	-d, --set-push-default
		Set remote.pushDefault to the name of the new git remote.

## Examples:
		$ hub fork
		[ repo forked on GitHub ]
		> git remote add -f USER git@github.com:USER/REPO.git

		$ hub fork --org=ORGANIZATION
		[ repo forked on GitHub into the ORGANIZATION organization]
		> git remote add -f ORGANIZATION git@github.com:ORGANIZATION/REPO.git

## See also:

hub-clone(1), hub(1)
`,
}

func init() {
	CmdRunner.Use(cmdFork)
}

func fork(cmd *Command, args *Args) {
	localRepo, err := github.LocalRepo()
	utils.Check(err)

	project, err := localRepo.MainProject()
	utils.Check(err)

	config := github.CurrentConfig()
	host, err := config.PromptForHost(project.Host)
	utils.Check(github.FormatError("forking repository", err))

	params := map[string]interface{}{}
	forkOwner := host.User
	if flagForkOrganization := args.Flag.Value("--org"); flagForkOrganization != "" {
		forkOwner = flagForkOrganization
		params["organization"] = forkOwner
	}

	forkProject := github.NewProject(forkOwner, project.Name, project.Host)
	var newRemoteName string
	if flagForkRemoteName := args.Flag.Value("--remote-name"); flagForkRemoteName != "" {
		newRemoteName = flagForkRemoteName
	} else {
		newRemoteName = forkProject.Owner
	}

	client := github.NewClient(project.Host)
	existingRepo, err := client.Repository(forkProject)
	if err == nil {
		existingProject, err := github.NewProjectFromRepo(existingRepo)
		if err == nil && !existingProject.SameAs(forkProject) {
			existingRepo = nil
		}
	}
	if err == nil && existingRepo != nil {
		var parentURL *github.URL
		if parent := existingRepo.Parent; parent != nil {
			parentURL, _ = github.ParseURL(parent.HTMLURL)
		}
		if parentURL == nil || !project.SameAs(parentURL.Project) {
			err = fmt.Errorf("Error creating fork: %s already exists on %s",
				forkProject, forkProject.Host)
			utils.Check(err)
		}
	} else {
		if !args.Noop {
			newRepo, err := client.ForkRepository(project, params)
			utils.Check(err)
			forkProject.Owner = newRepo.Owner.Login
			forkProject.Name = newRepo.Name
		}
	}

	args.NoForward()
	if !args.Flag.Bool("--no-remote") {
		originURL := project.GitURL("", "", false)
		url := forkProject.GitURL("", "", true)

		// Check to see if the remote already exists.
		currentRemote, err := localRepo.RemoteByName(newRemoteName)
		if err == nil {
			currentProject, err := currentRemote.Project()
			if err == nil {
				if currentProject.SameAs(forkProject) {
					ui.Printf("existing remote: %s\n", newRemoteName)
					return
				}
				if newRemoteName == "origin" {
					// Assume user wants to follow github guides for collaboration
					ui.Printf("renaming existing \"origin\" remote to \"upstream\"\n")
					args.Before("git", "remote", "rename", "origin", "upstream")
				}
			}
		}

		args.Before("git", "remote", "add", "-f", newRemoteName, originURL)
		args.Before("git", "remote", "set-url", newRemoteName, url)
		if args.Flag.Bool("--set-push-default") {
			args.Before("git", "config", "remote.pushDefault", newRemoteName)
		}

		args.AfterFn(func() error {
			ui.Printf("new remote: %s\n", newRemoteName)
			return nil
		})
	}
}
