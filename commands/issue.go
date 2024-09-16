package commands

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/github/hub/v2/git"
	"github.com/github/hub/v2/github"
	"github.com/github/hub/v2/ui"
	"github.com/github/hub/v2/utils"
)

var (
	cmdIssue = &Command{
		Run: listIssues,
		Usage: `
issue [-a <ASSIGNEE>] [-c <CREATOR>] [-@ <USER>] [-s <STATE>] [-f <FORMAT>] [-M <MILESTONE>] [-l <LABELS>] [-d <DATE>] [-o <SORT_KEY> [-^]] [-L <LIMIT>] [-r REMOTE_URL]
issue show [-f <FORMAT>] <NUMBER>
issue create [-oc] [-m <MESSAGE>|-F <FILE>] [--edit] [-a <USERS>] [-M <MILESTONE>] [-l <LABELS>]
issue update <NUMBER> [-m <MESSAGE>|-F <FILE>] [--edit] [-a <USERS>] [-M <MILESTONE>] [-l <LABELS>] [-s <STATE>]
issue labels [--color]
issue transfer <NUMBER> <REPO>
`,
		Long: `Manage GitHub Issues for the current repository.

## Commands:

With no arguments, show a list of open issues.

	* _show_:
		Show an existing issue specified by <NUMBER>.

	* _create_:
		Open an issue in the current repository.

	* _update_:
		Update fields of an existing issue specified by <NUMBER>. Use ''--edit''
		to edit the title and message interactively in the text editor.

	* _labels_:
		List the labels available in this repository.

	* _transfer_:
		Transfer an issue to another repository.

## Options:
	-a, --assignee <ASSIGNEE>
		In list mode, display only issues assigned to <ASSIGNEE>.

	-a, --assign <USERS>
		A comma-separated list of GitHub handles to assign to the created issue.

	-c, --creator <CREATOR>
		Display only issues created by <CREATOR>.

	-@, --mentioned <USER>
		Display only issues mentioning <USER>.

	-s, --state <STATE>
		Display issues with state <STATE> (default: "open").

	-f, --format <FORMAT>
		Pretty print the contents of the issues using format <FORMAT> (default:
		"%sC%>(8)%i%Creset  %t%  l%n"). See the "PRETTY FORMATS" section of
		git-log(1) for some additional details on how placeholders are used in
		format. The available placeholders for issues are:

		%I: issue number

		%i: issue number prefixed with "#"

		%U: the URL of this issue

		%S: state (i.e. "open", "closed")

		%sC: set color to red or green, depending on issue state.

		%t: title

		%l: colored labels

		%L: raw, comma-separated labels

		%b: body

		%au: login name of author

		%as: comma-separated list of assignees

		%Mn: milestone number

		%Mt: milestone title

		%NC: number of comments

		%Nc: number of comments wrapped in parentheses, or blank string if zero.

		%cD: created date-only (no time of day)

		%cr: created date, relative

		%ct: created date, UNIX timestamp

		%cI: created date, ISO 8601 format

		%uD: updated date-only (no time of day)

		%ur: updated date, relative

		%ut: updated date, UNIX timestamp

		%uI: updated date, ISO 8601 format

		%n: newline

		%%: a literal %

	--color[=<WHEN>]
		Enable colored output even if stdout is not a terminal. <WHEN> can be one
		of "always" (default for ''--color''), "never", or "auto" (default).

	-m, --message <MESSAGE>
		The text up to the first blank line in <MESSAGE> is treated as the issue
		title, and the rest is used as issue description in Markdown format.

		When multiple ''--message'' are passed, their values are concatenated with a
		blank line in-between.

		When neither ''--message'' nor ''--file'' were supplied to ''issue create'', a
		text editor will open to author the title and description in.

	-F, --file <FILE>
		Read the issue title and description from <FILE>. Pass "-" to read from
		standard input instead. See ''--message'' for the formatting rules.

	-e, --edit
		Open the issue title and description in a text editor before submitting.
		This can be used in combination with ''--message'' or ''--file''.

	-o, --browse
		Open the new issue in a web browser.

	-c, --copy
		Put the URL of the new issue to clipboard instead of printing it.

	-M, --milestone <NAME>
		Display only issues for a GitHub milestone with the name <NAME>.

		When opening an issue, add this issue to a GitHub milestone with the name <NAME>.
		Passing the milestone number is deprecated.

	-l, --labels <LABELS>
		Display only issues with certain labels.

		When opening an issue, add a comma-separated list of labels to this issue.

	-d, --since <DATE>
		Display only issues updated on or after <DATE> in ISO 8601 format.

	-o, --sort <KEY>
		Sort displayed issues by "created" (default), "updated" or "comments".

	-^ --sort-ascending
		Sort by ascending dates instead of descending.

	-L, --limit <LIMIT>
		Display only the first <LIMIT> issues.

	-r, --remote-url <REMOTE_URL>
		Avoid requiring a local clone in order to get issue data. Simply provide REMOTE_URL.

	--include-pulls
		Include pull requests as well as issues.

	--color
		Enable colored output for labels list.

## See also:

hub-pr(1), hub(1)
`,
		KnownFlags: `
		-a, --assignee USER
		-s, --state STATE
		-f, --format FMT
		-M, --milestone NAME
		-c, --creator USER
		-@, --mentioned USER
		-l, --labels LIST
		-d, --since DATE
		-o, --sort KEY
		-^, --sort-ascending
		--include-pulls
		-L, --limit N
		-r, --remote-url REMOTE_URL
		--color
`,
	}

	cmdCreateIssue = &Command{
		Key: "create",
		Run: createIssue,
		KnownFlags: `
		-m, --message MSG
		-F, --file FILE
		-M, --milestone NAME
		-l, --labels LIST
		-a, --assign USER
		-o, --browse
		-c, --copy
		-e, --edit
`,
	}

	cmdShowIssue = &Command{
		Key: "show",
		Run: showIssue,
		KnownFlags: `
		-f, --format FMT
		--color
`,
	}

	cmdLabel = &Command{
		Key: "labels",
		Run: listLabels,
		KnownFlags: `
		--color
`,
	}

	cmdTransfer = &Command{
		Key: "transfer",
		Run: transferIssue,
	}

	cmdUpdate = &Command{
		Key: "update",
		Run: updateIssue,
		KnownFlags: `
		-m, --message MSG
		-F, --file FILE
		-M, --milestone NAME
		-l, --labels LIST
		-a, --assign USER
		-e, --edit
		-s, --state STATE
`,
	}
)

