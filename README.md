# Tinkerbell

[![Build Status](https://cloud.drone.io/api/badges/tinkerbell/tink/status.svg)](https://cloud.drone.io/tinkerbell/tink)

It is comprised of following five major components:

1.  A DHCP server ([boots](https://github.com/tinkerbell/boots))
2.  A workflow engine (tink, this repository)
3.  A metadata service ([hegel](https://github.com/tinkerbell/hegel))
4.  An in-memory installation environment([osie](https://github.com/tinkerbell/osie))
5.  A controller/handler of BMC interactions([pbnj](https://github.com/tinkerbell/pbnj))

The workflow engine is comprised of a server and a CLI, which communicates over gRPC.
The CLI is used to create a workflow and its building blocks: templates and targeted hardware.

## Packet Workflow

A Packet Workflow is an open-source microservice that’s responsible for handling flexible, bare metal provisioning workflows, that is...

-   standalone and does not need the Packet API to function
-   contains `Boots`, `Tink`, `Hegel`, `OSIE`, `PBnJ` and workers
-   can bootstrap any remote worker using `Boots + Hegel + OSIE + PBnJ`
-   can run any set of actions as Docker container runtimes
-   receive, manipulate, and save runtime data

## Content

-   [Setup](docs/setup.md)
-   [Components](docs/components.md)
    -   [Boots](docs/components.md#boots)
    -   [OSIE](docs/components.md#osie)
    -   [PBnJ](docs/components.md#pbnj)
    -   [Tink](docs/components.md#tink)
    -   [Hegel](docs/components.md#hegel)
    -   [Database](docs/components.md#database)
    -   [Image Registry](docs/components.md#registry)
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
