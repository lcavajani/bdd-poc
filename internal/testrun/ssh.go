package testrun

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

const (
	defaultSshPort = 22
)

type Host struct {
	User      string
	Hostname  string
	Port      int
	Sudo      bool
	ClientSsh *ssh.Client
}

// initClient initializes the ssh client to the host
func (h *Host) initClient() error {
	socket := os.Getenv("SSH_AUTH_SOCK")
	if len(socket) == 0 {
		return errors.New("SSH_AUTH_SOCK is undefined. Make sure ssh-agent is running")
	}

	conn, err := net.Dial("unix", socket)
	if err != nil {
		return err
	}
	agentClient := agent.NewClient(conn)

	// check a precondition: there must be some SSH keys loaded in the ssh agent
	keys, err := agentClient.List()
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		fmt.Println("no keys have been loaded in the ssh-agent.")
		return errors.New("no keys loaded in the ssh-agent")
	}

	config := &ssh.ClientConfig{
		User: h.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeysCallback(agentClient.Signers),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	h.ClientSsh, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", h.Hostname, h.Port), config)
	if err != nil {
		// crypto/ssh does not provide constants for some common errors, so we
		// must "pattern match" the error strings in order to guess what failed
		if strings.Contains(err.Error(), "unable to authenticate") {
			fmt.Println("ssh authentication error: please make sure you have added to "+
				"your ssh-agent a ssh key that is authorized in %q.", h.Hostname)
			return errors.Errorf("authentication error")
		}
		return err
	}
	return nil
}

//TODO: revisit this...
func (t *TestRun) ISsh(user, hostname string, port int, sudo bool, command string, args ...string) (err error) {
	h := Host{User: user, Hostname: hostname, Port: port, Sudo: sudo}
	t.StdOut, t.StdErr, err = h.Ssh(command, args...)
	if err != nil {
		return err
	}

	t.CombinedOutput = concatByteSlices([][]byte{t.StdOut, t.StdErr})
	return nil
}

func (h *Host) Ssh(command string, args ...string) (stdout []byte, stderr []byte, error error) {
	return h.internalSshWithStdin("", command, args...)
}

//func (t *Target) internalSshWithStdin(stdin string, command string, args ...string) (stdout string, stderr string, error error) {
func (h *Host) internalSshWithStdin(stdin string, command string, args ...string) (stdout []byte, stderr []byte, error error) {
	if h.ClientSsh == nil {
		if err := h.initClient(); err != nil {
			return nil, nil, errors.Wrap(err, "failed to initialize client")
		}
	}
	session, err := h.ClientSsh.NewSession()
	if err != nil {
		return nil, nil, err
	}
	if len(stdin) > 0 {
		session.Stdin = bytes.NewBufferString(stdin)
	}
	stdoutReader, err := session.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	stderrReader, err := session.StderrPipe()
	if err != nil {
		return nil, nil, err
	}
	finalCommand := strings.Join(append([]string{command}, args...), " ")
	if h.Sudo {
		finalCommand = fmt.Sprintf("sudo sh -c '%s'", finalCommand)
	}
	//if !silent {
	//	klog.V(2).Infof("running command: %q", finalCommand)
	//}
	if err := session.Start(finalCommand); err != nil {
		return nil, nil, err
	}
	stdoutChan := make(chan []byte)
	stderrChan := make(chan []byte)
	go readerStreamer(stdoutReader, stdoutChan)
	go readerStreamer(stderrReader, stderrChan)
	if err := session.Wait(); err != nil {
		return nil, nil, err
	}
	stdout = <-stdoutChan
	stderr = <-stderrChan
	return stdout, stderr, nil
}

//func readerStreamer(reader io.Reader, outputChan chan<- string, description string, silent bool) {
func readerStreamer(reader io.Reader, outputChan chan<- []byte) {
	result := bytes.Buffer{}
	scanner := bufio.NewScanner(reader)
	// Define a token to split on new lines
	scanner.Split(bufio.ScanBytes)

	for scanner.Scan() {
		result.Write(scanner.Bytes())
		//fmt.Printf(scanner.Text()) -> use for debug
	}

	outputChan <- result.Bytes()
}
