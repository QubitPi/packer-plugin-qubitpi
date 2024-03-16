# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

source "amazon-ebs" "kong" {
  ami_name = "my-kong-api-gateway"
  force_deregister = "true"
  force_delete_snapshot = "true"
  skip_create_ami = "false"

  instance_type = "t2.large"
  launch_block_device_mappings {
    device_name = "/dev/sda1"
    volume_size = 8
    volume_type = "gp2"
    delete_on_termination = true
  }
  region = "us-east-1"
  source_ami_filter {
    filters = {
      name = "ubuntu/images/*ubuntu-*-22.04-amd64-server-*"
      root-device-type = "ebs"
      virtualization-type = "hvm"
    }
    most_recent = true
    owners = ["099720109477"]
  }
  ssh_username = "ubuntu"
}

build {
  sources = [
    "source.amazon-ebs.kong"
  ]

  provisioner "hashicorp-aws-kong-api-gateway-provisioner" {
    homeDir = "/home/ubuntu"
    sslCertSource = "/abs/or/rel/path/to/ssl-cert-file"
    sslCertKeySource = "/abs/or/rel/path/to/ssl-cert-key-file"
    kongApiGatewayDomain = "mykongdomain.com"
  }
}
