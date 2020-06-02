package testrun

import (
	"log"
	"os/exec"
	"strings"
	//"k8s.io/apimachinery/pkg/api/errors"
)

var (
	kubectlFlags = map[string]string{
		"allNamespaces": "--all-namespaces",
		"container":     "--container",
		"filename":      "--filename",
		"kubeconfig":    "--kubeconfig",
		"namespace":     "--namespace",
	}

	kubectlActions = map[string]string{
		"apply": "apply",
		"exec":  "exec",
		"path":  "patch",
		"wait":  "wait",
	}
)

const (
	emptyContainerName   string = ""
	defaultWaitCondition string = "condition=Ready"
	defaultWaitTimeout   string = "300s"

	//	httpbinManifestPath  string = "manifests/httpbin/httpbin-pod-svc.yaml"
	//	httpbinManifestPath  string = "manifests/httpbin/httpbin-svc.yaml"
	httpbinManifestPath  string = "manifests/httpbin/httpbin-pod.yaml"
	tblshootManifestPath string = "manifests/tblshoot/tblshoot-pod.yaml"
)

type KubectlOptions struct {
	//	Command       []string
	AllNamespaces      bool   // -A
	Namespace          string // -n default
	Action             string // get
	ResourceKind       string // pod
	IsResourceManifest bool   // -f
	Resource           string // httpbin, manifest url
	Args               []string
}

type KubectlApplyOptions struct {
	Prune bool
}

type KubectlWaitOptions struct {
	Condition string
	Timeout   string
}

type KubectlExecOptions struct {
	Container string
	Command   []string
}

type KubectlPatchOptions struct {
	StrategicMergePatch string
}

func (t *TestRun) RunKubectl(options *KubectlOptions) ([]byte, error) {
	var resourceFlags []string
	kubeconfig, err := GetKubeconfigPath()
	if err != nil {
		log.Fatal(err)
	}

	// NAMESPACES
	var nsFlags []string
	if options.AllNamespaces {
		nsFlags = []string{kubectlFlags["allNamespaces"]}
	}
	if options.Namespace != "" {
		nsFlags = []string{kubectlFlags["namespace"], options.Namespace}
	}

	// we need the filename flag if we wait after a manifest file
	if options.IsResourceManifest {
		resourceFlags = append(resourceFlags, kubectlFlags["filename"], options.Resource)
	} else {
		resourceFlags = append(resourceFlags, options.Resource)
	}

	defaultCmd := []string{
		"kubectl",
		kubectlFlags["kubeconfig"],
		kubeconfig,
		options.Action,
		options.ResourceKind,
	}

	fullCmd := concatStringSlices([][]string{defaultCmd, nsFlags, resourceFlags, options.Args})
	deleteEmtptyInStringSlice(&fullCmd)
	cmd := exec.Command(fullCmd[0], fullCmd[1:]...)
	combinedOutput, err := cmd.CombinedOutput()

	logDebug("run kubectl", strings.Join(fullCmd, " "))
	if err != nil {
		return nil, logError(err, "failed to run kubectl", strings.Join(fullCmd, " "))
	}

	return combinedOutput, err
}

func (t *TestRun) RunKubectlExecInPod(options *KubectlOptions, execOptions *KubectlExecOptions) ([]byte, error) {
	var containerFlags []string

	if execOptions.Container != emptyContainerName {
		containerFlags = []string{kubectlFlags["container"], execOptions.Container}
	}

	options.Action = kubectlActions["exec"]
	options.Args = concatStringSlices([][]string{
		containerFlags,
		[]string{"--"},
		execOptions.Command,
	})

	return t.RunKubectl(options)
}

func (t *TestRun) IKubectlExecCommandInPod(ns, pod, command string) error {
	return t.IKubectlExecCommandInPodContainer(ns, pod, emptyContainerName, command)
}

