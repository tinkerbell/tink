# Writing a [Workflow](concepts.md#workflow)

Any workflow comprises two building blocks: target and template. 

### Creating a [target](concepts.md#target)

A target is referred with MAC or IP address. Here is a sample target definition using the MAC address:

```json
{
    "machine1" :  {
        "mac_addr": "98:03:9b:4b:c5:34"
    }
}
```

The command below creates a workflow target and returns its UUID:
```shell
 $ tinkerbell target create '{"targets": {"machine1": {"mac_addr": "98:03:9b:4b:c5:34"}}}' 
```


### Creating a [template](concepts.md#template)

Consider a sample template like the following saved as `/tmp/sample.tmpl`.

```yaml
version: '0.1'
name: ubuntu_provisioning
global_timeout: 2500
tasks:
- name: "os-installation"
  worker: "{{index .Targets "machine1" "mac_addr"}}"
  volumes:
    - /dev:/dev
    - /lib/firmware:/lib/firmware:ro
  environment:
    MIRROR_HOST: 192.168.1.2
  actions:
  - name: "disk-partition"
    image: disk-partition
    timeout: 600
    environment:
       MIRROR_HOST: 192.168.1.3
    volumes:
      - /statedir:/statedir
  - name: "install-root-fs"
    image: install-root-fs
    timeout: 600
  - name: "install-grub"
    image: install-grub
    timeout: 600
    volumes:
      - /statedir:/statedir
```

Key points:
 - `global_timeout` is in seconds.
 - Any worflow that exceeds the global timeout will be terminated and marked as failed.
 - Each action has its own timeout. If an action reaches its timeout, it is marked as failed and so is the workflow.
 - An action cannot have space (` `) in its name.
 - Environment variables and volumes at action level overwrites the values for duplicate keys defined at task level.
 
A target can be accessed in a template like:

```
{{ index .Targets "machine1" "ip_addr"}}
{{ index .Targets "machine2" "mac_addr"}}
```

The following command creates a workflow template and returns a UUID:
```shell
 $ tinkerbell template create -n sample -p /tmp/sample.tmpl
``` 


### Creating a [workflow](concepts.md#workflow)

We can create a workflow using the above created (or existing) template and target. 
```shell
 $ tinkerbell workflow create -t <template-uuid> -r <target-uuid>
 $ tinkerbell workflow create -t edb80a56-b1f2-4502-abf9-17326324192b -r 9356ae1d-6165-4890-908d-7860ed04b421
```

The above command returns a UUID for the workflow thus created. The workflow ID can be used for getting further details about a workflow. Please refer the [Tinkerbell CLI reference](cli.md) for the same.

It's a good practice to verify that the targets have been well substituted in the template. In order to do so, use the following command:
```yaml
 $ tinkerbell workflow get edb80a56-b1f2-4502-abf9-17326324192b

version: '0.1'
name: ubuntu_provisioning
global_timeout: 2500
tasks:
- name: "os-installation"
  worker: ""
  volumes:
    - /dev:/dev
    - /lib/firmware:/lib/firmware:ro
  environment:
    MIRROR_HOST: 192.168.1.2
  actions:
  - name: "disk-partition"
    image: disk-partition
    timeout: 600
    environment:
       MIRROR_HOST: 192.168.1.3
    volumes:
      - /statedir:/statedir
  - name: "install-root-fs"
    image: install-root-fs
    timeout: 600
  - name: "install-grub"
    image: install-grub
    timeout: 600
    volumes:
      - /statedir:/statedir
```

Notice how `worker` is set to the MAC address we had defined in the target.

