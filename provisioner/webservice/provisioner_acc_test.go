// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package webservice

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
)

//go:embed test-fixtures/template.pkr.hcl
var testProvisionerHCL2Basic string

func TestAccWebserviceProvisioner(t *testing.T) {
	tempFile, err := os.CreateTemp(t.TempDir(), "my-webservice.war")
	if err != nil {
		return
	}

	testCase := &acctest.PluginTestCase{
		Name: "webservice_provisioner_basic_test",
		Setup: func() error {
			return nil
		},
		Teardown: func() error {
			return nil
		},
		Template: strings.Replace(testProvisionerHCL2Basic, "my-webservice.war", tempFile.Name(), -1),
		Type:     "hashicorp-aws-webservice-provisioner",
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
				t.Fatalf("Acceptance tests for %s failed. Please search for '%s' in log file at %s", "webservice provisioner", errorString, logfile)
			}

			provisionerOutputLog := "amazon-ebs.hashicorp-aws: AMIs were created:"
			if matched, _ := regexp.MatchString(provisionerOutputLog+".*", logsString); !matched {
				t.Fatalf("logs doesn't contain expected output %q", logsString)
			}

			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}
