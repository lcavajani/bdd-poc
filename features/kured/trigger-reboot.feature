Feature: Check if reboot triggered
  Scenario: Change reboot sentinel period
    When I set kured reboot sentinel period to "10s"
    And wait for kured Daemonset to be up-to-date
    Then I get kured pods logs
    And the output should contain "Reboot Sentinel: /var/run/reboot-required every 10s"

  Scenario: Trigger a reboot to a node
    #Given there is no kured sentinel file on worker 0
    When I run ssh command with user "sles" on host "10.84.72.233" on port 22 with sudo the command "stat /var/run/reboot-required" and fails
    Then the stderr should contain "No such file or directory"

    When I run ssh command with user "sles" on host "10.84.72.233" on port 22 with sudo the command "touch /var/run/reboot-required"
    #When I create kured sentinel file on worker 0
    Then I wait "15 seconds"

    Then wait for the node "lcavajani-worker-0" not to be schedulable
    #Then wait for the worker 0 not to be schedulable
    And wait for the node "lcavajani-worker-0" not to be ready
    #And wait for the worker 0 not to be ready

    Then wait for the node "lcavajani-worker-0" to be schedulable
    #Then wait for the worker 0 to be schedulable
    And wait for the node "lcavajani-worker-0" to be ready
    #And wait for the worker 0 to be ready

    When I run ssh command with user "sles" on host "10.84.72.233" on port 22 with sudo the command "stat /var/run/reboot-required" and fails
    #Then there is no kured sentinel file on worker 0
    Then the stderr should contain "No such file or directory"
