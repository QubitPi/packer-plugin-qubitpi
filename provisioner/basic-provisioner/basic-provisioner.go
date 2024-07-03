// Copyright (c) Jiaqi Liu
// SPDX-License-Identifier: MPL-2.0

// This package implements a provisioner for Packer that executes a specified list of shell commands within the remote
// machine. It doesn't use Packer's original shell provisioner
// (https://github.com/hashicorp/packer/blob/main/provisioner/shell/provisioner.go), which is essentially not fully
// exported for public use due to the unexported config parameter -
// https://github.com/hashicorp/packer/blob/1e446de977e93b7119ebbaa6f55268bd29240e4f/provisioner/shell/provisioner.go#L73
// This parameter is unexported because is not capitalized
//
// This provisioner works by loading all provided amiConfigCommands into a shell scripts. We do this because executing
// amiConfigCommands separately in the following way simply doesn't work:
//
//	if len(amiConfigCommands) > 0 {
//		   for _, command := range amiConfigCommands {
//		   	   err := (&packersdk.RemoteCmd{Command: command}).RunWithUi(ctx, communicator, ui)
//		   	   if err != nil {
//		   	   	   return fmt.Errorf("CMD error: %s", err)
//		   	   }
//		   }
//	}
//
// because each command is executed in separate shell, meaning their state is not preserved unless the state is written
// to the remote machines hard disk. For example, the regular env variable export like "export JAVA_HOME=..." won't
// carry over to the next command. The only way to preserve all in-memory states is to run everything in a one-time
// script, which is how this provisioner is implemented
package basicProvisioner

import (
	"bufio"
	"context"
	"fmt"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/retry"
	"github.com/hashicorp/packer-plugin-sdk/tmp"
	"math/rand"
	"os"
)

func Provision(ctx context.Context, ui packersdk.Ui, communicator packersdk.Communicator, amiConfigCommands []string) error {
	scriptFile, err := tmp.File("packer-shell")
	if err != nil {
		return fmt.Errorf("Error while trying to load commands into a shell script: %s", err)
	}
	defer os.Remove(scriptFile.Name())

	writer := bufio.NewWriter(scriptFile)
	writer.WriteString("#!/bin/bash\n")
	writer.WriteString("set -x\n")
	writer.WriteString("set -e\n")
	writer.WriteString("\n")
	for _, command := range amiConfigCommands {
		if _, err := writer.WriteString(command + "\n"); err != nil {
			return fmt.Errorf("Error flushing command '%s' into a shell script: %s", command, err)
		}
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("Error flushing shell script: %s", err)
	}

	scriptFile.Close()

	ui.Say(fmt.Sprintf("Provisioning with %s", amiConfigCommands))

	f, err := os.Open(scriptFile.Name())
	if err != nil {
		return fmt.Errorf("Error opening shell script: %s", err)
	}
	defer f.Close()

	var cmd *packersdk.RemoteCmd
	err = retry.Config{}.Run(ctx, func(ctx context.Context) error {
		if _, err := f.Seek(0, 0); err != nil {
			return err
		}

		remotePath := fmt.Sprintf("%s/%s", "/tmp", fmt.Sprintf("script_%d.sh", rand.Intn(9999)))
		if err := communicator.Upload(remotePath, f, nil); err != nil {
			return fmt.Errorf("Error uploading script: %s", err)
		}

		cmd = &packersdk.RemoteCmd{
			Command: fmt.Sprintf("chmod 0755 %s", remotePath),
		}
		if err := communicator.Start(ctx, cmd); err != nil {
			return fmt.Errorf("Error chmodding script file to 0755 in remote machine: %s", err)
		}
		cmd.Wait()

		cmd = &packersdk.RemoteCmd{Command: fmt.Sprintf("chmod +x %s; %s", remotePath, remotePath)}
		return cmd.RunWithUi(ctx, communicator, ui)
	})

	if err != nil {
		return err
	}

	return nil
}
