// Copyright (c) Jiaqi Liu
// SPDX-License-Identifier: MPL-2.0

package shell

import (
	"os"
	"testing"
)

func Test_loadCommandsIntoScript(t *testing.T) {
	actualScript, err := loadCommandsIntoScript([]string{
		"sudo apt update && sudo apt upgrade -y",
		"sudo apt install software-properties-common -y",
	})
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(actualScript.Name())

	b, err := os.ReadFile(actualScript.Name())
	if err != nil {
		t.Error(err)
	}

	expectedScript := `#!/bin/bash
set -x
set -e

sudo apt update && sudo apt upgrade -y
sudo apt install software-properties-common -y
`

	if string(b) != expectedScript {
		t.Errorf("Expected and actual scripts do not match: %s\n\n%s", expectedScript, string(b))
	}
}
