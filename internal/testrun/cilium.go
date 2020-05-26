package testrun

import (
	"github.com/cucumber/messages-go/v10"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ciliumConfigMap = map[string]string{
		"name":      "cilium-config",
		"namespace": metav1.NamespaceSystem,
	}
)

// CiliumConfigMapDoesHaveTheOptions returns if cilium ConfigMap does have the key/value
// has the key/value provided in the content of a PickleStepArgument_PickleDocString
func (t *TestRun) CiliumConfigMapDoesHaveTheOptions(options *messages.PickleStepArgument_PickleDocString) error {
	return t.ConfigMapDoesHaveTheOptions(ciliumConfigMap["namespace"], ciliumConfigMap["name"], options.Content)
}

// CiliumConfigMapDoesNotHaveTheOptions returns if cilium ConfigMap
// has the key/value provided in the content of a PickleStepArgument_PickleDocString
func (t *TestRun) CiliumConfigMapDoesNotHaveTheOptions(options *messages.PickleStepArgument_PickleDocString) error {
	return t.ConfigMapDoesNotHaveTheOptions(ciliumConfigMap["namespace"], ciliumConfigMap["name"], options.Content)
}
