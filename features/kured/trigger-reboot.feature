# TC:   https://github.com/fgerling/bdd-poc
# This is a basic test for kured (no PR or BSC provided)

Feature: Check if reboot triggered
#  Scenario: Change reboot sentinel period
#    When I set kured reboot sentinel period to "20s"
#    And wait for kured Daemonset to be up-to-date
#    Then I get kured pod logs
#    And the output should contain "Reboot Sentinel: /var/run/reboot-required every 20s"

  Scenario: Trigger a reboot to a node
    #When I create kured sentinel file on a worker
    When I run ssh command with "sles" user on "10.84.72.233" on port 22 with sudo the command "touch /var/run/reboot-required"
    #Then the worker should have "SchedulingDisabled" condition
    #And the worker should not be ready
    #Then the worker should not have "SchedulingDisabled" condition
    #And the worker should be ready
    #Then the sentinel file does exist on a worker

    #Then a worker is rebooted
    #And the worker is ready
