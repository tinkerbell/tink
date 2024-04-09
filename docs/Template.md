# Templating a Template

For this document we will be explaining how to templatize your Template. This will allow you to use data that lives in your Hardware objects to populate values in your Template object. Let's get started.

## Background

In Tinkerbell, we a use Kubernetes Customer Resources to define our data. This includes Hardware, Template, and Workflow definitions. Understanding these will be important to understanding how to template your Template.

- **Hardware**: describes your physical hardware. See the CRD [here](../config/crd/bases/tinkerbell.org_hardware.yaml) and an example Hardware object [here](../config/crd/examples/hardware.yaml)
- **Template**: describes the steps that will run during a workflow. See the CRD [here](../config/crd/bases/tinkerbell.org_templates.yaml) and an example Template object [here](../config/crd/examples/template.yaml)
- **Workflow**: describes the steps that will run during a workflow. See the CRD [here](../config/crd/bases/tinkerbell.org_workflows.yaml) and an example Workflow object [here](../config/crd/examples/workflow.yaml)

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

## Including data from a Workflow

Arbitrary data can be included in a Workflow object and then accessed from a Template. This data is set via `spec.hardwareMap` in the Workflow. For example below, `device_1` is the arbitrary key and `3c:ec:ef:4c:4f:54` is the value. It can then be accessed in the Template via `{{ .device_1 }}`. See the example below.

```yaml
apiVersion: "tinkerbell.org/v1alpha1"
kind: Workflow
metadata:
  name: wf1
spec:
  templateRef: debian
  hardwareRef: sm01
  hardwareMap:
    device_1: 3c:ec:ef:4c:4f:54
```

```yaml
apiVersion: "tinkerbell.org/v1alpha1"
kind: Template
metadata:
  name: debian
spec:
  data: |
    version: "0.1"
    name: debian
    global_timeout: 1800
    tasks:
      - name: "os-installation"
        worker: "{{.device_1}}"
```

## Templating functions

There are a number of built in functions that Go provides and that can be used in your templating. See [here](https://developer.hashicorp.com/nomad/tutorials/templates/go-template-syntax#function-list). Tinkerbell has also defined a few custom functions that can be used.

| Function Name     | Description | Use     | Examples |
| -------------     | ----------- | ------- | -------- |
| `contains`        | contains returns a bool for whether `substr` is within `s`. | `{{ contains "HELLO" "H" }}` | `contains <s> <substr>` |
| `hasPrefix`       | hasPrefix returns a bool for whether the string s begins with prefix. | `{{ hasPrefix "HELLO" "HE" }}` | `hasPrefix <s> <prefix>` |
| `hasSuffix`       | hasSuffix returns a bool for whether the string s ends with suffix. | `{{ hasPrefix "HELLO" "HE" }}` | `hasSuffix <s> <suffix>` |
| `formatPartition` | formatPartition formats a device path with partition for the specific device type. Supported devices: `/dev/nvme`, `/dev/sd`, `/dev/vd`, `/dev/xvd`, `/dev/hd`. | `{{ formatPartition ( index .Hardware.Disks 0 ) 2 }}` | `formatPartition("/dev/nvme0n1", 0) -> /dev/nvme0n1p1`, `formatPartition("/dev/sda", 1) -> /dev/sda1` |
