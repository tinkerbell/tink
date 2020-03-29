# Setup the Packet Workflow Environment with Terraform

 - Clone the `tinkerbell` repository for latest code:
```shell
$ git clone https://github.com/packethost/tinkerbell.git
$ cd tinkerbell/terraform
```

 - Update the `input.tf` file with actual username and password of GitHub and quay.io 
 - Add your Packet `auth_token` in `input.tf`
 - Run the following commands
```shell
$ terraform init
$ terraform apply
``` 

The above commands will create a complete setup with `tf-provisioner` and `tf-worker` machines for the `packet` provider which you can run any workflow. As an output it returns the IP address of the provisioner and MAC address of the worker machine.


**_Note_**: The default names of machines created by Terraform are `tf-provisioner` and `tf-worker`. If you prefer other names, you need to replace `tf-provisioner` and `tf-worker` with the new ones at all places in `main.tf`.


# Setup the Provisioner Machine directly with docker-compose.yml file


## Install git and git lfs as follows

1. ### Setup git and git lfs
    ```shell
    $ sudo apt install -y git  
    $ wget https://github.com/git-lfs/git-lfs/releases/download/v2.9.0/git-lfs-linux-amd64-v2.9.0.tar.gz  
    $ tar -C /usr/local/bin -xzf git-lfs-linux-amd64-v2.9.0.tar.gz  
    $ rm git-lfs-linux-amd64-v2.9.0.tar.gz  
    $ git lfs install  
2. ### Setup go
    ```shell
    $ wget https://dl.google.com/go/go1.12.13.linux-amd64.tar.gz
    $ tar -C /usr/local -xzf go1.12.13.linux-amd64.tar.gz go/
    $ rm go1.12.13.linux-amd64.tar.gz
3. ### Set GOPATH
    ```shell
    $ echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    $ echo 'export GOPATH=$GOPATH:$HOME/go' >> ~/.bashrc
    $ echo 'export PATH=$PATH:$GOPATH' >> ~/.bashrc
    $ source ~/.bashrc
4. ### Clone the tinkerbell repo in the $GOPATH
    ```shell
    $ mkdir -p ~/go/src/github.com/packethost
    $ cd ~/go/src/github.com/packethost
    $ git clone https://github.com/packethost/tinkerbell.git
    $ cd tinkerbell
    $ git checkout setup-with-docker-compose

3. ### Provide the input details in "input.sh" file

4. ### Run the following command
    ```
    $ sudo ./setup_with_docker_compose.sh

