package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/github/hub/cmd"
)

const AuthorSignatureHeader = "Signed-off-by: "

var GlobalFlags []string

func Version() (string, error) {
	output, err := gitOutput("version")
	if err != nil {
		return "", fmt.Errorf("Can't load git version")
	}

	return output[0], nil
}

func Dir() (string, error) {
	output, err := gitOutput("rev-parse", "-q", "--git-dir")
	if err != nil {
		return "", fmt.Errorf("Not a git repository (or any of the parent directories): .git")
	}

	gitDir := output[0]
	gitDir, err = filepath.Abs(gitDir)
	if err != nil {
		return "", err
	}

	return gitDir, nil
}

func HasFile(segments ...string) bool {
	dir, err := Dir()
	if err != nil {
		return false
	}

	s := []string{dir}
	s = append(s, segments...)
	path := filepath.Join(s...)
	if _, err := os.Stat(path); err == nil {
		return true
	}

	return false
}

func BranchAtRef(paths ...string) (name string, err error) {
	dir, err := Dir()
	if err != nil {
		return
	}

	segments := []string{dir}
	segments = append(segments, paths...)
	path := filepath.Join(segments...)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	n := string(b)
	refPrefix := "ref: "
	if strings.HasPrefix(n, refPrefix) {
		name = strings.TrimPrefix(n, refPrefix)
		name = strings.TrimSpace(name)
	} else {
		err = fmt.Errorf("No branch info in %s: %s", path, n)
	}

	return
}

func Editor() (string, error) {
	output, err := gitOutput("var", "GIT_EDITOR")
	if err != nil {
		return "", fmt.Errorf("Can't load git var: GIT_EDITOR")
	}

	return output[0], nil
}

func Head() (string, error) {
	return BranchAtRef("HEAD")
}

func SymbolicFullName(name string) (string, error) {
	output, err := gitOutput("rev-parse", "--symbolic-full-name", name)
	if err != nil {
		return "", fmt.Errorf("Unknown revision or path not in the working tree: %s", name)
	}

	return output[0], nil
}

func Ref(ref string) (string, error) {
	output, err := gitOutput("rev-parse", "-q", ref)
	if err != nil {
		return "", fmt.Errorf("Unknown revision or path not in the working tree: %s", ref)
	}

	return output[0], nil
}

func RefList(a, b string) ([]string, error) {
	ref := fmt.Sprintf("%s...%s", a, b)
	output, err := gitOutput("rev-list", "--cherry-pick", "--right-only", "--no-merges", ref)
	if err != nil {
		return []string{}, fmt.Errorf("Can't load rev-list for %s", ref)
	}

	return output, nil
}

func CommentChar() string {
	char, err := Config("core.commentchar")
	if err != nil {
		char = "#"
	}

	return char
}

func Show(sha string) (string, error) {
	cmd := cmd.New("git")
	cmd.WithArg("show").WithArg("-s").WithArg("--format=%s%n%+b").WithArg(sha)

	output, err := cmd.CombinedOutput()
	output = strings.TrimSpace(output)

	return output, err
}

func Log(sha1, sha2 string) (string, error) {
	execCmd := cmd.New("git")
	execCmd.WithArg("log").WithArg("--no-color")
	execCmd.WithArg("--format=%h (%aN, %ar)%n%w(78,3,3)%s%n%+b")
	execCmd.WithArg("--cherry")
	shaRange := fmt.Sprintf("%s...%s", sha1, sha2)
	execCmd.WithArg(shaRange)

	outputs, err := execCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Can't load git log %s..%s", sha1, sha2)
	}

	return outputs, nil
}

func Remotes() ([]string, error) {
	return gitOutput("remote", "-v")
}

func Config(name string) (string, error) {
	return gitGetConfig(name)
}

func BoolConfig(name string) bool {
	v, err := gitGetConfig(name)
	if err != nil {
		return false
	}

	v = strings.ToLower(v)
	if v == "" ||
		v == "true" ||
		v == "yes" ||
		v == "on" {
		return true
	}

	if v == "false" ||
		v == "no" ||
		v == "off" ||
		v[0] == '0' {
		return false
	}

	i, err := strconv.Atoi(v)
	if err != nil {
		return false
	}

	return i > 0
}

func SetConfig(name, value string) error {
	_, err := gitConfig(name, value)
	return err
}

func GlobalConfig(name string) (string, error) {
	return gitGetConfig("--global", name)
}

func SetGlobalConfig(name, value string) error {
	_, err := gitConfig("--global", name, value)
	return err
}

func gitGetConfig(args ...string) (string, error) {
	output, err := gitOutput(gitConfigCommand(args)...)
	if err != nil {
		return "", fmt.Errorf("Unknown config %s", args[len(args)-1])
	}

	return output[0], nil
}

func gitConfig(args ...string) ([]string, error) {
	return gitOutput(gitConfigCommand(args)...)
}

func gitConfigCommand(args []string) []string {
	cmd := []string{"config"}
	return append(cmd, args...)
}

func Alias(name string) (string, error) {
	return Config(fmt.Sprintf("alias.%s", name))
}

func AuthorSignature() (string, error) {
	ident, err := authorIdent()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%s", AuthorSignatureHeader, ident), nil
}

func authorIdent() (string, error) {
	name, err := gitGetConfig("user.name")
	if err != nil {
		name = os.Getenv("GIT_AUTHOR_NAME")
		if name == "" {
			return "", fmt.Errorf("Can't load git config: user.name")
		}
	}

	email, err := gitGetConfig("user.email")
	if err != nil {
		email = os.Getenv("GIT_AUTHOR_EMAIL")
		if email == "" {
			return "", fmt.Errorf("Can't load git config: user.email")
		}
	}

	return fmt.Sprintf("%s <%s>", name, email), nil
}

func Run(command string, args ...string) error {
	cmd := cmd.New("git")

	for _, v := range GlobalFlags {
		cmd.WithArg(v)
	}

	cmd.WithArg(command)

	for _, a := range args {
		cmd.WithArg(a)
	}

	return cmd.Run()
}

func gitOutput(input ...string) (outputs []string, err error) {
	cmd := cmd.New("git")

	for _, v := range GlobalFlags {
		cmd.WithArg(v)
	}

	for _, i := range input {
		cmd.WithArg(i)
	}

	out, err := cmd.CombinedOutput()
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			outputs = append(outputs, string(line))
		}
	}

	return outputs, err
}
