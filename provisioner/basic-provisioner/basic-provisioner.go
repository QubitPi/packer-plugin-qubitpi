// Copyright (c) Jiaqi Liu
// SPDX-License-Identifier: MPL-2.0

// This package implements a provisioner for Packer that executes a specified list of shell commands within the remote
// machine
package basicProvisioner

import (
	"context"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func Provision(ctx context.Context, ui packersdk.Ui, communicator packersdk.Communicator, amiConfigCommands []string) error {
	if len(amiConfigCommands) > 0 {
		for _, command := range amiConfigCommands {
			err := (&packersdk.RemoteCmd{Command: command}).RunWithUi(ctx, communicator, ui)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
