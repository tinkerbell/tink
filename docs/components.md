# Components

### Boots

Handles DHCP requests, hands out IPs, and serves up iPXE.
It also uses the Tinkerbell client to pull and push hardware data.
`boots` will only respond to a predefined set of MAC addresses so it can be deployed in an existing network without interfering with existing DHCP infrastructure.

### OSIE

Installs operating systems and handles deprovisioning.

### PBnJ

Communicates with BMCs to control power and boot settings.

### Tink

Service responsible for processing workflows. It is comprised of a server and a CLI, which communicate over gRPC. The CLI is used to create a workflow along with its building blocks, i.e., a template and targeted hardware devices.

### Hegel

Metadata service used by Tinkerbell and Osie during provisioning. It collects data from Tinkerbell and transforms it into a JSON format to be consumed as metadata.

### Database

We use [PostgreSQL](https://www.postgresql.org/), also known as Postgres, as our data store.
Postgres is a free and open-source relational database management system that emphasizes extensibility and technical standards compliance.
It is designed to handle a range of workloads, from single machines to data warehouses or Web services with many concurrent users.

### Image Repository

Depending on your use case, you can choose to use [Quay](https://quay.io/) or [DockerHub](https://hub.docker.com/) as the registry to store component images.
You can use the same registry to store all of the action images used for a workflow.

On the other hand, if you want to keep things local, you can also setup a secure private Docker registry to hold all your images locally.
