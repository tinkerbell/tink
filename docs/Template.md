# Templating a Template

For this document we will be explaining how to templatize your Template. This will allow you to use data that lives in your Hardware objects to populate values in your Template object. Let's get started.

## Background

In Tinkerbell, we have the concept of a Template. This is a defined as a Kubernetes Custom Resource (CR). It describes the steps that will run during a workflow. Look at the CR Definition (CRD) [here](../config/crd/bases/tinkerbell.org_templates.yaml) and an example Template object [here](../config/crd/examples/template.yaml). We also have the concept of Hardware. This is also a Kubernetes Custom Resource (CR). It describes your physical hardware. Look at the CR Definition (CRD) [here](../config/crd/bases/tinkerbell.org_hardware.yaml) and an example Hardware object [here](../config/crd/examples/hardware.yaml).

## Templating

Inside of this Template there is the ability to use a templating language to get data from your Hardware objects. See the docs [here](https://developer.hashicorp.com/nomad/tutorials/templates/go-template-syntax) for more details on use of the templating language. We don't expose the full Hardware object to the templating engine. We have created an abstraction on top of the Hardware object that is available to you. This was done so that we can change the underlying Hardware spec without breaking your Templates. Right now we expose the following fields:

### v1alpha1.Hardware

| Spec Field                            | Template Access                               | Field Type    | Example                                                   |
| ----------                            | ---------------                               | ----------    | -------                                                   |
| `spec.Disks`                          | `.Hardware.Disks[]`                           | string array  | `{{ index .Hardware.Disks 0 }}`                           |
| `spec.Interfaces`                     | `.Hardware.Interfaces[]`                      | string array  | `{{ index .Hardware.Interfaces 0 }}`                      |
| `spec.Interfaces[].DHCP.MAC`          | `.Hardware.Interfaces[].MAC`                  | string        | `{{ (index .Hardware.Interfaces 0).MAC }}`                |
| `spec.Interfaces[].DHCP.VLANID`       | `.Hardware.Interfaces[].VLANID`               | string        | `{{ (index .Hardware.Interfaces 0).VLANID }}`             |
| `spec.Interfaces[].DHCP.IP.Address`   | `.Hardware.Interfaces[].DHCP.IP`              | string        | `{{ (index .Hardware.Interfaces 0).DHCP.IP }}`            |
| `spec.Interfaces[].DHCP.IP.Gateway`   | `.Hardware.Interfaces[].DHCP.Gateway`         | string        | `{{ (index .Hardware.Interfaces 0).DHCP.Gateway }}`       |
| `spec.Interfaces[].DHCP.IP.Netmask`   | `.Hardware.Interfaces[].DHCP.Netmask`         | string        | `{{ (index .Hardware.Interfaces 0).DHCP.Netmask }}`       |
| `spec.Interfaces[].DHCP.Hostname`     | `.Hardware.Interfaces[].DHCP.Hostname`        | string        | `{{ (index .Hardware.Interfaces 0).DHCP.Hostname }}`      |
| `spec.Interfaces[].DHCP.NameServers`  | `.Hardware.Interfaces[].DHCP.Nameservers`     | string array  | `{{ (index .Hardware.Interfaces 0).DHCP.Nameservers }}`   |
| `spec.Interfaces[].DHCP.TimeServers`  | `.Hardware.Interfaces[].DHCP.Timeservers`     | string array  | `{{ (index .Hardware.Interfaces 0).DHCP.Timeservers }}`   |
