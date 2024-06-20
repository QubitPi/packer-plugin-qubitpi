// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	sonatypeNexusRepository "github.com/QubitPi/packer-plugin-hashicorp-aws/provisioner/sonatype-nexus-repository"
	"os"

	dockerMailServerProv "github.com/QubitPi/packer-plugin-hashicorp-aws/provisioner/docker-mailserver"
	kongApiGatewayProv "github.com/QubitPi/packer-plugin-hashicorp-aws/provisioner/kong-api-gateway"
	pluginVersion "github.com/QubitPi/packer-plugin-hashicorp-aws/version"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterProvisioner("docker-mailserver-provisioner", new(dockerMailServerProv.Provisioner))
	pps.RegisterProvisioner("kong-api-gateway-provisioner", new(kongApiGatewayProv.Provisioner))
	pps.RegisterProvisioner("sonatype-nexus-repository", new(sonatypeNexusRepository.Provisioner))
	pps.SetVersion(pluginVersion.PluginVersion)
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