func init() {
	cmdIssue.Use(cmdShowIssue)
	cmdIssue.Use(cmdCreateIssue)
	cmdIssue.Use(cmdLabel)
	cmdIssue.Use(cmdTransfer)
	cmdIssue.Use(cmdUpdate)
	CmdRunner.Use(cmdIssue)
}

func calculateProjectFromRFlag(rawRemote string) *github.Project {
	var remote *github.Remote

	if remoteURL, err := url.Parse(rawRemote); err == nil {
		remote = &github.Remote{"origin", remoteURL, remoteURL}
	} else {
		fmt.Fprintf(os.Stderr, "invalid remote URL string: %s", rawRemote)
		os.Exit(1)
	}

	if _, err := remote.Project(); err != nil {
		fmt.Fprintf(os.Stderr, "no project for: %s because %s", rawRemote, err)
		os.Exit(1)
	}

	validProject, _ := remote.Project()
	return validProject
}

func listIssues(cmd *Command, args *Args) {
	var project *github.Project

	if rawRemote := args.Flag.Value("--remote-url"); args.Flag.HasReceived("--remote-url") {
		project = calculateProjectFromRFlag(rawRemote)
	}

	if project == nil {
		localRepo, err := github.LocalRepo()
		utils.Check(err)

		project, err = localRepo.MainProject()
		utils.Check(err)
	}

	gh := github.NewClient(project.Host)

	if args.Noop {
		ui.Printf("Would request list of issues for %s\n", project)
	} else {
		filters := map[string]interface{}{}
		if args.Flag.HasReceived("--state") {
			filters["state"] = args.Flag.Value("--state")
		}
		if args.Flag.HasReceived("--assignee") {
			filters["assignee"] = args.Flag.Value("--assignee")
		}
		if args.Flag.HasReceived("--milestone") {
			milestoneValue := args.Flag.Value("--milestone")
			if milestoneValue == "none" {
				filters["milestone"] = milestoneValue
			} else {
				milestoneNumber, err := milestoneValueToNumber(milestoneValue, gh, project)
				utils.Check(err)
				if milestoneNumber > 0 {
					filters["milestone"] = milestoneNumber
				}
			}
		}
		if args.Flag.HasReceived("--creator") {
			filters["creator"] = args.Flag.Value("--creator")
		}
		if args.Flag.HasReceived("--mentioned") {
			filters["mentioned"] = args.Flag.Value("--mentioned")
		}
		if args.Flag.HasReceived("--labels") {
			labels := commaSeparated(args.Flag.AllValues("--labels"))
			filters["labels"] = strings.Join(labels, ",")
		}
		if args.Flag.HasReceived("--sort") {
			filters["sort"] = args.Flag.Value("--sort")
		}

		if args.Flag.Bool("--sort-ascending") {
			filters["direction"] = "asc"
		} else {
			filters["direction"] = "desc"
		}

		if args.Flag.HasReceived("--since") {
			flagIssueSince := args.Flag.Value("--since")
			if sinceTime, err := time.ParseInLocation("2006-01-02", flagIssueSince, time.Local); err == nil {
				filters["since"] = sinceTime.Format(time.RFC3339)
			} else {
				filters["since"] = flagIssueSince
			}
		}

		flagIssueLimit := args.Flag.Int("--limit")
		flagIssueIncludePulls := args.Flag.Bool("--include-pulls")
		flagIssueFormat := "%sC%>(8)%i%Creset  %t%  l%n"
		if args.Flag.HasReceived("--format") {
			flagIssueFormat = args.Flag.Value("--format")
		}

		issues, err := gh.FetchIssues(project, filters, flagIssueLimit, func(issue *github.Issue) bool {
			return issue.PullRequest == nil || flagIssueIncludePulls
		})
		utils.Check(err)

		maxNumWidth := 0
		for _, issue := range issues {
			if numWidth := len(strconv.Itoa(issue.Number)); numWidth > maxNumWidth {
				maxNumWidth = numWidth
			}
		}

		colorize := colorizeOutput(args.Flag.HasReceived("--color"), args.Flag.Value("--color"))
		for _, issue := range issues {
			ui.Print(formatIssue(issue, flagIssueFormat, colorize))
		}
	}

	args.NoForward()
}

