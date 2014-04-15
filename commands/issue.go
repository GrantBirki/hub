package commands

import (
	"fmt"
	"github.com/jingweno/gh/github"
	"github.com/jingweno/gh/utils"
)

var (
	cmdIssue = &Command{
		Run:   issue,
		Usage: "issue",
		Short: "List issues on GitHub",
		Long:  `List summary of the open issues for the project that the "origin" remote points to.`,
	}

	cmdCreateIssue = &Command{
		Key:   "create",
		Run:   createIssue,
		Usage: "issue create [-m <MESSAGE>|-f <FILE>] [-l <LABEL-1>,<LABEL-2>...,<LABEL-N>]",
		Short: "Create an issue on GitHub",
		Long: `Create an issue for the project that the "origin" remote points to.

Without <MESSAGE> or <FILE>, a text editor will open in which title and body
of the release can be entered in the same manner as git commit message.

Specify one or more labels via "-a".
`,
	}

	flagIssueMessage,
	flagIssueFile string

	flagIssueLabels listFlag
)

func init() {
	cmdCreateIssue.Flag.StringVarP(&flagIssueMessage, "message", "m", "", "MESSAGE")
	cmdCreateIssue.Flag.StringVarP(&flagIssueFile, "file", "f", "", "FILE")
	cmdCreateIssue.Flag.VarP(&flagIssueLabels, "label", "l", "LABEL")

	cmdIssue.Use(cmdCreateIssue)
	CmdRunner.Use(cmdIssue)
}

/*
  $ gh issue
*/
func issue(cmd *Command, args *Args) {
	runInLocalRepo(func(localRepo *github.GitHubRepo, project *github.Project, gh *github.Client) {
		if args.Noop {
			fmt.Printf("Would request list of issues for %s\n", project)
		} else {
			issues, err := gh.Issues(project)
			utils.Check(err)
			for _, issue := range issues {
				var url string
				// use the pull request URL if we have one
				if issue.PullRequest.HTMLURL != "" {
					url = issue.PullRequest.HTMLURL
				} else {
					url = issue.HTMLURL
				}
				// "nobody" should have more than 1 million github issues
				fmt.Printf("% 7d] %s ( %s )\n", issue.Number, issue.Title, url)
			}
		}
	})
}

func createIssue(cmd *Command, args *Args) {
	runInLocalRepo(func(localRepo *github.GitHubRepo, project *github.Project, gh *github.Client) {
		if args.Noop {
			fmt.Printf("Would create an issue for %s\n", project)
		} else {
			title, body, err := getTitleAndBodyFromFlags(flagIssueMessage, flagIssueFile)
			utils.Check(err)

			if title == "" {
				title, body, err = writeIssueTitleAndBody(project)
				utils.Check(err)
			}

			issue, err := gh.CreateIssue(project, title, body, flagIssueLabels)
			utils.Check(err)

			fmt.Println(issue.HTMLURL)
		}
	})
}

func writeIssueTitleAndBody(project *github.Project) (string, string, error) {
	message := `
# Creating issue for %s.
#
# Write a message for this issue. The first block
# of the text is the title and the rest is description.
`
	message = fmt.Sprintf(message, project.Name)

	editor, err := github.NewEditor("ISSUE", message)
	if err != nil {
		return "", "", err
	}

	return editor.EditTitleAndBody()
}
