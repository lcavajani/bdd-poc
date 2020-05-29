package testrun

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	//"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

const (

	// APICallRetryInterval defines how long we wait before retrying a failed API operation (kubeadm const)
	APICallRetryInterval = 500 * time.Millisecond

	DefaultPollTimeout = 5 * time.Minute
)

var (
	kindAppsReplicaSet = schema.GroupKind{Group: "apps", Kind: "ReplicaSet"}
	kindAppsDeployment = schema.GroupKind{Group: "apps", Kind: "Deployment"}
	kindAppsDaemonSet  = schema.GroupKind{Group: "apps", Kind: "DaemonSet"}
	kindBatchJob       = schema.GroupKind{Group: "batch", Kind: "Job"}
	kindConfigMap      = schema.GroupKind{Group: "core", Kind: "ConfigMap"}
	kindSecret         = schema.GroupKind{Group: "core", Kind: "Secret"}
	kindService        = schema.GroupKind{Group: "core", Kind: "Service"}

	tblshoot = map[string]string{
		"pod":       "tblshoot",
		"container": "tblshoot",
		"ns":        metav1.NamespaceDefault,
	}
)

func GetKubeconfigPath() (string, error) {
	// Read KUBECONFIG only from env var
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		return "", fmt.Errorf("KUBECONFIG variable not exported")
	}

	return kubeconfig, nil
}

// IExecCommandInPod execute a command inside the first container of a pod.
// It returns an error if the command failed.
func (t *TestRun) IExecCommandInPod(ns, pod, command string) error {
	// empty string means we want to use the first container
	return t.IExecCommandInPodContainer(ns, pod, emptyContainerName, command)
}

// IExecCommandInPodContainer execute a command inside a container of a pod using kubectl
// It returns an error if the command failed.
func (t *TestRun) IExecCommandInPodContainer(ns, pod, container, command string) (err error) {
	t.CombinedOutput, err = t.ExecuteCommandInPodWithCombinedOutput(ns, pod, container, strings.Split(command, " "))

	return err
}

// ISendRequestTo sends http requests within a conatiner using curl.
// The requests are based on a method and path.
// The requests are send from tblshoot pod to httpbin kubernetes service
// It returns an error if the command failed.
func (t *TestRun) ISendRequestTo(method, path string) (err error) {
	cmd := []string{"curl", "-s", "-o", "/dev/null", "-w", "'%{http_code}'", "--connect-timeout", "3", "-X", method, ("http://httpbin" + path)}
	t.CombinedOutput, err = t.ExecuteCommandInPodWithCombinedOutput(tblshoot["ns"], tblshoot["pod"], tblshoot["container"], cmd)

	return err
}

// IResolve resolves an fqdn within a container using dig.
// The resolution is made from tblshoot pod.
// It returns an error if the command failed.
func (t *TestRun) IResolve(fqdn string) (err error) {
	cmd := []string{"dig", "+timeout=2", "+tries=1", "+short", fqdn}
	t.CombinedOutput, err = t.ExecuteCommandInPodWithCombinedOutput(tblshoot["ns"], tblshoot["pod"], tblshoot["container"], cmd)

	return err
}

// IReverseResolve resolves an IP within a container using dig.
// The reverse resolution is made from tblshoot pod.
// It returns an error if the command failed.
func (t *TestRun) IReverseResolve(ip string) (err error) {
	cmd := []string{"dig", "+timeout=2", "+tries=1", "+short", "-x", ip}
	t.CombinedOutput, err = t.ExecuteCommandInPodWithCombinedOutput(tblshoot["ns"], tblshoot["pod"], tblshoot["container"], cmd)

	return err
}