func formatIssuePlaceholders(issue github.Issue, colorize bool) map[string]string {
	var stateColorSwitch string
	if colorize {
		issueColor := 32
		if issue.State == "closed" {
			issueColor = 31
		}
		stateColorSwitch = fmt.Sprintf("\033[%dm", issueColor)
	}

	var labelStrings []string
	var rawLabels []string
	for _, label := range issue.Labels {
		if colorize {
			color, err := utils.NewColor(label.Color)
			utils.Check(err)
			labelStrings = append(labelStrings, colorizeLabel(label, color))
		} else {
			labelStrings = append(labelStrings, fmt.Sprintf(" %s ", label.Name))
		}
		rawLabels = append(rawLabels, label.Name)
	}

	var assignees []string
	for _, assignee := range issue.Assignees {
		assignees = append(assignees, assignee.Login)
	}

	var milestoneNumber, milestoneTitle string
	if issue.Milestone != nil {
		milestoneNumber = fmt.Sprintf("%d", issue.Milestone.Number)
		milestoneTitle = issue.Milestone.Title
	}

	var numCommentsWrapped string
	numComments := fmt.Sprintf("%d", issue.Comments)
	if issue.Comments > 0 {
		numCommentsWrapped = fmt.Sprintf("(%d)", issue.Comments)
	}

	var createdDate, createdAtISO8601, createdAtUnix, createdAtRelative,
		updatedDate, updatedAtISO8601, updatedAtUnix, updatedAtRelative string
	if !issue.CreatedAt.IsZero() {
		createdDate = issue.CreatedAt.Format("02 Jan 2006")
		createdAtISO8601 = issue.CreatedAt.Format(time.RFC3339)
		createdAtUnix = fmt.Sprintf("%d", issue.CreatedAt.Unix())
		createdAtRelative = utils.TimeAgo(issue.CreatedAt)
	}
	if !issue.UpdatedAt.IsZero() {
		updatedDate = issue.UpdatedAt.Format("02 Jan 2006")
		updatedAtISO8601 = issue.UpdatedAt.Format(time.RFC3339)
		updatedAtUnix = fmt.Sprintf("%d", issue.UpdatedAt.Unix())
		updatedAtRelative = utils.TimeAgo(issue.UpdatedAt)
	}

	return map[string]string{
		"I":  fmt.Sprintf("%d", issue.Number),
		"i":  fmt.Sprintf("#%d", issue.Number),
		"U":  issue.HTMLURL,
		"S":  issue.State,
		"sC": stateColorSwitch,
		"t":  issue.Title,
		"l":  strings.Join(labelStrings, " "),
		"L":  strings.Join(rawLabels, ", "),
		"b":  issue.Body,
		"au": issue.User.Login,
		"as": strings.Join(assignees, ", "),
		"Mn": milestoneNumber,
		"Mt": milestoneTitle,
		"NC": numComments,
		"Nc": numCommentsWrapped,
		"cD": createdDate,
		"cI": createdAtISO8601,
		"ct": createdAtUnix,
		"cr": createdAtRelative,
		"uD": updatedDate,
		"uI": updatedAtISO8601,
		"ut": updatedAtUnix,
		"ur": updatedAtRelative,
	}
}

