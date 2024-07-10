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

source "docker" "hashicorp-aws" {
  image  = "jack20191124/packer-plugin-hashicorp-aws-acc-test-base:latest"
  discard = true
}

build {
  sources = [
    "source.docker.hashicorp-aws"
  ]

  provisioner "hashicorp-aws-react-provisioner" {
    distSource       = "/my/path/to/dist"
    homeDir          = "/"
    sslCertBase64    = "YXNkZnNnaHRkeWhyZXJ3ZGZydGV3ZHNmZ3RoeTY0cmV3ZGZyZWd0cmV3d2ZyZw=="
    sslCertKeyBase64 = "MzI0NXRnZjk4dmJoIGNsO2VbNDM1MHRdzszNDM1b2l0cmo="
    appDomain        = "app.mycompany.com"
  }
}
