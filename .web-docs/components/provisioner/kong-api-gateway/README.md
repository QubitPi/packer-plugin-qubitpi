  Include a short description about the provisioner. This is a good place
  to call out what the provisioner does, and any additional text that might
  be helpful to a user. See https://www.packer.io/docs/provisioner/null
-->

The `kong-api-gateway` provisioner is used to install [Kong API gateway] server in AWS AMI image

hashistack deploys Kong API Gateway in the following way:

- Deploys the gateway in **HTTP** mode
- Deploys a reverse proxy Nginx in front of the gateway in the same EC2 to redirect all HTTPS request to
  gateway's [corresponding][Kong gateway - various ports] HTTP ports

The diagrams below illustrates the resulting deployment

![Deployment diagram](img/kong-deployment-diagram.png "Error loading kong-deployment-diagram.png")

> [!NOTE]
>
> hashistack uses a [customized fork of docker-kong](https://github.com/QubitPi/docker-kong) to
> [fully separate the app and SSL](https://github.com/QubitPi/docker-kong/pull/1), and, therefore, the Nginx config needs
> multiple [servers](https://www.nginx.com/resources/wiki/start/topics/examples/server_blocks/)
> to ensure all HTTPS ports are mapped to their corresponding HTTP ports

All relevant HTTP and HTTPS ports are listed in [Kong's documentation here][Kong gateway - various ports]. In general,
our Nginx should **listen on an HTTPS port and `proxy_pass` to an HTTP port. For example, ports 8443 and 8444 are
`proxy_pass`ed to 8000 and 8001, respectively, both of which are listed in the doc.

One special case is HTTP port 8000, which is the redirect port. hashistack maps the standard SSL 443 port to 8000 so
that any downstream (such as UI web app) simply needs to hit the domain without specifying port number and have its
request be reidrected to upstream services (such as database webservice)

![Port mapping diagram](img/kong-ports-diagram.png "Error loading kong-ports-diagram.png")

<!-- Provisioner Configuration Fields -->

**Required**

- `kongApiGatewayDomain` (string) - the SSL-enabled domain that will serve the deployed HTTP Nexus instance. For
  example, `api.mycompany.com`
- `sslCertBase64` (string) - is a _base64 encoded_ string of the content of __SSL certificate file__ for the SSL-enabled
  `kongApiGatewayDomain` above
- `sslCertKeyBase64` (string) - is a _base64 encoded_ string of the content of __SSL certificate key file__ for the
  SSL-enabled domain above

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
packer {
  required_plugins {
    amazon = {
      version = ">= 0.0.2"
      source  = "github.com/hashicorp/amazon"
    }
    hashistack = {
      version = ">= 0.0.45"
      source = "github.com/QubitPi/hashistack"
    }
  }
}

source "amazon-ebs" "hashistack" {
  ami_name              = "my-kong-api-gateway"
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

  provisioner "hashistack-kong-api-gateway-provisioner" {
    homeDir              = "/home/ubuntu"
    sslCertBase64        = "YXNkZnNnaHRkeWhyZXJ3ZGZydGV3ZHNmZ3RoeTY0cmV3ZGZyZWd0cmV3d2ZyZw=="
    sslCertKeyBase64     = "MzI0NXRnZjk4dmJoIGNsO2VbNDM1MHRdzszNDM1b2l0cmo="
    kongApiGatewayDomain = "api.mycompany.com"
  }
}
```

[AWS AMI]: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/AMIs.html
[AWS EC2 instance type]: https://aws.amazon.com/ec2/instance-types/
[AWS regions]: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html#Concepts.RegionsAndAvailabilityZones.Availability
[AWS Security Group]: https://docs.aws.amazon.com/vpc/latest/userguide/vpc-security-groups.html

[Certbot]: https://certbot.eff.org/

[hashistack/hashicorp/kong-api-gateway/images]: https://github.com/QubitPi/hashistack/tree/master/hashicorp/kong-api-gateway/images
[hashistack/hashicorp/kong-api-gateway/instances]: https://github.com/QubitPi/hashistack/tree/master/hashicorp/kong-api-gateway/instances
[HashiCorp Packer - Install]: https://packer.qubitpi.org/packer/install
[HashiCorp Packer variable values file]: https://packer.qubitpi.org/packer/guides/hcl/variables#from-a-file
[HashiCorp Terraform - Install]: https://terraform.qubitpi.org/terraform/install
[HashiCorp Terraform variable values file]: https://terraform.qubitpi.org/terraform/language/values/variables#variable-definitions-tfvars-files

[Kong API Gateway]: https://qubitpi.github.io/docs.konghq.com/gateway/latest/
[Kong manager UI]: https://qubitpi.github.io/docs.konghq.com/gateway/latest/kong-manager/
[Kong gateway - various ports]: https://qubitpi.github.io/docs.konghq.com/gateway/latest/production/networking/default-ports/

[Let's Encrypt]: https://qubitpi.github.io/letsencrypt-website/
