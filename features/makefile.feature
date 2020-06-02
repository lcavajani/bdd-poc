# doc: https://github.com/SUSE/skuba/blob/master/README.md

Feature: Skuba repository
  Scenario: checkout skuba
    Given there is no "skuba" directory
    When I git clone "https://github.com/SUSE/skuba.git" into "skuba"
    Then the file "skuba/Makefile" should exist

  Scenario: skuba make test
    Given the "skuba" git repository should exist
    When I run "make test" in "skuba"
    Then the file "skuba/coverage.out" should exist

  Scenario: skuba make install
    Given the "skuba" git repository should exist
    And I remove "skuba" from gopath
    #    And the directory "bin" should exist
    #    And I set "GOPATH" to "./bin"
    #    And the "GOPATH" is set to "./bin"
    When I run "make install" in "skuba"
    Then "skuba" should exist in gopath

  Scenario Template: skuba files
    Given the "skuba" git repository should exist
    When I run "ls ." in "skuba"
    Then the file <file> should exist

    Scenarios:
      | file |
      | "skuba/coverage.out" |
      | "skuba/Makefile"     |
      | "skuba/THE_CAKE"     |
