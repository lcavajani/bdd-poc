Feature: cilium-basic
  Scenario: cilium is properly deployed and ready
    Given In namespace "kube-system" DaemonSet "cilium" should exist
    Then In namespace "kube-system" DaemonSet "cilium" should be ready
    Given In namespace "kube-system" Deployment "cilium-operator" should exist
    Then In namespace "kube-system" Deployment "cilium-operator" should be ready

    When I exec with kubectl in namespace "kube-system" in pod "ds/cilium" the command "cilium version"
    Then the output should contain "Client: 1.6.6"
    And the output should contain "Daemon: 1.6.6"

  Scenario: no leftovers from a migration
    Given In namespace "kube-system" no pods with labels "k8s-app=cilium-pre-flight-check" should exist

  Scenario: cilium uses CRD instead of etcd
    Given In namespace "kube-system" ConfigMap "cilium-config" should exist
    Then cilium ConfigMap should have the options:
      """
      {
        "bpf-ct-global-any-max": "262144",
        "bpf-ct-global-tcp-max": "524288",
        "debug": "false",
        "enable-ipv4": "true",
        "enable-ipv6": "false",
        "identity-allocation-mode": "crd",
        "preallocate-bpf-maps": "false"
      }
      """
    Then cilium ConfigMap should not have the options:
      """
      {
        "etcd-config": "",
        "kvstore": "",
        "kvstore-opt": ""
      }
      """
