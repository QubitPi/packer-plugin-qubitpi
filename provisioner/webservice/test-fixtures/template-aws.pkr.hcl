# Copyright (c) Jiaqi
# SPDX-License-Identifier: MPL-2.0

packer {
  required_plugins {
    amazon = {
      version = ">= 0.0.2"
      source  = "github.com/hashicorp/amazon"
    }
  }
}

source "amazon-ebs" "hashistack" {
  ami_name              = "packer-plugin-hashistack-acc-test-ami-webservice"
  force_deregister      = "true"
  force_delete_snapshot = "true"

  instance_type = "t2.micro"
  launch_block_device_mappings {
    device_name           = "/dev/sda1"
    volume_size           = 8
    volume_type           = "gp2"
    delete_on_termination = true
  }
  region = "us-west-1"
  source_ami_filter {
    filters = {
      name                = "ubuntu/images/*ubuntu-*-22.04-amd64-server-*"
      root-device-type    = "ebs"
      virtualization-type = "hvm"
    }
    most_recent = true
    owners      = ["099720109477"]
  }
  ssh_username = "ubuntu"
}

build {
  sources = [
    "source.amazon-ebs.hashistack"
  ]

  provisioner "hashistack-webservice-provisioner" {
    homeDir   = "/home/ubuntu"
    warSource = "my-webservice.war"
  }
}
