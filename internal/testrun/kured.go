package testrun

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

var (
	kuredDaemonset = map[string]string{
		"name":      "kured",
		"namespace": metav1.NamespaceSystem,
		"podLabels": "name=kured",
	}
)

// SetKuredRebootSentinelPeriod will change the reboot sentinel period
// in the kured DaemonSet
func (t *TestRun) SetKuredRebootSentinelPeriod(period string) error {
	dsClient := t.ClientSet.AppsV1().DaemonSets(kuredDaemonset["namespace"])

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		ds, getErr := t.GetDaemonSet(kuredDaemonset["namespace"], kuredDaemonset["name"])
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

// WaitKuredDaemonSetToBeUpToDate is a wrapper to wait for
// kured DaemonSet to be up-to-date. This is used for example after
// changing the reboot sentinel period
func (t *TestRun) WaitKuredDaemonSetToBeUpToDate() error {
	return t.WaitDaemonSetToBeUpToDate(kuredDaemonset["namespace"], kuredDaemonset["name"])
}

// GetKuredPodLogs returns the logs of every kured pod in the TestRun CombinedOutput field
func (t *TestRun) GetKuredPodsLogs() (err error) {
	//ds, getErr := t.GetDaemonset(kuredDaemonset["namespace"], kuredDaemonset["name"])
	_, getErr := t.GetDaemonSet(kuredDaemonset["namespace"], kuredDaemonset["name"])
	if getErr != nil {
		return fmt.Errorf("Failed to get latest version of DaemonSet: %v", getErr)
	}

	t.CombinedOutput, err = t.GetPodsLogsByLabels(kuredDaemonset["namespace"], kuredDaemonset["podLabels"])
	if err != nil {
		return err
	}

	return nil
}
