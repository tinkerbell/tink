# Say Hello World! with a Workflow

Here is an example to execute the most simple workflow that says "Hello World!".

### Prerequisite

You have a setup ready with a provisioner and a worker node.
If not, please follow the steps [here](setup.md) to complete the setup.

### Hardware Data

While the data model changes are in progress, the following data should be enough to get your workflow rolling at the moment:

```json
{
  "id": "0eba0bf8-3772-4b4a-ab9f-6ebe93b90a94",
  "metadata": {
    "facility": {
      "facility_code": "ewr1",
      "plan_slug": "c2.medium.x86",
      "plan_version_slug": ""
    },
    "instance": {},
    "state": ""
  },
  "network": {
    "interfaces": [
      {
        "dhcp": {
          "arch": "x86_64",
          "ip": {
            "address": "192.168.1.5",
            "gateway": "192.168.1.1",
            "netmask": "255.255.255.248"
          },
          "mac": "00:00:00:00:00:00",
          "uefi": false
        },
        "netboot": {
          "allow_pxe": true,
          "allow_workflow": true
        }
      }
    ]
  }
}
```

A few details to note:

-   `id` is the hardware UUID
-   `network.interfaces[].netboot.allow_workflow` must be `true` to boot into workflow mode
-   `network.interfaces[].dhcp.ip.address` is the IP to hand out
-   `network.interfaces[].dhcp.mac` is the worker MAC address

Now that we have the hardware data, we need to push it into the database.
In order to do so, remove the extra spaces in the above JSON and use the following command to push the data:

```shell
$ tink hardware push '<json-data-here>'
```

Or, if you have your json data in a file (e.g. `data.json`), you can use the following:

```shell
$ tink hardware push --file data.json
```

Verify that the data was actually pushed using the command:

```shell
$ tink hardware mac <worker-mac-address>
```

### Action images

The workflow will have a single task that will have a single action.
The action here is to say `Hello-world!`, so we will push the action image to the registry running on the provisioner:

```shell
$ docker pull hello-world
$ docker tag hello-world <registry-host>/hello-world
$ docker push <registry-host>/hello-world
```

### Workflow

We can now define a workflow with the following steps:

-   Create a template:

```shell
 # get the template from examples/hello-world.tmpl and save it
 $ tink template create -n hello-world -p hello-world.tmpl
```

-   Create a workflow:

```shell
 $ tink workflow create -t <template-uuid> -r '{"device_1":"mac/IP"}'
```

-   Reboot the worker machine
