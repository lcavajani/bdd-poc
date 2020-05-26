package testrun

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
)

var (
	kuredDaemonset = map[string]string{
		"name":      "kured",
		"namespace": metav1.NamespaceSystem,
		"podLabels": "name=kured",
	}
)

func (t *TestRun) SetKuredRebootSentinelPeriod(period string) error {
	dsClient := t.ClientSet.AppsV1().DaemonSets(kuredDaemonset["namespace"])

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		ds, getErr := t.GetDaemonset(kuredDaemonset["namespace"], kuredDaemonset["name"])
		if getErr != nil {
			return fmt.Errorf("Failed to get latest version of DaemonSet: %v", getErr)
		}

		// Update period
		ds.Spec.Template.Spec.Containers[0].Command = []string{"/usr/bin/kured", ("--period=" + period)}
		_, updateErr := dsClient.Update(ds)
		return updateErr
	})

	if retryErr != nil {
		return fmt.Errorf("Update failed: %v", retryErr)
	}

	fmt.Println("Updated daemonset...")

	return nil
}

func (t *TestRun) WaitKuredDaemonSetToBeUpToDate() error {
	return t.WaitDaemonSetToBeUpToDate(kuredDaemonset["namespace"], kuredDaemonset["name"])
}

func (t *TestRun) WaitDaemonSetToBeUpToDate(ns, name string) error {
	return wait.PollImmediate(APICallRetryInterval, Timeout, func() (bool, error) {
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

func (t *TestRun) GetKuredPodLogs() (err error) {
	//ds, getErr := t.GetDaemonset(kuredDaemonset["namespace"], kuredDaemonset["name"])
	_, getErr := t.GetDaemonset(kuredDaemonset["namespace"], kuredDaemonset["name"])
	if getErr != nil {
		return fmt.Errorf("Failed to get latest version of DaemonSet: %v", getErr)
	}

	t.CombinedOutput, err = t.GetPodsLogsByLabels(kuredDaemonset["namespace"], kuredDaemonset["podLabels"])
	if err != nil {
		return err
	}

	return nil
}
