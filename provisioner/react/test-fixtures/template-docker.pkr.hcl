# Copyright (c) Jiaqi
# SPDX-License-Identifier: MPL-2.0

packer {
  required_plugins {
    docker = {
      version = ">= 0.0.7"
      source  = "github.com/hashicorp/docker"
    }
  }
}

source "docker" "qubitpi" {
  image  = "jack20191124/packer-plugin-qubitpi-acc-test-base:latest"
  discard = true
}

build {
  sources = [
    "source.docker.qubitpi"
  ]

  provisioner "qubitpi-react-provisioner" {
    distSource       = "/my/path/to/dist"
    homeDir          = "/"
    sslCertBase64    = "VGhpcyBpcyBhIHRlc3QgY2VydA=="
    sslCertKeyBase64 = "VGhpcyBpcyBhIHRlc3QgY2VydA=="
    appDomain        = "app.mycompany.com"
  }
}
