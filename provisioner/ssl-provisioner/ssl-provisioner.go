// Copyright (c) Jiaqi Liu
// SPDX-License-Identifier: MPL-2.0

package sslProvisioner

import (
	"context"
	"encoding/base64"
	"fmt"
	basicProvisioner "github.com/QubitPi/packer-plugin-hashicorp-aws/provisioner/basic-provisioner"
	fileProvisioner "github.com/QubitPi/packer-plugin-hashicorp-aws/provisioner/file-provisioner"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer-plugin-sdk/tmp"
	"path/filepath"
)

const SSL_CERT_PATH string = "/etc/ssl/certs/server.crt"
const SSL_CERT_KEY_PATH string = "/etc/ssl/private/server.key"
const NGINX_CONFIG_PATH string = "/etc/nginx/sites-enabled/default"
const DEFAULT_HOME_DIR string = "/home/ubuntu"

// GetHomeDir Returns the home directory in Packer image builder. If a directory is specified, it is returned as it;
// otherwise the default Ubuntu home "/home/ubuntu" is returned
//
// configValue: A directory that can be either empty or a valid directory to be used unchanged.
//
// Returns:
// The actual home directory of the running Pakcer image builder
func GetHomeDir(configValue string) string {
	if configValue == "" {
		return DEFAULT_HOME_DIR
	}

	return configValue
}

// WriteToFile Flushes a specified string into a temporary file and returns the path of that file.
//
// content: The provided file content
//
// Returns:
// A path to the generated temporary file
func WriteToFile(content string) (string, error) {
	file, err := tmp.File("ssl-provisioner")
	if err != nil {
		return "", err
	}
	defer file.Close()
	if _, err := file.WriteString(content); err != nil {
		return "", err
	}

	return file.Name(), nil
}

func Provision(ctx context.Context, interCtx interpolate.Context, ui packersdk.Ui, communicator packersdk.Communicator, homeDir string, sslCertBase64 string, sslCertKeyBase64 string, nginxConfig string, amiConfigCommands []string) error {
	sslCert, err := decodeBase64(sslCertBase64)
	sslCertSource, err := WriteToFile(sslCert)
	sslCertDestination := fmt.Sprintf(filepath.Join(homeDir, "ssl.crt"))
	err = fileProvisioner.Provision(interCtx, ui, communicator, sslCertSource, sslCertDestination)
	if err != nil {
		return fmt.Errorf("error uploading '%s' to '%s': %s", sslCertSource, sslCertDestination, err)
	}

	sslCertKey, err := decodeBase64(sslCertKeyBase64)
	sslCertKeySource, err := WriteToFile(sslCertKey)
	sslCertKeyDestination := fmt.Sprintf(filepath.Join(homeDir, "ssl.key"))
	err = fileProvisioner.Provision(interCtx, ui, communicator, sslCertKeySource, sslCertKeyDestination)
	if err != nil {
		return fmt.Errorf("error uploading '%s' to '%s': %s", sslCertKeySource, sslCertKeyDestination, err)
	}

	if nginxConfig != "" {
		nginxSource, err := WriteToFile(nginxConfig)
		nginxDst := fmt.Sprintf(filepath.Join(homeDir, "nginx-ssl.conf"))
		err = fileProvisioner.Provision(interCtx, ui, communicator, nginxSource, nginxDst)
		if err != nil {
			return fmt.Errorf("error uploading '%s' to '%s': %s", nginxSource, nginxDst, err)
		}
	}

	return basicProvisioner.Provision(ctx, ui, communicator, amiConfigCommands)
}

// Decodes a base64-encoded string and returns the string representation of it
func decodeBase64(encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("error interpolating destination: %s", err)
	}
	return string(data), nil
}
