package testrun

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"

	//"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// GetDaemonSet returns a DaemonSet based on its namespace and name
func (t *TestRun) GetDaemonSet(ns, name string) (*appsv1.DaemonSet, error) {
	ds, err := t.ClientSet.AppsV1().DaemonSets(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return ds, nil
}

func (t *TestRun) WaitDaemonSetToBeUpToDate(ns, name string) error {
	return wait.PollImmediate(APICallRetryInterval, DefaultPollTimeout, func() (bool, error) {
		ds, getErr := t.GetDaemonSet(ns, name)
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

// DaemonSetExists returns an error if DaemonSet does not exist
func (t *TestRun) DaemonSetExists(ns, name string) error {
	_, err := t.GetDaemonSet(ns, name)
	return err
}

// DaemonSetIsReady returns an error if DaemonSet is not ready
func (t *TestRun) DaemonSetIsReady(ns, name string) error {
	ds, err := t.GetDaemonSet(ns, name)
	if err != nil {
		return err
	}

	if ds.Status.DesiredNumberScheduled < ds.Status.NumberReady {
		return fmt.Errorf("Some pods are not ready %d/%d", ds.Status.DesiredNumberScheduled, ds.Status.NumberReady)
	}

	return nil
	//	return IsRuntimeObjectReady(ds)
}