func formatPullRequestPlaceholders(pr github.PullRequest, colorize bool) map[string]string {
	prState := pr.State
	if prState == "open" && pr.Draft {
		prState = "draft"
	} else if !pr.MergedAt.IsZero() {
		prState = "merged"
	}

	var stateColorSwitch string
	var prColor int
	if colorize {
		switch prState {
		case "draft":
			prColor = 37
		case "merged":
			prColor = 35
		case "closed":
			prColor = 31
		default:
			prColor = 32
		}
		stateColorSwitch = fmt.Sprintf("\033[%dm", prColor)
	}

	base := pr.Base.Ref
	head := pr.Head.Label
	if pr.IsSameRepo() {
		head = pr.Head.Ref
	}

	var requestedReviewers []string
	for _, requestedReviewer := range pr.RequestedReviewers {
		requestedReviewers = append(requestedReviewers, requestedReviewer.Login)
	}
	for _, requestedTeam := range pr.RequestedTeams {
		teamSlug := fmt.Sprintf("%s/%s", pr.Base.Repo.Owner.Login, requestedTeam.Slug)
		requestedReviewers = append(requestedReviewers, teamSlug)
	}

	var mergedDate, mergedAtISO8601, mergedAtUnix, mergedAtRelative string
	if !pr.MergedAt.IsZero() {
		mergedDate = pr.MergedAt.Format("02 Jan 2006")
		mergedAtISO8601 = pr.MergedAt.Format(time.RFC3339)
		mergedAtUnix = fmt.Sprintf("%d", pr.MergedAt.Unix())
		mergedAtRelative = utils.TimeAgo(pr.MergedAt)
	}

	return map[string]string{
		"pS": prState,
		"pC": stateColorSwitch,
		"B":  base,
		"H":  head,
		"sB": pr.Base.Sha,
		"sH": pr.Head.Sha,
		"sm": pr.MergeCommitSha,
		"rs": strings.Join(requestedReviewers, ", "),
		"mD": mergedDate,
		"mI": mergedAtISO8601,
		"mt": mergedAtUnix,
		"mr": mergedAtRelative,
	}
}

func formatIssue(issue github.Issue, format string, colorize bool) string {
	placeholders := formatIssuePlaceholders(issue, colorize)
	return ui.Expand(format, placeholders, colorize)
}

