# Writing a [Workflow](concepts.md#workflow)

Any workflow comprises two building blocks: hardware device (worker) and a template. 

### Creating a [worker](concepts.md#worker)

A hardware device is a worker machine on which workflow needs to run.
User need to push the hardware details as per the below command:
```shell
 $ tink hardware push "<Hardware Data in json format>" 
```
### Creating a [template](concepts.md#template)

Consider a sample template like the following saved as `/tmp/sample.tmpl`.

```yaml
version: '0.1'
name: ubuntu_provisioning
global_timeout: 2500
tasks:
- name: "os-installation"
  worker: "{{.device_1}}"
  volumes:
    - /dev:/dev
    - /lib/firmware:/lib/firmware:ro
  environment:
    MIRROR_HOST: <MIRROR_HOST_IP>
  actions:
  - name: "disk-partition"
    image: disk-partition
    timeout: 600
    environment:
       MIRROR_HOST: <MIRROR_HOST_IP>
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
 
A worker can be accessed in a template like:

```
{{.device_1}}
{{.device_2}}
```

The following command creates a workflow template and returns a UUID:
```shell
 $ tink template create -n sample -p /tmp/sample.tmpl
``` 


### Creating a [workflow](concepts.md#workflow)

We can create a workflow using the above created (or existing) template and worker. 
```shell
 $ tink workflow create -t <template-uuid> -r '{worker machines in json format}'
 $ tink workflow create -t edb80a56-b1f2-4502-abf9-17326324192b -r '{"device_1":"mac/IP"}'
```

The above command returns a UUID for the workflow thus created. The workflow ID can be used for getting further details about a workflow. Please refer the [Tinkerbell CLI reference](cli/workflow.md) for the same.

It's a good practice to verify that the worker have been well substituted in the template. In order to do so, use the following command:
```yaml
 $ tink workflow get <workflow Id returns from the above command>

version: '0.1'
name: ubuntu_provisioning
global_timeout: 2500
tasks:
- name: "os-installation"
  worker: "98:03:9b:4b:c5:34"
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

Notice how `worker` is set to the MAC address we had defined in the input while creating a workflow.

