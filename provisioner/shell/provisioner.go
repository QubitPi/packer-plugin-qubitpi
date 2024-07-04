// Copyright (c) Jiaqi Liu
// SPDX-License-Identifier: MPL-2.0

// Package shell This package implements an internal provisioner for Packer that executes a specified list of shell
// commands within the remote machine
package shell

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

func loadCommandsIntoScript(commands []string) (*os.File, error) {
	scriptFile, err := tmp.File("packer-shell")
	if err != nil {
		return nil, fmt.Errorf("error while trying to load commands into a shell script: %s", err)
	}

	writer := bufio.NewWriter(scriptFile)
	writer.WriteString("#!/bin/bash\n")
	writer.WriteString("set -x\n")
	writer.WriteString("set -e\n")
	writer.WriteString("\n")
	for _, command := range commands {
		if _, err := writer.WriteString(command + "\n"); err != nil {
			return nil, fmt.Errorf("error flushing command '%s' into a shell script: %s", command, err)
		}
	}
	if err := writer.Flush(); err != nil {
		return nil, fmt.Errorf("error flushing shell script: %s", err)
	}

	scriptFile.Close()

	return scriptFile, err
}

func executeScript(ctx context.Context, ui packersdk.Ui, communicator packersdk.Communicator, scriptFile *os.File) error {
	f, err := os.Open(scriptFile.Name())
	if err != nil {
		return fmt.Errorf("error opening shell script: %s", err)
	}
	defer f.Close()

	var cmd *packersdk.RemoteCmd
	err = retry.Config{}.Run(ctx, func(ctx context.Context) error {
		if _, err := f.Seek(0, 0); err != nil {
			return err
		}

		remotePath := fmt.Sprintf("%s/%s", "/tmp", fmt.Sprintf("script_%d.sh", rand.Intn(9999)))
		if err := communicator.Upload(remotePath, f, nil); err != nil {
			return fmt.Errorf("error uploading script: %s", err)
		}

		cmd = &packersdk.RemoteCmd{
			Command: fmt.Sprintf("chmod 0755 %s", remotePath),
		}
		if err := communicator.Start(ctx, cmd); err != nil {
			return fmt.Errorf("error chmodding script file to 0755 in remote machine: %s", err)
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

// Provision Batch executes a list of ordered bash shell commands.
//
// It doesn't reuse Packer's original shell provisioner
// (https://github.com/hashicorp/packer/blob/main/provisioner/shell/provisioner.go), which is not fully exported for
// public use due to the unexported config parameter -
// https://github.com/hashicorp/packer/blob/1e446de977e93b7119ebbaa6f55268bd29240e4f/provisioner/shell/provisioner.go#L73
// This parameter is unexported because is not capitalized
//
// This provisioner works by loading all provided commands into a shell script. We do this because executing commands
// separately in the following way simply doesn't work:
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
// each command is executed in a separate shell, meaning their state is not preserved unless the state is flushed to the
// hard disk of the remote machine. For example, the regular env variable export like "export JAVA_HOME=..." won't carry
// over to the next command's execution context. The only way to preserve all in-memory states is to run everything in a
// one-time script, which is how this function is implemented
func Provision(ctx context.Context, ui packersdk.Ui, communicator packersdk.Communicator, commands []string) error {
	scriptFile, err := loadCommandsIntoScript(commands)
	if err != nil {
		return err
	}
	defer os.Remove(scriptFile.Name())

	ui.Say(fmt.Sprintf("Provisioning with %s", commands))

	return executeScript(ctx, ui, communicator, scriptFile)
}
