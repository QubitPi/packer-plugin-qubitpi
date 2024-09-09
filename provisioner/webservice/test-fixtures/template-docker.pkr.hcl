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

source "docker" "hashistack" {
  image  = "jack20191124/packer-plugin-hashistack-acc-test-base:latest"
  discard = true
}

build {
  sources = [
    "source.docker.hashistack"
  ]

  provisioner "hashistack-webservice-provisioner" {
    homeDir   = "/"
    warSource = "my-webservice.war"
  }
}
