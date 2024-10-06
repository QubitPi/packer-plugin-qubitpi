// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gateway

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
)

//go:embed test-fixtures/template-aws.pkr.hcl
var testProvisionerHCL2AWS string

//go:embed test-fixtures/template-docker.pkr.hcl
var testProvisionerHCL2Docker string

func TestAccKongAPIGatewayProvisioner(t *testing.T) {
	testCaseAws := &acctest.PluginTestCase{
		Name: "kong_api_gateway_provisioner_aws_test",
		Setup: func() error {
			return nil
		},
		Teardown: func() error {
			return nil
		},
		Template: testProvisionerHCL2AWS,
		Type:     "qubitpi-kong-api-gateway-provisioner",
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}

			logs, err := os.Open(logfile)
			if err != nil {
				return fmt.Errorf("Unable find %s", logfile)
			}
			defer logs.Close()

			logsBytes, err := ioutil.ReadAll(logs)
			if err != nil {
				return fmt.Errorf("Unable to read %s", logfile)
			}
			logsString := string(logsBytes)

			errorString := "\\[ERROR\\] Remote command exited with"
			if matched, _ := regexp.MatchString(".*"+errorString+".*", logsString); matched {
				t.Fatalf("%s\n Acceptance tests for %s failed. Please search for '%s' in log file at %s", logsString, "kong-api-gateway provisioner", errorString, logfile)
			}

			provisionerOutputLog := "amazon-ebs.qubitpi: AMIs were created:"
			if matched, _ := regexp.MatchString(provisionerOutputLog+".*", logsString); !matched {
				t.Fatalf("%s\n logs doesn't contain expected output %q", logsString, provisionerOutputLog)
			}
			return nil
		},
	}
	acctest.TestPlugin(t, testCaseAws)

	testCaseDocker := &acctest.PluginTestCase{
		Name: "kong_api_gateway_provisioner_docker_test",
		Setup: func() error {
			return nil
		},
		Teardown: func() error {
			return nil
		},
		Template: testProvisionerHCL2Docker,
		Type:     "qubitpi-kong-api-gateway-provisioner",
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}

			logs, err := os.Open(logfile)
			if err != nil {
				return fmt.Errorf("Unable find %s", logfile)
			}
			defer logs.Close()

			logsBytes, err := ioutil.ReadAll(logs)
			if err != nil {
				return fmt.Errorf("Unable to read %s", logfile)
			}
			logsString := string(logsBytes)

			errorString := "error(s) occurred"
			if matched, _ := regexp.MatchString(".*"+errorString+".*", logsString); matched {
				t.Fatalf("Acceptance tests for %s failed. Please search for '%s' in log file at %s", "kong-api-gateway-provisioner provisioner", errorString, logfile)
			}

			provisionerOutputLog := "docker.qubitpi: Exported Docker file:"
			if matched, _ := regexp.MatchString(provisionerOutputLog+".*", logsString); !matched {
				t.Fatalf("logs doesn't contain expected output %q", logsString)
			}
			return nil
		},
	}
	acctest.TestPlugin(t, testCaseDocker)
}
