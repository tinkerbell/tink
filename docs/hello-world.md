# Say Hello World! with a Workflow

Here is an example to execute the most simple workflow that says "Hello World!".

### Prerequisite

You have a setup ready with a provisioner and a worker node. If not, please follow the steps [here](setup.md) to complete the setup.

### Hardware Data

While the data model changes are in progress, the following data should be enough to get your workflow rolling at the moment:
```json
{
  "id": "ce2e62ed-826f-4485-a39f-a82bb74338e2",
  "arch": "x86_64",
  "allow_pxe": true,
  "allow_workflow": true,
  "facility_code": "onprem",
  "ip_addresses": [
    {
      "address": "192.168.1.5",
      "address_family": 4,
      "enabled": true,
      "gateway": "192.168.1.1",
      "management": true,
      "netmask": "255.255.255.248",
      "public": false
    }
  ],
  "network_ports": [
    {
      "data": {
        "mac": "ec:0d:9a:bf:ff:dc"
      },
      "name": "eth0",
      "type": "data"
    }
  ]
}
```

A few details to note:
 - `id` is the hardware UUID
 - `allow_workflow` must be `true` to boot into workflow mode
 - `ip_addresses.address` is the IP to hand out
 - `network_ports` is the worker MAC address

Now that we have the hardware data, we need to push it into the database. In order to do so, remove the extra spaces in the above JSON and use the following command to push the data:
```
$ tink hardware push '<json-data-here>'
```

Verify that the data was actually pushed using the command:
```shell
$ tink hardware mac <worker-mac-address>
```

### Action images

The workflow will have a single task that will have a single action. The action here is to say `Hello-world!`, so we will push the action image to the registry running on the provisioner:
```shell
$ docker pull hello-world
$ docker tag hello-world <registry-host>/hello-world
$ docker push <registry-host>/hello-world
```

### Workflow

We can now define a workflow with the following steps:
 - Create a target:
 ```shell
  $ tink target create '{"targets": {"machine1": {"mac_addr": "<worker-mac-address>"}}}'
 ```
 - Create a template:
 ```shell
  # get the template from examples/hello-world.tmpl and save it
  $ tink template create -n hello-world -p hello-world.tmpl
 ```
 - Create a workflow:
 ```shell
  $ tink workflow create -t <template-uuid> -r <target-uuid>
 ```
 - Reboot the worker machine

