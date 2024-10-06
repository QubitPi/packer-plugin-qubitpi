// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"github.com/QubitPi/packer-plugin-qubitpi/provisioner/react"
	artifactory "github.com/QubitPi/packer-plugin-qubitpi/provisioner/sonatype-nexus-repository"
	"github.com/QubitPi/packer-plugin-qubitpi/provisioner/webservice"
	"os"

	mailserver "github.com/QubitPi/packer-plugin-qubitpi/provisioner/docker-mailserver"
	gateway "github.com/QubitPi/packer-plugin-qubitpi/provisioner/kong-api-gateway"
	pluginVersion "github.com/QubitPi/packer-plugin-qubitpi/version"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterProvisioner("docker-mailserver-provisioner", new(mailserver.Provisioner))
	pps.RegisterProvisioner("kong-api-gateway-provisioner", new(gateway.Provisioner))
	pps.RegisterProvisioner("sonatype-nexus-repository-provisioner", new(artifactory.Provisioner))
	pps.RegisterProvisioner("webservice-provisioner", new(webservice.Provisioner))
	pps.RegisterProvisioner("react-provisioner", new(react.Provisioner))
	pps.SetVersion(pluginVersion.PluginVersion)
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
