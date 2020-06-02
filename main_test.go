package main

import (
	"github.com/cucumber/godog"
	suse "github.com/fgerling/bdd-poc/internal/suse"
	"github.com/fgerling/bdd-poc/internal/testrun"
)

func FeatureContext(s *godog.Suite) {
	// Init
	var t testrun.TestRun
	t.InitLogger()
	t.InitKubernetesClient()

	// Kubernetes resources
	s.Step(`^In namespace "([^"]*)" DaemonSet "([^"]*)" should exist$`, t.DaemonSetExists)
	s.Step(`^In namespace "([^"]*)" DaemonSet "([^"]*)" should be ready$`, t.DaemonSetIsReady)
	s.Step(`^In namespace "([^"]*)" Deployment "([^"]*)" should exist$`, t.DeploymentExists)
	s.Step(`^In namespace "([^"]*)" Deployment "([^"]*)" should be ready$`, t.DeploymentIsReady)
	s.Step(`^In namespace "([^"]*)" ConfigMap "([^"]*)" should exist$`, t.ConfigMapExists)
	s.Step(`^In namespace "([^"]*)" no pods with labels "([^"]*)" should exist$`, t.NoPodWithLabelsExist)
	s.Step(`^cilium ConfigMap should have the options:$`, t.CiliumConfigMapDoesHaveTheOptions)
	s.Step(`^cilium ConfigMap should not have the options:$`, t.CiliumConfigMapDoesNotHaveTheOptions)
	s.Step(`^wait in namespace "([^"]*)" for Daemonset "([^"]*)" to be up-to-date$`, t.WaitDaemonSetToBeUpToDate)

	s.Step(`^node "([^"]*)" should be ready$`, t.IsNodeReady)
	s.Step(`^node "([^"]*)" should not be ready$`, t.IsNodeNotReady)

	// Testing workloads
	s.Step(`^httpbin must be ready$`, t.HttpbinMustBeReady)
	s.Step(`^tblshoot must be ready$`, t.TblshootMustBeReady)

	// SSH
	s.Step(`^I run ssh command with user "([^"]*)" on host "([^"]*)" on port (\d+) without sudo the command "([^"]*)"$`, func(user, hostname string, port int, commandPlusArgs string) error {
		return t.ISsh(user, hostname, port, false, commandPlusArgs)
	})

	s.Step(`^I run ssh command with user "([^"]*)" on host "([^"]*)" on port (\d+) with sudo the command "([^"]*)"$`, func(user, hostname string, port int, commandPlusArgs string) error {
		return t.ISsh(user, hostname, port, true, commandPlusArgs)
	})

	s.Step(`^I run ssh command with user "([^"]*)" on host "([^"]*)" on port (\d+) with sudo the command "([^"]*)" and fails$`, func(user, hostname string, port int, commandPlusArgs string) error {
		return testrun.Silent(t.ISsh(user, hostname, port, true, commandPlusArgs))
	})

	// kured
	s.Step(`^I set kured reboot sentinel period to "([^"]*)"$`, t.SetKuredRebootSentinelPeriod)
	s.Step(`^wait for kured Daemonset to be up-to-date$`, t.WaitKuredDaemonSetToBeUpToDate)
	s.Step(`^I get kured pods logs$`, t.GetKuredPodsLogs)

	// Kubernetes client-go
	s.Step(`^I exec in namespace "([^"]*)" in pod "([^"]*)" the command "([^"]*)"$`, t.IExecCommandInPod)
	s.Step(`^I exec in namespace "([^"]*)" in pod "([^"]*)" in container "([^"]*)" the command "([^"]*)"$`, t.IExecCommandInPodContainer)
	s.Step(`^wait for the node "([^"]*)" to be ready$`, t.WaitNodeToBeReady)
	s.Step(`^wait for the node "([^"]*)" not to be ready$`, t.WaitNodeToBeNotReady)
	s.Step(`^wait for the node "([^"]*)" to be schedulable$`, t.WaitNodeToBeSchedulable)
	s.Step(`^wait for the node "([^"]*)" not to be schedulable$`, t.WaitNodeToBeNotSchedulable)

	// kubectl
	s.Step(`^I exec with kubectl in namespace "([^"]*)" in pod "([^"]*)" the command "([^"]*)"$`, t.IKubectlExecCommandInPod)
	s.Step(`^I exec with kubectl in namespace "([^"]*)" in pod "([^"]*)" in container "([^"]*)" the command "([^"]*)"$`, t.IKubectlExecCommandInPodContainer)

	s.Step(`^I apply the manifest "([^"]*)"$`, t.IApplyManifest)
	s.Step(`^I apply the manifest "([^"]*)" and wait for it to be ready$`, t.IApplyManifest)

	// Test resolving http using curl within the tblshoot container
	s.Step(`^I send "([^"]*)" request to "([^"]*)"$`, t.ISendRequestTo)
	s.Step(`^I send "([^"]*)" request to "([^"]*)" and fails$`, func(method, path string) error {
		return testrun.Silent(t.ISendRequestTo(method, path))
	})

	//s.Step(`^there is no resource "([^"]*)" in "([^"]*)" namespace$`, t.ThereIsNoResourceInNamespace)

	// Test resolving dns using dig within the tblshoot container
	s.Step(`^I resolve "([^"]*)"$`, t.IResolve)
	s.Step(`^I resolve "([^"]*)" and fails$`, func(fqdn string) error {
		return testrun.Silent(t.IResolve(fqdn))
	})
	s.Step(`^I reverse resolve "([^"]*)"$`, t.IReverseResolve)
	s.Step(`^I reverse resolve "([^"]*)" and fails$`, func(ip string) error {
		return testrun.Silent(t.IReverseResolve(ip))
	})

	// misc
	s.Step(`^I wait "([^"]*)"$`, testrun.WaitDuration)
	s.Step(`^I install the pattern "([^"]*)"$`, func() error { return t.IRunInDirectory("zypper -n in -t pattern SUSE-CaaSP-Management", ".") })
	s.Step(`^I set "([^"]*)" to "([^"]*)"$`, t.ISetTo)
	s.Step(`^my workstation fulfill the requirements$`, func() error { return t.IRunInDirectory("./check_requirement_workstation.sh", "scripts") })
	s.Step(`^the "([^"]*)" is set to "([^"]*)"$`, t.TheIsSetTo)

	// git
	s.Step(`^I git clone "([^"]*)" into "([^"]*)"$`, t.IGitCloneInto)
	s.Step(`^the "([^"]*)" git repository should exist$`, t.TheRepositoryExist)

	// test file/dir exist
	s.Step(`^I should have "([^"]*)" in PATH$`, suse.IHaveInPATH)
	s.Step(`^the directory "([^"]*)" should exist$`, t.TheDirectoryExist)
	s.Step(`^the file "([^"]*)" should exist$`, t.TheFileExist)
	s.Step(`^there is "([^"]*)" directory$`, t.TheDirectoryExist)
	s.Step(`^there is no "([^"]*)" directory$`, t.ThereIsNoDirectory)

	// test outputs
	s.Step(`^the ([^"]*) should contain "([^"]*)"$`, t.ShouldContain)
	s.Step(`^the output should contain "([^"]*)" and "([^"]*)"$`, t.TheOutputContainsAnd)
	s.Step(`^the output should contain "([^"]*)" or "([^"]*)"$`, t.TheOutputContainsOr)
	s.Step(`^the output should contain a valid ip address$`, t.TheOutputContainsAValidIpAddress)
	s.Step(`^the output shoud match the output the command "([^"]*)"$`, t.TheOutputShoudMatchTheOutputTheCommand)

	// checks related to go
	s.Step(`^"([^"]*)" should exist in gopath$`, t.ExistInGopath)
	s.Step(`^I remove "([^"]*)" from gopath$`, t.IRemoveFromGopath)
	s.Step(`^I have the correct go version$`, func() error { return t.IRunInDirectory("make go-version-check", "skuba") })

	// run local commands
	s.Step(`^I run "([^"]*)" in "([^"]*)" directory$`, t.IRunInDirectory)
	s.Step(`^I run "([^"]*)"$`, func(command string) error { return t.IRunInDirectory(command, ".") })
	s.Step(`^I run "([^"]*)" and fails$`, func(command string) error {
		return testrun.Silent(t.IRunInDirectory(command, "."))
	})
}
