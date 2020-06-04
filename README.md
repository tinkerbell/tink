# Tinkerbell

[![Build Status](https://cloud.drone.io/api/badges/tinkerbell/tink/status.svg)](https://cloud.drone.io/tinkerbell/tink)

Tinkerbell is a bare metal provisioning and workflow engine comprised of three major components: a DHCP server ([boots](https://github.com/packethost/boots)), a workflow engine (tinkerbell, this repository), and a metadata service ([hegel](https://github.com/packethost/hegel)).
The workflow engine is comprised of a server and a CLI, which communicate over gRPC.
The CLI is used to create a workflow along with its building blocks, templates and targeted hardware.

# Packet Workflow

A Packet Workflow is an open-source microservice that’s responsible for handling flexible, bare metal provisioning workflows, that is...

-   standalone and does not need the Packet API to function
-   contains `Boots`, `Tinkerbell`, `Osie`, and workers
-   can bootstrap any remote worker using `Boots + Osie`
-   can run any set of actions as Docker container runtimes
-   receive, manipulate, and save runtime data

## Content

-   [Setup](docs/setup.md)
-   [Components](docs/components.md)
    -   [Boots](docs/components.md#boots)
    -   [Osie](docs/components.md#osie)
    -   [Tinkerbell](docs/components.md#tinkerbell)
    -   [Hegel](docs/components.md#hegel)
    -   [Database](docs/components.md#database)
    -   [Image Registry](docs/components.md#registry)
    -   [Elasticsearch](docs/components.md#elastic)
    -   [Fluent Bit](docs/components.md#fluent-bit)
    -   [Kibana](docs/components.md#kibana)
-   [Architecture](docs/architecture.md)
-   [Say "Hello-World!" with a Workflow](docs/hello-world.md)
-   [Concepts](docs/concepts.md)
    -   [Template](docs/concepts.md#template)
    -   [Provisioner](docs/concepts.md#provisioner)
    -   [Worker](docs/concepts.md#worker)
    -   [Ephemeral Data](docs/concepts.md#ephemeral-data)
-   [Writing a Workflow](docs/writing-workflow.md)
-   [Tinkerbell CLI Reference](docs/cli/README.md)
-   [Troubleshooting](docs/troubleshoot.md)

## Website

The Tinkerbell project is hosted at [tinkerbell.org](https://tinkerbell.org).