func showIssue(cmd *Command, args *Args) {
	issueNumber := ""
	if args.ParamsSize() > 0 {
		issueNumber = args.GetParam(0)
	}
	if issueNumber == "" {
		utils.Check(cmd.UsageError(""))
	}

	localRepo, err := github.LocalRepo()
	utils.Check(err)

	project, err := localRepo.MainProject()
	utils.Check(err)

	gh := github.NewClient(project.Host)

	var issue = &github.Issue{}
	issue, err = gh.FetchIssue(project, issueNumber)
	utils.Check(err)

	args.NoForward()

	colorize := colorizeOutput(args.Flag.HasReceived("--color"), args.Flag.Value("--color"))
	if args.Flag.HasReceived("--format") {
		flagShowIssueFormat := args.Flag.Value("--format")
		ui.Print(formatIssue(*issue, flagShowIssueFormat, colorize))
		return
	}

	var closed = ""
	if issue.State != "open" {
		closed = "[CLOSED] "
	}
	commentsList, err := gh.FetchComments(project, issueNumber)
	utils.Check(err)

	ui.Printf("# %s%s\n\n", closed, issue.Title)
	ui.Printf("* created by @%s on %s\n", issue.User.Login, issue.CreatedAt.String())

	if len(issue.Assignees) > 0 {
		var assignees []string
		for _, user := range issue.Assignees {
			assignees = append(assignees, user.Login)
		}
		ui.Printf("* assignees: %s\n", strings.Join(assignees, ", "))
	}

	ui.Printf("\n%s\n", issue.Body)

	if issue.Comments > 0 {
		ui.Printf("\n## Comments:\n")
		for _, comment := range commentsList {
			ui.Printf("\n### comment by @%s on %s\n\n%s\n", comment.User.Login, comment.CreatedAt.String(), comment.Body)
		}
	}
}

func createIssue(cmd *Command, args *Args) {
	localRepo, err := github.LocalRepo()
	utils.Check(err)

	project, err := localRepo.MainProject()
	utils.Check(err)

	gh := github.NewClient(project.Host)

	messageBuilder := &github.MessageBuilder{
		Filename: "ISSUE_EDITMSG",
		Title:    "issue",
	}

	messageBuilder.AddCommentedSection(fmt.Sprintf(`Creating an issue for %s

Write a message for this issue. The first block of
text is the title and the rest is the description.`, project))

	flagIssueEdit := args.Flag.Bool("--edit")
	flagIssueMessage := args.Flag.AllValues("--message")
	if len(flagIssueMessage) > 0 {
		messageBuilder.Message = strings.Join(flagIssueMessage, "\n\n")
		messageBuilder.Edit = flagIssueEdit
	} else if args.Flag.HasReceived("--file") {
		messageBuilder.Message, err = msgFromFile(args.Flag.Value("--file"))
		utils.Check(err)
		messageBuilder.Edit = flagIssueEdit
	} else {
		messageBuilder.Edit = true

		workdir, _ := git.WorkdirName()
		if workdir != "" {
			template, err := github.ReadTemplate(github.IssueTemplate, workdir)
			utils.Check(err)
			if template != "" {
				messageBuilder.Message = template
			}
		}

	}

	title, body, err := messageBuilder.Extract()
	utils.Check(err)

	if title == "" {
		utils.Check(fmt.Errorf("Aborting creation due to empty issue title"))
	}

	params := map[string]interface{}{
		"title": title,
		"body":  body,
	}

	setLabelsFromArgs(params, args)

	setAssigneesFromArgs(params, args)

	setMilestoneFromArgs(params, args, gh, project)

	args.NoForward()
	if args.Noop {
		ui.Printf("Would create issue `%s' for %s\n", params["title"], project)
	} else {
		issue, err := gh.CreateIssue(project, params)
		utils.Check(err)

		flagIssueBrowse := args.Flag.Bool("--browse")
		flagIssueCopy := args.Flag.Bool("--copy")
		printBrowseOrCopy(args, issue.HTMLURL, flagIssueBrowse, flagIssueCopy)
	}

	messageBuilder.Cleanup()
}

