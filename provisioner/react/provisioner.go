// Copyright (c) Jiaqi Liu
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package react

import (
	"bytes"
	"context"
	"fmt"
	fileProvisioner "github.com/QubitPi/packer-plugin-hashicorp-aws/provisioner/file-provisioner"
	sslProvisioner "github.com/QubitPi/packer-plugin-hashicorp-aws/provisioner/ssl-provisioner"
	"github.com/hashicorp/hcl/v2/hcldec"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"path/filepath"
	"text/template"
)

// PORT Default port of Sonatype Nexus
const PORT string = "3000"

type Config struct {
	DistSource       string `mapstructure:"distSource" required:"true"`
	SslCertBase64    string `mapstructure:"sslCertBase64" required:"true"`
	SslCertKeyBase64 string `mapstructure:"sslCertKeyBase64" required:"true"`
	AppDomain        string `mapstructure:"appDomain" required:"true"`
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

	distFileDst := fmt.Sprintf(filepath.Join(p.config.HomeDir, "dist"))
	err := fileProvisioner.Provision(p.config.ctx, ui, communicator, p.config.DistSource, distFileDst)
	if err != nil {
		return err
	}

	return sslProvisioner.Provision(ctx, p.config.ctx, ui, communicator, p.config.HomeDir, p.config.SslCertBase64, p.config.SslCertKeyBase64, getNginxConfig(p.config.AppDomain), getCommands(p.config.HomeDir))
}

func getNginxConfig(domain string) string {
	var sslConfigs = struct {
		Domain         string
		SslCertPath    string
		SslCertKeyPath string
		Port           string
	}{domain, sslProvisioner.SSL_CERT_PATH, sslProvisioner.SSL_CERT_KEY_PATH, PORT}
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
        proxy_pass http://localhost:{{.Port}};
    }

    listen [::]:443 ssl ipv6only=on;
    listen 443 ssl;
    ssl_certificate {{.SslCertPath}};
    ssl_certificate_key {{.SslCertKeyPath}};
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
	`))

	if err := t.Execute(&buf, sslConfigs); err != nil {
		panic(err)
	}

	return buf.String()
}

func getCommands(homeDir string) []string {
	return append(
		getCommandsUpdatingUbuntu(),
		append(getCommandsInstallingNode(), getSslCommands(homeDir)...)...,
	)
}

func getCommandsUpdatingUbuntu() []string {
	return []string{
		"sudo apt update && sudo apt upgrade -y",
		"sudo apt install software-properties-common -y",
	}
}

func getCommandsInstallingNode() []string {
	return []string{
		"sudo apt install -y curl",
		"curl -fsSL https://deb.nodesource.com/setup_16.x | sudo -E bash -",
		"sudo apt install -y nodejs",

		"sudo npm install -g yarn",

		"sudo npm install -g serve",
	}
}

// Install and configure Jetty (for JDK 17) container
func getSslCommands(homeDir string) []string {
	return []string{
		"sudo apt install -y nginx",
		fmt.Sprintf("sudo mv %s/nginx-ssl.conf %s", homeDir, sslProvisioner.NGINX_CONFIG_PATH),
		fmt.Sprintf("sudo mv %s/ssl.crt %s", homeDir, sslProvisioner.SSL_CERT_PATH),
		fmt.Sprintf("sudo mv %s/ssl.key %s", homeDir, sslProvisioner.SSL_CERT_KEY_PATH),
	}
}
