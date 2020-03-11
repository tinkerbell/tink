# Packet Workflow 

A Packet Workflow is an open-source microservice that’s responsible for handling flexible, bare metal
provisioning workflows, that is...
 - standalone and does not need the Packet API to function
 - contains `Cacher`, `Tinkerbell`, `Osie`, and local workers
 - can bootstrap any remote worker using `Tink + Osie`
 - can run any set of actions as Docker container runtimes
 - receive, manipulate, and save runtime data


## Content
 
 - [Setup](setup.md)
 - [Components](components.md)
   - [Cacher](components.md#cacher)
   - [Tinkerbell](components.md#tinkerbell)
   - [Osie](components.md#osie)
   - [Rover](components.md#rover)
   - [Hegel](components.md#hegel)
   - [Database](components.md#database)
   - [Image Registry](components.md#registry)
   - [Elasticsearch](components.md#elastic)
   - [Fluent Bit](components.md#cacher)
   - [Kibana](components.md#kibana)
 - [Architecture](architecture.md)
 - [Example: First Good Workflow](first-good-workflow.md)
 - [Concepts](concepts.md)
   - [Template](concepts.md#template)
   - [Target](concepts.md#target)
   - [Provisioner](concepts.md#provisioner)
   - [Worker](concepts.md#worker)
   - [Ephemeral Data](concepts.md#ephemeral-data)
 - [Writing a Workflow](writing-workflow.md)
 - [Rover CLI Reference](cli/README.md)
 - [Troubleshooting](troubleshoot.md)


