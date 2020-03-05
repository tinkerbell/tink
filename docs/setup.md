# Setup the Packet Workflow Environment with Terraform

 - Clone the `rover` repository for latest code:
```shell
$ git clone https://github.com/packethost/rover.git
$ cd rover/terraform
```

 - Update the `input.tf` file with actual username and password of GitHub and quay.io 
 - Add your Packet `auth_token` in `input.tf`
 - Run the following commands
```shell
$ terraform init
$ terraform apply
``` 

The above commands will create a complete setup with `tf-provisioner` and `tf-worker` machines on which you can run any workflow. As an output it returns the IP address of the provisioner and MAC address of the worker machine.


**_Note_**: The default names of machines created by Terraform are `tf-provisioner` and `tf-worker`. If you prefer other names, you need to replace `tf-provisioner` and `tf-worker` with the new ones at all places in `main.tf`.


