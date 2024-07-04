// Copyright (c) Jiaqi Liu
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package kongApiGateway

import (
	"bytes"
	"context"
	sslProvisioner "github.com/QubitPi/packer-plugin-hashicorp-aws/provisioner/ssl-provisioner"
	"github.com/hashicorp/hcl/v2/hcldec"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"text/template"
)

type Config struct {
	SslCertBase64        string `mapstructure:"sslCertBase64" required:"true"`
	SslCertKeyBase64     string `mapstructure:"sslCertKeyBase64" required:"true"`
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
	p.config.HomeDir = sslProvisioner.GetHomeDir(p.config.HomeDir)
	return sslProvisioner.Provision(ctx, p.config.ctx, ui, communicator, p.config.HomeDir, p.config.SslCertBase64, p.config.SslCertKeyBase64, getNginxConfig(p.config.KongApiGatewayDomain), getCommands(p.config.HomeDir))
}

func getCommands(homeDir string) []string {
	return []string{
		"sudo apt update && sudo apt upgrade -y",
		"sudo apt install software-properties-common -y",

		"curl -fsSL https://get.docker.com -o get-docker.sh",
		"sh get-docker.sh",

		"git clone https://github.com/QubitPi/docker-kong.git",
	}
}

func getNginxConfig(domain string) string {
	var sslConfigs = struct {
		Domain        string
		SslCertDst    string
		SslCertKeyDst string
	}{domain, sslProvisioner.SslCertDst, sslProvisioner.SslCertKeyDst}
	var buf bytes.Buffer
	t := template.Must(template.New("Nginx Config").Parse(`
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
    server_name {{.Domain}};

    location / {
        proxy_pass http://localhost:8000;
    }

    listen [::]:443 ssl ipv6only=on;
    listen 443 ssl;
    ssl_certificate {{.SslCertDst}};
    ssl_certificate_key {{.SslCertKeyDst}};
}
server {
    if ($host = {{.Domain}}) {
        return 301 https://$host$request_uri;
    }

    listen 80 ;
    listen [::]:80 ;
    server_name {{.Domain}};
    return 404;
}

server {
    root /var/www/html;

    index index.html index.htm index.nginx-debian.html;
    server_name {{.Domain}};

    location / {
        proxy_pass http://localhost:8001;
    }

    listen [::]:8444 ssl ipv6only=on;
    listen 8444 ssl;
    ssl_certificate {{.SslCertDst}};
    ssl_certificate_key {{.SslCertKeyDst}};
}
server {
    root /var/www/html;

    index index.html index.htm index.nginx-debian.html;
    server_name {{.Domain}};

    location / {
        proxy_pass http://localhost:8002;
    }

    listen [::]:8445 ssl ipv6only=on;
    listen 8445 ssl;
    ssl_certificate {{.SslCertDst}};
    ssl_certificate_key {{.SslCertKeyDst}};
}
	`))

	if err := t.Execute(&buf, sslConfigs); err != nil {
		panic(err)
	}

	return buf.String()
}
