# doc: https://documentation.suse.com/suse-caasp/4.1/single-html/caasp-deployment/#_requirements

Feature: Requirements

  Scenario: Prepare Management Workstation
    Given my workstation fulfill the requirements
    When I install the pattern "SUSE-CaaSP-Management" 
    Then I should have "skuba" in PATH
    And I should have "terraform" in PATH
