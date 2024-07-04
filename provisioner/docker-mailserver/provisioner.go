// Copyright (c) Jiaqi Liu
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package dockerMailserver

import (
	"context"
	"fmt"
	basicProvisioner "github.com/QubitPi/packer-plugin-hashicorp-aws/provisioner/basic-provisioner"
	fileProvisioner "github.com/QubitPi/packer-plugin-hashicorp-aws/provisioner/file-provisioner"
	sslProvisioner "github.com/QubitPi/packer-plugin-hashicorp-aws/provisioner/ssl-provisioner"
	"github.com/hashicorp/hcl/v2/hcldec"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"path/filepath"
	"strings"
)

type Config struct {
	SslCertBase64    string `mapstructure:"sslCertBase64" required:"true"`
	SslCertKeyBase64 string `mapstructure:"sslCertKeyBase64" required:"true"`
	BaseDomain       string `mapstructure:"baseDomain" required:"true"`
	HomeDir          string `mapstructure:"homeDir" required:"false"`

	ctx interpolate.Context
}

type Provisioner struct {
	config Config
}

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec {
	return p.config.FlatMapstructure().HCL2Spec()
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := config.Decode(&p.config, nil, raws...)
	if err != nil {
		return err
	}

	return nil
}

func (p *Provisioner) Provision(ctx context.Context, ui packersdk.Ui, communicator packersdk.Communicator, generatedData map[string]interface{}) error {
	p.config.HomeDir = sslProvisioner.GetHomeDir(p.config.HomeDir)

	mailServerDomain := "mail." + p.config.BaseDomain

	composeFile := strings.Replace(getDockerComposeFileTemplate(), "mail.domain.com", mailServerDomain, -1)
	composeFileDst := fmt.Sprintf(filepath.Join(p.config.HomeDir, "compose.yaml"))
	composeFileSource, err := sslProvisioner.WriteToFile(composeFile)
	err = fileProvisioner.Provision(p.config.ctx, ui, communicator, composeFileSource, composeFileDst)
	if err != nil {
		return fmt.Errorf("error uploading '%s' to '%s': %s", composeFileSource, composeFileDst, err)
	}

	sslCert, err := sslProvisioner.DecodeBase64(p.config.SslCertBase64)
	sslCertSource, err := sslProvisioner.WriteToFile(sslCert)
	sslCertDestination := fmt.Sprintf(filepath.Join(p.config.HomeDir, "fullchain.pem"))
	err = fileProvisioner.Provision(p.config.ctx, ui, communicator, sslCertSource, sslCertDestination)
	if err != nil {
		return fmt.Errorf("error uploading '%s' to '%s': %s", sslCertSource, sslCertDestination, err)
	}

	sslCertKey, err := sslProvisioner.DecodeBase64(p.config.SslCertKeyBase64)
	sslCertKeySource, err := sslProvisioner.WriteToFile(sslCertKey)
	sslCertKeyDestination := fmt.Sprintf(filepath.Join(p.config.HomeDir, "privkey.pem"))
	err = fileProvisioner.Provision(p.config.ctx, ui, communicator, sslCertKeySource, sslCertKeyDestination)
	if err != nil {
		return fmt.Errorf("error uploading '%s' to '%s': %s", sslCertKeySource, sslCertKeyDestination, err)
	}

	return basicProvisioner.Provision(ctx, ui, communicator, getCommands(p.config.HomeDir, mailServerDomain, sslCertDestination, sslCertKeyDestination))
}

func getDockerComposeFileTemplate() string {
	return `
services:
  mailserver:
    image: ghcr.io/docker-mailserver/docker-mailserver:latest
    container_name: mailserver
    hostname: mail.domain.com
    env_file: mailserver.env
    ports:
      - "25:25"
      - "143:143"
      - "465:465"
      - "587:587"
      - "993:993"
    volumes:
      - ./docker-data/dms/mail-data/:/var/mail/
      - ./docker-data/dms/mail-state/:/var/mail-state/
      - ./docker-data/dms/mail-logs/:/var/log/mail/
      - ./docker-data/dms/config/:/tmp/docker-mailserver/
      - /etc/localtime:/etc/localtime:ro
      - ./docker-data/certbot/certs/:/etc/letsencrypt
    restart: always
    stop_grace_period: 1m
    healthcheck:
      test: "ss --listening --tcp | grep -P 'LISTEN.+:smtp' || exit 1"
      timeout: 3s
      retries: 0
    environment:
      - SSL_TYPE=letsencrypt
    `
}

func getCommands(homeDir string, domain string, sslCertDestination string, sslCertKeyDestination string) []string {
	certsDir := filepath.Join(homeDir, fmt.Sprintf("docker-data/certbot/certs/live/%s", domain))

	return []string{
		"sudo apt update && sudo apt upgrade -y",
		"sudo apt install software-properties-common -y",

		"curl -fsSL https://get.docker.com -o get-docker.sh",
		"sh get-docker.sh",

		fmt.Sprintf("sudo mkdir -p %s", certsDir),
		fmt.Sprintf("sudo mv %s %s", sslCertDestination, certsDir),
		fmt.Sprintf("sudo mv %s %s", sslCertKeyDestination, certsDir),

		"wget \"https://raw.githubusercontent.com/docker-mailserver/docker-mailserver/master/mailserver.env\"",
	}
}
