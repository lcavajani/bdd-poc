package testrun

import (
	"fmt"
	"log"
	"os"
	"time"

	//"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (

	// APICallRetryInterval defines how long we wait before retrying a failed API operation (kubeadm const)
	APICallRetryInterval = 500 * time.Millisecond

	DefaultPollTimeout = 5 * time.Minute
)

var (
	tblshoot = map[string]string{
		"pod":       "tblshoot",
		"container": "tblshoot",
		"ns":        metav1.NamespaceDefault,
	}
)

func (t *TestRun) InitKubernetesClient() {
	var err error
	t.RestConfig, err = GetClientRestConfig()
	if err != nil {
		log.Fatal(err)
	}

	t.ClientSet, err = t.CreateClientSet()
	if err != nil {
		log.Fatal(err)
	}
}

func GetKubeconfigPath() (string, error) {
	// Read KUBECONFIG only from env var
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		return "", fmt.Errorf("KUBECONFIG variable not exported")
	}

	return kubeconfig, nil
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
