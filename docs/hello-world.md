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
In order to do so, save the above JSON into a file (e.g., `data.json`) and use either of the following commands to push the data:

```shell
$ tink hardware push --file data.json
$ cat data.json | tink hardware push
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
 # the value of device_1 could be either the MAC address or IP address of the hardware/device
 $ tink workflow create -t <template-uuid> -r '{"device_1":"mac/IP"}'
```

-   Reboot the worker machine
