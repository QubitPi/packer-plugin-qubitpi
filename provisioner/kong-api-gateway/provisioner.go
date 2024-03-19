// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package kongApiGateway

import (
	"context"
	"fmt"
	"github.com/QubitPi/packer-plugin-hashicorp-aws/provisioner"
	"github.com/hashicorp/hcl/v2/hcldec"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer-plugin-sdk/tmp"
	"strings"
)

type Config struct {
	SslCertSource      string `mapstructure:"sslCertSource" required:"true"`
	SslCertDestination string `mapstructure:"sslCertDestination" required:"false"`

	SslCertKeySource      string `mapstructure:"sslCertKeySource" required:"true"`
	SslCertKeyDestination string `mapstructure:"sslCertKeyDestination" required:"false"`

	KongApiGatewayDomain string `mapstructure:"kongApiGatewayDomain" required:"true"`
	HomeDir              string `mapstructure:"homeDir" required:"false"`

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
	if p.config.HomeDir == "" {
		p.config.HomeDir = "/home/ubuntu"
	}

	if p.config.SslCertDestination == "" {
		p.config.SslCertDestination = fmt.Sprintf("%s/ssl.crt", p.config.HomeDir)
	}
	err := p.ProvisionUpload(ui, communicator, p.config.SslCertSource, p.config.SslCertDestination)
	if err != nil {
		return fmt.Errorf("error uploading '%s' to '%s': %s", p.config.SslCertSource, p.config.SslCertDestination, err)
	}

	if p.config.SslCertKeyDestination == "" {
		p.config.SslCertDestination = fmt.Sprintf("%s/ssl.key", p.config.HomeDir)
	}
	err = p.ProvisionUpload(ui, communicator, p.config.SslCertKeySource, p.config.SslCertKeyDestination)
	if err != nil {
		return fmt.Errorf("error uploading '%s' to '%s': %s", p.config.SslCertKeySource, p.config.SslCertKeyDestination, err)
	}

	nginxConfig := strings.Replace(getNginxConfigTemplate(), "kong.domain.com", p.config.KongApiGatewayDomain, -1)
	file, err := tmp.File("nginx-config-file")
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.WriteString(nginxConfig); err != nil {
		return err
	}
	nginxConfig = ""
	err = p.ProvisionUpload(ui, communicator, file.Name(), "/home/ubuntu/nginx-ssl.conf")
	if err != nil {
		return fmt.Errorf("error uploading '%s' to '%s': %s", file.Name(), "/home/ubuntu/nginx-ssl.conf", err)
	}
	for _, command := range getCommands(p.config.HomeDir) {
		err := (&packersdk.RemoteCmd{Command: command}).RunWithUi(ctx, communicator, ui)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Provisioner) ProvisionUpload(ui packersdk.Ui, communicator packersdk.Communicator, source string, destination string) error {
	src, err := interpolate.Render(source, &p.config.ctx)
	if err != nil {
		return fmt.Errorf("error interpolating source: %s", err)
	}

	dst, err := interpolate.Render(destination, &p.config.ctx)
	if err != nil {
		return fmt.Errorf("error interpolating destination: %s", err)
	}

	return provisioner.ProvisionUpload(ui, communicator, src, dst)
}

func getCommands(homeDir string) []string {
	return []string{
		"sudo apt update && sudo apt upgrade -y",
		"sudo apt install software-properties-common -y",

		"curl -fsSL https://get.docker.com -o get-docker.sh",
		"sh get-docker.sh",

		"git clone https://github.com/QubitPi/docker-kong.git",

		"sudo apt install -y nginx",
		fmt.Sprintf("sudo mv %s/nginx-ssl.conf /etc/nginx/sites-enabled/default", homeDir),
		fmt.Sprintf("sudo mv %s/ssl.crt /etc/ssl/certs/server.crt", homeDir),
		fmt.Sprintf("sudo mv %s/ssl.key /etc/ssl/private/server.key", homeDir),
	}
}

func getNginxConfigTemplate() string {
	return `
server {
    listen 80 default_server;
    listen [::]:80 default_server;

    root /var/www/html;

    index index.html index.htm index.nginx-debian.html;

    server_name _;

    location / {
        try_files $uri $uri/ =404;
    }
}

server {
    root /var/www/html;

    index index.html index.htm index.nginx-debian.html;
    server_name kong.domain.com;

    location / {
        proxy_pass http://localhost:8000;
    }

    listen [::]:443 ssl ipv6only=on;
    listen 443 ssl;
    ssl_certificate /etc/ssl/certs/server.crt;
    ssl_certificate_key /etc/ssl/private/server.key;
}
server {
    if ($host = kong.domain.com) {
        return 301 https://$host$request_uri;
    }

    listen 80 ;
    listen [::]:80 ;
    server_name kong.domain.com;
    return 404;
}

server {
    root /var/www/html;

    index index.html index.htm index.nginx-debian.html;
    server_name kong.domain.com;

    location / {
        proxy_pass http://localhost:8001;
    }

    listen [::]:8444 ssl ipv6only=on;
    listen 8444 ssl;
    ssl_certificate /etc/ssl/certs/server.crt;
    ssl_certificate_key /etc/ssl/private/server.key;
}
server {
    root /var/www/html;

    index index.html index.htm index.nginx-debian.html;
    server_name kong.domain.com;

    location / {
        proxy_pass http://localhost:8002;
    }

    listen [::]:8445 ssl ipv6only=on;
    listen 8445 ssl;
    ssl_certificate /etc/ssl/certs/server.crt;
    ssl_certificate_key /etc/ssl/private/server.key;
}
    `
}
