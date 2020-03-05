# Components

### Cacher

Cacher is the data store that Rover pulls from. It uses Postgres as its data store. The data stored is structured in a way such that the information is split into two groups: information directly required by Rover to operate and anything else that will be relevant to other parts of the system, called metadata. 


### Tinkerbell

Handles DHCP requests, hands out IPs, serves up iPXE. It also uses Cacher to pull and push the data. 

### Osie

Installs all of our operating systems and handles deprovisioning.

### Hegel

It is the metadata service used by Rover and Osie during provisioning.

### Rover

It is the service responsible for handling the workflows. It comprises of a server and a CLI, which communicate over gRPC. The CLI used to create a workflow along with its building blocks, i.e., template and target.

