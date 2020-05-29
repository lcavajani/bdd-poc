Feature: kured-basic
  Scenario: cilium is properly deployed and working
    Given In namespace "kube-system" DaemonSet "kured" exists
    Then In namespace "kube-system" DaemonSet "kured" should be ready
