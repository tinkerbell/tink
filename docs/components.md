# Components

### Cacher

Cacher is the data store that Rover pulls from. It uses Postgres as its data store. The data stored is structured in a way such that the information is split into two groups: information directly required by Rover to operate and anything else that will be relevant to other parts of the system, called metadata. 

### Boots

Handles DHCP requests, hands out IPs, serves up iPXE. It also uses Cacher to pull and push the data. 

### Osie

Installs all of our operating systems and handles deprovisioning.

### Tinkerbell

It is the service responsible for handling the workflows. It comprises of a server and a CLI, which communicate over gRPC. The CLI used to create a workflow along with its building blocks, i.e., template and target.

### Hegel

It is the metadata service used by Rover and Osie during provisioning. It collects the data from Cacher and transforms it into a different structure that is used as the metadata. The data thus received is in JSON format. 

### Database

We use [PostgreSQL](https://www.postgresql.org/), also known as Postgres as our data store. It is a free and open-source relational database management system emphasizing extensibility and technical standards compliance. It is designed to handle a range of workloads, from single machines to data warehouses or Web services with many concurrent users.

### Image Repository

Depending on your use case, you can choose to use [Quay](https://quay.io/) or [DockerHub](https://hub.docker.com/) as the registry to store the component images. You can use the same registry to store all the action images used for a workflow. 

On the hand, if you want to keep things local, you can also setup a secure private Docker registry to hold all your images. 

### Fluent Bit

[Fluent Bit](https://fluentbit.io/) is an open source and multi-platform Log Processor and Forwarder which allows you to collect data/logs from different sources, unify and send them to multiple destinations. The components write their logs to `stdout`. These logs are then collected by Fluent Bit and pushed to Elasticsearch. 

### Elasticsearch

[Elasticsearch](https://www.elastic.co/) is a distributed, open source search and analytics engine for all types of data, including textual, numerical, geospatial, structured, and unstructured. Fluent Bit collects the logs from each component and pushes into Elasticsearch. 

### Kibana

[Kibana](https://www.elastic.co/kibana) lets you visualize your Elasticsearch data and navigate the Elastic Stack so you can do anything from tracking query load to understanding the way requests flow through your apps.