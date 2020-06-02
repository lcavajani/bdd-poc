Feature: kured-basic
  Scenario: kured is properly deployed and ready
    Given In namespace "kube-system" DaemonSet "kured" should exist
    Then In namespace "kube-system" DaemonSet "kured" should be ready
