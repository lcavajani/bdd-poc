package testrun

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	//"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

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
