// Copyright (c) Jiaqi Liu
// SPDX-License-Identifier: MPL-2.0

package react

import (
	"reflect"
	"testing"
)

func Test_getCommandsInstallingNode(t *testing.T) {
	actualCommands := getCommandsInstallingNode("18")

	expectedCommands := []string{
		"sudo apt install -y curl",
		"curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -",
		"sudo apt install -y nodejs",

		"sudo npm install -g yarn",

		"sudo npm install -g serve",
	}

	if !reflect.DeepEqual(expectedCommands, actualCommands) {
		t.Errorf("Expected and actual commands do not match: %s\n\n%s", expectedCommands, actualCommands)
	}
}
