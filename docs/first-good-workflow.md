# First Good Workflow (Example)

Here is an example workflow that provisions a worker machine with Ubuntu 18.04.

### Prerequisite

You have a setup ready with a Provisioner and a Worker node. If not, please follow the [steps](setup.md) here to complete the setup.


### Steps
 - Login to the `provisioner` machine
 - Copy the following folders/files from `test-provisioner` machine to the current machine at the `/packet/nginx/misc/osie/current/` location:
 ```shell 
  /var/www/html/misc/osie/current/grub/
  /var/www/html/misc/osie/current/ubuntu_18_04 
  /var/www/html/misc/osie/current/modloop-x86_64
 ```
 - Change directory to `tinkerbell`:
 ```shell
 $ cd ~/go/src/github.com/tinkerbell/tink/
 ```
 - switch to `first-good-workflow` branch
 ```shell
 $ git checkout first-good-workflow
 ```
 - Replace the MAC addresses in `push.sh` file with the MAC address of `tf-worker` machine which you have got in the terraform output
 - Push the hardware data
 ```shell
  $ ./push.sh
 ```
 - Create action images and push them in to the private registry:
 ```shell
  $ cd ~/go/src/github.com/tinkerbell/tink/workflow-samples/ubuntu/v3/
  $ ./create_image.sh
 ```
 - Create a target:
 ```shell
  $ tink target create '{"targets": {"machine1": {"mac_addr": "<mac address of tf-worker>"}}}'
 ```
 - Create a template:
 ```shell
  $ tink template create -n ubuntu-sample -p /root/go/src/github.com/tinkerbell/tink/workflow-samples/ubuntu/ubuntu.tmpl
 ```
 - Create a workflow:
 ```shell
  $ tink workflow create -t <template Id> -r <target_id>
 ```
 - Reboot the worker machine