func GetClientRestConfig() (*rest.Config, error) {
	kubeconfig, err := GetKubeconfigPath()
	if err != nil {
		log.Fatal(err)
	}

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// create the clientSet
func (t *TestRun) CreateClientSet() (*kubernetes.Clientset, error) {
	clientSet, err := kubernetes.NewForConfig(t.RestConfig)
	if err != nil {
		return nil, err
	}

	return clientSet, nil
}

func (t *TestRun) GetPodByName(ns, name string) (*corev1.Pod, error) {
	pod, err := t.ClientSet.CoreV1().Pods(ns).Get(name, metav1.GetOptions{})
	//if !(errors.IsNotFound(err)) {
	//	fmt.Printf("Found pod %s in namespace %s\n", name, ns)
	//}
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

// GetDeployment returns a Deployment based on its namespace and name
func (t *TestRun) GetDeployment(ns, name string) (*appsv1.Deployment, error) {
	dp, err := t.GetRuntimeObjectForKind(kindAppsDeployment, ns, name)
	if err != nil {
		return nil, err
	}

	return dp.(*appsv1.Deployment), nil
}

// GetDaemonSet returns a DaemonSet based on its namespace and name
func (t *TestRun) GetDaemonset(ns, name string) (*appsv1.DaemonSet, error) {
	ds, err := t.GetRuntimeObjectForKind(kindAppsDaemonSet, ns, name)
	if err != nil {
		return nil, err
	}

	return ds.(*appsv1.DaemonSet), nil
}

func (t *TestRun) WaitDaemonSetToBeUpToDate(ns, name string) error {
	return wait.PollImmediate(APICallRetryInterval, DefaultPollTimeout, func() (bool, error) {
		ds, getErr := t.GetDaemonset(ns, name)
		if getErr != nil {
			return false, fmt.Errorf("Failed to get latest version of Daemonset: %v", getErr)
		}

		// first check if pods are all updated
		if ds.Status.DesiredNumberScheduled != ds.Status.UpdatedNumberScheduled {
			return false, nil
		}

		// and check if they are ready after
		if err := t.DaemonSetIsReady(ns, name); err != nil {
			return false, nil
		}

		return true, nil
	})
}

// GetConfigMap returns a ConfigMap based on it namespace and name
func (t *TestRun) GetConfigMap(ns, name string) (*corev1.ConfigMap, error) {
	cm, err := t.GetRuntimeObjectForKind(kindConfigMap, ns, name)
	if err != nil {
		return nil, err
	}

	return cm.(*corev1.ConfigMap), nil
}

// DaemonSetExists returns an error if DaemonSet does not exist
func (t *TestRun) DaemonSetExists(ns, name string) error {
	_, err := t.GetDaemonset(ns, name)
	return err
}

// DaemonSetIsReady returns an error if DaemonSet is not ready
func (t *TestRun) DaemonSetIsReady(ns, name string) error {
	ds, err := t.GetRuntimeObjectForKind(kindAppsDaemonSet, ns, name)
	if err != nil {
		return err
	}

	return IsRuntimeObjectReady(ds)
}

// DeploymentExists returns an error if Deployment does not exist
func (t *TestRun) DeploymentExists(ns, name string) error {
	_, err := t.GetDeployment(ns, name)
	return err
}

// DeploymentIsReady returns an error if Deployment is not ready
func (t *TestRun) DeploymentIsReady(ns, name string) error {
	dp, err := t.GetRuntimeObjectForKind(kindAppsDeployment, ns, name)
	if err != nil {
		return err
	}

	return IsRuntimeObjectReady(dp)
}

// ConfigMapExists returns an error if a ConfigMap does not exist
func (t *TestRun) ConfigMapExists(ns, name string) error {
	_, err := t.GetConfigMap(ns, name)
	return err
}

//func (t *TestRun) ConfigMapDoesHaveTheOptions(ns, configMap string, options *messages.PickleStepArgument_PickleDocString) error {
func (t *TestRun) ConfigMapDoesHaveTheOptions(ns, name, optionsJson string) error {
	var expected map[string]string
	c, err := t.GetConfigMap(ns, name)
	if err != nil {
		return err
	}

	if err = json.Unmarshal([]byte(optionsJson), &expected); err != nil {
		return err
	}

	for k, v := range expected {
		if c.Data[k] != v {
			return fmt.Errorf("incorrect option in current config, %v: %v", k, v)
		}
	}

	return nil
}

//func (t *TestRun) ConfigMapDoesNotHaveTheOptions(ns, configMap string, options *messages.PickleStepArgument_PickleDocString) error {
func (t *TestRun) ConfigMapDoesNotHaveTheOptions(ns, name, optionsJson string) error {
	var expected map[string]string
	c, err := t.GetConfigMap(ns, name)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(optionsJson), &expected); err != nil {
		return err
	}

	for k, _ := range expected {
		if _, err := c.Data[k]; err {
			return fmt.Errorf("non expected options exists in current config, %v", k)
		}
	}

	return nil
}

// GetRuntimeObjectForKind returns a runtime.Object based on its GroupKind,
// name and namespace
func (t *TestRun) GetRuntimeObjectForKind(kind schema.GroupKind, ns, name string) (runtime.Object, error) {
	switch kind {
	case kindAppsReplicaSet:
		return t.ClientSet.AppsV1().ReplicaSets(ns).Get(name, metav1.GetOptions{})
	case kindAppsDeployment:
		return t.ClientSet.AppsV1().Deployments(ns).Get(name, metav1.GetOptions{})
	case kindAppsDaemonSet:
		return t.ClientSet.AppsV1().DaemonSets(ns).Get(name, metav1.GetOptions{})
	case kindBatchJob:
		return t.ClientSet.BatchV1().Jobs(ns).Get(name, metav1.GetOptions{})
	case kindConfigMap:
		return t.ClientSet.CoreV1().ConfigMaps(ns).Get(name, metav1.GetOptions{})
	default:
		return nil, fmt.Errorf("Unsupported kind when getting runtime object: %v", kind)
	}
}

// IsRuntimeObjectReady returns if a given object is ready
func IsRuntimeObjectReady(obj runtime.Object) error {
	switch typed := obj.(type) {
	case *appsv1.ReplicaSet:
		if typed.Status.Replicas >= typed.Status.ReadyReplicas {
			return nil
		}
		return fmt.Errorf("Some pods are not ready %d/%d", typed.Status.Replicas, typed.Status.ReadyReplicas)
	case *appsv1.Deployment:
		if typed.Status.Replicas >= typed.Status.ReadyReplicas {
			return nil
		}
		return fmt.Errorf("Some pods are not ready %d/%d", typed.Status.Replicas, typed.Status.ReadyReplicas)
	case *appsv1.DaemonSet:
		if typed.Status.DesiredNumberScheduled >= typed.Status.NumberReady {
			return nil
		}
		return fmt.Errorf("Some pods are not ready %d/%d", typed.Status.DesiredNumberScheduled, typed.Status.NumberReady)
	default:
		return fmt.Errorf("Unsupported kind when getting number of replicas: %v", obj)
	}
}

// GetNode returns a node
func (t *TestRun) GetNode(name string) (*corev1.Node, error) {
	node, err := t.ClientSet.CoreV1().Nodes().Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return node, nil
}

// IsNodeSchedulable returns if a node is Schedulable
func (t *TestRun) IsNodeSchedulable(name string) error {
	node, err := t.GetNode(name)
	if err != nil {
		return err
	}

	if !node.Spec.Unschedulable {
		return nil
	}

	return fmt.Errorf("Node %v is not schedulable", name)
}

func (t *TestRun) WaitNodeToBeSchedulable(name string) error {
	return wait.PollImmediate(APICallRetryInterval, DefaultPollTimeout, func() (bool, error) {
		node, err := t.GetNode(name)
		if err != nil {
			return false, err
		}

		if !node.Spec.Unschedulable {
			return true, nil
		}

		return false, nil
	})
}

// IsNodeUnschedulable returns if a node is Unschedulable
func (t *TestRun) IsNodeUnschedulable(name string) error {
	node, err := t.GetNode(name)
	if err != nil {
		return err
	}

	if node.Spec.Unschedulable {
		return nil
	}

	return fmt.Errorf("Node %v is schedulable", name)
}

func (t *TestRun) WaitNodeToBeNotSchedulable(name string) error {
	return wait.PollImmediate(APICallRetryInterval, DefaultPollTimeout, func() (bool, error) {
		node, err := t.GetNode(name)
		if err != nil {
			return false, err
		}

		if node.Spec.Unschedulable {
			return true, nil
		}

		return false, nil
	})
}

func (t *TestRun) IsNodeReady(name string) error {
	node, err := t.GetNode(name)
	if err != nil {
		return err
	}

	if getNodeReadyStatus(node) {
		return nil
	}

	return fmt.Errorf("Node %v is not ready", name)
}

func (t *TestRun) WaitNodeToBeReady(name string) error {
	return wait.PollImmediate(APICallRetryInterval, DefaultPollTimeout, func() (bool, error) {
		node, err := t.GetNode(name)
		if err != nil {
			return false, err
		}

		if getNodeReadyStatus(node) {
			return true, nil
		}

		return false, nil
	})
}

func getNodeReadyStatus(node *corev1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

func (t *TestRun) IsNodeNotReady(name string) error {
	node, err := t.GetNode(name)
	if err != nil {
		return err
	}

	if !getNodeReadyStatus(node) {
		return nil
	}

	return fmt.Errorf("Node %v is ready", name)
}

func (t *TestRun) WaitNodeToBeNotReady(name string) error {
	return wait.PollImmediate(APICallRetryInterval, DefaultPollTimeout, func() (bool, error) {
		node, err := t.GetNode(name)
		if err != nil {
			return false, err
		}

		if !getNodeReadyStatus(node) {
			return true, nil
		}

		return false, nil
	})
}

//func (t *TestRun) WaitUntilNodeReady(name string) error {
//
//}

func (t *TestRun) ExecuteCommandInPodWithCombinedOutput(ns, pod, container string, cmd []string) ([]byte, error) {
	stdout, stderr, err := t.ExecuteCommandInPod(ns, pod, container, cmd)
	combinedOutput := concatByteSlices([][]byte{stdout, stderr})

	return combinedOutput, err
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
		//		Param("container", pod.Spec.Containers[0].Name)

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

//type Deployment struct {
//	*appsv1.Deployment
//}
//
//func (d Deployment) IsReady() error {
//	if d.Status.Replicas == d.Status.ReadyReplicas {
//		return nil
//	}
//
//	return errors.New(fmt.Sprintf("Some pods are not ready %d/%d", d.Status.Replicas, d.Status.ReadyReplicas))
//}

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
	//str := buf.String()
	//fmt.Println(str)

	return buf.Bytes(), nil
}
