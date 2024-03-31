<!--
  Include a short overview about the plugin.

  This document is a great location for creating a table of contents for each
  of the components the plugin may provide. This document should load automatically
  when navigating to the docs directory for a plugin.

-->

### Installation

To install this plugin, copy and paste this code into your Packer configuration, then run [`packer init`](https://www.packer.io/docs/commands/init).

```hcl
packer {
  required_plugins {
    hashicorp-aws = {
      version = ">= 0.0.1"
      source = "github.com/QubitPi/hashicorp-aws"
    }
  }
}
```

Alternatively, you can use `packer plugins install` to manage installation of this plugin.

```sh
$ packer plugins install github.com/QubitPi/hashicorp-aws
```

### Components

#### Provisioners

- [Kong API Gateway](./provisioners/kong-api-gateway.mdx)
- [Docker Mailserver](./provisioners/docker-mailserver.mdx)
