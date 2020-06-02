package testrun

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"

	//"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetDeployment returns a Deployment based on its namespace and name
func (t *TestRun) GetDeployment(ns, name string) (*appsv1.Deployment, error) {
	dp, err := t.ClientSet.AppsV1().Deployments(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return dp, nil
}

// DeploymentExists returns an error if Deployment does not exist
func (t *TestRun) DeploymentExists(ns, name string) error {
	_, err := t.GetDeployment(ns, name)
	return err
}

// DeploymentIsReady returns an error if Deployment is not ready
func (t *TestRun) DeploymentIsReady(ns, name string) error {
	dp, err := t.GetDeployment(ns, name)
	if err != nil {
		return err
	}

	if dp.Status.Replicas < dp.Status.ReadyReplicas {
		return fmt.Errorf("Some pods are not ready %d/%d", dp.Status.Replicas, dp.Status.ReadyReplicas)
	}

	return nil
}