func updateIssue(cmd *Command, args *Args) {
	issueNumber := 0
	if args.ParamsSize() > 0 {
		issueNumber, _ = strconv.Atoi(args.GetParam(0))
	}
	if issueNumber == 0 {
		utils.Check(cmd.UsageError(""))
	}
	if !hasField(args, "--message", "--file", "--labels", "--milestone", "--assign", "--state", "--edit") {
		utils.Check(cmd.UsageError("please specify fields to update"))
	}

	localRepo, err := github.LocalRepo()
	utils.Check(err)

	project, err := localRepo.MainProject()
	utils.Check(err)

	gh := github.NewClient(project.Host)

	params := map[string]interface{}{}
	setLabelsFromArgs(params, args)
	setAssigneesFromArgs(params, args)
	setMilestoneFromArgs(params, args, gh, project)

	if args.Flag.HasReceived("--state") {
		params["state"] = args.Flag.Value("--state")
	}

	if hasField(args, "--message", "--file", "--edit") {
		messageBuilder := &github.MessageBuilder{
			Filename: "ISSUE_EDITMSG",
			Title:    "issue",
		}

		messageBuilder.AddCommentedSection(fmt.Sprintf(`Editing issue #%d for %s

Update the message for this issue. The first block of
text is the title and the rest is the description.`, issueNumber, project))

		messageBuilder.Edit = args.Flag.Bool("--edit")
		flagIssueMessage := args.Flag.AllValues("--message")
		if len(flagIssueMessage) > 0 {
			messageBuilder.Message = strings.Join(flagIssueMessage, "\n\n")
		} else if args.Flag.HasReceived("--file") {
			messageBuilder.Message, err = msgFromFile(args.Flag.Value("--file"))
			utils.Check(err)
		} else {
			issue, err := gh.FetchIssue(project, strconv.Itoa(issueNumber))
			utils.Check(err)
			existingMessage := fmt.Sprintf("%s\n\n%s", issue.Title, issue.Body)
			messageBuilder.Message = strings.Replace(existingMessage, "\r\n", "\n", -1)
		}

		title, body, err := messageBuilder.Extract()
		utils.Check(err)
		if title == "" {
			utils.Check(fmt.Errorf("Aborting creation due to empty issue title"))
		}
		params["title"] = title
		params["body"] = body
		defer messageBuilder.Cleanup()
	}

	args.NoForward()
	if args.Noop {
		ui.Printf("Would update issue #%d for %s\n", issueNumber, project)
	} else {
		err := gh.UpdateIssue(project, issueNumber, params)
		utils.Check(err)
	}
}

func listLabels(cmd *Command, args *Args) {
	localRepo, err := github.LocalRepo()
	utils.Check(err)

	project, err := localRepo.MainProject()
	utils.Check(err)

	gh := github.NewClient(project.Host)

	args.NoForward()
	if args.Noop {
		ui.Printf("Would request list of labels for %s\n", project)
		return
	}

	labels, err := gh.FetchLabels(project)
	utils.Check(err)

	flagLabelsColorize := colorizeOutput(args.Flag.HasReceived("--color"), args.Flag.Value("--color"))
	for _, label := range labels {
		ui.Print(formatLabel(label, flagLabelsColorize))
	}
}

func hasField(args *Args, names ...string) bool {
	found := false
	for _, name := range names {
		if args.Flag.HasReceived(name) {
			found = true
		}
	}
	return found
}

func setLabelsFromArgs(params map[string]interface{}, args *Args) {
	if !args.Flag.HasReceived("--labels") {
		return
	}
	params["labels"] = commaSeparated(args.Flag.AllValues("--labels"))
}

func setAssigneesFromArgs(params map[string]interface{}, args *Args) {
	if !args.Flag.HasReceived("--assign") {
		return
	}
	params["assignees"] = commaSeparated(args.Flag.AllValues("--assign"))
}

func setMilestoneFromArgs(params map[string]interface{}, args *Args, gh *github.Client, project *github.Project) {
	if !args.Flag.HasReceived("--milestone") {
		return
	}
	milestoneNumber, err := milestoneValueToNumber(args.Flag.Value("--milestone"), gh, project)
	utils.Check(err)
	if milestoneNumber == 0 {
		params["milestone"] = nil
	} else {
		params["milestone"] = milestoneNumber
	}
}

func colorizeOutput(colorSet bool, when string) bool {
	if !colorSet || when == "auto" {
		colorConfig, _ := git.Config("color.ui")
		switch colorConfig {
		case "false", "never":
			return false
		case "always":
			return true
		}
		return ui.IsTerminal(os.Stdout)
	} else if when == "never" {
		return false
	} else {
		return true // "always"
	}
}

