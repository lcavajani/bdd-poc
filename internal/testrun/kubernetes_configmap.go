package testrun

import (
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"

	//"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetConfigMap returns a ConfigMap based on it namespace and name
func (t *TestRun) GetConfigMap(ns, name string) (*corev1.ConfigMap, error) {
	cm, err := t.ClientSet.CoreV1().ConfigMaps(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return cm, nil
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
	c, err := t.ClientSet.CoreV1().ConfigMaps(ns).Get(name, metav1.GetOptions{})
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
