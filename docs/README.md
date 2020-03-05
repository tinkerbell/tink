# Packet Workflow 

A Packet Workflow is an open-source microservice that’s responsible for handling flexible, bare metal
provisioning workflows, that is...
 - standalone and does not need the Packet API to function
 - contains `Cacher`, `Tinkerbell`, `Osie`, and local workers
 - can bootstrap any remote worker using `Tink + Osie`
 - can run any set of actions as Docker container runtimes
 - receive, manipulate, and save runtime data


## Content
 
 - Setup & Get started
 - [Architecture](architecture.md)
  - [Components](architecture.md#components)
  - [10,000ft View](architecture.md#10000ft-view)
 - Example: First Good Workflow 
 - Writing a Workflow
  - Concepts 
  - Useful CLI commands
 - [Common Errors](errors.md)


