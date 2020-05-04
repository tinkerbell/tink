# Provisioning Flatcar Container Linux

This document describes the process of provisioning Flatcar Container Linux machines using Tinkerbell.

## Preparing flatcar-install Docker image

To install [Flatcar Container Linux](https://www.flatcar-linux.org/) as the OS on the worker, the Docker image from the following `Dockerfile` needs to be pushed into the Tinkerbell registry:

```
FROM ubuntu

RUN apt update && apt install -y udev gpg wget

RUN wget https://raw.githubusercontent.com/flatcar-linux/init/flatcar-master/bin/flatcar-install -O /usr/local/bin/flatcar-install && \
    chmod +x /usr/local/bin/flatcar-install

ENTRYPOINT ["/usr/local/bin/flatcar-install"]
```

You should build the image using the following command:

```
docker build -t $TINKERBELL_HOST_IP:flatcar-install .
```

Then, you should push the image to Tinkerbell registry, so it can be pulled by workers during workflow execution:

```
docker push $TINKERBELL_HOST_IP:flatcar-install
```

## Preparing template

You can find example template in [./examples/flatcar-install.tmpl](./examples/flatcar-install.tmpl) file. It should be used as a base for the next steps.

### Ignition configuration

Recommended way of configuring Flatcar is via Ignition file during installation.

Example template supports providing Ignition configuration in a form of Base64 encoded string injected directly into the template.

See [Flatcar documentation about provisioning](https://docs.flatcar-linux.org/os/provisioning/) for more details.

To create sample configuration, create `config.yaml` file with the following content:

```yaml
passwd:
  users:
  - name: core
    ssh_authorized_keys:
    - YOUR_SSH_KEY
```

Then, use the following command to validate and encode your configuration:

```
ct -in-file config.yaml  | base64 -w0; echo
```

Returned string can be then put in the template, by replacing default `eyJpZ25pdGlvbiI6eyJjb25maWciOnt9LCJzZWN1cml0eSI6eyJ0bHMiOnt9fSwidGltZW91dHMiOnt9LCJ2ZXJzaW9uIjoiMi4yLjAifSwibmV0d29ya2QiOnt9LCJwYXNzd2QiOnt9LCJzdG9yYWdlIjp7fSwic3lzdGVtZCI6e319` Ignition configuration.

### Target disk

By default the template installs Flatcar to smallest disk found on the worker using `-s` flag.

### Extra installation parameters

See [Flatcar documentation](https://docs.flatcar-linux.org/os/installing-to-disk/) to see all available parameters.

### Creating

After modifying the example template, you need to create it in Tinkerbell using the following command:

```
tink template create -n flatcar-install -p /tmp/flatcar-install.tmpl
```

Please keep the `ID` of created template, as it will be used in next steps.

## Creating target and workflow

After creating the template, you need to create target and workflow object. This can be done using the following commands:

```
tink target create '{"targets": {"machine1": {"mac_addr": "02:42:db:98:4b:1e"},"machine2": {"ipv4_addr": "192.168.1.5"}}}'
```

Again, save the `ID` of created target, as it is required for creating the workflow.

To create a workflow, execute the following command:

```
tink workflow create -r TARGET_ID -t TEMPLATE_ID
```

This command will return you workflow ID, which is needed to check the status of the workflow. Target ID and Template ID will no longer be used.

## Provisioning

With workflow created, you can boot/reboot your worker(s), so it picks up the new workflow.

You can track the progress of workflow execution using the following commands:

```
tink workflow state WORKFLOW_ID
tink workflow events WORKFLOW_ID
```

Once the workflow is finished, you can reboot the worker and it should now boot Flatcar Container Linux from disk.



