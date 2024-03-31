  Include a short description about the provisioner. This is a good place
  to call out what the provisioner does, and any additional text that might
  be helpful to a user. See https://www.packer.io/docs/provisioner/null
-->

The docker-mailserver provisioner is used to provision Packer builds for docker-mailserver AMI.


<!-- Provisioner Configuration Fields -->

**Required**

- `sslCertSource` (string) - The path to the local SSL certificate file to upload to the machine. The path can be
  absolute or relative. If it is relative, it is relative to the working directory when Packer is executed.
- `sslCertKeySource` (string) - The path to the local SSL certificate key file to upload to the machine. The path can be
  absolute or relative. If it is relative, it is relative to the working directory when Packer is executed.
- `baseDomain` (string) - The base domain name of the MX record. For example, if base domain is 'mycompany.com', the
  generated MX record will be 'mail.mycompany.com'

<!--
  Optional Configuration Fields

  Configuration options that are not required or have reasonable defaults
  should be listed under the optionals section. Defaults values should be
  noted in the description of the field
-->

**Optional**

- `homeDir` (string) - The `$Home` directory in AMI image; default to `/home/ubuntu`

<!--
  A basic example on the usage of the provisioner. Multiple examples
  can be provided to highlight various configurations.

-->

### Example Usage

```hcl
source "amazon-ebs" "docker-mailserver" {
  ami_name = "my-docker-mailserver-ami"
  force_deregister = "true"
  force_delete_snapshot = "true"
  skip_create_ami = "false"

  instance_type = "t2.micro"
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
  name = "install-docker-mailserver"
  sources = [
    "amazon-ebs.docker-mailserver"
  ]

  provisioner "hashicorp-aws-docker-mailserver-provisioner" {
    homeDir = "/home/ubuntu"
    sslCertSource = "/abs/or/rel/path/to/ssl-cert-file"
    sslCertKeySource = "/abs/or/rel/path/to/ssl-cert-key-file"
    baseDomain = "mycompany.com"
  }
}
```
