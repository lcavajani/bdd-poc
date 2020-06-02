package testrun

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"

	corev1 "k8s.io/api/core/v1"

	//"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

// IExecCommandInPod execute a command inside the first container of a pod.
// It returns an error if the command failed.
func (t *TestRun) IExecCommandInPod(ns, pod, command string) error {
	// empty string means we want to use the first container
	return t.IExecCommandInPodContainer(ns, pod, emptyContainerName, command)
}

// IExecCommandInPodContainer execute a command inside a container of a pod using kubectl
// It returns an error if the command failed.
func (t *TestRun) IExecCommandInPodContainer(ns, pod, container, command string) (err error) {
	t.CombinedOutput, t.StdOut, t.StdErr, err = t.ExecuteCommandInPodWithCombinedOutput(ns, pod, container, strings.Split(command, " "))

	return err
}

// ISendRequestTo sends http requests within a conatiner using curl.
// The requests are based on a method and path.
// The requests are send from tblshoot pod to httpbin kubernetes service
// It returns an error if the command failed.
func (t *TestRun) ISendRequestTo(method, path string) (err error) {
	cmd := []string{"curl", "-s", "-o", "/dev/null", "-w", "'%{http_code}'", "--connect-timeout", "2", "-X", method, ("http://httpbin" + path)}
	t.CombinedOutput, t.StdOut, t.StdErr, err = t.ExecuteCommandInPodWithCombinedOutput(tblshoot["ns"], tblshoot["pod"], tblshoot["container"], cmd)
	return err
}

// IResolve resolves an fqdn within a container using dig.
// The resolution is made from tblshoot pod.
// It returns an error if the command failed.
func (t *TestRun) IResolve(fqdn string) (err error) {
	cmd := []string{"dig", "+timeout=2", "+tries=1", "+short", fqdn}
	t.CombinedOutput, t.StdOut, t.StdErr, err = t.ExecuteCommandInPodWithCombinedOutput(tblshoot["ns"], tblshoot["pod"], tblshoot["container"], cmd)
	return err
}

// IReverseResolve resolves an IP within a container using dig.
// The reverse resolution is made from tblshoot pod.
// It returns an error if the command failed.
func (t *TestRun) IReverseResolve(ip string) (err error) {
	cmd := []string{"dig", "+timeout=2", "+tries=1", "+short", "-x", ip}
	t.CombinedOutput, t.StdOut, t.StdErr, err = t.ExecuteCommandInPodWithCombinedOutput(tblshoot["ns"], tblshoot["pod"], tblshoot["container"], cmd)

	return err
}

func (t *TestRun) GetPodByName(ns, name string) (*corev1.Pod, error) {
	pod, err := t.ClientSet.CoreV1().Pods(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return pod, nil
}

// GetPodsByLabels returns a list of pods based on their namespace and labels
func (t *TestRun) GetPodsByLabels(ns, labels string) (*corev1.PodList, error) {
	pods, err := t.ClientSet.CoreV1().Pods(ns).List(metav1.ListOptions{
		LabelSelector: labels,
	})

	if err != nil {
		return nil, err
	}

	return pods, nil
}

// PodWithLabelsExist returns an error if a pod does not exist
// based on its namespace and labels
func (t *TestRun) PodWithLabelsExist(ns, labels string) error {
	pods, err := t.GetPodsByLabels(ns, labels)
	if err != nil {
		return err
	}

	if len(pods.Items) == 0 {
		return fmt.Errorf("no pods in namespace, %v with labels, %v exist ", ns, labels)
	}

	return nil
}

/// NoPodWithLabelsExist returns an error if a pod exists
// based on its namespace and labels
func (t *TestRun) NoPodWithLabelsExist(ns, labels string) error {
	pods, err := t.GetPodsByLabels(ns, labels)
	if err != nil {
		return err
	}

	if len(pods.Items) != 0 {
		return fmt.Errorf("pods in namespace, %v with labels, %v exist ", ns, labels)
	}

	return nil
}

func (t *TestRun) ExecuteCommandInPodWithCombinedOutput(ns, pod, container string, cmd []string) ([]byte, []byte, []byte, error) {
	stdout, stderr, err := t.ExecuteCommandInPod(ns, pod, container, cmd)
	combinedOutput := concatByteSlices([][]byte{stdout, stderr})

	return combinedOutput, stdout, stderr, err
}

func (t *TestRun) ExecuteCommandInPod(ns, pod, container string, cmd []string) ([]byte, []byte, error) {
	var tty = true
	var stdin io.Reader

	if container != "" {
		pod, err := t.GetPodByName(ns, pod)
		if err != nil {
			log.Fatal(err)
		}
		container = pod.Spec.Containers[0].Name
	}

	var stdout, stderr bytes.Buffer

	req := t.ClientSet.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod).
		Namespace(ns).
		SubResource("exec").
		Param("container", container)

	options := &corev1.PodExecOptions{
		Container: container,
		Command:   cmd,
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       tty,
	}

	req.VersionedParams(options, scheme.ParameterCodec)
	err := t.executeCmd("POST", req.URL(), stdin, &stdout, &stderr, tty)

	return stdout.Bytes(), stderr.Bytes(), err
}

func (t *TestRun) executeCmd(method string, url *url.URL, stdin io.Reader, stdout, stderr io.Writer, tty bool) error {
	exec, err := remotecommand.NewSPDYExecutor(t.RestConfig, "POST", url)
	if err != nil {
		return err
	}

	return exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		Tty:    tty,
	})
}

// GetPodsLogsByLabels get the logs from all the pods
// maching some labels within a namespace
func (t *TestRun) GetPodsLogsByLabels(ns, labels string) ([]byte, error) {
	pods, err := t.GetPodsByLabels(ns, labels)
	if err != nil {
		return nil, err
	}

	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("no pods in namespace, %v with labels, %v exist ", ns, labels)
	}

	var podsLogs []byte

	for _, pod := range pods.Items {
		podLogs, getLogErr := t.GetPodLogs(ns, pod.ObjectMeta.Name, pod.Spec.Containers[0].Name)
		if getLogErr != nil {
			return nil, getLogErr
		}

		podsLogs = concatByteSlices([][]byte{podsLogs, podLogs})
	}

	return podsLogs, nil
}

// GetPodLogs get the logs from a container in a pod within a namespace
func (t *TestRun) GetPodLogs(ns, pod, container string) ([]byte, error) {
	if container != "" {
		pod, err := t.GetPodByName(ns, pod)
		if err != nil {
			log.Fatal(err)
		}
		container = pod.Spec.Containers[0].Name
	}

	req := t.ClientSet.CoreV1().RESTClient().Get().
		Namespace(ns).
		Name(pod).
		Resource("pods").
		SubResource("log").
		Param("container", container).
		VersionedParams(&corev1.PodLogOptions{Container: container}, scheme.ParameterCodec)

	podLogs, err := req.Stream()
	if err != nil {
		return nil, fmt.Errorf("error in opening stream")
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return nil, fmt.Errorf("error in copy information from podLogs to buf")
	}

	return buf.Bytes(), nil
}
