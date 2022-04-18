# cenk1cenk2/do-dyndns

[![pipeline status](https://gitlab.kilic.dev/docker/do-dyndns/badges/master/pipeline.svg)](https://gitlab.kilic.dev/docker/do-dyndns/-/commits/master) [![Docker Pulls](https://img.shields.io/docker/pulls/cenk1cenk2/do-dyndns)](https://hub.docker.com/repository/docker/cenk1cenk2/do-dyndns) [![Docker Image Size (latest by date)](https://img.shields.io/docker/image-size/cenk1cenk2/do-dyndns)](https://hub.docker.com/repository/docker/cenk1cenk2/do-dyndns) [![Docker Image Version (latest by date)](https://img.shields.io/docker/v/cenk1cenk2/do-dyndns)](https://hub.docker.com/repository/docker/cenk1cenk2/do-dyndns) [![GitHub last commit](https://img.shields.io/github/last-commit/cenk1cenk2/do-dyndns)](https://github.com/cenk1cenk2/do-dyndns)

---

## Description

This utility provides a native way to update "A" records of a domain that is managed by Digital Ocean name-servers. Simply create an API token to access and update domain records depending on the IP address of the host that is running this utility.

<!-- toc -->

- [Install](#install)
  - [Deploy with docker-compose](#deploy-with-docker-compose)
  - [Get Binary From Releases](#get-binary-from-releases)
- [Setup](#setup)
  - [Utilizing the Configuration File](#utilizing-the-configuration-file)
  - [Utilizing the Environment Variables](#utilizing-the-environment-variables)

<!-- tocstop -->

## Install

Platforms `linux-amd64`, `linux-arm64`, `linux-armv7` is supported.

### Deploy with docker-compose

Published image is available in Docker Hub as `cenk1cenk2/do-dyndns`. Will work on every platform defined.

```docker
version: "3.7"
services:
  do-dyndns:
    image: cenk1cenk2/do-dyndns
    environment:
      - DYNDNS_DOMAINS=example1.com
      - DYNDNS_SUBDOMAINS=1.example1.com,2.example2.com
      - DYNDNS_TOKEN=$DIGITAL_OCEAN_TOKEN
```

### Get Binary From Releases

You can find the natively compiled versions in the [releases](https://github.com/cenk1cenk2/do-dyndns/releases/latest).

## Setup

### Utilizing the Configuration File

Create a `.yml` file in any of the locations `.`, `/etc/do-dyndns`, `~/.config/do-dyndns` named `.do-dyndns.yml`.

You can also pass in `--config` flag to pass in the absolute path of the configuration file.

The configuration file structure is as below:

```yaml
domains:
  - example.com
subdomains:
  - 1.example.com
token: DIGITAL_OCEAN_TOKEN
repeat: 3600 # this is optional and 3600 is the default if you want repeat
```

### Utilizing the Environment Variables

To run with environment variables just pass in the variables with `DYNDNS_` prefix.

| Environment Variable | Description                                                                                    | Default |
| -------------------- | ---------------------------------------------------------------------------------------------- | ------- |
| DYNDNS_TOKEN         | Digital Ocean API token.                                                                       |         |
| DYNDNS_DOMAINS       | Comma-separated list of domains to run against.                                                |         |
| DYNDNS_SUBDOMAINS    | Comma-separated list of subdomains, which must match with the domains or else will be skipped. |         |
| DYNDNS_REPEAT        | Check repeat time in seconds.                                                                  | 3600    |