func (t *TestRun) IKubectlExecCommandInPodContainer(ns, pod, container, command string) (err error) {
	t.CombinedOutput, err = t.RunKubectlExecInPod(
		&KubectlOptions{
			Namespace: ns,
			Resource:  pod},
		&KubectlExecOptions{
			Container: container,
			Command:   strings.Split(command, " ")},
	)

	return err
}

//func (t *TestRun) RunKubectlApply(options *KubectlOptions, manifestPath string) ([]byte, error) {
func (t *TestRun) RunKubectlApply(options *KubectlOptions, applyOptions *KubectlApplyOptions) ([]byte, error) {
	var applyFlags []string

	if applyOptions.Prune {
		applyFlags = append(applyFlags, "--prune")
	}

	options.Action = kubectlActions["apply"]
	options.Args = concatStringSlices([][]string{options.Args, applyFlags})

	return t.RunKubectl(options)
}

func (t *TestRun) RunKubectlPatch(options *KubectlOptions, patchOptions *KubectlPatchOptions) ([]byte, error) {
	options.Action = kubectlActions["patch"]
	return t.RunKubectl(options)
}

func (t *TestRun) RunKubectlWait(options *KubectlOptions, waitOptions *KubectlWaitOptions) ([]byte, error) {

	if waitOptions.Condition == "" {
		waitOptions.Condition = defaultWaitCondition
	}

	if waitOptions.Timeout == "" {
		waitOptions.Timeout = defaultWaitTimeout
	}

	options.Action = kubectlActions["wait"]
	options.Args = append(options.Args, "--for", waitOptions.Condition, "--timeout", waitOptions.Timeout)

	return t.RunKubectl(options)
}

// RunKubectlApplyAndWaitForReady is a wrapper around RunKubectlApply and RunKubectlWait.
// It applies a manifest and wait for a specific condition.
// It returns an error if applying or waiting failed.
func (t *TestRun) RunKubectlApplyAndWaitForReady(options *KubectlOptions, waitOptions *KubectlWaitOptions, applyOptions *KubectlApplyOptions) (err error) {
	t.CombinedOutput, err = t.RunKubectlApply(options, applyOptions)
	if err != nil {
		log.Fatal(err)
	}

	t.CombinedOutput, err = t.RunKubectlWait(options, waitOptions)
	if err != nil {
		log.Fatal(err)
	}

	return err
}

// IApplyManifest applies a manifest using kubectl.
// It returns an error if applying failed.
func (t *TestRun) IApplyManifest(manifestPath string) (err error) {
	t.CombinedOutput, err = t.RunKubectlApply(
		&KubectlOptions{IsResourceManifest: true, Resource: manifestPath},
		&KubectlApplyOptions{},
	)

	return err
}

// IApplyManifestAndWaitForReay applies a manifest using kubectl
// and wait for a specific condition.
// It returns an error if applying failed.
func (t *TestRun) IApplyManifestAndWaitForReady(manifestPath string) (err error) {
	return t.RunKubectlApplyAndWaitForReady(
		&KubectlOptions{IsResourceManifest: true, Resource: manifestPath},
		&KubectlWaitOptions{},
		&KubectlApplyOptions{},
	)

	return err
}

// HttpbinMustBeReady deploys httpbin in the default namespace from a manifest.
// It returns an error if applying or waiting failed.
func (t *TestRun) HttpbinMustBeReady() error {
	return t.RunKubectlApplyAndWaitForReady(
		&KubectlOptions{IsResourceManifest: true, Resource: httpbinManifestPath},
		&KubectlWaitOptions{},
		&KubectlApplyOptions{},
	)
}

// TblshootMustBeReady deploys tblshoot pod in the default namespace from a manifest.
// It returns an error if applying or waiting failed.
func (t *TestRun) TblshootMustBeReady() error {
	return t.RunKubectlApplyAndWaitForReady(
		&KubectlOptions{IsResourceManifest: true, Resource: tblshootManifestPath},
		&KubectlWaitOptions{},
		&KubectlApplyOptions{},
	)
}