func formatLabel(label github.IssueLabel, colorize bool) string {
	if colorize {
		if color, err := utils.NewColor(label.Color); err == nil {
			return fmt.Sprintf("%s\n", colorizeLabel(label, color))
		}
	}
	return fmt.Sprintf("%s\n", label.Name)
}

func colorizeLabel(label github.IssueLabel, color *utils.Color) string {
	bgColorCode := utils.RgbToTermColorCode(color)
	fgColor := pickHighContrastTextColor(color)
	fgColorCode := utils.RgbToTermColorCode(fgColor)
	return fmt.Sprintf("\033[38;%s;48;%sm %s \033[m",
		fgColorCode, bgColorCode, label.Name)
}

type contrastCandidate struct {
	color    *utils.Color
	contrast float64
}

func pickHighContrastTextColor(color *utils.Color) *utils.Color {
	candidates := []contrastCandidate{}
	appendCandidate := func(c *utils.Color) {
		candidates = append(candidates, contrastCandidate{
			color:    c,
			contrast: color.ContrastRatio(c),
		})
	}

	appendCandidate(utils.White)
	appendCandidate(utils.Black)

	for _, candidate := range candidates {
		if candidate.contrast >= 7.0 {
			return candidate.color
		}
	}
	for _, candidate := range candidates {
		if candidate.contrast >= 4.5 {
			return candidate.color
		}
	}
	return utils.Black
}

func milestoneValueToNumber(value string, client *github.Client, project *github.Project) (int, error) {
	if value == "" {
		return 0, nil
	}

	if milestoneNumber, err := strconv.Atoi(value); err == nil {
		return milestoneNumber, nil
	}

	milestones, err := client.FetchMilestones(project)
	if err != nil {
		return 0, err
	}
	for _, milestone := range milestones {
		if strings.EqualFold(milestone.Title, value) {
			return milestone.Number, nil
		}
	}

	return 0, fmt.Errorf("error: no milestone found with name '%s'", value)
}

func transferIssue(cmd *Command, args *Args) {
	if args.ParamsSize() < 2 {
		utils.Check(cmd.UsageError(""))
	}

	localRepo, err := github.LocalRepo()
	utils.Check(err)

	project, err := localRepo.MainProject()
	utils.Check(err)

	issueNumber, err := strconv.Atoi(args.GetParam(0))
	utils.Check(err)
	targetOwner := project.Owner
	targetRepo := args.GetParam(1)
	if strings.Contains(targetRepo, "/") {
		parts := strings.SplitN(targetRepo, "/", 2)
		targetOwner = parts[0]
		targetRepo = parts[1]
	}

	gh := github.NewClient(project.Host)

	nodeIDsResponse := struct {
		Source struct {
			Issue struct {
				ID string
			}
		}
		Target struct {
			ID string
		}
	}{}
	err = gh.GraphQL(`
	query($issue: Int!, $sourceOwner: String!, $sourceRepo: String!, $targetOwner: String!, $targetRepo: String!) {
		source: repository(owner: $sourceOwner, name: $sourceRepo) {
			issue(number: $issue) {
				id
			}
		}
		target: repository(owner: $targetOwner, name: $targetRepo) {
			id
		}
	}`, map[string]interface{}{
		"issue":       issueNumber,
		"sourceOwner": project.Owner,
		"sourceRepo":  project.Name,
		"targetOwner": targetOwner,
		"targetRepo":  targetRepo,
	}, &nodeIDsResponse)
	utils.Check(err)

	issueResponse := struct {
		TransferIssue struct {
			Issue struct {
				URL string
			}
		}
	}{}
	err = gh.GraphQL(`
	mutation($issue: ID!, $repo: ID!) {
		transferIssue(input: {issueId: $issue, repositoryId: $repo}) {
			issue {
				url
			}
		}
	}`, map[string]interface{}{
		"issue": nodeIDsResponse.Source.Issue.ID,
		"repo":  nodeIDsResponse.Target.ID,
	}, &issueResponse)
	utils.Check(err)

	ui.Println(issueResponse.TransferIssue.Issue.URL)
	args.NoForward()
}
