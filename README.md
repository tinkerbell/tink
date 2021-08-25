# Tinkerbell

[![Build Status](https://github.com/tinkerbell/tink/workflows/For%20each%20commit%20and%20PR/badge.svg)](https://github.com/tinkerbell/tink/actions?query=workflow%3A%22For+each+commit+and+PR%22+branch%3Amaster)
[![codecov](https://codecov.io/gh/tinkerbell/tink/branch/master/graph/badge.svg)](https://codecov.io/gh/tinkerbell/tink)
![](https://img.shields.io/badge/Stability-Experimental-red.svg)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/4512/badge)](https://bestpractices.coreinfrastructure.org/projects/4512)

This repository is [Experimental](https://github.com/packethost/standards/blob/master/experimental-statement.md) meaning that it's based on untested ideas or techniques and not yet established or finalized or involves a radically new and innovative style! This means that support is best effort (at best!) and we strongly encourage you to NOT use this in production.

## License

Tinkerbell is licensed under the Apache License, Version 2.0. See [LICENSE](./LICENSE) for the full license text. Some of the projects used by the Tinkerbell project may be governed by a different license, please refer to its specific license.

Tinkerbell is part of the CNCF Projects.

[![CNCF](https://github.com/cncf/artwork/blob/master/other/cncf/horizontal/color/cncf-color.png)](https://landscape.cncf.io/?selected=tinkerbell)

## Community

The Tinkerbell community meets bi-weekly on Tuesday. The meeting details can be found [here][7].

Community Resources:

-   [Community Slack](https://eqix-metal-community.slack.com/)
-   [CNCF #tinkerbell](https://app.slack.com/client/T08PSQ7BQ/C01SRB41GMT)
-   [YouTube Channel (demos, meeting recordings, virtual meetups)](https://www.youtube.com/channel/UCTzWInTQPvzH21KHS8jrq7A/featured)

## What's Powering Tinkerbell?

The Tinkerbell stack consists of several microservices, and a gRPC API:

### Tink

[Tink][1] is the short-hand name for the tink-server, tink-worker, and tink-cli.
`tink-worker` and `tink-server` communicate over gRPC, and are responsible for processing workflows.
The CLI is the user-interactive piece for creating workflows and their building blocks, templates and hardware data.

### Boots

[Boots][2] is Tinkerbell's DHCP server.
It handles DHCP requests, hands out IPs, and serves up iPXE.
It uses the Tinkerbell client to pull and push hardware data.
It only responds to a predefined set of MAC addresses so it can be deployed in an existing network without interfering with existing DHCP infrastructure.

### Hegel

[Hegel][3] is the metadata service used by Tinkerbell and OSIE.
It collects data from both and transforms it into a JSON format to be consumed as metadata.

### OSIE

[OSIE][4] is Tinkerbell's default an in-memory installation environment for bare metal.
It installs operating systems and handles deprovisioning.

### Hook

[Hook][5] is the newly introduced alternative to OSIE.
It's the next iteration of the in-memory installation environment to handle operating system installation and deprovisioning.

### PBnJ

[PBnJ][6] is an optional microservice that can communicate with baseboard management controllers (BMCs) to control power and boot settings.

## Building

Use `make help`.
The most interesting targets are `make all` (or just `make`) and `make images`.
`make all` builds all the binaries for your host OS and CPU to enable running directly.
`make images` will build all the binaries for Linux/x86_64 and build docker images with them.

## Website

For complete documentation, please visit the Tinkerbell project hosted at [tinkerbell.org](https://tinkerbell.org).

[1]: https://github.com/tinkerbell/tink
[2]: https://github.com/tinkerbell/boots
[3]: https://github.com/tinkerbell/hegel
[4]: https://github.com/tinkerbell/osie
[5]: https://github.com/tinkerbell/hook
[6]: https://github.com/tinkerbell/pbnj
[7]: https://docs.google.com/document/d/1cEObfvQ9Tdp8zIIIg9O7P5i3CKaSj2t3JTxEufDxwWs/
