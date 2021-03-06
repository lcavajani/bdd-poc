package testrun

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
)

func (t *TestRun) ExistInGopath(arg1 string) error {
	return t.TheFileExist(path.Join(os.Getenv("GOPATH"), "bin"))
}

func (t *TestRun) IGitCloneInto(url, target string) error {
	_, err := git.PlainClone(target, false, &git.CloneOptions{
		URL: url,
	})
	return err
}

func (t *TestRun) IRemoveFromGopath(file string) error {
	os.Remove(path.Join(os.Getenv("GOPATH"), "bin", file))
	return nil
}

func (t *TestRun) ISetTo(variable, value string) error {
	os.Setenv(variable, value)
	return t.TheIsSetTo(variable, value)
}

func (t *TestRun) TheIsSetTo(variable, value string) error {
	if os.Getenv(variable) != value {
		return fmt.Errorf("Env %v is not set to %v", variable, value)
	}
	return nil
}

func (t *TestRun) TheGitRepositoryExist(repository string) error {
	return t.TheFileExist(path.Join(repository, ".git/"))
}

func (t *TestRun) TheDirectoryExist(dir string) error {
	return t.TheFileExist(dir)
}

func (t *TestRun) TheFileExist(file string) error {
	_, err := os.Stat(file)
	return err
}

func (t *TestRun) ThereIsNoDirectory(target string) error {
	return os.RemoveAll(target)
}

func (t *TestRun) IRunInDirectory(command, workdir string) error {
	var err error
	args := strings.Split(command, " ")
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = workdir
	t.CombinedOutput, err = cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(t.CombinedOutput))
	}
	return err
}

func (t *TestRun) ShouldContain(output, substr string) error {
	//TODO: show output when debug enable, currently use to check output during dev
	var str string
	switch output {
	case "output":
		str = string(t.CombinedOutput)
	case "stdout":
		str = string(t.StdOut)
	case "stderr":
		str = string(t.StdErr)
	default:
		return errors.New(fmt.Sprintf("%v not recognized, choose between [output|stdout|stderr]", output))
	}

	logDebug("compare strings", fmt.Sprintf("Current: %v\nExpected: %v", str, substr))
	if !strings.Contains(str, substr) {
		return logError(errors.New("string does not contain expected substring"),
			"compare strings",
			fmt.Sprintf("Current: %v\nExpected: %v", str, substr))
	}

	return nil
}

func (t *TestRun) TheOutputContainsAnd(arg1, arg2 string) error {
	if !strings.Contains(fmt.Sprintf("%s", string(t.CombinedOutput)), arg1) && strings.Contains(fmt.Sprintf("%s", string(t.CombinedOutput)), arg2) {
		return errors.New("Output does not contain expected arguments")
	}
	return nil
}

func (t *TestRun) TheOutputContainsOr(arg1, arg2 string) error {
	if strings.Contains(fmt.Sprintf("%s", string(t.CombinedOutput)), arg1) || strings.Contains(fmt.Sprintf("%s", string(t.CombinedOutput)), arg2) {
		return nil
	}
	return errors.New("Output does not contain expected arguments")
}

func (t *TestRun) TheOutputContainsAValidIpAddress() error {
	IP := net.ParseIP(string(t.CombinedOutput))
	if IP == nil {
		return errors.New(fmt.Sprintf("%s is not a valid textual representation of an IP address", string(t.CombinedOutput)))
	}
	return nil
}

func (t *TestRun) TheOutputShoudMatchTheOutputTheCommand(command2 string) error {
	args := strings.Split(command2, " ")
	cmd := exec.Command(args[0], args[1:]...)
	cmd2Output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(cmd2Output))
	}
	return t.ShouldContain("output", string(cmd2Output))
}

func WaitDuration(duration string) (err error) {
	temp := strings.Split(duration, " ")
	if len(temp) != 2 {
		return fmt.Errorf("Sorry... you've mistaken the format of time input (it's <NUMBER><1*EMPTYSPACE><WORD[seconds:minutes:hours]>")
	}

	switch temp[1] {
	case "seconds":
		d, err := strconv.Atoi(temp[0])
		if err != nil {
			return err
		}
		time.Sleep(time.Duration(d) * time.Second)
	case "minutes":
		d, err := strconv.Atoi(temp[0])
		if err != nil {
			return err
		}
		time.Sleep(time.Duration(d) * time.Minute)
	case "hours":
		d, err := strconv.Atoi(temp[0])
		if err != nil {
			return err
		}
		time.Sleep(time.Duration(d) * time.Hour)
	}

	return nil
}

// Silent returns nil if it get an error as a parameter
// and an error if the parameter is not nil
func Silent(err error) error {
	if err == nil {
		return fmt.Errorf("command was supposed to fail")
	}

	return nil
}

// concatByteSlices contats slices of []Byte into one
func concatByteSlices(slices [][]byte) []byte {
	var totalLen int
	for _, f := range slices {
		totalLen += len(f)
	}

	tmp := make([]byte, totalLen)

	var i int
	for _, f := range slices {
		i += copy(tmp[i:], f)
	}
	return tmp
}

// concatStringSlices contats slices of []string into one
func concatStringSlices(slices [][]string) []string {
	var totalLen int
	for _, f := range slices {
		totalLen += len(f)
	}

	tmp := make([]string, totalLen)

	var i int
	for _, f := range slices {
		i += copy(tmp[i:], f)
	}
	return tmp
}

// deleteEmtptyInStringSlice deletes empty slices from a string slice
func deleteEmtptyInStringSlice(s *[]string) {
	var tmp []string
	for _, str := range *s {
		if str != "" {
			tmp = append(tmp, str)
		}
	}
	*s = tmp
}
